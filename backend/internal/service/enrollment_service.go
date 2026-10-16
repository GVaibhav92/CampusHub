package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"campushub/backend/internal/domain"
	"campushub/backend/internal/repository"

	"github.com/google/uuid"
)

var (
	ErrBookingNotFound           = errors.New("booking not found")
	ErrBookingNotConfirmed       = errors.New("booking is not active or has been cancelled")
	ErrEventExpired              = errors.New("event has already ended")
	ErrCapacityExceeded          = errors.New("event has reached maximum capacity")
	ErrAlreadyEnrolled           = errors.New("already enrolled in this event")
	ErrNotEnrolled               = errors.New("not enrolled in this event")
	ErrEnrollmentRevokedByFaculty = errors.New("enrollment was cancelled by faculty and cannot be re-activated")
)

type EnrollmentService struct {
	db          *sql.DB
	enrollments repository.EnrollmentRepository
	bookings    repository.BookingRepository
	resources   repository.ResourceRepository
	logger      *slog.Logger
}

func NewEnrollmentService(db *sql.DB, enrollments repository.EnrollmentRepository, bookings repository.BookingRepository, resources repository.ResourceRepository, logger *slog.Logger) *EnrollmentService {
	return &EnrollmentService{
		db:          db,
		enrollments: enrollments,
		bookings:    bookings,
		resources:   resources,
		logger:      logger,
	}
}

func (s *EnrollmentService) Enroll(ctx context.Context, tenantID, bookingID, studentID uuid.UUID) (*domain.EventEnrollment, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Lock booking row and retrieve capacity in one atomic query
	var (
		status   domain.BookingStatus
		endTime  time.Time
		capacity int
	)
	query := `
		SELECT b.status, b.end_time, r.capacity
		FROM bookings b
		JOIN resources r ON b.resource_id = r.id
		WHERE b.tenant_id = $1 AND b.id = $2
		FOR UPDATE`
	err = tx.QueryRowContext(ctx, query, tenantID, bookingID).Scan(&status, &endTime, &capacity)
	if err == sql.ErrNoRows {
		return nil, ErrBookingNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("locking booking for enrollment: %w", err)
	}

	if status != domain.BookingConfirmed {
		return nil, ErrBookingNotConfirmed
	}
	if endTime.Before(time.Now()) {
		return nil, ErrEventExpired
	}

	// 2. Check for existing enrollment record
	existing, err := s.enrollments.GetByStudentAndBooking(ctx, tenantID, bookingID, studentID)
	if err != nil {
		return nil, fmt.Errorf("checking student enrollment: %w", err)
	}

	if existing != nil {
		if existing.Status == domain.EnrollmentCancelledByFaculty {
			return nil, ErrEnrollmentRevokedByFaculty
		}
		if existing.Status == domain.EnrollmentEnrolled {
			return nil, ErrAlreadyEnrolled
		}
	}

	// 3. Check capacity inside transaction
	activeCount, err := s.enrollments.GetActiveCount(ctx, tx, tenantID, bookingID)
	if err != nil {
		return nil, fmt.Errorf("checking active enrollment count: %w", err)
	}
	if activeCount >= capacity {
		return nil, ErrCapacityExceeded
	}

	// 4. Reactivate or create record
	if existing != nil && existing.Status == domain.EnrollmentCancelledByStudent {
		if err := s.enrollments.UpdateStatus(ctx, tx, tenantID, bookingID, studentID, domain.EnrollmentEnrolled); err != nil {
			return nil, fmt.Errorf("reactivating enrollment: %w", err)
		}
		existing.Status = domain.EnrollmentEnrolled
		existing.UpdatedAt = time.Now()
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("committing enrollment transaction: %w", err)
		}
		return existing, nil
	}

	enrollment := &domain.EventEnrollment{
		ID:        uuid.New(),
		TenantID:  tenantID,
		BookingID: bookingID,
		StudentID: studentID,
		Status:    domain.EnrollmentEnrolled,
	}
	if err := s.enrollments.Create(ctx, tx, enrollment); err != nil {
		return nil, fmt.Errorf("creating enrollment: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("committing enrollment transaction: %w", err)
	}
	return enrollment, nil
}

func (s *EnrollmentService) CancelStudentEnrollment(ctx context.Context, tenantID, bookingID, studentID uuid.UUID) error {
	existing, err := s.enrollments.GetByStudentAndBooking(ctx, tenantID, bookingID, studentID)
	if err != nil {
		return fmt.Errorf("checking enrollment: %w", err)
	}
	if existing == nil || existing.Status != domain.EnrollmentEnrolled {
		return ErrNotEnrolled
	}
	return s.enrollments.UpdateStatus(ctx, nil, tenantID, bookingID, studentID, domain.EnrollmentCancelledByStudent)
}

func (s *EnrollmentService) RemoveStudentFromRoster(ctx context.Context, tenantID, bookingID, callerID uuid.UUID, callerRole domain.Role, studentID uuid.UUID) error {
	booking, err := s.bookings.GetByID(ctx, tenantID, bookingID)
	if err != nil {
		return fmt.Errorf("retrieving booking: %w", err)
	}
	if booking == nil {
		return ErrBookingNotFound
	}
	if booking.UserID != callerID && callerRole != domain.RoleAdmin {
		return ErrForbidden
	}

	existing, err := s.enrollments.GetByStudentAndBooking(ctx, tenantID, bookingID, studentID)
	if err != nil {
		return fmt.Errorf("checking student enrollment: %w", err)
	}
	if existing == nil || existing.Status != domain.EnrollmentEnrolled {
		return ErrNotEnrolled
	}

	return s.enrollments.UpdateStatus(ctx, nil, tenantID, bookingID, studentID, domain.EnrollmentCancelledByFaculty)
}

func (s *EnrollmentService) GetRoster(ctx context.Context, tenantID, bookingID, callerID uuid.UUID, callerRole domain.Role) ([]domain.RosterStudent, error) {
	booking, err := s.bookings.GetByID(ctx, tenantID, bookingID)
	if err != nil {
		return nil, fmt.Errorf("retrieving booking: %w", err)
	}
	if booking == nil {
		return nil, ErrBookingNotFound
	}
	if booking.UserID != callerID && callerRole != domain.RoleAdmin {
		return nil, ErrForbidden
	}

	return s.enrollments.GetRoster(ctx, tenantID, bookingID)
}

func (s *EnrollmentService) ListUpcomingEvents(ctx context.Context, tenantID uuid.UUID, studentID *uuid.UUID) ([]domain.EventSession, error) {
	return s.enrollments.ListUpcomingEvents(ctx, tenantID, studentID)
}

func (s *EnrollmentService) ListMyEnrolledEvents(ctx context.Context, tenantID, studentID uuid.UUID) ([]domain.EventSession, error) {
	return s.enrollments.ListEnrolledByStudent(ctx, tenantID, studentID)
}
