package repository

import (
	"context"
	"database/sql"
	"fmt"

	"campushub/backend/internal/domain"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, u *domain.User) error
	GetByCognitoSub(ctx context.Context, sub string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	List(ctx context.Context, tenantID uuid.UUID, role *domain.Role) ([]domain.User, error)
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, u *domain.User) error {
	query := `
		INSERT INTO users (id, tenant_id, cognito_sub, email, password_hash, role, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, now())
		RETURNING created_at`
	err := r.db.QueryRowContext(ctx, query, u.ID, u.TenantID, u.CognitoSub, u.Email, u.PasswordHash, u.Role).Scan(&u.CreatedAt)
	if err != nil {
		return fmt.Errorf("creating user: %w", err)
	}
	return nil
}

// GetByCognitoSub is THE query the tenant middleware runs on every
// authenticated request. It is the entire implementation of the
// "server-side cognito_sub -> tenant_id" decision — nothing more is needed.
func (r *userRepository) GetByCognitoSub(ctx context.Context, sub string) (*domain.User, error) {
	var u domain.User
	query := `SELECT id, tenant_id, cognito_sub, email, password_hash, role, created_at
	          FROM users WHERE cognito_sub = $1`
	err := r.db.QueryRowContext(ctx, query, sub).Scan(&u.ID, &u.TenantID, &u.CognitoSub, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting user by cognito sub: %w", err)
	}
	return &u, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User
	query := `SELECT id, tenant_id, cognito_sub, email, password_hash, role, created_at
	          FROM users WHERE email = $1`
	err := r.db.QueryRowContext(ctx, query, email).Scan(&u.ID, &u.TenantID, &u.CognitoSub, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting user by email: %w", err)
	}
	return &u, nil
}

// List returns tenant-scoped users, optionally filtered by role.
func (r *userRepository) List(ctx context.Context, tenantID uuid.UUID, role *domain.Role) ([]domain.User, error) {
	query := `SELECT id, tenant_id, cognito_sub, email, password_hash, role, created_at
	          FROM users WHERE tenant_id = $1`
	args := []any{tenantID}
	if role != nil {
		query += ` AND role = $2`
		args = append(args, *role)
	}
	query += ` ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing users: %w", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.TenantID, &u.CognitoSub, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning user: %w", err)
		}
		users = append(users, u)
	}
	return users, rows.Err()
}
