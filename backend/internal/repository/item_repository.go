package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/renovation/renovation-budget-api/internal/model"
)

// ItemRepository 预算项数据访问接口。
type ItemRepository interface {
	Create(ctx context.Context, item *model.BudgetItem) error
	FindByID(ctx context.Context, id uint) (*model.BudgetItem, error)
	ListByBudgetID(ctx context.Context, budgetSheetID uint) ([]model.BudgetItem, error)
	Update(ctx context.Context, item *model.BudgetItem) error
	Delete(ctx context.Context, id uint) error
}

type itemRepository struct {
	db *gorm.DB
}

// NewItemRepository 构造预算项仓储。
func NewItemRepository(db *gorm.DB) ItemRepository {
	return &itemRepository{db: db}
}

func (r *itemRepository) Create(ctx context.Context, item *model.BudgetItem) error {
	if err := r.db.WithContext(ctx).Create(item).Error; err != nil {
		return fmt.Errorf("create budget item: %w", err)
	}
	return nil
}

func (r *itemRepository) FindByID(ctx context.Context, id uint) (*model.BudgetItem, error) {
	var item model.BudgetItem
	if err := r.db.WithContext(ctx).First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find budget item %d: %w", id, err)
	}
	return &item, nil
}

func (r *itemRepository) ListByBudgetID(ctx context.Context, budgetSheetID uint) ([]model.BudgetItem, error) {
	var items []model.BudgetItem
	if err := r.db.WithContext(ctx).Where("budget_sheet_id = ?", budgetSheetID).Order("sort_order ASC, id ASC").Find(&items).Error; err != nil {
		return nil, fmt.Errorf("list budget items for sheet %d: %w", budgetSheetID, err)
	}
	return items, nil
}

func (r *itemRepository) Update(ctx context.Context, item *model.BudgetItem) error {
	if err := r.db.WithContext(ctx).Save(item).Error; err != nil {
		return fmt.Errorf("update budget item %d: %w", item.ID, err)
	}
	return nil
}

func (r *itemRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&model.BudgetItem{}, id).Error; err != nil {
		return fmt.Errorf("delete budget item %d: %w", id, err)
	}
	return nil
}
