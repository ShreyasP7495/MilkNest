package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	headerRequestID = "X-Request-ID"
	ctxKeyRequestID = "request_id"
)

// RequestID attaches an X-Request-ID either from the incoming header or a
// freshly generated UUID. Downstream logging uses GetRequestID() to retrieve
// it.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(headerRequestID)
		if id == "" {
			id = uuid.NewString()
		}
		c.Set(ctxKeyRequestID, id)
		c.Writer.Header().Set(headerRequestID, id)
		c.Next()
	}
}

// GetRequestID extracts the request id, returning "" if absent.
func GetRequestID(c *gin.Context) string {
	if v, ok := c.Get(ctxKeyRequestID); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
