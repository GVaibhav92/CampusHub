package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"campushub/backend/internal/middleware"
	"campushub/backend/internal/repository"
	"campushub/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type BookingHandler struct {
	bookingService *service.BookingService
	bookings       repository.BookingRepository
}

func NewBookingHandler(bookingService *service.BookingService, bookings repository.BookingRepository) *BookingHandler {
	return &BookingHandler{bookingService: bookingService, bookings: bookings}
}

type createBookingRequest struct {
	ResourceID string    `json:"resource_id" binding:"required"`
	StartTime  time.Time `json:"start_time" binding:"required"`
	EndTime    time.Time `json:"end_time" binding:"required"`
}

func (h *BookingHandler) Create(c *gin.Context) {
	tenantID, _ := middleware.TenantIDFromContext(c)
	sub, _ := middleware.SubFromContext(c)
	_ = sub

	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}

	var req createBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resourceID, err := uuid.Parse(req.ResourceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid resource_id"})
		return
	}

	booking, err := h.bookingService.CreateBooking(c.Request.Context(), tenantID, userID, resourceID, req.StartTime, req.EndTime)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrBookingConflict):
			c.JSON(http.StatusConflict, gin.H{"error": "resource is already booked for that time"})
		case errors.Is(err, service.ErrResourceNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "resource not found"})
		case errors.Is(err, service.ErrResourceUnavailable):
			c.JSON(http.StatusConflict, gin.H{"error": "resource is unavailable"})
		case errors.Is(err, service.ErrInvalidTimeRange):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create booking"})
		}
		return
	}
	c.JSON(http.StatusCreated, booking)
}

func (h *BookingHandler) ListMine(c *gin.Context) {
	tenantID, _ := middleware.TenantIDFromContext(c)
	userIDVal, _ := c.Get("user_id")
	userID := userIDVal.(uuid.UUID)

	bookings, err := h.bookings.ListByUser(c.Request.Context(), tenantID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list bookings"})
		return
	}
	c.JSON(http.StatusOK, bookings)
}

// returns every booking in the tenant.
func (h *BookingHandler) ListAll(c *gin.Context) {
	tenantID, _ := middleware.TenantIDFromContext(c)
	bookings, err := h.bookings.ListByTenant(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list bookings"})
		return
	}
	c.JSON(http.StatusOK, bookings)
}

func (h *BookingHandler) Cancel(c *gin.Context) {
	tenantID, _ := middleware.TenantIDFromContext(c)
	role, _ := middleware.RoleFromContext(c)
	userIDVal, _ := c.Get("user_id")
	userID := userIDVal.(uuid.UUID)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid booking id"})
		return
	}

	err = h.bookingService.CancelBooking(c.Request.Context(), tenantID, userID, role, id)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
		case errors.Is(err, service.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "not permitted to cancel this booking"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to cancel booking"})
		}
		return
	}
	c.Status(http.StatusNoContent)
}
