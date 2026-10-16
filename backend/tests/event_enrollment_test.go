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

func TestEventEnrollmentWorkflow(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	ctx := context.Background()
	tenantRepo := repository.NewTenantRepository(db)
	userRepo := repository.NewUserRepository(db)
	resourceRepo := repository.NewResourceRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	enrollmentRepo := repository.NewEnrollmentRepository(db)

	logger := slog.Default()
	bookingSvc := service.NewBookingService(db, bookingRepo, resourceRepo, nil, logger)
	enrollmentSvc := service.NewEnrollmentService(db, enrollmentRepo, bookingRepo, resourceRepo, logger)

	// 1. Setup Tenant
	tenant := &domain.Tenant{ID: uuid.New(), Name: "Tech University", Slug: "tech-" + uuid.NewString()[:8]}
	must(t, tenantRepo.Create(ctx, tenant))

	// 2. Setup Users
	faculty := &domain.User{
		ID: uuid.New(), TenantID: tenant.ID, CognitoSub: uuid.NewString(),
		Email: "prof.sharma@tech.edu", PasswordHash: "x", Role: domain.RoleFaculty,
	}
	must(t, userRepo.Create(ctx, faculty))

	otherFaculty := &domain.User{
		ID: uuid.New(), TenantID: tenant.ID, CognitoSub: uuid.NewString(),
		Email: "prof.rao@tech.edu", PasswordHash: "x", Role: domain.RoleFaculty,
	}
	must(t, userRepo.Create(ctx, otherFaculty))

	student1 := &domain.User{
		ID: uuid.New(), TenantID: tenant.ID, CognitoSub: uuid.NewString(),
		Email: "alice@tech.edu", PasswordHash: "x", Role: domain.RoleStudent,
	}
	must(t, userRepo.Create(ctx, student1))

	student2 := &domain.User{
		ID: uuid.New(), TenantID: tenant.ID, CognitoSub: uuid.NewString(),
		Email: "bob@tech.edu", PasswordHash: "x", Role: domain.RoleStudent,
	}
	must(t, userRepo.Create(ctx, student2))

	student3 := &domain.User{
		ID: uuid.New(), TenantID: tenant.ID, CognitoSub: uuid.NewString(),
		Email: "charlie@tech.edu", PasswordHash: "x", Role: domain.RoleStudent,
	}
	must(t, userRepo.Create(ctx, student3))

	// 3. Admin creates resource: Classroom (Capacity: 2)
	classroom := &domain.Resource{
		ID: uuid.New(), TenantID: tenant.ID, Name: "Classroom 101", Type: "CLASSROOM",
		Capacity: 2, Location: "Academic Block 1", Status: domain.ResourceAvailable,
	}
	must(t, resourceRepo.Create(ctx, classroom))

	// 4. Faculty reserves Classroom for "Guest Lecture"
	startTime := time.Now().Add(24 * time.Hour).Truncate(time.Hour)
	endTime := startTime.Add(2 * time.Hour)
	courseCode := "CS500"
	booking, err := bookingSvc.CreateBooking(ctx, tenant.ID, faculty.ID, classroom.ID, "Guest Lecture: AI Systems", &courseCode, startTime, endTime)
	must(t, err)

	// 5. Student 1 and Student 2 enroll
	_, err = enrollmentSvc.Enroll(ctx, tenant.ID, booking.ID, student1.ID)
	must(t, err)

	_, err = enrollmentSvc.Enroll(ctx, tenant.ID, booking.ID, student2.ID)
	must(t, err)

	// 6. Student 3 attempts to enroll -> capacity exceeded (classroom capacity = 2)
	_, err = enrollmentSvc.Enroll(ctx, tenant.ID, booking.ID, student3.ID)
	if !errors.Is(err, service.ErrCapacityExceeded) {
		t.Fatalf("expected ErrCapacityExceeded, got %v", err)
	}

	// 7. Non-owner faculty attempts to view roster -> forbidden
	_, err = enrollmentSvc.GetRoster(ctx, tenant.ID, booking.ID, otherFaculty.ID, domain.RoleFaculty)
	if !errors.Is(err, service.ErrForbidden) {
		t.Fatalf("expected ErrForbidden for non-owner faculty roster access, got %v", err)
	}

	// 8. Hosting Faculty views roster
	roster, err := enrollmentSvc.GetRoster(ctx, tenant.ID, booking.ID, faculty.ID, domain.RoleFaculty)
	must(t, err)
	if len(roster) != 2 {
		t.Fatalf("expected 2 roster entries, got %d", len(roster))
	}

	// 9. Faculty removes Student 1 (terminal removal)
	err = enrollmentSvc.RemoveStudentFromRoster(ctx, tenant.ID, booking.ID, faculty.ID, domain.RoleFaculty, student1.ID)
	must(t, err)

	// 10. Student 3 can now enroll into the freed seat
	_, err = enrollmentSvc.Enroll(ctx, tenant.ID, booking.ID, student3.ID)
	must(t, err)

	// 11. Student 1 attempts to re-enroll -> terminal removal blocks re-enrollment
	_, err = enrollmentSvc.Enroll(ctx, tenant.ID, booking.ID, student1.ID)
	if !errors.Is(err, service.ErrEnrollmentRevokedByFaculty) {
		t.Fatalf("expected ErrEnrollmentRevokedByFaculty, got %v", err)
	}

	// 12. Student 2 self-cancels and re-enrolls
	err = enrollmentSvc.CancelStudentEnrollment(ctx, tenant.ID, booking.ID, student2.ID)
	must(t, err)

	// Student 2 re-enrolls successfully
	_, err = enrollmentSvc.Enroll(ctx, tenant.ID, booking.ID, student2.ID)
	must(t, err)
}
