package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const (
	_rateLimitWindow  = time.Minute
	_rateLimitMaxReqs = 10
)

func RateLimiter(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		key := fmt.Sprintf("rate_limit:%s", c.ClientIP())

		pipe := rdb.Pipeline()
		incr := pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, _rateLimitWindow)

		if _, err := pipe.Exec(ctx); err != nil {
			c.Next()
			return
		}

		if incr.Val() > _rateLimitMaxReqs {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			return
		}

		c.Next()
	}
}
