package repository

import (
	"context"
	"database/sql"
	"fmt"

	"campushub/backend/internal/domain"

	"github.com/google/uuid"
)

type IssueRepository interface {
	Create(ctx context.Context, i *domain.Issue) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Issue, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]domain.Issue, error)
	ListByReporter(ctx context.Context, tenantID, userID uuid.UUID) ([]domain.Issue, error)
	UpdateStatus(ctx context.Context, tenantID, id uuid.UUID, fromStatus, toStatus domain.IssueStatus, assignedTo *uuid.UUID) error
}

type issueRepository struct {
	db *sql.DB
}

func NewIssueRepository(db *sql.DB) IssueRepository {
	return &issueRepository{db: db}
}

func (r *issueRepository) Create(ctx context.Context, i *domain.Issue) error {
	query := `
		INSERT INTO issues (id, tenant_id, resource_id, reported_by, category, priority, description, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, now(), now())
		RETURNING created_at, updated_at`
	err := r.db.QueryRowContext(ctx, query, i.ID, i.TenantID, i.ResourceID, i.ReportedBy, i.Category, i.Priority, i.Description, i.Status).
		Scan(&i.CreatedAt, &i.UpdatedAt)
	if err != nil {
		return fmt.Errorf("creating issue: %w", err)
	}
	return nil
}

func (r *issueRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Issue, error) {
	var i domain.Issue
	query := `SELECT id, tenant_id, resource_id, reported_by, assigned_to, category, priority,
	                 description, status, attachment_key, created_at, updated_at
	          FROM issues WHERE tenant_id = $1 AND id = $2`
	err := r.db.QueryRowContext(ctx, query, tenantID, id).Scan(
		&i.ID, &i.TenantID, &i.ResourceID, &i.ReportedBy, &i.AssignedTo, &i.Category, &i.Priority,
		&i.Description, &i.Status, &i.AttachmentKey, &i.CreatedAt, &i.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting issue: %w", err)
	}
	return &i, nil
}

func (r *issueRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]domain.Issue, error) {
	return r.list(ctx, `SELECT id, tenant_id, resource_id, reported_by, assigned_to, category, priority,
	                            description, status, attachment_key, created_at, updated_at
	                     FROM issues WHERE tenant_id = $1 ORDER BY created_at DESC`, tenantID)
}

func (r *issueRepository) ListByReporter(ctx context.Context, tenantID, userID uuid.UUID) ([]domain.Issue, error) {
	return r.list(ctx, `SELECT id, tenant_id, resource_id, reported_by, assigned_to, category, priority,
	                            description, status, attachment_key, created_at, updated_at
	                     FROM issues WHERE tenant_id = $1 AND reported_by = $2 ORDER BY created_at DESC`, tenantID, userID)
}

func (r *issueRepository) list(ctx context.Context, query string, args ...any) ([]domain.Issue, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing issues: %w", err)
	}
	defer rows.Close()

	var issues []domain.Issue
	for rows.Next() {
		var i domain.Issue
		if err := rows.Scan(&i.ID, &i.TenantID, &i.ResourceID, &i.ReportedBy, &i.AssignedTo, &i.Category,
			&i.Priority, &i.Description, &i.Status, &i.AttachmentKey, &i.CreatedAt, &i.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning issue: %w", err)
		}
		issues = append(issues, i)
	}
	return issues, rows.Err()
}

func (r *issueRepository) UpdateStatus(ctx context.Context, tenantID, id uuid.UUID, fromStatus, toStatus domain.IssueStatus, assignedTo *uuid.UUID) error {
	query := `
		UPDATE issues
		SET status = $4, assigned_to = COALESCE($5, assigned_to), updated_at = now()
		WHERE tenant_id = $1 AND id = $2 AND status = $3`
	result, err := r.db.ExecContext(ctx, query, tenantID, id, fromStatus, toStatus, assignedTo)
	if err != nil {
		return fmt.Errorf("updating issue status: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}
