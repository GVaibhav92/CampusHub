package domain

import (
	"time"

	"github.com/google/uuid"
)

type EnrollmentStatus string

const (
	EnrollmentEnrolled           EnrollmentStatus = "ENROLLED"
	EnrollmentCancelledByStudent EnrollmentStatus = "CANCELLED_BY_STUDENT"
	EnrollmentCancelledByFaculty EnrollmentStatus = "CANCELLED_BY_FACULTY"
)

type EventEnrollment struct {
	ID         uuid.UUID        `json:"id"`
	TenantID   uuid.UUID        `json:"tenant_id"`
	BookingID  uuid.UUID        `json:"booking_id"`
	StudentID  uuid.UUID        `json:"student_id"`
	Status     EnrollmentStatus `json:"status"`
	EnrolledAt time.Time        `json:"enrolled_at"`
	UpdatedAt  time.Time        `json:"updated_at"`
}

type RosterStudent struct {
	StudentID  uuid.UUID        `json:"student_id"`
	Email      string           `json:"email"`
	Status     EnrollmentStatus `json:"status"`
	EnrolledAt time.Time        `json:"enrolled_at"`
}

type EventSession struct {
	ID               uuid.UUID `json:"id"`
	TenantID         uuid.UUID `json:"tenant_id"`
	ResourceID       uuid.UUID `json:"resource_id"`
	ResourceName     string    `json:"resource_name"`
	ResourceLocation string    `json:"resource_location"`
	Capacity         int       `json:"capacity"`
	FacultyID        uuid.UUID `json:"faculty_id"`
	FacultyEmail     string    `json:"faculty_email"`
	Title            string    `json:"title"`
	CourseCode       *string   `json:"course_code,omitempty"`
	StartTime        time.Time `json:"start_time"`
	EndTime          time.Time `json:"end_time"`
	Status           string    `json:"status"`
	ActiveEnrolled   int       `json:"active_enrolled"`
	IsEnrolled       bool      `json:"is_enrolled"`
}
