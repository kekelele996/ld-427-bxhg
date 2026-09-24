package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/redis/go-redis/v9"

	"github.com/renovation/renovation-budget-api/internal/dto"
	"github.com/renovation/renovation-budget-api/internal/model"
	"github.com/renovation/renovation-budget-api/internal/repository"
)

// ItemService 预算项业务逻辑。
type ItemService struct {
	repo       repository.ItemRepository
	budgetRepo repository.BudgetRepository
	audit      *AuditService
	rdb        *redis.Client
	logger     *slog.Logger
}

// NewItemService 构造预算项服务。
func NewItemService(repo repository.ItemRepository, budgetRepo repository.BudgetRepository, audit *AuditService, rdb *redis.Client, logger *slog.Logger) *ItemService {
	return &ItemService{repo: repo, budgetRepo: budgetRepo, audit: audit, rdb: rdb, logger: logger}
}

// Create 创建预算项。
func (s *ItemService) Create(ctx context.Context, actor model.Actor, budgetSheetID uint, req dto.CreateItemRequest) (*model.BudgetItem, error) {
	if _, err := s.budgetRepo.FindByID(ctx, budgetSheetID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get budget sheet %d: %w", budgetSheetID, err)
	}
	item := &model.BudgetItem{
		BudgetSheetID:  budgetSheetID,
		Category:       req.Category,
		SubCategory:    req.SubCategory,
		BudgetAmount:   req.BudgetAmount,
		VarianceAmount: CalculateVariance(0, req.BudgetAmount),
		SortOrder:      req.SortOrder,
		Remark:         req.Remark,
	}
	if err := s.repo.Create(ctx, item); err != nil {
		return nil, fmt.Errorf("create budget item: %w", err)
	}
	s.invalidateBudget(ctx, budgetSheetID)
	s.audit.Record(ctx, actor, "budget_item_create", "budget_item", item.ID, fmt.Sprintf("budget_sheet_id=%d category=%s", budgetSheetID, item.Category))
	return item, nil
}

// List 查询预算项列表。
func (s *ItemService) List(ctx context.Context, budgetSheetID uint) ([]model.BudgetItem, error) {
	items, err := s.repo.ListByBudgetID(ctx, budgetSheetID)
	if err != nil {
		return nil, fmt.Errorf("list budget items: %w", err)
	}
	return items, nil
}

// Update 更新预算项。
func (s *ItemService) Update(ctx context.Context, actor model.Actor, budgetSheetID, itemID uint, req dto.UpdateItemRequest) (*model.BudgetItem, error) {
	item, err := s.repo.FindByID(ctx, itemID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get budget item %d: %w", itemID, err)
	}
	if item.BudgetSheetID != budgetSheetID {
		return nil, fmt.Errorf("budget item %d belongs to sheet %d: %w", itemID, item.BudgetSheetID, ErrNotFound)
	}
	if req.Category != "" {
		item.Category = req.Category
	}
	if req.SubCategory != "" {
		item.SubCategory = req.SubCategory
	}
	if req.BudgetAmount > 0 {
		item.BudgetAmount = req.BudgetAmount
	}
	if req.SortOrder != nil {
		item.SortOrder = *req.SortOrder
	}
	if req.Remark != "" {
		item.Remark = req.Remark
	}
	item.VarianceAmount = CalculateVariance(item.SpentAmount, item.BudgetAmount)
	if err := s.repo.Update(ctx, item); err != nil {
		return nil, fmt.Errorf("update budget item %d: %w", itemID, err)
	}
	s.invalidateBudget(ctx, budgetSheetID)
	s.audit.Record(ctx, actor, "budget_item_update", "budget_item", item.ID, fmt.Sprintf("budget_sheet_id=%d", budgetSheetID))
	return item, nil
}

// Delete 删除预算项。
func (s *ItemService) Delete(ctx context.Context, actor model.Actor, budgetSheetID, itemID uint) error {
	item, err := s.repo.FindByID(ctx, itemID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("get budget item %d: %w", itemID, err)
	}
	if item.BudgetSheetID != budgetSheetID {
		return fmt.Errorf("budget item %d belongs to sheet %d: %w", itemID, item.BudgetSheetID, ErrNotFound)
	}
	if err := s.repo.Delete(ctx, itemID); err != nil {
		return fmt.Errorf("delete budget item %d: %w", itemID, err)
	}
	s.invalidateBudget(ctx, budgetSheetID)
	s.audit.Record(ctx, actor, "budget_item_delete", "budget_item", itemID, fmt.Sprintf("budget_sheet_id=%d", budgetSheetID))
	return nil
}

func (s *ItemService) invalidateBudget(ctx context.Context, budgetSheetID uint) {
	if s.rdb == nil {
		return
	}
	if err := s.rdb.Del(ctx, budgetSnapshotKey(budgetSheetID)).Err(); err != nil {
		s.logger.Warn("invalidate budget snapshot failed", slog.Uint64("budget_id", uint64(budgetSheetID)), slog.String("error", err.Error()))
	}
}
