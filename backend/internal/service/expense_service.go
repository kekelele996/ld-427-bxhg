package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/renovation/renovation-budget-api/internal/constants"
	"github.com/renovation/renovation-budget-api/internal/dto"
	"github.com/renovation/renovation-budget-api/internal/model"
	"github.com/renovation/renovation-budget-api/internal/repository"
)

// ExpenseService 支出记录业务逻辑。
type ExpenseService struct {
	repo       repository.ExpenseRepository
	itemRepo   repository.ItemRepository
	budgetRepo repository.BudgetRepository
	audit      *AuditService
	rdb        *redis.Client
	logger     *slog.Logger
}

// NewExpenseService 构造支出记录服务。
func NewExpenseService(repo repository.ExpenseRepository, itemRepo repository.ItemRepository, budgetRepo repository.BudgetRepository, audit *AuditService, rdb *redis.Client, logger *slog.Logger) *ExpenseService {
	return &ExpenseService{repo: repo, itemRepo: itemRepo, budgetRepo: budgetRepo, audit: audit, rdb: rdb, logger: logger}
}

// Create 创建草稿支出记录。
func (s *ExpenseService) Create(ctx context.Context, actor model.Actor, req dto.CreateExpenseRequest) (*model.ExpenseRecord, error) {
	item, err := s.itemRepo.FindByID(ctx, req.BudgetItemID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get budget item %d: %w", req.BudgetItemID, err)
	}
	expenseDate, err := time.Parse("2006-01-02", req.ExpenseDate)
	if err != nil {
		return nil, fmt.Errorf("parse expense date: %w", err)
	}
	record := &model.ExpenseRecord{
		BudgetItemID:  item.ID,
		Amount:        req.Amount,
		ExpenseDate:   expenseDate,
		PaymentMethod: req.PaymentMethod,
		SupplierID:    req.SupplierID,
		InvoiceNo:     req.InvoiceNo,
		Description:   req.Description,
		AttachmentURL: req.AttachmentURL,
		Status:        constants.ExpenseStatusDraft,
		ApplicantID:   actor.UserID,
	}
	if err := s.repo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("create expense record: %w", err)
	}
	s.audit.Record(ctx, actor, "expense_create", "expense", record.ID, fmt.Sprintf("budget_item_id=%d amount=%.2f", item.ID, req.Amount))
	return record, nil
}

// List 分页查询支出记录。
func (s *ExpenseService) List(ctx context.Context, filter dto.ExpenseFilter) ([]model.ExpenseRecord, int64, error) {
	records, total, err := s.repo.List(ctx, repository.ExpenseListFilter{
		Status:       filter.Status,
		BudgetItemID: filter.BudgetItemID,
		SupplierID:   filter.SupplierID,
		Page:         filter.Page,
		PageSize:     filter.PageSize,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list expense records: %w", err)
	}
	return records, total, nil
}

// Get 获取支出记录。
func (s *ExpenseService) Get(ctx context.Context, id uint) (*model.ExpenseRecord, error) {
	record, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get expense record %d: %w", id, err)
	}
	return record, nil
}

// Submit 提交支出并冻结预算。
func (s *ExpenseService) Submit(ctx context.Context, actor model.Actor, id uint) (*model.ExpenseRecord, error) {
	record, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get expense record %d: %w", id, err)
	}
	if record.Status != constants.ExpenseStatusDraft {
		return nil, fmt.Errorf("submit expense record %d: %w", id, ErrInvalidState)
	}
	item, err := s.itemRepo.FindByID(ctx, record.BudgetItemID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get budget item %d: %w", record.BudgetItemID, err)
	}
	budget, err := s.budgetRepo.FindByID(ctx, item.BudgetSheetID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get budget sheet %d: %w", item.BudgetSheetID, err)
	}
	if CalculateAvailable(budget.TotalAmount, budget.SpentAmount, budget.FrozenAmount) < record.Amount {
		return nil, fmt.Errorf("submit expense record %d: %w", id, ErrInsufficientBalance)
	}

	record.Status = constants.ExpenseStatusSubmitted
	budget.FrozenAmount += record.Amount
	budget.AvailableAmount = CalculateAvailable(budget.TotalAmount, budget.SpentAmount, budget.FrozenAmount)

	if err := s.repo.Update(ctx, record); err != nil {
		return nil, fmt.Errorf("submit expense record %d: %w", id, err)
	}
	if err := s.budgetRepo.Update(ctx, budget); err != nil {
		return nil, fmt.Errorf("freeze budget sheet %d: %w", budget.ID, err)
	}
	s.invalidateBudget(ctx, budget.ID)
	s.audit.Record(ctx, actor, "expense_submit", "expense", id, fmt.Sprintf("amount=%.2f budget_sheet_id=%d", record.Amount, budget.ID))
	return record, nil
}

