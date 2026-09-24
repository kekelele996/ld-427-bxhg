package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/renovation/renovation-budget-api/internal/model"
)

// SupplierListFilter 供应商列表过滤条件。
type SupplierListFilter struct {
	Status   string
	Category string
	Page     int
	PageSize int
}

// SupplierRepository 供应商数据访问接口。
type SupplierRepository interface {
	Create(ctx context.Context, supplier *model.Supplier) error
	FindByID(ctx context.Context, id uint) (*model.Supplier, error)
	List(ctx context.Context, filter SupplierListFilter) ([]model.Supplier, int64, error)
	Update(ctx context.Context, supplier *model.Supplier) error
	Delete(ctx context.Context, id uint) error
}

type supplierRepository struct {
	db *gorm.DB
}

// NewSupplierRepository 构造供应商仓储。
func NewSupplierRepository(db *gorm.DB) SupplierRepository {
	return &supplierRepository{db: db}
}

func (r *supplierRepository) Create(ctx context.Context, supplier *model.Supplier) error {
	if err := r.db.WithContext(ctx).Create(supplier).Error; err != nil {
		return fmt.Errorf("create supplier: %w", err)
	}
	return nil
}

func (r *supplierRepository) FindByID(ctx context.Context, id uint) (*model.Supplier, error) {
	var supplier model.Supplier
	if err := r.db.WithContext(ctx).First(&supplier, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find supplier %d: %w", id, err)
	}
	return &supplier, nil
}

func (r *supplierRepository) List(ctx context.Context, filter SupplierListFilter) ([]model.Supplier, int64, error) {
	page, pageSize := normalizePage(filter.Page, filter.PageSize)
	q := r.db.WithContext(ctx).Model(&model.Supplier{})
	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}
	if filter.Category != "" {
		q = q.Where("category = ?", filter.Category)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count suppliers: %w", err)
	}
	var suppliers []model.Supplier
	if err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&suppliers).Error; err != nil {
		return nil, 0, fmt.Errorf("list suppliers: %w", err)
	}
	return suppliers, total, nil
}

func (r *supplierRepository) Update(ctx context.Context, supplier *model.Supplier) error {
	if err := r.db.WithContext(ctx).Save(supplier).Error; err != nil {
		return fmt.Errorf("update supplier %d: %w", supplier.ID, err)
	}
	return nil
}

func (r *supplierRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&model.Supplier{}, id).Error; err != nil {
		return fmt.Errorf("delete supplier %d: %w", id, err)
	}
	return nil
}
