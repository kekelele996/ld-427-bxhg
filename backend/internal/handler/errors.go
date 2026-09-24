package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/renovation/renovation-budget-api/internal/constants"
	"github.com/renovation/renovation-budget-api/internal/response"
	"github.com/renovation/renovation-budget-api/internal/service"
)

// handleError 将服务层错误转换为统一 HTTP 响应。
func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrNotFound):
		response.Abort(c, http.StatusNotFound, constants.CodeNotFound, "resource not found")
	case errors.Is(err, service.ErrInvalidLogin):
		response.Abort(c, http.StatusUnauthorized, constants.CodeUnauthorized, "invalid username or password")
	case errors.Is(err, service.ErrInvalidState), errors.Is(err, service.ErrInsufficientBalance), errors.Is(err, service.ErrForbiddenTransition):
		response.Abort(c, http.StatusConflict, constants.CodeConflict, err.Error())
	default:
		response.Abort(c, http.StatusInternalServerError, constants.CodeInternal, "internal server error")
	}
}

// bindJSON 绑定并校验 JSON 请求体。
func bindJSON(c *gin.Context, obj any) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		response.Abort(c, http.StatusBadRequest, constants.CodeBadRequest, err.Error())
		return false
	}
	return true
}

// bindQuery 绑定并校验查询参数。
func bindQuery(c *gin.Context, obj any) bool {
	if err := c.ShouldBindQuery(obj); err != nil {
		response.Abort(c, http.StatusBadRequest, constants.CodeBadRequest, err.Error())
		return false
	}
	return true
}

// pathUint 解析路径中的无符号整数。
func pathUint(c *gin.Context, key string) (uint, bool) {
	value, err := strconv.ParseUint(c.Param(key), 10, 64)
	if err != nil || value == 0 {
		response.Abort(c, http.StatusBadRequest, constants.CodeBadRequest, "invalid path parameter: "+key)
		return 0, false
	}
	return uint(value), true
}
