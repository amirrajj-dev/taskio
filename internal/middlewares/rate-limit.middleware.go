// internal/middlewares/rate_limit.go
package middlewares

import (
	"fmt"
	"net/http"
	"time"

	"github.com/amirrajj-dev/taskio/internal/errors"
	"github.com/amirrajj-dev/taskio/internal/infrastructure/redis"
	"github.com/gin-gonic/gin"
)

func RateLimit(limit int, window time.Duration, keyFunc func(c *gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := keyFunc(c)
		ctx := c.Request.Context()

		val, err := redis.RedisClient.Incr(ctx, key).Result()
		if err != nil {
			c.Next()
			return
		}

		if val == 1 {
			redis.RedisClient.Expire(ctx, key, window)
		}

		if val > int64(limit) {
			ttl, _ := redis.RedisClient.TTL(ctx, key).Result()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, errors.NewBasicError(
				fmt.Sprintf("rate limit exceeded. Try again in %s seconds", ttl.Round(time.Second)),
				c.Request.URL.Path,
			))
			return
		}

		c.Next()
	}
}

// 5 request per 24 hours for each user (ip based)
func RegisterLimit() gin.HandlerFunc {
	return RateLimit(5, 24*time.Hour, func(c *gin.Context) string {
		return fmt.Sprintf("register:%s", c.ClientIP())
	})
}

// 10 request per 15 minutes for each user (ip based)
func LoginLimit() gin.HandlerFunc {
	return RateLimit(10, 15*time.Minute, func(c *gin.Context) string {
		return fmt.Sprintf("login:%s", c.ClientIP())
	})
}