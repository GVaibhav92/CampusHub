package repository

import (
	"context"
	"database/sql"
	"fmt"

	"campushub/backend/internal/domain"

	"github.com/google/uuid"
)

type EnrollmentRepository interface {
	GetByStudentAndBooking(ctx context.Context, tenantID, bookingID, studentID uuid.UUID) (*domain.EventEnrollment, error)
	Create(ctx context.Context, tx *sql.Tx, e *domain.EventEnrollment) error
	UpdateStatus(ctx context.Context, tx *sql.Tx, tenantID, bookingID, studentID uuid.UUID, status domain.EnrollmentStatus) error
	GetActiveCount(ctx context.Context, tx *sql.Tx, tenantID, bookingID uuid.UUID) (int, error)
	GetRoster(ctx context.Context, tenantID, bookingID uuid.UUID) ([]domain.RosterStudent, error)
	ListUpcomingEvents(ctx context.Context, tenantID uuid.UUID, studentID *uuid.UUID) ([]domain.EventSession, error)
	ListEnrolledByStudent(ctx context.Context, tenantID, studentID uuid.UUID) ([]domain.EventSession, error)
}

type enrollmentRepository struct {
	db *sql.DB
}

func NewEnrollmentRepository(db *sql.DB) EnrollmentRepository {
	return &enrollmentRepository{db: db}
}

func (r *enrollmentRepository) GetByStudentAndBooking(ctx context.Context, tenantID, bookingID, studentID uuid.UUID) (*domain.EventEnrollment, error) {
	var e domain.EventEnrollment
	query := `SELECT id, tenant_id, booking_id, student_id, status, enrolled_at, updated_at
	          FROM event_enrollments
	          WHERE tenant_id = $1 AND booking_id = $2 AND student_id = $3`
	err := r.db.QueryRowContext(ctx, query, tenantID, bookingID, studentID).Scan(
		&e.ID, &e.TenantID, &e.BookingID, &e.StudentID, &e.Status, &e.EnrolledAt, &e.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting enrollment: %w", err)
	}
	return &e, nil
}

func (r *enrollmentRepository) Create(ctx context.Context, tx *sql.Tx, e *domain.EventEnrollment) error {
	query := `INSERT INTO event_enrollments (id, tenant_id, booking_id, student_id, status, enrolled_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, now(), now())
	          RETURNING enrolled_at, updated_at`
	var err error
	if tx != nil {
		err = tx.QueryRowContext(ctx, query, e.ID, e.TenantID, e.BookingID, e.StudentID, e.Status).Scan(&e.EnrolledAt, &e.UpdatedAt)
	} else {
		err = r.db.QueryRowContext(ctx, query, e.ID, e.TenantID, e.BookingID, e.StudentID, e.Status).Scan(&e.EnrolledAt, &e.UpdatedAt)
	}
	if err != nil {
		return fmt.Errorf("creating enrollment: %w", err)
	}
	return nil
}

