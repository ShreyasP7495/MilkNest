package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"github.com/milknest/backend/pkg/logger"
	"github.com/milknest/backend/pkg/utils"
)

// Recover catches panics, logs the stack with the request id, and returns a
// generic 500 envelope.
func Recover() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				logger.L().Error().
					Str("request_id", GetRequestID(c)).
					Interface("panic", r).
					Bytes("stack", debug.Stack()).
					Msg("panic recovered")
				utils.Fail(c, http.StatusInternalServerError, "internal_error", "internal server error")
			}
		}()
		c.Next()
	}
}
