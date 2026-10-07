package repository

import (
	"context"
	"database/sql"
	"fmt"

	"campushub/backend/internal/domain"

	"github.com/google/uuid"
)

type TenantRepository interface {
	Create(ctx context.Context, t *domain.Tenant) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Tenant, error)
	List(ctx context.Context) ([]domain.Tenant, error)
}

type tenantRepository struct {
	db *sql.DB
}

func NewTenantRepository(db *sql.DB) TenantRepository {
	return &tenantRepository{db: db}
}

func (r *tenantRepository) Create(ctx context.Context, t *domain.Tenant) error {
	query := `INSERT INTO tenants (id, name, slug, created_at) VALUES ($1, $2, $3, now()) RETURNING created_at`
	if err := r.db.QueryRowContext(ctx, query, t.ID, t.Name, t.Slug).Scan(&t.CreatedAt); err != nil {
		return fmt.Errorf("creating tenant: %w", err)
	}
	return nil
}

func (r *tenantRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	var t domain.Tenant
	query := `SELECT id, name, slug, created_at FROM tenants WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(&t.ID, &t.Name, &t.Slug, &t.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting tenant by id: %w", err)
	}
	return &t, nil
}

func (r *tenantRepository) GetBySlug(ctx context.Context, slug string) (*domain.Tenant, error) {
	var t domain.Tenant
	query := `SELECT id, name, slug, created_at FROM tenants WHERE slug = $1`
	err := r.db.QueryRowContext(ctx, query, slug).Scan(&t.ID, &t.Name, &t.Slug, &t.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting tenant by slug: %w", err)
	}
	return &t, nil
}

func (r *tenantRepository) List(ctx context.Context) ([]domain.Tenant, error) {
	query := `SELECT id, name, slug, created_at FROM tenants ORDER BY name ASC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("listing tenants: %w", err)
	}
	defer rows.Close()

	var tenants []domain.Tenant
	for rows.Next() {
		var t domain.Tenant
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning tenant: %w", err)
		}
		tenants = append(tenants, t)
	}
	return tenants, rows.Err()
}
