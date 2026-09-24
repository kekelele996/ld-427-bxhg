package service

import (
	"context"
	"encoding/json"
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

const budgetSnapshotTTL = 5 * time.Minute

func budgetSnapshotKey(id uint) string {
	return fmt.Sprintf("budget_snapshot:%d", id)
}

// BudgetService 预算表业务逻辑。
type BudgetService struct {
	repo     repository.BudgetRepository
	itemRepo repository.ItemRepository
	audit    *AuditService
	rdb      *redis.Client
	logger   *slog.Logger
}

// NewBudgetService 构造预算表服务。
func NewBudgetService(repo repository.BudgetRepository, itemRepo repository.ItemRepository, audit *AuditService, rdb *redis.Client, logger *slog.Logger) *BudgetService {
	return &BudgetService{repo: repo, itemRepo: itemRepo, audit: audit, rdb: rdb, logger: logger}
}

// Create 创建预算表。
func (s *BudgetService) Create(ctx context.Context, actor model.Actor, req dto.CreateBudgetRequest) (*model.BudgetSheet, error) {
	status := req.Status
	if status == "" {
		status = constants.BudgetStatusDraft
	}
	sheet := &model.BudgetSheet{
		ProjectID:       req.ProjectID,
		Name:            req.Name,
		TotalAmount:     req.TotalAmount,
		AvailableAmount: req.TotalAmount,
		Status:          status,
		CreatedByID:     actor.UserID,
		Version:         1,
	}
	if err := s.repo.Create(ctx, sheet); err != nil {
		return nil, fmt.Errorf("create budget sheet: %w", err)
	}
	s.audit.Record(ctx, actor, "budget_create", "budget", sheet.ID, fmt.Sprintf("name=%s total=%.2f", sheet.Name, sheet.TotalAmount))
	return sheet, nil
}

// Get 获取预算表（优先读取 Redis 快照）。
func (s *BudgetService) Get(ctx context.Context, id uint) (*model.BudgetSheet, error) {
	if s.rdb != nil {
		key := budgetSnapshotKey(id)
		raw, err := s.rdb.Get(ctx, key).Result()
		if err == nil {
			var sheet model.BudgetSheet
			if json.Unmarshal([]byte(raw), &sheet) == nil {
				return &sheet, nil
			}
		}
	}

	sheet, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get budget sheet %d: %w", id, err)
	}

	if s.rdb != nil {
		if raw, err := json.Marshal(sheet); err == nil {
			_ = s.rdb.Set(ctx, budgetSnapshotKey(id), raw, budgetSnapshotTTL).Err()
		}
	}
	return sheet, nil
}

// List 分页查询预算表。
func (s *BudgetService) List(ctx context.Context, filter dto.BudgetFilter) ([]model.BudgetSheet, int64, error) {
	sheets, total, err := s.repo.List(ctx, repository.BudgetListFilter{
		ProjectID: filter.ProjectID,
		Status:    filter.Status,
		Page:      filter.Page,
		PageSize:  filter.PageSize,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list budget sheets: %w", err)
	}
	return sheets, total, nil
}

// Update 更新预算表基本信息。
func (s *BudgetService) Update(ctx context.Context, actor model.Actor, id uint, req dto.UpdateBudgetRequest) (*model.BudgetSheet, error) {
	sheet, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get budget sheet %d: %w", id, err)
	}
	if req.Name != "" {
		sheet.Name = req.Name
	}
	if req.TotalAmount > 0 {
		sheet.TotalAmount = req.TotalAmount
	}
	if req.Status != "" {
		sheet.Status = req.Status
	}
	sheet.AvailableAmount = CalculateAvailable(sheet.TotalAmount, sheet.SpentAmount, sheet.FrozenAmount)
	if err := s.repo.Update(ctx, sheet); err != nil {
		return nil, fmt.Errorf("update budget sheet %d: %w", id, err)
	}
	s.invalidateSnapshot(ctx, id)
	s.audit.Record(ctx, actor, "budget_update", "budget", id, fmt.Sprintf("name=%s status=%s", sheet.Name, sheet.Status))
	return sheet, nil
}

// Adjust 调整预算总额并递增版本号。
func (s *BudgetService) Adjust(ctx context.Context, actor model.Actor, id uint, req dto.AdjustBudgetRequest) (*model.BudgetSheet, error) {
	sheet, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get budget sheet %d: %w", id, err)
	}
	if sheet.Status == constants.BudgetStatusArchived {
		return nil, fmt.Errorf("adjust budget sheet %d: %w", id, ErrInvalidState)
	}
	sheet.TotalAmount = req.TotalAmount
	sheet.Version++
	sheet.AvailableAmount = CalculateAvailable(sheet.TotalAmount, sheet.SpentAmount, sheet.FrozenAmount)
	if err := s.repo.Update(ctx, sheet); err != nil {
		return nil, fmt.Errorf("adjust budget sheet %d: %w", id, err)
	}
	s.invalidateSnapshot(ctx, id)
	s.audit.Record(ctx, actor, "budget_adjust", "budget", id, fmt.Sprintf("total=%.2f version=%d reason=%s", sheet.TotalAmount, sheet.Version, req.Reason))
	return sheet, nil
}

// Delete 删除预算表，仅允许 Draft 状态。
func (s *BudgetService) Delete(ctx context.Context, actor model.Actor, id uint) error {
	sheet, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("get budget sheet %d: %w", id, err)
	}
	if sheet.Status != constants.BudgetStatusDraft {
		return fmt.Errorf("delete budget sheet %d: %w", id, ErrInvalidState)
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete budget sheet %d: %w", id, err)
	}
	s.invalidateSnapshot(ctx, id)
	s.audit.Record(ctx, actor, "budget_delete", "budget", id, fmt.Sprintf("name=%s", sheet.Name))
	return nil
}

// EnsureBudgetExists 校验预算表存在，供其他服务复用。
func (s *BudgetService) EnsureBudgetExists(ctx context.Context, id uint) (*model.BudgetSheet, error) {
	sheet, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get budget sheet %d: %w", id, err)
	}
	return sheet, nil
}

func (s *BudgetService) invalidateSnapshot(ctx context.Context, id uint) {
	if s.rdb == nil {
		return
	}
	if err := s.rdb.Del(ctx, budgetSnapshotKey(id)).Err(); err != nil {
		s.logger.Warn("invalidate budget snapshot failed", slog.Uint64("budget_id", uint64(id)), slog.String("error", err.Error()))
	}
}
