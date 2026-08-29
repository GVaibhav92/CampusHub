package handler

import (
	"net/http"

	"campushub/backend/internal/domain"
	"campushub/backend/internal/middleware"
	"campushub/backend/internal/repository"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	users repository.UserRepository
}

func NewUserHandler(users repository.UserRepository) *UserHandler {
	return &UserHandler{users: users}
}

// List returns tenant-scoped users, optionally filtered by ?role=. Its
// primary purpose is technician discovery for issue assignment — not
// general user management, which was explicitly deferred. Password hashes
// are stripped from the response; nothing else about a user is sensitive
// enough to warrant a separate response DTO at this scale, but the hash
// specifically must never leave the server.
func (h *UserHandler) List(c *gin.Context) {
	tenantID, _ := middleware.TenantIDFromContext(c)

	var rolePtr *domain.Role
	if roleParam := c.Query("role"); roleParam != "" {
		role := domain.Role(roleParam)
		rolePtr = &role
	}

	users, err := h.users.List(c.Request.Context(), tenantID, rolePtr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list users"})
		return
	}

	type userSummary struct {
		ID    string      `json:"id"`
		Email string      `json:"email"`
		Role  domain.Role `json:"role"`
	}
	out := make([]userSummary, 0, len(users))
	for _, u := range users {
		out = append(out, userSummary{ID: u.ID.String(), Email: u.Email, Role: u.Role})
	}
	c.JSON(http.StatusOK, out)
}
