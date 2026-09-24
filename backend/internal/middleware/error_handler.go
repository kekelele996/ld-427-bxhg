package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"github.com/renovation/renovation-budget-api/internal/constants"
	"github.com/renovation/renovation-budget-api/internal/response"
)

// ErrorHandlerMiddleware 统一捕获 panic，并将 gin 错误转换为标准 JSON 响应。
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("panic recovered", slog.Any("panic", r), slog.String("stack", string(debug.Stack())))
				response.Abort(c, http.StatusInternalServerError, constants.CodeInternal, "internal server error")
			}
		}()

		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			response.Abort(c, http.StatusInternalServerError, constants.CodeInternal, fmt.Sprintf("internal server error: %v", err.Err))
		}
	}
}