func (r *enrollmentRepository) UpdateStatus(ctx context.Context, tx *sql.Tx, tenantID, bookingID, studentID uuid.UUID, status domain.EnrollmentStatus) error {
	query := `UPDATE event_enrollments
	          SET status = $4, updated_at = now()
	          WHERE tenant_id = $1 AND booking_id = $2 AND student_id = $3`
	var result sql.Result
	var err error
	if tx != nil {
		result, err = tx.ExecContext(ctx, query, tenantID, bookingID, studentID, status)
	} else {
		result, err = r.db.ExecContext(ctx, query, tenantID, bookingID, studentID, status)
	}
	if err != nil {
		return fmt.Errorf("updating enrollment status: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *enrollmentRepository) GetActiveCount(ctx context.Context, tx *sql.Tx, tenantID, bookingID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM event_enrollments
	          WHERE tenant_id = $1 AND booking_id = $2 AND status = 'ENROLLED'`
	var count int
	var err error
	if tx != nil {
		err = tx.QueryRowContext(ctx, query, tenantID, bookingID).Scan(&count)
	} else {
		err = r.db.QueryRowContext(ctx, query, tenantID, bookingID).Scan(&count)
	}
	if err != nil {
		return 0, fmt.Errorf("getting active enrollment count: %w", err)
	}
	return count, nil
}

func (r *enrollmentRepository) GetRoster(ctx context.Context, tenantID, bookingID uuid.UUID) ([]domain.RosterStudent, error) {
	query := `SELECT e.student_id, u.email, e.status, e.enrolled_at
	          FROM event_enrollments e
	          JOIN users u ON e.student_id = u.id
	          WHERE e.tenant_id = $1 AND e.booking_id = $2
	          ORDER BY e.enrolled_at ASC`
	rows, err := r.db.QueryContext(ctx, query, tenantID, bookingID)
	if err != nil {
		return nil, fmt.Errorf("querying roster: %w", err)
	}
	defer rows.Close()

	var roster []domain.RosterStudent
	for rows.Next() {
		var s domain.RosterStudent
		if err := rows.Scan(&s.StudentID, &s.Email, &s.Status, &s.EnrolledAt); err != nil {
			return nil, fmt.Errorf("scanning roster student: %w", err)
		}
		roster = append(roster, s)
	}
	return roster, rows.Err()
}

func (r *enrollmentRepository) ListUpcomingEvents(ctx context.Context, tenantID uuid.UUID, studentID *uuid.UUID) ([]domain.EventSession, error) {
	var targetStudent uuid.UUID
	if studentID != nil {
		targetStudent = *studentID
	}

	query := `
		SELECT 
			b.id, b.tenant_id, b.resource_id, r.name, r.location, r.capacity,
			b.user_id, u.email, b.title, b.course_code, b.start_time, b.end_time, b.status,
			COALESCE(e_agg.active_count, 0) AS active_enrolled,
			CASE WHEN user_e.student_id IS NOT NULL THEN true ELSE false END AS is_enrolled
		FROM bookings b
		JOIN resources r ON b.resource_id = r.id
		JOIN users u ON b.user_id = u.id
		LEFT JOIN (
			SELECT booking_id, COUNT(*) AS active_count
			FROM event_enrollments
			WHERE status = 'ENROLLED'
			GROUP BY booking_id
		) e_agg ON e_agg.booking_id = b.id
		LEFT JOIN event_enrollments user_e ON user_e.booking_id = b.id AND user_e.student_id = $2 AND user_e.status = 'ENROLLED'
		WHERE b.tenant_id = $1 AND b.status = 'CONFIRMED' AND b.end_time >= now()
		ORDER BY b.start_time ASC`

	rows, err := r.db.QueryContext(ctx, query, tenantID, targetStudent)
	if err != nil {
		return nil, fmt.Errorf("querying upcoming events: %w", err)
	}
	defer rows.Close()

	var events []domain.EventSession
	for rows.Next() {
		var ev domain.EventSession
		if err := rows.Scan(
			&ev.ID, &ev.TenantID, &ev.ResourceID, &ev.ResourceName, &ev.ResourceLocation, &ev.Capacity,
			&ev.FacultyID, &ev.FacultyEmail, &ev.Title, &ev.CourseCode, &ev.StartTime, &ev.EndTime, &ev.Status,
			&ev.ActiveEnrolled, &ev.IsEnrolled,
		); err != nil {
			return nil, fmt.Errorf("scanning event: %w", err)
		}
		events = append(events, ev)
	}
	return events, rows.Err()
}

func (r *enrollmentRepository) ListEnrolledByStudent(ctx context.Context, tenantID, studentID uuid.UUID) ([]domain.EventSession, error) {
	query := `
		SELECT 
			b.id, b.tenant_id, b.resource_id, r.name, r.location, r.capacity,
			b.user_id, u.email, b.title, b.course_code, b.start_time, b.end_time, b.status,
			COALESCE(e_agg.active_count, 0) AS active_enrolled,
			true AS is_enrolled
		FROM event_enrollments e
		JOIN bookings b ON e.booking_id = b.id
		JOIN resources r ON b.resource_id = r.id
		JOIN users u ON b.user_id = u.id
		LEFT JOIN (
			SELECT booking_id, COUNT(*) AS active_count
			FROM event_enrollments
			WHERE status = 'ENROLLED'
			GROUP BY booking_id
		) e_agg ON e_agg.booking_id = b.id
		WHERE e.tenant_id = $1 AND e.student_id = $2 AND e.status = 'ENROLLED' AND b.status = 'CONFIRMED' AND b.end_time >= now()
		ORDER BY b.start_time ASC`

	rows, err := r.db.QueryContext(ctx, query, tenantID, studentID)
	if err != nil {
		return nil, fmt.Errorf("querying student enrolled events: %w", err)
	}
	defer rows.Close()

	var events []domain.EventSession
	for rows.Next() {
		var ev domain.EventSession
		if err := rows.Scan(
			&ev.ID, &ev.TenantID, &ev.ResourceID, &ev.ResourceName, &ev.ResourceLocation, &ev.Capacity,
			&ev.FacultyID, &ev.FacultyEmail, &ev.Title, &ev.CourseCode, &ev.StartTime, &ev.EndTime, &ev.Status,
			&ev.ActiveEnrolled, &ev.IsEnrolled,
		); err != nil {
			return nil, fmt.Errorf("scanning enrolled event: %w", err)
		}
		events = append(events, ev)
	}
	return events, rows.Err()
}
