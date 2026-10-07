package handler

import (
	"net/http"

	"campushub/backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func Me(c *gin.Context) {
	userID, _ := middleware.UserIDFromContext(c)
	email, _ := middleware.EmailFromContext(c)
	tenantID, _ := middleware.TenantIDFromContext(c)
	role, _ := middleware.RoleFromContext(c)

	c.JSON(http.StatusOK, gin.H{
		"id":        userID,
		"email":     email,
		"tenant_id": tenantID,
		"role":      role,
	})
}
