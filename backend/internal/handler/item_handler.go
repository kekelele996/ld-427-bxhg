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

// ItemHandler 预算项处理层。
type ItemHandler struct {
	service *service.ItemService
	logger  *slog.Logger
}

// NewItemHandler 构造预算项处理层。
func NewItemHandler(service *service.ItemService, logger *slog.Logger) *ItemHandler {
	return &ItemHandler{service: service, logger: logger}
}

// Create 创建预算项。
// @Summary 创建预算项
// @Tags budget-items
// @Accept json
// @Produce json
// @Param id path int true "预算表ID"
// @Param request body dto.CreateItemRequest true "创建请求"
// @Success 200 {object} response.Body
// @Failure 400 {object} response.Body
// @Security BearerAuth
// @Router /budgets/{id}/items [post]
func (h *ItemHandler) Create(c *gin.Context) {
	actor, _ := middleware.CurrentActor(c)
	budgetID, ok := pathUint(c, "id")
	if !ok {
		return
	}
	var req dto.CreateItemRequest
	if !bindJSON(c, &req) {
		return
	}
	item, err := h.service.Create(context.Background(), actor, budgetID, req)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, item)
}

// List 查询预算项列表。
// @Summary 查询预算项列表
// @Tags budget-items
// @Produce json
// @Param id path int true "预算表ID"
// @Success 200 {object} response.Body
// @Security BearerAuth
// @Router /budgets/{id}/items [get]
func (h *ItemHandler) List(c *gin.Context) {
	budgetID, ok := pathUint(c, "id")
	if !ok {
		return
	}
	items, err := h.service.List(context.Background(), budgetID)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, items)
}

// Update 更新预算项。
// @Summary 更新预算项
// @Tags budget-items
// @Accept json
// @Produce json
// @Param id path int true "预算表ID"
// @Param item_id path int true "预算项ID"
// @Param request body dto.UpdateItemRequest true "更新请求"
// @Success 200 {object} response.Body
// @Security BearerAuth
// @Router /budgets/{id}/items/{item_id} [put]
func (h *ItemHandler) Update(c *gin.Context) {
	actor, _ := middleware.CurrentActor(c)
	budgetID, ok := pathUint(c, "id")
	if !ok {
		return
	}
	itemID, ok := pathUint(c, "item_id")
	if !ok {
		return
	}
	var req dto.UpdateItemRequest
	if !bindJSON(c, &req) {
		return
	}
	item, err := h.service.Update(context.Background(), actor, budgetID, itemID, req)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, item)
}

// Delete 删除预算项。
// @Summary 删除预算项
// @Tags budget-items
// @Produce json
// @Param id path int true "预算表ID"
// @Param item_id path int true "预算项ID"
// @Success 200 {object} response.Body
// @Security BearerAuth
// @Router /budgets/{id}/items/{item_id} [delete]
func (h *ItemHandler) Delete(c *gin.Context) {
	actor, _ := middleware.CurrentActor(c)
	budgetID, ok := pathUint(c, "id")
	if !ok {
		return
	}
	itemID, ok := pathUint(c, "item_id")
	if !ok {
		return
	}
	if err := h.service.Delete(context.Background(), actor, budgetID, itemID); err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, nil)
}
