package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/milknest/backend/pkg/utils"
)

// RedisRateLimit is a simple fixed-window limiter keyed by IP (or the
// keyFn override) using a Redis INCR + EXPIRE. Suitable for OTP endpoints
// where we want to throttle aggressive callers per phone.
func RedisRateLimit(rdb *redis.Client, key string, max int, window time.Duration, keyFn ...func(c *gin.Context) string) gin.HandlerFunc {
	getKey := func(c *gin.Context) string { return "rl:" + key + ":" + c.ClientIP() }
	if len(keyFn) > 0 {
		fn := keyFn[0]
		getKey = func(c *gin.Context) string { return "rl:" + key + ":" + fn(c) }
	}
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 200*time.Millisecond)
		defer cancel()

		k := getKey(c)
		count, err := rdb.Incr(ctx, k).Result()
		if err != nil {
			c.Next() // best-effort: do not fail open requests on Redis blip
			return
		}
		if count == 1 {
			_ = rdb.Expire(ctx, k, window).Err()
		}
		if int(count) > max {
			utils.Fail(c, http.StatusTooManyRequests, "rate_limited", "too many requests, please retry later")
			return
		}
		c.Next()
	}
}
