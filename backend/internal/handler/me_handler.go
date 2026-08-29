package handler

import (
	"net/http"

	"campushub/backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func Me(c *gin.Context) {
	tenantID, _ := middleware.TenantIDFromContext(c)
	role, _ := middleware.RoleFromContext(c)
	c.JSON(http.StatusOK, gin.H{"tenant_id": tenantID, "role": role})
}
