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

// SupplierService 供应商业务逻辑。
type SupplierService struct {
	repo   repository.SupplierRepository
	audit  *AuditService
	logger *slog.Logger
}

// NewSupplierService 构造供应商服务。
func NewSupplierService(repo repository.SupplierRepository, audit *AuditService, logger *slog.Logger) *SupplierService {
	return &SupplierService{repo: repo, audit: audit, logger: logger}
}

// Create 创建供应商。
func (s *SupplierService) Create(ctx context.Context, actor model.Actor, req dto.CreateSupplierRequest) (*model.Supplier, error) {
	status := req.Status
	if status == "" {
		status = constants.SupplierStatusActive
	}
	supplier := &model.Supplier{
		Name:        req.Name,
		Category:    req.Category,
		Contact:     req.Contact,
		Phone:       req.Phone,
		Address:     req.Address,
		BankName:    req.BankName,
		BankAccount: req.BankAccount,
		Status:      status,
		Rating:      req.Rating,
	}
	if err := s.repo.Create(ctx, supplier); err != nil {
		return nil, fmt.Errorf("create supplier: %w", err)
	}
	s.audit.Record(ctx, actor, "supplier_create", "supplier", supplier.ID, fmt.Sprintf("name=%s category=%s", supplier.Name, supplier.Category))
	return supplier, nil
}

// List 分页查询供应商。
func (s *SupplierService) List(ctx context.Context, filter dto.SupplierFilter) ([]model.Supplier, int64, error) {
	suppliers, total, err := s.repo.List(ctx, repository.SupplierListFilter{
		Status:   filter.Status,
		Category: filter.Category,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list suppliers: %w", err)
	}
	return suppliers, total, nil
}

// Get 获取供应商。
func (s *SupplierService) Get(ctx context.Context, id uint) (*model.Supplier, error) {
	supplier, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get supplier %d: %w", id, err)
	}
	return supplier, nil
}

// Update 更新供应商。
func (s *SupplierService) Update(ctx context.Context, actor model.Actor, id uint, req dto.UpdateSupplierRequest) (*model.Supplier, error) {
	supplier, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get supplier %d: %w", id, err)
	}
	if req.Name != "" {
		supplier.Name = req.Name
	}
	if req.Category != "" {
		supplier.Category = req.Category
	}
	if req.Contact != "" {
		supplier.Contact = req.Contact
	}
	if req.Phone != "" {
		supplier.Phone = req.Phone
	}
	if req.Address != "" {
		supplier.Address = req.Address
	}
	if req.BankName != "" {
		supplier.BankName = req.BankName
	}
	if req.BankAccount != "" {
		supplier.BankAccount = req.BankAccount
	}
	if req.Status != "" {
		supplier.Status = req.Status
	}
	if req.Rating != 0 {
		supplier.Rating = req.Rating
	}
	if err := s.repo.Update(ctx, supplier); err != nil {
		return nil, fmt.Errorf("update supplier %d: %w", id, err)
	}
	s.audit.Record(ctx, actor, "supplier_update", "supplier", id, fmt.Sprintf("name=%s status=%s", supplier.Name, supplier.Status))
	return supplier, nil
}

// Delete 删除供应商。
func (s *SupplierService) Delete(ctx context.Context, actor model.Actor, id uint) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("get supplier %d: %w", id, err)
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete supplier %d: %w", id, err)
	}
	s.audit.Record(ctx, actor, "supplier_delete", "supplier", id, "")
	return nil
}
