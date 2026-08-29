package domain

import (
	"time"

	"github.com/google/uuid"
)

type ResourceStatus string

const (
	ResourceAvailable   ResourceStatus = "AVAILABLE"
	ResourceMaintenance ResourceStatus = "MAINTENANCE"
)

type Resource struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	Name      string
	Type      string
	Capacity  int
	Location  string
	Status    ResourceStatus
	CreatedAt time.Time
}
