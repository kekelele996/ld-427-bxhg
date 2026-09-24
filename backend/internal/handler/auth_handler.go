package handler

import (
	"context"
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/renovation/renovation-budget-api/internal/dto"
	"github.com/renovation/renovation-budget-api/internal/response"
	"github.com/renovation/renovation-budget-api/internal/service"
)

// AuthHandler 认证处理层。
type AuthHandler struct {
	service *service.AuthService
	logger  *slog.Logger
}

// NewAuthHandler 构造认证处理层。
func NewAuthHandler(service *service.AuthService, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{service: service, logger: logger}
}

// Login 登录并签发 JWT。
// @Summary 登录
// @Description 使用预置角色账号登录并获取 JWT。
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "登录请求"
// @Success 200 {object} response.Body{data=dto.LoginResponse}
// @Failure 400 {object} response.Body
// @Failure 401 {object} response.Body
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if !bindJSON(c, &req) {
		return
	}
	result, err := h.service.Login(context.Background(), req)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, result)
}
