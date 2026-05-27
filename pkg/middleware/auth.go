package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/milknest/backend/pkg/utils"
)

// Claims is the JWT payload shared by access + refresh tokens.
type Claims struct {
	UserID string `json:"uid"`
	Role   string `json:"role"`
	Type   string `json:"typ"` // "access" | "refresh"
	jwt.RegisteredClaims
}

const (
	ctxKeyUserID = "auth_user_id"
	ctxKeyRole   = "auth_role"
)

// JWTAuth validates the Authorization: Bearer <token> header against the
// provided secret. On success it stores the user id + role in the gin context.
func JWTAuth(accessSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := c.GetHeader("Authorization")
		if raw == "" {
			utils.FailErr(c, utils.NewUnauthorized("missing Authorization header"))
			return
		}
		tokenStr := strings.TrimPrefix(raw, "Bearer ")
		if tokenStr == raw { // no "Bearer " prefix
			utils.FailErr(c, utils.NewUnauthorized("malformed Authorization header"))
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(accessSecret), nil
		})
		if err != nil || !token.Valid {
			utils.FailErr(c, utils.NewUnauthorized("invalid or expired token"))
			return
		}
		if claims.Type != "access" {
			utils.FailErr(c, utils.NewUnauthorized("wrong token type"))
			return
		}

		c.Set(ctxKeyUserID, claims.UserID)
		c.Set(ctxKeyRole, claims.Role)
		c.Next()
	}
}

// UserID returns the authenticated user id, panicking if missing - the
// middleware guarantees it.
func UserID(c *gin.Context) string {
	v, _ := c.Get(ctxKeyUserID)
	s, _ := v.(string)
	return s
}

// Role returns the authenticated user role.
func Role(c *gin.Context) string {
	v, _ := c.Get(ctxKeyRole)
	s, _ := v.(string)
	return s
}
