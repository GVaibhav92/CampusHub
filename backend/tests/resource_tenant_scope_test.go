package tests

import (
	"context"
	"testing"

	"campushub/backend/internal/domain"
	"campushub/backend/internal/repository"

	"github.com/google/uuid"
)

// TestResourceTenantScope proves a resource created under tenant A is
// invisible to a GetByID call scoped to tenant B — same pattern as Phase
// 1's tenant isolation test, applied to the resources table.
func TestResourceTenantScope(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	tenantRepo := repository.NewTenantRepository(db)
	resourceRepo := repository.NewResourceRepository(db)
	ctx := context.Background()

	tenantA := &domain.Tenant{ID: uuid.New(), Name: "VIT", Slug: "vit-" + uuid.NewString()[:8]}
	tenantB := &domain.Tenant{ID: uuid.New(), Name: "SRM", Slug: "srm-" + uuid.NewString()[:8]}
	must(t, tenantRepo.Create(ctx, tenantA))
	must(t, tenantRepo.Create(ctx, tenantB))

	res := &domain.Resource{
		ID: uuid.New(), TenantID: tenantA.ID, Name: "Lab 204", Type: "LAB",
		Capacity: 60, Location: "AB-2", Status: domain.ResourceAvailable,
	}
	must(t, resourceRepo.Create(ctx, res))

	// Correct tenant: found
	found, err := resourceRepo.GetByID(ctx, tenantA.ID, res.ID)
	must(t, err)
	if found == nil {
		t.Fatal("expected to find resource under its own tenant")
	}

	// Wrong tenant: not found, not an error
	notFound, err := resourceRepo.GetByID(ctx, tenantB.ID, res.ID)
	must(t, err)
	if notFound != nil {
		t.Fatal("tenant isolation violated: resource visible from another tenant")
	}
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