// Approve 审批通过支出，冻结金额转为已支出金额。
func (s *ExpenseService) Approve(ctx context.Context, actor model.Actor, id uint, req dto.ApproveExpenseRequest) (*model.ExpenseRecord, error) {
	record, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get expense record %d: %w", id, err)
	}
	if record.Status != constants.ExpenseStatusSubmitted {
		return nil, fmt.Errorf("approve expense record %d: %w", id, ErrInvalidState)
	}
	item, err := s.itemRepo.FindByID(ctx, record.BudgetItemID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get budget item %d: %w", record.BudgetItemID, err)
	}
	budget, err := s.budgetRepo.FindByID(ctx, item.BudgetSheetID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get budget sheet %d: %w", item.BudgetSheetID, err)
	}

	record.Status = constants.ExpenseStatusApproved
	record.ApprovedByID = &actor.UserID
	record.ApprovalComment = req.ApprovalComment
	budget.FrozenAmount -= record.Amount
	budget.SpentAmount += record.Amount
	budget.AvailableAmount = CalculateAvailable(budget.TotalAmount, budget.SpentAmount, budget.FrozenAmount)
	item.SpentAmount += record.Amount
	item.VarianceAmount = CalculateVariance(item.SpentAmount, item.BudgetAmount)

	if err := s.repo.Update(ctx, record); err != nil {
		return nil, fmt.Errorf("approve expense record %d: %w", id, err)
	}
	if err := s.itemRepo.Update(ctx, item); err != nil {
		return nil, fmt.Errorf("update budget item %d: %w", item.ID, err)
	}
	if err := s.budgetRepo.Update(ctx, budget); err != nil {
		return nil, fmt.Errorf("update budget sheet %d: %w", budget.ID, err)
	}
	s.invalidateBudget(ctx, budget.ID)
	s.audit.Record(ctx, actor, "expense_approve", "expense", id, fmt.Sprintf("amount=%.2f comment=%s", record.Amount, req.ApprovalComment))
	return record, nil
}

// Reject 驳回支出并释放冻结金额。
func (s *ExpenseService) Reject(ctx context.Context, actor model.Actor, id uint, req dto.RejectExpenseRequest) (*model.ExpenseRecord, error) {
	record, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get expense record %d: %w", id, err)
	}
	if record.Status != constants.ExpenseStatusSubmitted {
		return nil, fmt.Errorf("reject expense record %d: %w", id, ErrInvalidState)
	}
	item, err := s.itemRepo.FindByID(ctx, record.BudgetItemID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get budget item %d: %w", record.BudgetItemID, err)
	}
	budget, err := s.budgetRepo.FindByID(ctx, item.BudgetSheetID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get budget sheet %d: %w", item.BudgetSheetID, err)
	}

	record.Status = constants.ExpenseStatusRejected
	record.ApprovedByID = &actor.UserID
	record.ApprovalComment = req.ApprovalComment
	budget.FrozenAmount -= record.Amount
	budget.AvailableAmount = CalculateAvailable(budget.TotalAmount, budget.SpentAmount, budget.FrozenAmount)

	if err := s.repo.Update(ctx, record); err != nil {
		return nil, fmt.Errorf("reject expense record %d: %w", id, err)
	}
	if err := s.budgetRepo.Update(ctx, budget); err != nil {
		return nil, fmt.Errorf("update budget sheet %d: %w", budget.ID, err)
	}
	s.invalidateBudget(ctx, budget.ID)
	s.audit.Record(ctx, actor, "expense_reject", "expense", id, fmt.Sprintf("amount=%.2f comment=%s", record.Amount, req.ApprovalComment))
	return record, nil
}

// Pay 确认付款，仅允许已审批通过的支出。
func (s *ExpenseService) Pay(ctx context.Context, actor model.Actor, id uint, req dto.PayExpenseRequest) (*model.ExpenseRecord, error) {
	record, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get expense record %d: %w", id, err)
	}
	if record.Status != constants.ExpenseStatusApproved {
		return nil, fmt.Errorf("pay expense record %d: %w", id, ErrInvalidState)
	}
	record.Status = constants.ExpenseStatusPaid
	if req.PaymentDate != "" {
		paymentDate, err := time.Parse("2006-01-02", req.PaymentDate)
		if err != nil {
			return nil, fmt.Errorf("parse payment date: %w", err)
		}
		record.PaymentDate = &paymentDate
	} else {
		now := time.Now()
		record.PaymentDate = &now
	}
	if err := s.repo.Update(ctx, record); err != nil {
		return nil, fmt.Errorf("pay expense record %d: %w", id, err)
	}
	s.audit.Record(ctx, actor, "expense_pay", "expense", id, fmt.Sprintf("amount=%.2f", record.Amount))
	return record, nil
}

func (s *ExpenseService) invalidateBudget(ctx context.Context, budgetSheetID uint) {
	if s.rdb == nil {
		return
	}
	if err := s.rdb.Del(ctx, budgetSnapshotKey(budgetSheetID)).Err(); err != nil {
		s.logger.Warn("invalidate budget snapshot failed", slog.Uint64("budget_id", uint64(budgetSheetID)), slog.String("error", err.Error()))
	}
}
