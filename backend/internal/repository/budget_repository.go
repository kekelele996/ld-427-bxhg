package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/renovation/renovation-budget-api/internal/model"
)

// BudgetListFilter 预算表列表过滤条件。
type BudgetListFilter struct {
	ProjectID string
	Status    string
	Page      int
	PageSize  int
}

// BudgetRepository 预算表数据访问接口。
type BudgetRepository interface {
	Create(ctx context.Context, sheet *model.BudgetSheet) error
	FindByID(ctx context.Context, id uint) (*model.BudgetSheet, error)
	List(ctx context.Context, filter BudgetListFilter) ([]model.BudgetSheet, int64, error)
	Update(ctx context.Context, sheet *model.BudgetSheet) error
	Delete(ctx context.Context, id uint) error
}

type budgetRepository struct {
	db *gorm.DB
}

// NewBudgetRepository 构造预算表仓储。
func NewBudgetRepository(db *gorm.DB) BudgetRepository {
	return &budgetRepository{db: db}
}

func (r *budgetRepository) Create(ctx context.Context, sheet *model.BudgetSheet) error {
	if err := r.db.WithContext(ctx).Create(sheet).Error; err != nil {
		return fmt.Errorf("create budget sheet: %w", err)
	}
	return nil
}

func (r *budgetRepository) FindByID(ctx context.Context, id uint) (*model.BudgetSheet, error) {
	var sheet model.BudgetSheet
	if err := r.db.WithContext(ctx).Preload("Items").First(&sheet, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find budget sheet %d: %w", id, err)
	}
	return &sheet, nil
}

func (r *budgetRepository) List(ctx context.Context, filter BudgetListFilter) ([]model.BudgetSheet, int64, error) {
	page, pageSize := normalizePage(filter.Page, filter.PageSize)
	q := r.db.WithContext(ctx).Model(&model.BudgetSheet{})
	if filter.ProjectID != "" {
		q = q.Where("project_id = ?", filter.ProjectID)
	}
	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count budget sheets: %w", err)
	}
	var sheets []model.BudgetSheet
	if err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&sheets).Error; err != nil {
		return nil, 0, fmt.Errorf("list budget sheets: %w", err)
	}
	return sheets, total, nil
}

func (r *budgetRepository) Update(ctx context.Context, sheet *model.BudgetSheet) error {
	if err := r.db.WithContext(ctx).Save(sheet).Error; err != nil {
		return fmt.Errorf("update budget sheet %d: %w", sheet.ID, err)
	}
	return nil
}

func (r *budgetRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&model.BudgetSheet{}, id).Error; err != nil {
		return fmt.Errorf("delete budget sheet %d: %w", id, err)
	}
	return nil
}

func normalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}
