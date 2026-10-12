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

// TestBookingConflictRejected proves the exclusion constraint blocks an
// overlapping booking on the same resource, and that a non-overlapping
// booking on the same resource still succeeds.
func TestBookingConflictRejected(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	ctx := context.Background()
	tenantRepo := repository.NewTenantRepository(db)
	userRepo := repository.NewUserRepository(db)
	resourceRepo := repository.NewResourceRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	bookingSvc := service.NewBookingService(db, bookingRepo, resourceRepo, nil, slog.Default())

	tenant := &domain.Tenant{ID: uuid.New(), Name: "VIT", Slug: "vit-" + uuid.NewString()[:8]}
	must(t, tenantRepo.Create(ctx, tenant))

	user := &domain.User{
		ID: uuid.New(), TenantID: tenant.ID, CognitoSub: uuid.NewString(),
		Email: "s1@vit.edu", PasswordHash: "x", Role: domain.RoleStudent,
	}
	must(t, userRepo.Create(ctx, user))

	resource := &domain.Resource{
		ID: uuid.New(), TenantID: tenant.ID, Name: "Lab 204", Type: "LAB",
		Capacity: 60, Location: "AB-2", Status: domain.ResourceAvailable,
	}
	must(t, resourceRepo.Create(ctx, resource))

	base := time.Now().Add(24 * time.Hour).Truncate(time.Hour)

	// First booking: 10:00-12:00 — should succeed
	_, err := bookingSvc.CreateBooking(ctx, tenant.ID, user.ID, resource.ID, base, base.Add(2*time.Hour))
	must(t, err)

	// Overlapping: 11:00-13:00 — should be rejected
	_, err = bookingSvc.CreateBooking(ctx, tenant.ID, user.ID, resource.ID, base.Add(1*time.Hour), base.Add(3*time.Hour))
	if !errors.Is(err, repository.ErrBookingConflict) {
		t.Fatalf("expected ErrBookingConflict, got %v", err)
	}

	// Non-overlapping: 12:00-13:00 — should succeed (back-to-back is allowed)
	_, err = bookingSvc.CreateBooking(ctx, tenant.ID, user.ID, resource.ID, base.Add(2*time.Hour), base.Add(3*time.Hour))
	must(t, err)
}
