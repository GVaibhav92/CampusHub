package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"campushub/backend/internal/domain"
	"campushub/backend/internal/repository"

	"github.com/google/uuid"
)

var (
	ErrResourceNotFound    = errors.New("resource not found")
	ErrResourceUnavailable = errors.New("resource is under maintenance")
	ErrInvalidTimeRange    = errors.New("end time must be after start time")
)

// BookingService is Phase 2's first service — introduced specifically
// because booking creation needs to orchestrate two things atomically
// (resource validity check + conflict-safe insert) inside one transaction.
// Resources and Users have no service because their handlers are pure
// pass-through CRUD with nothing to orchestrate.
type BookingService struct {
	db        *sql.DB
	bookings  repository.BookingRepository
	resources repository.ResourceRepository
}

func NewBookingService(db *sql.DB, bookings repository.BookingRepository, resources repository.ResourceRepository) *BookingService {
	return &BookingService{db: db, bookings: bookings, resources: resources}
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
	// The exclusion constraint is the actual conflict guard; this insert
	// either succeeds or returns ErrBookingConflict — there is no
	// separate "check for conflicts" query to get out of sync with it.
	if err := s.bookings.Create(ctx, tx, booking); err != nil {
		return nil, err // repository.ErrBookingConflict passes through as-is
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("committing booking: %w", err)
	}
	return booking, nil
}

// CancelBooking enforces that only the booking's owner or an ADMIN can
// cancel it. Ownership is checked here in the service, not left to the
// handler, because it's booking-specific business logic, not a generic
// authorization concern the role middleware can express.
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

var ErrForbidden = errors.New("not permitted to cancel this booking")
