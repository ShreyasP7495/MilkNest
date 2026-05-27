package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/milknest/backend/pkg/utils"
)

// Role constants - keep in sync with the user_role enum in 0001_init.up.sql.
const (
	RoleCustomer = "customer"
	RoleAdmin    = "admin"
	RoleDelivery = "delivery"
)

// Roles enforces that the authenticated user has one of the allowed roles.
// Must run after JWTAuth.
func Roles(allowed ...string) gin.HandlerFunc {
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, r := range allowed {
		allowedSet[r] = struct{}{}
	}
	return func(c *gin.Context) {
		role := Role(c)
		if _, ok := allowedSet[role]; !ok {
			utils.FailErr(c, utils.NewForbidden("role not permitted"))
			return
		}
		c.Next()
	}
}
