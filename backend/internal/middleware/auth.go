package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/renovation/renovation-budget-api/internal/auth"
	"github.com/renovation/renovation-budget-api/internal/constants"
	"github.com/renovation/renovation-budget-api/internal/model"
	"github.com/renovation/renovation-budget-api/internal/response"
)

// JWTAuthMiddleware 校验 Bearer Token 并注入操作者。
func JWTAuthMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			response.Abort(c, http.StatusUnauthorized, constants.CodeUnauthorized, "missing authorization header")
			return
		}
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
			response.Abort(c, http.StatusUnauthorized, constants.CodeUnauthorized, "invalid authorization header")
			return
		}
		claims, err := auth.ParseToken(secret, strings.TrimSpace(parts[1]))
		if err != nil {
			response.Abort(c, http.StatusUnauthorized, constants.CodeUnauthorized, "invalid or expired token")
			return
		}
		SetActor(c, model.Actor{
			UserID:   claims.UserID,
			Username: claims.Username,
			Role:     claims.Role,
			IP:       c.ClientIP(),
		})
		c.Next()
	}
}
