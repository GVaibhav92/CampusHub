package tests

import (
	"context"
	"testing"

	"campushub/backend/internal/domain"
	"campushub/backend/internal/repository"

	"github.com/google/uuid"
)

// TestUserListingByRole proves the technician-discovery query returns only
// users of the requested role, scoped to the correct tenant.
func TestUserListingByRole(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	ctx := context.Background()

	tenantRepo := repository.NewTenantRepository(db)
	userRepo := repository.NewUserRepository(db)

	tenant := &domain.Tenant{ID: uuid.New(), Name: "VIT", Slug: "vit-" + uuid.NewString()[:8]}
	must(t, tenantRepo.Create(ctx, tenant))

	student := &domain.User{ID: uuid.New(), TenantID: tenant.ID, CognitoSub: uuid.NewString(), Email: "s@vit.edu", PasswordHash: "x", Role: domain.RoleStudent}
	tech1 := &domain.User{ID: uuid.New(), TenantID: tenant.ID, CognitoSub: uuid.NewString(), Email: "t1@vit.edu", PasswordHash: "x", Role: domain.RoleTechnician}
	tech2 := &domain.User{ID: uuid.New(), TenantID: tenant.ID, CognitoSub: uuid.NewString(), Email: "t2@vit.edu", PasswordHash: "x", Role: domain.RoleTechnician}
	must(t, userRepo.Create(ctx, student))
	must(t, userRepo.Create(ctx, tech1))
	must(t, userRepo.Create(ctx, tech2))

	techRole := domain.RoleTechnician
	technicians, err := userRepo.List(ctx, tenant.ID, &techRole)
	must(t, err)
	if len(technicians) != 2 {
		t.Fatalf("expected 2 technicians, got %d", len(technicians))
	}
	for _, u := range technicians {
		if u.Role != domain.RoleTechnician {
			t.Fatalf("expected only TECHNICIAN role, got %s", u.Role)
		}
	}

	all, err := userRepo.List(ctx, tenant.ID, nil)
	must(t, err)
	if len(all) != 3 {
		t.Fatalf("expected 3 total users, got %d", len(all))
	}
}
