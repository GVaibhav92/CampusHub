package tests

import (
	"context"
	"testing"

	"campushub/backend/internal/domain"
	"campushub/backend/internal/repository"

	"github.com/google/uuid"
)

func TestIssueTenantScope(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	ctx := context.Background()

	tenantRepo := repository.NewTenantRepository(db)
	userRepo := repository.NewUserRepository(db)
	resourceRepo := repository.NewResourceRepository(db)
	issueRepo := repository.NewIssueRepository(db)

	tenantA := &domain.Tenant{ID: uuid.New(), Name: "VIT", Slug: "vit-" + uuid.NewString()[:8]}
	tenantB := &domain.Tenant{ID: uuid.New(), Name: "SRM", Slug: "srm-" + uuid.NewString()[:8]}
	must(t, tenantRepo.Create(ctx, tenantA))
	must(t, tenantRepo.Create(ctx, tenantB))

	reporter := &domain.User{ID: uuid.New(), TenantID: tenantA.ID, CognitoSub: uuid.NewString(), Email: "s@vit.edu", PasswordHash: "x", Role: domain.RoleStudent}
	must(t, userRepo.Create(ctx, reporter))

	resource := &domain.Resource{ID: uuid.New(), TenantID: tenantA.ID, Name: "Lab 204", Type: "LAB", Status: domain.ResourceAvailable}
	must(t, resourceRepo.Create(ctx, resource))

	issue := &domain.Issue{
		ID: uuid.New(), TenantID: tenantA.ID, ResourceID: resource.ID, ReportedBy: reporter.ID,
		Category: "EQUIPMENT", Priority: domain.PriorityHigh, Description: "Projector not working", Status: domain.IssueOpen,
	}
	must(t, issueRepo.Create(ctx, issue))

	found, err := issueRepo.GetByID(ctx, tenantA.ID, issue.ID)
	must(t, err)
	if found == nil {
		t.Fatal("expected to find issue under its own tenant")
	}

	notFound, err := issueRepo.GetByID(ctx, tenantB.ID, issue.ID)
	must(t, err)
	if notFound != nil {
		t.Fatal("tenant isolation violated: issue visible from another tenant")
	}
}
