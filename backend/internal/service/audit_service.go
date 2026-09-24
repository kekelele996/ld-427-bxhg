package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/renovation/renovation-budget-api/internal/model"
	"github.com/renovation/renovation-budget-api/internal/repository"
)

// AuditService 审计日志业务逻辑。
type AuditService struct {
	repo   repository.AuditRepository
	logger *slog.Logger
}

// NewAuditService 构造审计日志服务。
func NewAuditService(repo repository.AuditRepository, logger *slog.Logger) *AuditService {
	return &AuditService{repo: repo, logger: logger}
}

// Record 记录一次业务审计日志。审计写入失败不影响主流程。
func (s *AuditService) Record(ctx context.Context, actor model.Actor, action, resource string, resourceID uint, detail string) {
	entry := &model.AuditLog{
		UserID:     actor.UserID,
		Username:   actor.Username,
		Action:     action,
		Resource:   resource,
		ResourceID: fmt.Sprintf("%d", resourceID),
		Detail:     detail,
		IP:         actor.IP,
	}
	if err := s.repo.Create(ctx, entry); err != nil {
		s.logger.Warn("write business audit log failed", slog.String("action", action), slog.String("error", err.Error()))
	}
}

// List 查询审计日志。
func (s *AuditService) List(ctx context.Context, action, resource string, page, pageSize int) ([]model.AuditLog, int64, error) {
	logs, total, err := s.repo.List(ctx, repository.AuditListFilter{
		Action:   action,
		Resource: resource,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list audit logs: %w", err)
	}
	return logs, total, nil
}
