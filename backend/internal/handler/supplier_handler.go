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

// SupplierHandler 供应商处理层。
type SupplierHandler struct {
	service *service.SupplierService
	logger  *slog.Logger
}

// NewSupplierHandler 构造供应商处理层。
func NewSupplierHandler(service *service.SupplierService, logger *slog.Logger) *SupplierHandler {
	return &SupplierHandler{service: service, logger: logger}
}

// Create 创建供应商。
// @Summary 创建供应商
// @Tags suppliers
// @Accept json
// @Produce json
// @Param request body dto.CreateSupplierRequest true "创建请求"
// @Success 200 {object} response.Body
// @Failure 400 {object} response.Body
// @Security BearerAuth
// @Router /suppliers [post]
func (h *SupplierHandler) Create(c *gin.Context) {
	actor, _ := middleware.CurrentActor(c)
	var req dto.CreateSupplierRequest
	if !bindJSON(c, &req) {
		return
	}
	supplier, err := h.service.Create(context.Background(), actor, req)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, supplier)
}

// List 查询供应商列表。
// @Summary 查询供应商列表
// @Tags suppliers
// @Produce json
// @Param status query string false "合作状态"
// @Param category query string false "类别"
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Success 200 {object} response.Body
// @Security BearerAuth
// @Router /suppliers [get]
func (h *SupplierHandler) List(c *gin.Context) {
	var filter dto.SupplierFilter
	if !bindQuery(c, &filter) {
		return
	}
	suppliers, total, err := h.service.List(context.Background(), filter)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, dto.PageResult{List: suppliers, Total: total, Page: filter.Page, PageSize: filter.PageSize})
}

// Get 获取供应商详情。
// @Summary 获取供应商详情
// @Tags suppliers
// @Produce json
// @Param id path int true "供应商ID"
// @Success 200 {object} response.Body
// @Failure 404 {object} response.Body
// @Security BearerAuth
// @Router /suppliers/{id} [get]
func (h *SupplierHandler) Get(c *gin.Context) {
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	supplier, err := h.service.Get(context.Background(), id)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, supplier)
}

// Update 更新供应商。
// @Summary 更新供应商
// @Tags suppliers
// @Accept json
// @Produce json
// @Param id path int true "供应商ID"
// @Param request body dto.UpdateSupplierRequest true "更新请求"
// @Success 200 {object} response.Body
// @Security BearerAuth
// @Router /suppliers/{id} [put]
func (h *SupplierHandler) Update(c *gin.Context) {
	actor, _ := middleware.CurrentActor(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateSupplierRequest
	if !bindJSON(c, &req) {
		return
	}
	supplier, err := h.service.Update(context.Background(), actor, id, req)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, supplier)
}

// Delete 删除供应商。
// @Summary 删除供应商
// @Tags suppliers
// @Produce json
// @Param id path int true "供应商ID"
// @Success 200 {object} response.Body
// @Security BearerAuth
// @Router /suppliers/{id} [delete]
func (h *SupplierHandler) Delete(c *gin.Context) {
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
