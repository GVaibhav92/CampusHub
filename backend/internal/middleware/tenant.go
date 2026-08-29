package middleware

import (
	"net/http"

	"campushub/backend/internal/domain"
	"campushub/backend/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	ctxKeyTenantID = "tenant_id"
	ctxKeyRole     = "role"
	ctxKeyUserID   = "user_id"
)

func RequireTenant(users repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		sub, ok := SubFromContext(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
			return
		}

		user, err := users.GetByCognitoSub(c.Request.Context(), sub)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve tenant"})
			return
		}
		if user == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "no account found for this identity"})
			return
		}

		c.Set(ctxKeyTenantID, user.TenantID)
		c.Set(ctxKeyRole, user.Role)
		c.Set(ctxKeyUserID, user.ID)
		c.Next()
	}
}

func TenantIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	v, ok := c.Get(ctxKeyTenantID)
	if !ok {
		return uuid.Nil, false
	}
	id, ok := v.(uuid.UUID)
	return id, ok
}

func RoleFromContext(c *gin.Context) (domain.Role, bool) {
	v, ok := c.Get(ctxKeyRole)
	if !ok {
		return "", false
	}
	role, ok := v.(domain.Role)
	return role, ok
}

func UserIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	v, ok := c.Get(ctxKeyUserID)
	if !ok {
		return uuid.Nil, false
	}
	id, ok := v.(uuid.UUID)
	return id, ok
}
