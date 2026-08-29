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
	ID         uuid.UUID
	TenantID   uuid.UUID
	ResourceID uuid.UUID
	UserID     uuid.UUID
	StartTime  time.Time
	EndTime    time.Time
	Status     BookingStatus
	CreatedAt  time.Time
}
