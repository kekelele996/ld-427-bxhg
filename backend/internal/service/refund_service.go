package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/renovation/renovation-budget-api/internal/constants"
	"github.com/renovation/renovation-budget-api/internal/dto"
	"github.com/renovation/renovation-budget-api/internal/model"
	"github.com/renovation/renovation-budget-api/internal/repository"
)

// refundAmountEpsilon 金额比较容差，规避浮点累计误差。
const refundAmountEpsilon = 1e-9

// RefundService 退款登记业务逻辑。
type RefundService struct {
	tx          repository.TxManager
	refundRepo  repository.RefundRepository
	expenseRepo repository.ExpenseRepository
	itemRepo    repository.ItemRepository
	budgetRepo  repository.BudgetRepository
	audit       *AuditService
	rdb         *redis.Client
	logger      *slog.Logger
}

// NewRefundService 构造退款登记服务。
func NewRefundService(
	tx repository.TxManager,
	refundRepo repository.RefundRepository,
	expenseRepo repository.ExpenseRepository,
	itemRepo repository.ItemRepository,
	budgetRepo repository.BudgetRepository,
	audit *AuditService,
	rdb *redis.Client,
	logger *slog.Logger,
) *RefundService {
	return &RefundService{
		tx:          tx,
		refundRepo:  refundRepo,
		expenseRepo: expenseRepo,
		itemRepo:    itemRepo,
		budgetRepo:  budgetRepo,
		audit:       audit,
		rdb:         rdb,
		logger:      logger,
	}
}

// Register 针对已付款支出登记一笔退款，并回冲预算分项与预算表已用金额。
// 累计退款超过本笔实付金额时整体回滚；退款凭据号重复时只有第一次入账。
func (s *RefundService) Register(ctx context.Context, actor model.Actor, expenseID uint, req dto.CreateRefundRequest) (*model.RefundRecord, error) {
	receivedDate, err := time.Parse("2006-01-02", req.ReceivedDate)
	if err != nil {
		return nil, fmt.Errorf("parse received date: %w", err)
	}

	var saved *model.RefundRecord
	var budgetSheetID uint
	err = s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		record, err := s.expenseRepo.FindByIDForUpdate(txCtx, expenseID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return ErrNotFound
			}
			return fmt.Errorf("lock expense record %d: %w", expenseID, err)
		}
		if record.Status != constants.ExpenseStatusPaid {
			return fmt.Errorf("register refund for expense %d: %w", expenseID, ErrInvalidState)
		}
		if record.RefundedAmount+req.Amount > record.Amount+refundAmountEpsilon {
			return fmt.Errorf("register refund for expense %d: %w", expenseID, ErrRefundExceeded)
		}

		item, err := s.itemRepo.FindByID(txCtx, record.BudgetItemID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return ErrNotFound
			}
			return fmt.Errorf("get budget item %d: %w", record.BudgetItemID, err)
		}
		budget, err := s.budgetRepo.FindByID(txCtx, item.BudgetSheetID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return ErrNotFound
			}
			return fmt.Errorf("get budget sheet %d: %w", item.BudgetSheetID, err)
		}

		refund := &model.RefundRecord{
			ExpenseID:      record.ID,
			Amount:         req.Amount,
			ReceivedDate:   receivedDate,
			Reason:         req.Reason,
			VoucherNo:      req.VoucherNo,
			RegisteredByID: actor.UserID,
		}
		if err := s.refundRepo.Create(txCtx, refund); err != nil {
			if errors.Is(err, repository.ErrDuplicateVoucherNo) {
				return ErrDuplicateVoucher
			}
			return fmt.Errorf("create refund record: %w", err)
		}

		record.RefundedAmount = roundAmount(record.RefundedAmount + req.Amount)
		item.SpentAmount = roundAmount(item.SpentAmount - req.Amount)
		item.VarianceAmount = CalculateVariance(item.SpentAmount, item.BudgetAmount)
		budget.SpentAmount = roundAmount(budget.SpentAmount - req.Amount)
		budget.AvailableAmount = CalculateAvailable(budget.TotalAmount, budget.SpentAmount, budget.FrozenAmount)

		if err := s.expenseRepo.Update(txCtx, record); err != nil {
			return fmt.Errorf("update refunded amount of expense %d: %w", record.ID, err)
		}
		if err := s.itemRepo.Update(txCtx, item); err != nil {
			return fmt.Errorf("reverse budget item %d spent: %w", item.ID, err)
		}
		if err := s.budgetRepo.Update(txCtx, budget); err != nil {
			return fmt.Errorf("reverse budget sheet %d spent: %w", budget.ID, err)
		}

		saved = refund
		budgetSheetID = budget.ID
		return nil
	})
	if err != nil {
		return nil, err
	}

	s.invalidateBudget(ctx, budgetSheetID)
	s.audit.Record(ctx, actor, "expense_refund_register", "expense", expenseID,
		fmt.Sprintf("refund_id=%d amount=%.2f voucher_no=%s reason=%s", saved.ID, saved.Amount, saved.VoucherNo, saved.Reason))
	return saved, nil
}

// ListByExpense 查询某笔支出的退款记录。
func (s *RefundService) ListByExpense(ctx context.Context, expenseID uint) ([]model.RefundRecord, error) {
	if _, err := s.expenseRepo.FindByID(ctx, expenseID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get expense record %d: %w", expenseID, err)
	}
	refunds, err := s.refundRepo.ListByExpenseID(ctx, expenseID)
	if err != nil {
		return nil, fmt.Errorf("list refund records for expense %d: %w", expenseID, err)
	}
	if refunds == nil {
		refunds = []model.RefundRecord{}
	}
	return refunds, nil
}

// GetExpenseDetail 查询支出详情，附带退款记录、已退金额与剩余可退金额。
func (s *RefundService) GetExpenseDetail(ctx context.Context, expenseID uint) (*dto.ExpenseDetailResponse, error) {
	record, err := s.expenseRepo.FindByID(ctx, expenseID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get expense record %d: %w", expenseID, err)
	}
	refunds, err := s.refundRepo.ListByExpenseID(ctx, expenseID)
	if err != nil {
		return nil, fmt.Errorf("list refund records for expense %d: %w", expenseID, err)
	}
	if refunds == nil {
		refunds = []model.RefundRecord{}
	}
	refunded := 0.0
	for i := range refunds {
		refunded += refunds[i].Amount
	}
	if math.Abs(refunded-record.RefundedAmount) > refundAmountEpsilon {
		s.logger.Warn("refund total disagrees with expense refunded amount",
			slog.Uint64("expense_id", uint64(expenseID)),
			slog.Float64("summed", refunded),
			slog.Float64("persisted", record.RefundedAmount))
	}
	return &dto.ExpenseDetailResponse{
		ExpenseRecord: *record,
		Refund: dto.ExpenseRefundDetail{
			RefundedAmount: roundAmount(record.RefundedAmount),
			RefundableLeft: roundAmount(record.Amount - record.RefundedAmount),
			Refunds:        refunds,
		},
	}, nil
}

func (s *RefundService) invalidateBudget(ctx context.Context, budgetSheetID uint) {
	if s.rdb == nil {
		return
	}
	if err := s.rdb.Del(ctx, budgetSnapshotKey(budgetSheetID)).Err(); err != nil {
		s.logger.Warn("invalidate budget snapshot failed",
			slog.Uint64("budget_id", uint64(budgetSheetID)), slog.String("error", err.Error()))
	}
}

// roundAmount 将金额归一化到分，避免浮点累计误差。
func roundAmount(v float64) float64 {
	return math.Round(v*100) / 100
}
