package domain

import (
	"time"

	"github.com/google/uuid"
)

type BookingStatus string

const (
	BookingConfirmed BookingStatus = "CONFIRMED"
	BookingCancelled BookingStatus = "CANCELLED"
)

type Booking struct {
	ID         uuid.UUID  `json:"id"`
	TenantID   uuid.UUID  `json:"tenant_id"`
	ResourceID uuid.UUID  `json:"resource_id"`
	UserID     uuid.UUID  `json:"user_id"`
	Title      string     `json:"title"`
	CourseCode *string    `json:"course_code,omitempty"`
	StartTime  time.Time  `json:"start_time"`
	EndTime    time.Time  `json:"end_time"`
	Status     BookingStatus `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
}
