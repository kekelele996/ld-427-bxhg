package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/renovation/renovation-budget-api/internal/model"
	"github.com/renovation/renovation-budget-api/internal/repository"
)

// AuditLogMiddleware 对写操作记录通用 HTTP 审计日志。
func AuditLogMiddleware(auditRepo repository.AuditRepository, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead || c.Request.Method == http.MethodOptions {
			return
		}
		actor, ok := CurrentActor(c)
		if !ok {
			return
		}
		entry := &model.AuditLog{
			UserID:     actor.UserID,
			Username:   actor.Username,
			Action:     fmt.Sprintf("http.%s", c.Request.Method),
			Resource:   c.FullPath(),
			ResourceID: c.Param("id"),
			Detail:     fmt.Sprintf("%s %s status=%d", c.Request.Method, c.Request.URL.Path, c.Writer.Status()),
			IP:         actor.IP,
			CreatedAt:  time.Now(),
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := auditRepo.Create(ctx, entry); err != nil {
			logger.Warn("write audit log failed", slog.String("error", err.Error()))
		}
	}
}
