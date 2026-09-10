package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"campushub/backend/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

var ErrBookingConflict = errors.New("booking conflicts with an existing reservation")

const exclusionViolationCode = "23P01"

type BookingRepository interface {
	// Create runs inside the *sql.Tx the service layer manages, so resource
	// validation and the insert share one transaction.
	Create(ctx context.Context, tx *sql.Tx, b *domain.Booking) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Booking, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]domain.Booking, error)
	ListByUser(ctx context.Context, tenantID, userID uuid.UUID) ([]domain.Booking, error)
	Cancel(ctx context.Context, tenantID, id uuid.UUID) error
}

type bookingRepository struct {
	db *sql.DB
}

func NewBookingRepository(db *sql.DB) BookingRepository {
	return &bookingRepository{db: db}
}

func (r *bookingRepository) Create(ctx context.Context, tx *sql.Tx, b *domain.Booking) error {
	query := `
		INSERT INTO bookings (id, tenant_id, resource_id, user_id, start_time, end_time, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, now())
		RETURNING created_at`
	err := tx.QueryRowContext(ctx, query, b.ID, b.TenantID, b.ResourceID, b.UserID, b.StartTime, b.EndTime, b.Status).Scan(&b.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == exclusionViolationCode {
			return ErrBookingConflict
		}
		return fmt.Errorf("creating booking: %w", err)
	}
	return nil
}

func (r *bookingRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Booking, error) {
	var b domain.Booking
	query := `SELECT id, tenant_id, resource_id, user_id, start_time, end_time, status, created_at
	          FROM bookings WHERE tenant_id = $1 AND id = $2`
	err := r.db.QueryRowContext(ctx, query, tenantID, id).Scan(
		&b.ID, &b.TenantID, &b.ResourceID, &b.UserID, &b.StartTime, &b.EndTime, &b.Status, &b.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting booking: %w", err)
	}
	return &b, nil
}

func (r *bookingRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]domain.Booking, error) {
	return r.list(ctx, `SELECT id, tenant_id, resource_id, user_id, start_time, end_time, status, created_at
	                     FROM bookings WHERE tenant_id = $1 ORDER BY start_time DESC`, tenantID)
}

func (r *bookingRepository) ListByUser(ctx context.Context, tenantID, userID uuid.UUID) ([]domain.Booking, error) {
	return r.list(ctx, `SELECT id, tenant_id, resource_id, user_id, start_time, end_time, status, created_at
	                     FROM bookings WHERE tenant_id = $1 AND user_id = $2 ORDER BY start_time DESC`, tenantID, userID)
}

func (r *bookingRepository) list(ctx context.Context, query string, args ...any) ([]domain.Booking, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing bookings: %w", err)
	}
	defer rows.Close()

	var bookings []domain.Booking
	for rows.Next() {
		var b domain.Booking
		if err := rows.Scan(&b.ID, &b.TenantID, &b.ResourceID, &b.UserID, &b.StartTime, &b.EndTime, &b.Status, &b.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning booking: %w", err)
		}
		bookings = append(bookings, b)
	}
	return bookings, rows.Err()
}

// Cancel sets status to CANCELLED rather than deleting the row — booking
// history (who booked what, and when it was cancelled) is worth keeping
// for the analytics module in Week 4. Setting status also clears the row
// from the exclusion constraint's WHERE (status = 'CONFIRMED') scope,
// freeing the slot for a new booking.
func (r *bookingRepository) Cancel(ctx context.Context, tenantID, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE bookings SET status = 'CANCELLED' WHERE tenant_id = $1 AND id = $2 AND status = 'CONFIRMED'`,
		tenantID, id)
	if err != nil {
		return fmt.Errorf("cancelling booking: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}
