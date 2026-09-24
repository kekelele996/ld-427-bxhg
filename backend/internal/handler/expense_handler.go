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

// ExpenseHandler 支出记录处理层。
type ExpenseHandler struct {
	service *service.ExpenseService
	logger  *slog.Logger
}

// NewExpenseHandler 构造支出记录处理层。
func NewExpenseHandler(service *service.ExpenseService, logger *slog.Logger) *ExpenseHandler {
	return &ExpenseHandler{service: service, logger: logger}
}

// Create 创建支出记录。
// @Summary 创建支出记录
// @Tags expenses
// @Accept json
// @Produce json
// @Param request body dto.CreateExpenseRequest true "创建请求"
// @Success 200 {object} response.Body
// @Failure 400 {object} response.Body
// @Security BearerAuth
// @Router /expenses [post]
func (h *ExpenseHandler) Create(c *gin.Context) {
	actor, _ := middleware.CurrentActor(c)
	var req dto.CreateExpenseRequest
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

// List 查询支出记录列表。
// @Summary 查询支出记录列表
// @Tags expenses
// @Produce json
// @Param status query string false "审批状态"
// @Param budget_item_id query int false "预算项ID"
// @Param supplier_id query int false "供应商ID"
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Success 200 {object} response.Body
// @Security BearerAuth
// @Router /expenses [get]
func (h *ExpenseHandler) List(c *gin.Context) {
	var filter dto.ExpenseFilter
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

// Get 获取支出记录详情。
// @Summary 获取支出记录详情
// @Tags expenses
// @Produce json
// @Param id path int true "支出记录ID"
// @Success 200 {object} response.Body
// @Failure 404 {object} response.Body
// @Security BearerAuth
// @Router /expenses/{id} [get]
func (h *ExpenseHandler) Get(c *gin.Context) {
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

// Submit 提交支出并冻结预算。
// @Summary 提交支出
// @Tags expenses
// @Produce json
// @Param id path int true "支出记录ID"
// @Success 200 {object} response.Body
// @Failure 409 {object} response.Body
// @Security BearerAuth
// @Router /expenses/{id}/submit [post]
func (h *ExpenseHandler) Submit(c *gin.Context) {
	actor, _ := middleware.CurrentActor(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	record, err := h.service.Submit(context.Background(), actor, id)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, record)
}

// Approve 审批通过支出。
// @Summary 审批通过支出
// @Tags expenses
// @Accept json
// @Produce json
// @Param id path int true "支出记录ID"
// @Param request body dto.ApproveExpenseRequest true "审批意见"
// @Success 200 {object} response.Body
// @Failure 403 {object} response.Body
// @Failure 409 {object} response.Body
// @Security BearerAuth
// @Router /expenses/{id}/approve [post]
func (h *ExpenseHandler) Approve(c *gin.Context) {
	actor, _ := middleware.CurrentActor(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	var req dto.ApproveExpenseRequest
	if !bindJSON(c, &req) {
		return
	}
	record, err := h.service.Approve(context.Background(), actor, id, req)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, record)
}

// Reject 驳回支出。
// @Summary 驳回支出
// @Tags expenses
// @Accept json
// @Produce json
// @Param id path int true "支出记录ID"
// @Param request body dto.RejectExpenseRequest true "驳回意见"
// @Success 200 {object} response.Body
// @Security BearerAuth
// @Router /expenses/{id}/reject [post]
func (h *ExpenseHandler) Reject(c *gin.Context) {
	actor, _ := middleware.CurrentActor(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	var req dto.RejectExpenseRequest
	if !bindJSON(c, &req) {
		return
	}
	record, err := h.service.Reject(context.Background(), actor, id, req)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, record)
}

// Pay 确认付款。
// @Summary 确认付款
// @Tags expenses
// @Accept json
// @Produce json
// @Param id path int true "支出记录ID"
// @Param request body dto.PayExpenseRequest true "付款日期"
// @Success 200 {object} response.Body
// @Security BearerAuth
// @Router /expenses/{id}/pay [post]
func (h *ExpenseHandler) Pay(c *gin.Context) {
	actor, _ := middleware.CurrentActor(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	var req dto.PayExpenseRequest
	if !bindJSON(c, &req) {
		return
	}
	record, err := h.service.Pay(context.Background(), actor, id, req)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, record)
}
