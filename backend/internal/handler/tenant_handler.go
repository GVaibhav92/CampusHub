package handler

import (
	"net/http"
	"os"

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
	// Verify bootstrap secret
	expectedKey := os.Getenv("TENANT_PROVISIONING_KEY")

	if expectedKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "tenant provisioning is not configured",
		})
		return
	}

	providedKey := c.GetHeader("X-Tenant-Provisioning-Key")
	if providedKey != expectedKey {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized provisioning request"})
		return
	}

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

type tenantSummary struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Slug string    `json:"slug"`
}

// List returns a list of tenants for public discovery during user signup.
func (h *TenantHandler) List(c *gin.Context) {
	tenants, err := h.tenants.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list tenants"})
		return
	}

	out := make([]tenantSummary, 0, len(tenants))
	for _, t := range tenants {
		out = append(out, tenantSummary{
			ID:   t.ID,
			Name: t.Name,
			Slug: t.Slug,
		})
	}
	c.JSON(http.StatusOK, out)
}
