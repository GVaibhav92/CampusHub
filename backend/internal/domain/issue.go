package domain

import (
	"time"

	"github.com/google/uuid"
)

type IssueStatus string

const (
	IssueOpen       IssueStatus = "OPEN"
	IssueInProgress IssueStatus = "IN_PROGRESS"
	IssueResolved   IssueStatus = "RESOLVED"
	IssueClosed     IssueStatus = "CLOSED"
)

type IssuePriority string

const (
	PriorityLow    IssuePriority = "LOW"
	PriorityMedium IssuePriority = "MEDIUM"
	PriorityHigh   IssuePriority = "HIGH"
)

type Issue struct {
	ID            uuid.UUID
	TenantID      uuid.UUID
	ResourceID    uuid.UUID
	ReportedBy    uuid.UUID
	AssignedTo    *uuid.UUID
	Category      string
	Priority      IssuePriority
	Description   string
	Status        IssueStatus
	AttachmentKey *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
