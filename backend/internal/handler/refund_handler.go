package handler

import (
	"context"
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/renovation/renovation-budget-api/internal/dto"
	"github.com/renovation/renovation-budget-api/internal/middleware"
	"github.com/renovation/renovation-budget-api/internal/response"
	"github.com/renovation/renovation-budget-api/internal/service"
)

// RefundHandler 支出退款处理层。
type RefundHandler struct {
	service *service.RefundService
	logger  *slog.Logger
}

// NewRefundHandler 构造支出退款处理层。
func NewRefundHandler(service *service.RefundService, logger *slog.Logger) *RefundHandler {
	return &RefundHandler{service: service, logger: logger}
}

// Register 登记一笔支出退款。
// @Summary 登记退款
// @Tags expenses
// @Accept json
// @Produce json
// @Param id path int true "支出记录ID"
// @Param request body dto.RegisterRefundRequest true "退款登记请求"
// @Success 200 {object} response.Body
// @Failure 400 {object} response.Body
// @Failure 404 {object} response.Body
// @Failure 409 {object} response.Body
// @Security BearerAuth
// @Router /expenses/{id}/refunds [post]
func (h *RefundHandler) Register(c *gin.Context) {
	actor, _ := middleware.CurrentActor(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	var req dto.RegisterRefundRequest
	if !bindJSON(c, &req) {
		return
	}
	refund, err := h.service.Register(context.Background(), actor, id, req)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, refund)
}

// List 查询支出退款记录。
// @Summary 查询支出退款记录
// @Tags expenses
// @Produce json
// @Param id path int true "支出记录ID"
// @Success 200 {object} response.Body
// @Failure 404 {object} response.Body
// @Security BearerAuth
// @Router /expenses/{id}/refunds [get]
func (h *RefundHandler) List(c *gin.Context) {
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	summary, err := h.service.GetSummary(context.Background(), id)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, summary)
}
