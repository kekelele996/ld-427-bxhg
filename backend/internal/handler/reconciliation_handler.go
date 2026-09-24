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

// ReconciliationHandler 对账单处理层。
type ReconciliationHandler struct {
	service *service.ReconciliationService
	logger  *slog.Logger
}

// NewReconciliationHandler 构造对账单处理层。
func NewReconciliationHandler(service *service.ReconciliationService, logger *slog.Logger) *ReconciliationHandler {
	return &ReconciliationHandler{service: service, logger: logger}
}

// Create 创建对账单。
// @Summary 创建对账单
// @Tags reconciliations
// @Accept json
// @Produce json
// @Param request body dto.CreateReconciliationRequest true "创建请求"
// @Success 200 {object} response.Body
// @Failure 400 {object} response.Body
// @Security BearerAuth
// @Router /reconciliations [post]
func (h *ReconciliationHandler) Create(c *gin.Context) {
	actor, _ := middleware.CurrentActor(c)
	var req dto.CreateReconciliationRequest
	if !bindJSON(c, &req) {
		return
	}
	record, err := h.service.Create(context.Background(), actor, req)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, record)
}

// List 查询对账单列表。
// @Summary 查询对账单列表
// @Tags reconciliations
// @Produce json
// @Param project_id query string false "项目ID"
// @Param period query string false "对账期间"
// @Param supplier_id query int false "供应商ID"
// @Param status query string false "状态"
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Success 200 {object} response.Body
// @Security BearerAuth
// @Router /reconciliations [get]
func (h *ReconciliationHandler) List(c *gin.Context) {
	var filter dto.ReconciliationFilter
	if !bindQuery(c, &filter) {
		return
	}
	records, total, err := h.service.List(context.Background(), filter)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, dto.PageResult{List: records, Total: total, Page: filter.Page, PageSize: filter.PageSize})
}

// Get 获取对账单详情。
// @Summary 获取对账单详情
// @Tags reconciliations
// @Produce json
// @Param id path int true "对账单ID"
// @Success 200 {object} response.Body
// @Failure 404 {object} response.Body
// @Security BearerAuth
// @Router /reconciliations/{id} [get]
func (h *ReconciliationHandler) Get(c *gin.Context) {
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	record, err := h.service.Get(context.Background(), id)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, record)
}

// Update 更新对账单金额。
// @Summary 更新对账单金额
// @Tags reconciliations
// @Accept json
// @Produce json
// @Param id path int true "对账单ID"
// @Param request body dto.UpdateReconciliationRequest true "更新请求"
// @Success 200 {object} response.Body
// @Security BearerAuth
// @Router /reconciliations/{id} [put]
func (h *ReconciliationHandler) Update(c *gin.Context) {
	actor, _ := middleware.CurrentActor(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateReconciliationRequest
	if !bindJSON(c, &req) {
		return
	}
	record, err := h.service.Update(context.Background(), actor, id, req)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, record)
}

// Confirm 确认对账。
// @Summary 确认对账
// @Tags reconciliations
// @Produce json
// @Param id path int true "对账单ID"
// @Success 200 {object} response.Body
// @Failure 409 {object} response.Body
// @Security BearerAuth
// @Router /reconciliations/{id}/confirm [post]
func (h *ReconciliationHandler) Confirm(c *gin.Context) {
	actor, _ := middleware.CurrentActor(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	record, err := h.service.Confirm(context.Background(), actor, id)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, record)
}

// Dispute 标记对账争议。
// @Summary 标记对账争议
// @Tags reconciliations
// @Produce json
// @Param id path int true "对账单ID"
// @Success 200 {object} response.Body
// @Security BearerAuth
// @Router /reconciliations/{id}/dispute [post]
func (h *ReconciliationHandler) Dispute(c *gin.Context) {
	actor, _ := middleware.CurrentActor(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	record, err := h.service.Dispute(context.Background(), actor, id)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, record)
}

// Resolve 解决对账争议。
// @Summary 解决对账争议
// @Tags reconciliations
// @Produce json
// @Param id path int true "对账单ID"
// @Success 200 {object} response.Body
// @Security BearerAuth
// @Router /reconciliations/{id}/resolve [post]
func (h *ReconciliationHandler) Resolve(c *gin.Context) {
	actor, _ := middleware.CurrentActor(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	record, err := h.service.Resolve(context.Background(), actor, id)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, record)
}

// Delete 删除对账单。
// @Summary 删除对账单
// @Tags reconciliations
// @Produce json
// @Param id path int true "对账单ID"
// @Success 200 {object} response.Body
// @Security BearerAuth
// @Router /reconciliations/{id} [delete]
func (h *ReconciliationHandler) Delete(c *gin.Context) {
	actor, _ := middleware.CurrentActor(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	if err := h.service.Delete(context.Background(), actor, id); err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, nil)
}
