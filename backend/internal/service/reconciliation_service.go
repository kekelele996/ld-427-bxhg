package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/renovation/renovation-budget-api/internal/constants"
	"github.com/renovation/renovation-budget-api/internal/dto"
	"github.com/renovation/renovation-budget-api/internal/model"
	"github.com/renovation/renovation-budget-api/internal/repository"
)

// ReconciliationService 对账单业务逻辑。
type ReconciliationService struct {
	repo   repository.ReconciliationRepository
	audit  *AuditService
	logger *slog.Logger
}

// NewReconciliationService 构造对账单服务。
func NewReconciliationService(repo repository.ReconciliationRepository, audit *AuditService, logger *slog.Logger) *ReconciliationService {
	return &ReconciliationService{repo: repo, audit: audit, logger: logger}
}

// Create 创建对账单。
func (s *ReconciliationService) Create(ctx context.Context, actor model.Actor, req dto.CreateReconciliationRequest) (*model.Reconciliation, error) {
	record := &model.Reconciliation{
		ProjectID:     req.ProjectID,
		Period:        req.Period,
		SupplierID:    req.SupplierID,
		PayableAmount: req.PayableAmount,
		PaidAmount:    req.PaidAmount,
		UnpaidAmount:  CalculateUnpaid(req.PayableAmount, req.PaidAmount),
		Status:        constants.ReconciliationStatusPending,
	}
	if err := s.repo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("create reconciliation: %w", err)
	}
	s.audit.Record(ctx, actor, "reconciliation_create", "reconciliation", record.ID, fmt.Sprintf("period=%s supplier_id=%d", record.Period, record.SupplierID))
	return record, nil
}

// List 分页查询对账单。
func (s *ReconciliationService) List(ctx context.Context, filter dto.ReconciliationFilter) ([]model.Reconciliation, int64, error) {
	records, total, err := s.repo.List(ctx, repository.ReconciliationListFilter{
		ProjectID:  filter.ProjectID,
		Period:     filter.Period,
		SupplierID: filter.SupplierID,
		Status:     filter.Status,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list reconciliations: %w", err)
	}
	return records, total, nil
}

// Get 获取对账单。
func (s *ReconciliationService) Get(ctx context.Context, id uint) (*model.Reconciliation, error) {
	record, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get reconciliation %d: %w", id, err)
	}
	return record, nil
}

// Update 更新对账金额。
func (s *ReconciliationService) Update(ctx context.Context, actor model.Actor, id uint, req dto.UpdateReconciliationRequest) (*model.Reconciliation, error) {
	record, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get reconciliation %d: %w", id, err)
	}
	if req.PayableAmount != nil {
		record.PayableAmount = *req.PayableAmount
	}
	if req.PaidAmount != nil {
		record.PaidAmount = *req.PaidAmount
	}
	record.UnpaidAmount = CalculateUnpaid(record.PayableAmount, record.PaidAmount)
	if err := s.repo.Update(ctx, record); err != nil {
		return nil, fmt.Errorf("update reconciliation %d: %w", id, err)
	}
	s.audit.Record(ctx, actor, "reconciliation_update", "reconciliation", id, fmt.Sprintf("payable=%.2f paid=%.2f", record.PayableAmount, record.PaidAmount))
	return record, nil
}

// Confirm 确认对账。
func (s *ReconciliationService) Confirm(ctx context.Context, actor model.Actor, id uint) (*model.Reconciliation, error) {
	record, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get reconciliation %d: %w", id, err)
	}
	if record.Status != constants.ReconciliationStatusPending {
		return nil, fmt.Errorf("confirm reconciliation %d: %w", id, ErrInvalidState)
	}
	record.Status = constants.ReconciliationStatusConfirmed
	record.ConfirmedByID = &actor.UserID
	if err := s.repo.Update(ctx, record); err != nil {
		return nil, fmt.Errorf("confirm reconciliation %d: %w", id, err)
	}
	s.audit.Record(ctx, actor, "reconciliation_confirm", "reconciliation", id, fmt.Sprintf("period=%s", record.Period))
	return record, nil
}

// Dispute 将对账单标记为争议。
func (s *ReconciliationService) Dispute(ctx context.Context, actor model.Actor, id uint) (*model.Reconciliation, error) {
	record, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get reconciliation %d: %w", id, err)
	}
	if record.Status != constants.ReconciliationStatusPending && record.Status != constants.ReconciliationStatusConfirmed {
		return nil, fmt.Errorf("dispute reconciliation %d: %w", id, ErrInvalidState)
	}
	record.Status = constants.ReconciliationStatusDisputed
	if err := s.repo.Update(ctx, record); err != nil {
		return nil, fmt.Errorf("dispute reconciliation %d: %w", id, err)
	}
	s.audit.Record(ctx, actor, "reconciliation_dispute", "reconciliation", id, fmt.Sprintf("period=%s", record.Period))
	return record, nil
}

// Resolve 解决争议。
func (s *ReconciliationService) Resolve(ctx context.Context, actor model.Actor, id uint) (*model.Reconciliation, error) {
	record, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get reconciliation %d: %w", id, err)
	}
	if record.Status != constants.ReconciliationStatusDisputed {
		return nil, fmt.Errorf("resolve reconciliation %d: %w", id, ErrInvalidState)
	}
	record.Status = constants.ReconciliationStatusResolved
	record.ConfirmedByID = &actor.UserID
	if err := s.repo.Update(ctx, record); err != nil {
		return nil, fmt.Errorf("resolve reconciliation %d: %w", id, err)
	}
	s.audit.Record(ctx, actor, "reconciliation_resolve", "reconciliation", id, fmt.Sprintf("period=%s", record.Period))
	return record, nil
}

// Delete 删除对账单，仅允许 Pending 状态。
func (s *ReconciliationService) Delete(ctx context.Context, actor model.Actor, id uint) error {
	record, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("get reconciliation %d: %w", id, err)
	}
	if record.Status != constants.ReconciliationStatusPending {
		return fmt.Errorf("delete reconciliation %d: %w", id, ErrInvalidState)
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete reconciliation %d: %w", id, err)
	}
	s.audit.Record(ctx, actor, "reconciliation_delete", "reconciliation", id, fmt.Sprintf("period=%s", record.Period))
	return nil
}
