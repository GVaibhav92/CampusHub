package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"campushub/backend/internal/cloud"
	"campushub/backend/internal/domain"
	"campushub/backend/internal/repository"

	"github.com/google/uuid"
)

var (
	ErrResourceNotFound    = errors.New("resource not found")
	ErrResourceUnavailable = errors.New("resource is under maintenance")
	ErrInvalidTimeRange    = errors.New("end time must be after start time")
	ErrForbidden           = errors.New("not permitted to cancel this booking")
)

type BookingService struct {
	db        *sql.DB
	bookings  repository.BookingRepository
	resources repository.ResourceRepository
	publisher *cloud.EventPublisher
	logger    *slog.Logger
}

func NewBookingService(db *sql.DB, bookings repository.BookingRepository, resources repository.ResourceRepository, publisher *cloud.EventPublisher,
	logger *slog.Logger) *BookingService {
	return &BookingService{db: db, bookings: bookings, resources: resources, publisher: publisher, logger: logger}
}

func (s *BookingService) CreateBooking(ctx context.Context, tenantID, userID, resourceID uuid.UUID, start, end time.Time) (*domain.Booking, error) {
	if !end.After(start) {
		return nil, ErrInvalidTimeRange
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback() // no-op if committed

	resource, err := s.resources.GetByID(ctx, tenantID, resourceID)
	if err != nil {
		return nil, fmt.Errorf("checking resource: %w", err)
	}
	if resource == nil {
		return nil, ErrResourceNotFound
	}
	if resource.Status != domain.ResourceAvailable {
		return nil, ErrResourceUnavailable
	}

	booking := &domain.Booking{
		ID: uuid.New(), TenantID: tenantID, ResourceID: resourceID, UserID: userID,
		StartTime: start, EndTime: end, Status: domain.BookingConfirmed,
	}

	if err := s.bookings.Create(ctx, tx, booking); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("committing booking: %w", err)
	}

	if s.publisher != nil {
		event := cloud.NotificationEvent{
			Subject: "New Booking Created",
			Message: fmt.Sprintf("Resource %q booked from %s to %s.", resource.Name, start.Format(time.RFC1123), end.Format(time.RFC1123)),
		}
		if err := s.publisher.Publish(ctx, event); err != nil {
			s.logger.Error("failed to publish booking event", "error", err, "booking_id", booking.ID)
		}
	}
	return booking, nil
}

// CancelBooking enforces that only the booking's owner or an ADMIN can
// cancel it. Ownership is checked here in the service
func (s *BookingService) CancelBooking(ctx context.Context, tenantID, userID uuid.UUID, role domain.Role, bookingID uuid.UUID) error {
	booking, err := s.bookings.GetByID(ctx, tenantID, bookingID)
	if err != nil {
		return fmt.Errorf("looking up booking: %w", err)
	}
	if booking == nil {
		return sql.ErrNoRows
	}
	if booking.UserID != userID && role != domain.RoleAdmin {
		return ErrForbidden
	}
	return s.bookings.Cancel(ctx, tenantID, bookingID)
}
