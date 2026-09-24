package response

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/renovation/renovation-budget-api/internal/constants"
)

// Body 统一响应结构。
type Body struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// OK 返回成功响应。
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Body{Code: constants.CodeOK, Message: constants.MessageOK, Data: data})
}

// Fail 返回统一错误响应。
func Fail(c *gin.Context, httpStatus, code int, message string) {
	c.JSON(httpStatus, Body{Code: code, Message: message, Data: nil})
}

// Abort 终止请求并返回统一错误响应。
func Abort(c *gin.Context, httpStatus, code int, message string) {
	c.AbortWithStatusJSON(httpStatus, Body{Code: code, Message: message, Data: nil})
}
