package repository

import (
	"context"
	"database/sql"
	"fmt"

	"campushub/backend/internal/domain"

	"github.com/google/uuid"
)

type ResourceRepository interface {
	Create(ctx context.Context, r *domain.Resource) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Resource, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.Resource, error)
	Update(ctx context.Context, r *domain.Resource) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	HasActiveBookings(ctx context.Context, tenantID, id uuid.UUID) (bool, error)
}

type resourceRepository struct {
	db *sql.DB
}

func NewResourceRepository(db *sql.DB) ResourceRepository {
	return &resourceRepository{db: db}
}

func (r *resourceRepository) Create(ctx context.Context, res *domain.Resource) error {
	query := `
		INSERT INTO resources (id, tenant_id, name, type, capacity, location, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, now())
		RETURNING created_at`
	err := r.db.QueryRowContext(ctx, query, res.ID, res.TenantID, res.Name, res.Type, res.Capacity, res.Location, res.Status).Scan(&res.CreatedAt)
	if err != nil {
		return fmt.Errorf("creating resource: %w", err)
	}
	return nil
}

// GetByID always filters by tenant_id in the WHERE clause, never as a
// post-fetch check. This is the pattern every tenant-scoped query in the
// project follows — a row belonging to another tenant simply doesn't
// exist as far as this query is concerned, so a wrong tenant_id returns
// "not found," not "found but rejected."
func (r *resourceRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Resource, error) {
	var res domain.Resource
	query := `SELECT id, tenant_id, name, type, capacity, location, status, created_at
	          FROM resources WHERE tenant_id = $1 AND id = $2`
	err := r.db.QueryRowContext(ctx, query, tenantID, id).Scan(
		&res.ID, &res.TenantID, &res.Name, &res.Type, &res.Capacity, &res.Location, &res.Status, &res.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting resource: %w", err)
	}
	return &res, nil
}

func (r *resourceRepository) List(ctx context.Context, tenantID uuid.UUID) ([]domain.Resource, error) {
	query := `SELECT id, tenant_id, name, type, capacity, location, status, created_at
	          FROM resources WHERE tenant_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("listing resources: %w", err)
	}
	defer rows.Close()

	var resources []domain.Resource
	for rows.Next() {
		var res domain.Resource
		if err := rows.Scan(&res.ID, &res.TenantID, &res.Name, &res.Type, &res.Capacity, &res.Location, &res.Status, &res.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning resource: %w", err)
		}
		resources = append(resources, res)
	}
	return resources, rows.Err()
}

func (r *resourceRepository) Update(ctx context.Context, res *domain.Resource) error {
	query := `
		UPDATE resources SET name = $3, type = $4, capacity = $5, location = $6, status = $7
		WHERE tenant_id = $1 AND id = $2`
	result, err := r.db.ExecContext(ctx, query, res.TenantID, res.ID, res.Name, res.Type, res.Capacity, res.Location, res.Status)
	if err != nil {
		return fmt.Errorf("updating resource: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *resourceRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM resources WHERE tenant_id = $1 AND id = $2`, tenantID, id)
	if err != nil {
		return fmt.Errorf("deleting resource: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// HasActiveBookings guards deletion. Without this, deleting a resource
// would cascade-delete its bookings (ON DELETE CASCADE on bookings.resource_id),
// silently destroying booking history. Checked explicitly instead so the
// handler can return a clear 409 rather than a surprising cascade.
func (r *resourceRepository) HasActiveBookings(ctx context.Context, tenantID, id uuid.UUID) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(
		SELECT 1 FROM bookings
		WHERE tenant_id = $1 AND resource_id = $2 AND status = 'CONFIRMED' AND end_time > now()
	)`
	if err := r.db.QueryRowContext(ctx, query, tenantID, id).Scan(&exists); err != nil {
		return false, fmt.Errorf("checking active bookings: %w", err)
	}
	return exists, nil
}
