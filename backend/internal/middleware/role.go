package middleware

import (
	"net/http"

	"campushub/backend/internal/domain"

	"github.com/gin-gonic/gin"
)

// RequireRole is Phase 2's first real authorization check. It must run
// after RequireTenant, since it reads the role that middleware sets — Gin
// route groups enforce that ordering. Kept as a single small function
// rather than a policy engine: with four roles and a handful of write
// endpoints, an allow-list per route is simpler to read and audit than a
// rules table would be.
func RequireRole(roles ...domain.Role) gin.HandlerFunc {
	allowed := make(map[domain.Role]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}
	return func(c *gin.Context) {
		role, ok := RoleFromContext(c)
		if !ok || !allowed[role] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}
		c.Next()
	}
}
