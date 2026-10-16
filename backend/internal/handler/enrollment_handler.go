package handler

import (
	"errors"
	"net/http"

	"campushub/backend/internal/domain"
	"campushub/backend/internal/middleware"
	"campushub/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type EnrollmentHandler struct {
	enrollmentService *service.EnrollmentService
}

func NewEnrollmentHandler(enrollmentService *service.EnrollmentService) *EnrollmentHandler {
	return &EnrollmentHandler{enrollmentService: enrollmentService}
}

func (h *EnrollmentHandler) ListEvents(c *gin.Context) {
	tenantID, _ := middleware.TenantIDFromContext(c)
	var studentID *uuid.UUID
	if userID, ok := middleware.UserIDFromContext(c); ok {
		studentID = &userID
	}

	events, err := h.enrollmentService.ListUpcomingEvents(c.Request.Context(), tenantID, studentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list events"})
		return
	}
	if events == nil {
		events = []domain.EventSession{}
	}
	c.JSON(http.StatusOK, events)
}

func (h *EnrollmentHandler) ListMyEnrolled(c *gin.Context) {
	tenantID, _ := middleware.TenantIDFromContext(c)
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}

	events, err := h.enrollmentService.ListMyEnrolledEvents(c.Request.Context(), tenantID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list enrolled events"})
		return
	}
	if events == nil {
		events = []domain.EventSession{}
	}
	c.JSON(http.StatusOK, events)
}

func (h *EnrollmentHandler) Enroll(c *gin.Context) {
	tenantID, _ := middleware.TenantIDFromContext(c)
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}

	bookingID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}

	enrollment, err := h.enrollmentService.Enroll(c.Request.Context(), tenantID, bookingID, userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrBookingNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		case errors.Is(err, service.ErrBookingNotConfirmed):
			c.JSON(http.StatusBadRequest, gin.H{"error": "event is no longer confirmed"})
		case errors.Is(err, service.ErrEventExpired):
			c.JSON(http.StatusBadRequest, gin.H{"error": "event has already ended"})
		case errors.Is(err, service.ErrCapacityExceeded):
			c.JSON(http.StatusConflict, gin.H{"error": "event has reached maximum capacity"})
		case errors.Is(err, service.ErrAlreadyEnrolled):
			c.JSON(http.StatusConflict, gin.H{"error": "already enrolled in this event"})
		case errors.Is(err, service.ErrEnrollmentRevokedByFaculty):
			c.JSON(http.StatusForbidden, gin.H{"error": "enrollment was revoked by the instructor and cannot be reinstated"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to enroll in event"})
		}
		return
	}

	c.JSON(http.StatusCreated, enrollment)
}

func (h *EnrollmentHandler) Unenroll(c *gin.Context) {
	tenantID, _ := middleware.TenantIDFromContext(c)
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}

	bookingID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}

	err = h.enrollmentService.CancelStudentEnrollment(c.Request.Context(), tenantID, bookingID, userID)
	if err != nil {
		if errors.Is(err, service.ErrNotEnrolled) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not enrolled in this event"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to cancel enrollment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "enrollment cancelled successfully"})
}

func (h *EnrollmentHandler) GetRoster(c *gin.Context) {
	tenantID, _ := middleware.TenantIDFromContext(c)
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}
	role, _ := middleware.RoleFromContext(c)

	bookingID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}

	roster, err := h.enrollmentService.GetRoster(c.Request.Context(), tenantID, bookingID, userID, role)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrBookingNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		case errors.Is(err, service.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "you do not have permission to view this roster"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get roster"})
		}
		return
	}

	if roster == nil {
		roster = []domain.RosterStudent{}
	}
	c.JSON(http.StatusOK, roster)
}

func (h *EnrollmentHandler) RemoveStudent(c *gin.Context) {
	tenantID, _ := middleware.TenantIDFromContext(c)
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}
	role, _ := middleware.RoleFromContext(c)

	bookingID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}
	studentID, err := uuid.Parse(c.Param("student_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid student id"})
		return
	}

	err = h.enrollmentService.RemoveStudentFromRoster(c.Request.Context(), tenantID, bookingID, userID, role, studentID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrBookingNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		case errors.Is(err, service.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "you do not have permission to manage this roster"})
		case errors.Is(err, service.ErrNotEnrolled):
			c.JSON(http.StatusNotFound, gin.H{"error": "student is not actively enrolled in this event"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove student from roster"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "student enrollment cancelled by faculty"})
}
