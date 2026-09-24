package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/renovation/renovation-budget-api/internal/constants"
	"github.com/renovation/renovation-budget-api/internal/response"
)

// RateLimitMiddleware 基于 Redis 的固定窗口限流，按客户端 IP + 路由路径计数。
func RateLimitMiddleware(rdb *redis.Client, max int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := fmt.Sprintf("rate_limit:%s:%s", c.ClientIP(), c.FullPath())
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		count, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			// Redis 不可用时放行，避免限流组件拖垮业务。
			c.Next()
			return
		}
		if count == 1 {
			_, _ = rdb.Expire(ctx, key, window).Result()
		}
		if max > 0 && count > int64(max) {
			response.Abort(c, http.StatusTooManyRequests, constants.CodeRateLimited, "rate limit exceeded")
			return
		}
		c.Next()
	}
}
