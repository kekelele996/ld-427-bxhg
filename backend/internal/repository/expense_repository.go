package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/renovation/renovation-budget-api/internal/model"
)

// ExpenseListFilter 支出记录列表过滤条件。
type ExpenseListFilter struct {
	Status       string
	BudgetItemID uint
	SupplierID   uint
	Page         int
	PageSize     int
}

// ExpenseRepository 支出记录数据访问接口。
type ExpenseRepository interface {
	Create(ctx context.Context, record *model.ExpenseRecord) error
	FindByID(ctx context.Context, id uint) (*model.ExpenseRecord, error)
	List(ctx context.Context, filter ExpenseListFilter) ([]model.ExpenseRecord, int64, error)
	Update(ctx context.Context, record *model.ExpenseRecord) error
}

type expenseRepository struct {
	db *gorm.DB
}

// NewExpenseRepository 构造支出记录仓储。
func NewExpenseRepository(db *gorm.DB) ExpenseRepository {
	return &expenseRepository{db: db}
}

func (r *expenseRepository) Create(ctx context.Context, record *model.ExpenseRecord) error {
	if err := r.db.WithContext(ctx).Create(record).Error; err != nil {
		return fmt.Errorf("create expense record: %w", err)
	}
	return nil
}

func (r *expenseRepository) FindByID(ctx context.Context, id uint) (*model.ExpenseRecord, error) {
	var record model.ExpenseRecord
	if err := r.db.WithContext(ctx).First(&record, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find expense record %d: %w", id, err)
	}
	return &record, nil
}

func (r *expenseRepository) List(ctx context.Context, filter ExpenseListFilter) ([]model.ExpenseRecord, int64, error) {
	page, pageSize := normalizePage(filter.Page, filter.PageSize)
	q := r.db.WithContext(ctx).Model(&model.ExpenseRecord{})
	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}
	if filter.BudgetItemID != 0 {
		q = q.Where("budget_item_id = ?", filter.BudgetItemID)
	}
	if filter.SupplierID != 0 {
		q = q.Where("supplier_id = ?", filter.SupplierID)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count expense records: %w", err)
	}
	var records []model.ExpenseRecord
	if err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&records).Error; err != nil {
		return nil, 0, fmt.Errorf("list expense records: %w", err)
	}
	return records, total, nil
}

func (r *expenseRepository) Update(ctx context.Context, record *model.ExpenseRecord) error {
	if err := r.db.WithContext(ctx).Save(record).Error; err != nil {
		return fmt.Errorf("update expense record %d: %w", record.ID, err)
	}
	return nil
}
