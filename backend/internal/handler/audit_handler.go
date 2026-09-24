package handler

import (
	"context"
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/renovation/renovation-budget-api/internal/dto"
	"github.com/renovation/renovation-budget-api/internal/response"
	"github.com/renovation/renovation-budget-api/internal/service"
)

// AuditHandler 审计日志处理层。
type AuditHandler struct {
	service *service.AuditService
	logger  *slog.Logger
}

// NewAuditHandler 构造审计日志处理层。
func NewAuditHandler(service *service.AuditService, logger *slog.Logger) *AuditHandler {
	return &AuditHandler{service: service, logger: logger}
}

// List 查询审计日志。
// @Summary 查询审计日志
// @Tags audit-logs
// @Produce json
// @Param action query string false "动作"
// @Param resource query string false "资源"
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Success 200 {object} response.Body
// @Security BearerAuth
// @Router /audit-logs [get]
func (h *AuditHandler) List(c *gin.Context) {
	var filter dto.AuditLogFilter
	if !bindQuery(c, &filter) {
		return
	}
	logs, total, err := h.service.List(context.Background(), filter.Action, filter.Resource, filter.Page, filter.PageSize)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, dto.PageResult{List: logs, Total: total, Page: filter.Page, PageSize: filter.PageSize})
}
