package tests

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"campushub/backend/internal/database"
	"campushub/backend/internal/domain"
	"campushub/backend/internal/repository"

	"github.com/google/uuid"
)

// TestTenantIsolation proves a user created under one tenant can never be
// resolved into another tenant's scope. Requires local Postgres running
// (docker compose up -d) with migrations applied.
func TestTenantIsolation(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	tenantRepo := repository.NewTenantRepository(db)
	userRepo := repository.NewUserRepository(db)
	ctx := context.Background()

	tenantA := &domain.Tenant{ID: uuid.New(), Name: "VIT", Slug: "vit-" + uuid.NewString()[:8]}
	tenantB := &domain.Tenant{ID: uuid.New(), Name: "SRM", Slug: "srm-" + uuid.NewString()[:8]}
	if err := tenantRepo.Create(ctx, tenantA); err != nil {
		t.Fatalf("creating tenant A: %v", err)
	}
	if err := tenantRepo.Create(ctx, tenantB); err != nil {
		t.Fatalf("creating tenant B: %v", err)
	}

	userA := &domain.User{
		ID: uuid.New(), TenantID: tenantA.ID, CognitoSub: uuid.NewString(),
		Email: "student@vit.edu", PasswordHash: "hash", Role: domain.RoleStudent,
	}
	if err := userRepo.Create(ctx, userA); err != nil {
		t.Fatalf("creating user A: %v", err)
	}

	resolved, err := userRepo.GetByCognitoSub(ctx, userA.CognitoSub)
	if err != nil {
		t.Fatalf("resolving user A: %v", err)
	}
	if resolved == nil {
		t.Fatal("expected to find user A, got nil")
	}
	if resolved.TenantID != tenantA.ID {
		t.Fatalf("tenant isolation violated: got tenant %s, want %s", resolved.TenantID, tenantA.ID)
	}
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://campushub:campushub@localhost:5432/campushub?sslmode=disable"
	}
	db, err := database.NewPool(database.Config{DSN: dsn, MaxOpenConns: 5, MaxIdleConns: 2})
	if err != nil {
		t.Fatalf("connecting to test database: %v", err)
	}
	return db
}
