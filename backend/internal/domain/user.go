package domain

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleAdmin      Role = "ADMIN"
	RoleStudent    Role = "STUDENT"
	RoleFaculty    Role = "FACULTY"
	RoleTechnician Role = "TECHNICIAN"
)

type User struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	CognitoSub   string
	Email        string
	PasswordHash string
	Role         Role
	CreatedAt    time.Time
}
