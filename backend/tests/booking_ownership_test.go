package tests

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"campushub/backend/internal/domain"
	"campushub/backend/internal/repository"
	"campushub/backend/internal/service"

	"github.com/google/uuid"
)

// TestBookingOwnershipAndTenantIsolation closes the gap flagged in review:
// a non-owner, non-admin cannot cancel another user's booking, and a
// booking ID from another tenant is treated as not-found, not forbidden.
func TestBookingOwnershipAndTenantIsolation(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	ctx := context.Background()

	tenantRepo := repository.NewTenantRepository(db)
	userRepo := repository.NewUserRepository(db)
	resourceRepo := repository.NewResourceRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	bookingSvc := service.NewBookingService(db, bookingRepo, resourceRepo, nil, slog.Default())

	tenantA := &domain.Tenant{ID: uuid.New(), Name: "VIT", Slug: "vit-" + uuid.NewString()[:8]}
	tenantB := &domain.Tenant{ID: uuid.New(), Name: "SRM", Slug: "srm-" + uuid.NewString()[:8]}
	must(t, tenantRepo.Create(ctx, tenantA))
	must(t, tenantRepo.Create(ctx, tenantB))

	owner := &domain.User{ID: uuid.New(), TenantID: tenantA.ID, CognitoSub: uuid.NewString(), Email: "owner@vit.edu", PasswordHash: "x", Role: domain.RoleStudent}
	otherStudent := &domain.User{ID: uuid.New(), TenantID: tenantA.ID, CognitoSub: uuid.NewString(), Email: "other@vit.edu", PasswordHash: "x", Role: domain.RoleStudent}
	must(t, userRepo.Create(ctx, owner))
	must(t, userRepo.Create(ctx, otherStudent))

	resource := &domain.Resource{ID: uuid.New(), TenantID: tenantA.ID, Name: "Lab 204", Type: "LAB", Status: domain.ResourceAvailable}
	must(t, resourceRepo.Create(ctx, resource))

	start := time.Now().Add(48 * time.Hour).Truncate(time.Hour)
	booking, err := bookingSvc.CreateBooking(ctx, tenantA.ID, owner.ID, resource.ID, "Physics Lab", nil, start, start.Add(time.Hour))
	must(t, err)

	// Different student, same tenant, not the owner -> forbidden
	err = bookingSvc.CancelBooking(ctx, tenantA.ID, otherStudent.ID, domain.RoleStudent, booking.ID)
	if !errors.Is(err, service.ErrForbidden) {
		t.Fatalf("expected ErrForbidden for non-owner cancel, got %v", err)
	}

	// Different tenant entirely -> not found, not forbidden (isolation, not just permission)
	err = bookingSvc.CancelBooking(ctx, tenantB.ID, otherStudent.ID, domain.RoleStudent, booking.ID)
	if err == nil {
		t.Fatal("expected an error for cross-tenant cancel attempt")
	}

	// Owner cancels their own booking -> succeeds
	must(t, bookingSvc.CancelBooking(ctx, tenantA.ID, owner.ID, domain.RoleStudent, booking.ID))
}
