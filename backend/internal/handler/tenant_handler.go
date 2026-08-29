package handler

import (
	"net/http"

	"campushub/backend/internal/domain"
	"campushub/backend/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TenantHandler struct {
	tenants repository.TenantRepository
}

func NewTenantHandler(tenants repository.TenantRepository) *TenantHandler {
	return &TenantHandler{tenants: tenants}
}

type createTenantRequest struct {
	Name string `json:"name" binding:"required"`
	Slug string `json:"slug" binding:"required"`
}

// Create is a local-dev-only bootstrap endpoint for seeding tenants.
func (h *TenantHandler) Create(c *gin.Context) {
	var req createTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	t := &domain.Tenant{ID: uuid.New(), Name: req.Name, Slug: req.Slug}
	if err := h.tenants.Create(c.Request.Context(), t); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create tenant"})
		return
	}
	c.JSON(http.StatusCreated, t)
}
