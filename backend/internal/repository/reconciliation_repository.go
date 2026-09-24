package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/renovation/renovation-budget-api/internal/model"
)

// ReconciliationListFilter 对账单列表过滤条件。
type ReconciliationListFilter struct {
	ProjectID  string
	Period     string
	SupplierID uint
	Status     string
	Page       int
	PageSize   int
}

// ReconciliationRepository 对账单数据访问接口。
type ReconciliationRepository interface {
	Create(ctx context.Context, reconciliation *model.Reconciliation) error
	FindByID(ctx context.Context, id uint) (*model.Reconciliation, error)
	List(ctx context.Context, filter ReconciliationListFilter) ([]model.Reconciliation, int64, error)
	Update(ctx context.Context, reconciliation *model.Reconciliation) error
	Delete(ctx context.Context, id uint) error
}

type reconciliationRepository struct {
	db *gorm.DB
}

// NewReconciliationRepository 构造对账单仓储。
func NewReconciliationRepository(db *gorm.DB) ReconciliationRepository {
	return &reconciliationRepository{db: db}
}

func (r *reconciliationRepository) Create(ctx context.Context, reconciliation *model.Reconciliation) error {
	if err := r.db.WithContext(ctx).Create(reconciliation).Error; err != nil {
		return fmt.Errorf("create reconciliation: %w", err)
	}
	return nil
}

func (r *reconciliationRepository) FindByID(ctx context.Context, id uint) (*model.Reconciliation, error) {
	var reconciliation model.Reconciliation
	if err := r.db.WithContext(ctx).First(&reconciliation, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find reconciliation %d: %w", id, err)
	}
	return &reconciliation, nil
}

func (r *reconciliationRepository) List(ctx context.Context, filter ReconciliationListFilter) ([]model.Reconciliation, int64, error) {
	page, pageSize := normalizePage(filter.Page, filter.PageSize)
	q := r.db.WithContext(ctx).Model(&model.Reconciliation{})
	if filter.ProjectID != "" {
		q = q.Where("project_id = ?", filter.ProjectID)
	}
	if filter.Period != "" {
		q = q.Where("period = ?", filter.Period)
	}
	if filter.SupplierID != 0 {
		q = q.Where("supplier_id = ?", filter.SupplierID)
	}
	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count reconciliations: %w", err)
	}
	var records []model.Reconciliation
	if err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&records).Error; err != nil {
		return nil, 0, fmt.Errorf("list reconciliations: %w", err)
	}
	return records, total, nil
}

func (r *reconciliationRepository) Update(ctx context.Context, reconciliation *model.Reconciliation) error {
	if err := r.db.WithContext(ctx).Save(reconciliation).Error; err != nil {
		return fmt.Errorf("update reconciliation %d: %w", reconciliation.ID, err)
	}
	return nil
}

func (r *reconciliationRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&model.Reconciliation{}, id).Error; err != nil {
		return fmt.Errorf("delete reconciliation %d: %w", id, err)
	}
	return nil
}
