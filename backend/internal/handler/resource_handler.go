package handler

import (
	"database/sql"
	"net/http"

	"campushub/backend/internal/domain"
	"campushub/backend/internal/middleware"
	"campushub/backend/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ResourceHandler struct {
	resources repository.ResourceRepository
}

func NewResourceHandler(resources repository.ResourceRepository) *ResourceHandler {
	return &ResourceHandler{resources: resources}
}

type createResourceRequest struct {
	Name     string `json:"name" binding:"required"`
	Type     string `json:"type" binding:"required"`
	Capacity int    `json:"capacity"`
	Location string `json:"location"`
}

func (h *ResourceHandler) Create(c *gin.Context) {
	tenantID, _ := middleware.TenantIDFromContext(c)

	var req createResourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res := &domain.Resource{
		ID: uuid.New(), TenantID: tenantID, Name: req.Name, Type: req.Type,
		Capacity: req.Capacity, Location: req.Location, Status: domain.ResourceAvailable,
	}
	if err := h.resources.Create(c.Request.Context(), res); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create resource"})
		return
	}
	c.JSON(http.StatusCreated, res)
}

func (h *ResourceHandler) List(c *gin.Context) {
	tenantID, _ := middleware.TenantIDFromContext(c)
	resources, err := h.resources.List(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list resources"})
		return
	}
	c.JSON(http.StatusOK, resources)
}

func (h *ResourceHandler) Get(c *gin.Context) {
	tenantID, _ := middleware.TenantIDFromContext(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid resource id"})
		return
	}
	res, err := h.resources.GetByID(c.Request.Context(), tenantID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get resource"})
		return
	}
	if res == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "resource not found"})
		return
	}
	c.JSON(http.StatusOK, res)
}

type updateResourceRequest struct {
	Name     string                `json:"name" binding:"required"`
	Type     string                `json:"type" binding:"required"`
	Capacity int                   `json:"capacity"`
	Location string                `json:"location"`
	Status   domain.ResourceStatus `json:"status" binding:"required"`
}

func (h *ResourceHandler) Update(c *gin.Context) {
	tenantID, _ := middleware.TenantIDFromContext(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid resource id"})
		return
	}

	var req updateResourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res := &domain.Resource{
		ID: id, TenantID: tenantID, Name: req.Name, Type: req.Type,
		Capacity: req.Capacity, Location: req.Location, Status: req.Status,
	}
	if err := h.resources.Update(c.Request.Context(), res); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "resource not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update resource"})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *ResourceHandler) Delete(c *gin.Context) {
	tenantID, _ := middleware.TenantIDFromContext(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid resource id"})
		return
	}

	active, err := h.resources.HasActiveBookings(c.Request.Context(), tenantID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check bookings"})
		return
	}
	if active {
		c.JSON(http.StatusConflict, gin.H{"error": "resource has active bookings and cannot be deleted"})
		return
	}

	if err := h.resources.Delete(c.Request.Context(), tenantID, id); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "resource not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete resource"})
		return
	}
	c.Status(http.StatusNoContent)
}
