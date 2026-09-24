package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/renovation/renovation-budget-api/internal/model"
)

const actorContextKey = "actor"

// SetActor 将操作者写入请求上下文。
func SetActor(c *gin.Context, actor model.Actor) {
	c.Set(actorContextKey, actor)
}

// CurrentActor 从请求上下文读取操作者。
func CurrentActor(c *gin.Context) (model.Actor, bool) {
	v, ok := c.Get(actorContextKey)
	if !ok {
		return model.Actor{}, false
	}
	actor, ok := v.(model.Actor)
	return actor, ok
}
