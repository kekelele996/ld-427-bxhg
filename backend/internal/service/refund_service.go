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

// amountEpsilon 金额比较容忍误差，规避浮点累加误差。
const amountEpsilon = 1e-9

// RefundService 支出退款业务逻辑。
type RefundService struct {
	repo        repository.RefundRepository
	expenseRepo repository.ExpenseRepository
	itemRepo    repository.ItemRepository
	budgetRepo  repository.BudgetRepository
	audit       *AuditService
	rdb         *redis.Client
	logger      *slog.Logger
}

// NewRefundService 构造支出退款服务。
func NewRefundService(repo repository.RefundRepository, expenseRepo repository.ExpenseRepository, itemRepo repository.ItemRepository, budgetRepo repository.BudgetRepository, audit *AuditService, rdb *redis.Client, logger *slog.Logger) *RefundService {
	return &RefundService{
		repo:        repo,
		expenseRepo: expenseRepo,
		itemRepo:    itemRepo,
		budgetRepo:  budgetRepo,
		audit:       audit,
		rdb:         rdb,
		logger:      logger,
	}
}

// Register 针对已付款支出登记一笔退款，并回冲预算分项与预算表已用金额。
func (s *RefundService) Register(ctx context.Context, actor model.Actor, expenseID uint, req dto.RegisterRefundRequest) (*model.ExpenseRefund, error) {
	// 所有校验先于写入完成：任一条件不满足时不产生任何数据变更。
	record, err := s.expenseRepo.FindByID(ctx, expenseID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get expense record %d: %w", expenseID, err)
	}
	if record.Status != constants.ExpenseStatusPaid {
		return nil, fmt.Errorf("register refund for expense %d: %w", expenseID, ErrRefundNotAllowed)
	}

	// 凭据号重复时只入账第一次：已存在则本次操作失败。
	exists, err := s.repo.ExistsByVoucherNo(ctx, req.VoucherNo)
	if err != nil {
		return nil, fmt.Errorf("check refund voucher number: %w", err)
	}
	if exists {
		return nil, ErrDuplicateVoucherNo
	}

	refunded, err := s.repo.SumAmountByExpenseID(ctx, expenseID)
	if err != nil {
		return nil, fmt.Errorf("sum refunded amount: %w", err)
	}
	// 累计退款超过本笔实付金额时操作失败。
	if refunded+req.Amount > record.Amount+amountEpsilon {
		return nil, fmt.Errorf("register refund for expense %d: %w", expenseID, ErrRefundExceedsPaid)
	}

	receivedDate, err := time.Parse("2006-01-02", req.ReceivedDate)
	if err != nil {
		return nil, fmt.Errorf("parse received date: %w", err)
	}

	refund := &model.ExpenseRefund{
		ExpenseID:      expenseID,
		Amount:         req.Amount,
		ReceivedDate:   receivedDate,
		Reason:         req.Reason,
		VoucherNo:      req.VoucherNo,
		RegisteredByID: actor.UserID,
	}
	// 唯一索引兜底并发场景下的凭据号重复。
	if err := s.repo.Create(ctx, refund); err != nil {
		if errors.Is(err, repository.ErrDuplicateRefundVoucher) {
			return nil, ErrDuplicateVoucherNo
		}
		return nil, fmt.Errorf("create expense refund: %w", err)
	}

	// 退款成功后回冲预算分项与预算表已用金额。
	item, err := s.itemRepo.FindByID(ctx, record.BudgetItemID)
	if err != nil {
		return nil, fmt.Errorf("get budget item %d: %w", record.BudgetItemID, err)
	}
	budget, err := s.budgetRepo.FindByID(ctx, item.BudgetSheetID)
	if err != nil {
		return nil, fmt.Errorf("get budget sheet %d: %w", item.BudgetSheetID, err)
	}

	item.SpentAmount = roundAmount(item.SpentAmount - req.Amount)
	item.VarianceAmount = CalculateVariance(item.SpentAmount, item.BudgetAmount)
	budget.SpentAmount = roundAmount(budget.SpentAmount - req.Amount)
	budget.AvailableAmount = CalculateAvailable(budget.TotalAmount, budget.SpentAmount, budget.FrozenAmount)

	if err := s.itemRepo.Update(ctx, item); err != nil {
		return nil, fmt.Errorf("reverse budget item %d spent: %w", item.ID, err)
	}
	if err := s.budgetRepo.Update(ctx, budget); err != nil {
		return nil, fmt.Errorf("reverse budget sheet %d spent: %w", budget.ID, err)
	}
	s.invalidateBudget(ctx, budget.ID)
	s.audit.Record(ctx, actor, "expense_refund_register", "expense", expenseID,
		fmt.Sprintf("refund_id=%d amount=%.2f voucher_no=%s budget_sheet_id=%d", refund.ID, req.Amount, req.VoucherNo, budget.ID))
	return refund, nil
}

// GetSummary 汇总支出退款：退款记录、已退金额、剩余可退金额。
func (s *RefundService) GetSummary(ctx context.Context, expenseID uint) (*dto.ExpenseRefundSummary, error) {
	record, err := s.expenseRepo.FindByID(ctx, expenseID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get expense record %d: %w", expenseID, err)
	}
	refunds, err := s.repo.ListByExpenseID(ctx, expenseID)
	if err != nil {
		return nil, fmt.Errorf("list refunds for expense %d: %w", expenseID, err)
	}
	refunded, err := s.repo.SumAmountByExpenseID(ctx, expenseID)
	if err != nil {
		return nil, fmt.Errorf("sum refunded amount: %w", err)
	}
	if refunds == nil {
		refunds = []model.ExpenseRefund{}
	}
	return &dto.ExpenseRefundSummary{
		Records:        refunds,
		RefundedAmount: roundAmount(refunded),
		RefundableLeft: roundAmount(record.Amount - refunded),
	}, nil
}

// GetExpenseDetail 获取支出详情，附带退款记录、已退金额与剩余可退金额。
func (s *RefundService) GetExpenseDetail(ctx context.Context, expenseID uint) (*dto.ExpenseDetailResponse, error) {
	record, err := s.expenseRepo.FindByID(ctx, expenseID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get expense record %d: %w", expenseID, err)
	}
	refunds, err := s.repo.ListByExpenseID(ctx, expenseID)
	if err != nil {
		return nil, fmt.Errorf("list refunds for expense %d: %w", expenseID, err)
	}
	refunded, err := s.repo.SumAmountByExpenseID(ctx, expenseID)
	if err != nil {
		return nil, fmt.Errorf("sum refunded amount: %w", err)
	}
	if refunds == nil {
		refunds = []model.ExpenseRefund{}
	}
	return &dto.ExpenseDetailResponse{
		ExpenseRecord:  *record,
		RefundedAmount: roundAmount(refunded),
		RefundableLeft: roundAmount(record.Amount - refunded),
		Refunds:        refunds,
	}, nil
}

func (s *RefundService) invalidateBudget(ctx context.Context, budgetSheetID uint) {
	if s.rdb == nil {
		return
	}
	if err := s.rdb.Del(ctx, budgetSnapshotKey(budgetSheetID)).Err(); err != nil {
		s.logger.Warn("invalidate budget snapshot failed", slog.Uint64("budget_id", uint64(budgetSheetID)), slog.String("error", err.Error()))
	}
}

// roundAmount 保留两位小数，消除浮点误差。
func roundAmount(v float64) float64 {
	return math.Round(v*100) / 100
}
