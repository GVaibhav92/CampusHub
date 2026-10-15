package handler

import (
	"log"
	"net/http"
	"os"

	"campushub/backend/internal/auth"
	"campushub/backend/internal/domain"
	"campushub/backend/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	users    repository.UserRepository
	tenants  repository.TenantRepository
	provider auth.Provider
}

func NewAuthHandler(users repository.UserRepository, tenants repository.TenantRepository, provider auth.Provider) *AuthHandler {
	return &AuthHandler{users: users, tenants: tenants, provider: provider}
}

type registerRequest struct {
	TenantSlug string `json:"tenant_slug" binding:"required"`
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=8"`
	Role       string `json:"role"` // optional: defaults to STUDENT, only STUDENT or FACULTY allowed
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Enforce registration role allow-list
	assignedRole := domain.RoleStudent
	if req.Role != "" {
		switch domain.Role(req.Role) {
		case domain.RoleStudent:
			assignedRole = domain.RoleStudent
		case domain.RoleFaculty:
			assignedRole = domain.RoleFaculty
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid registration role: only STUDENT or FACULTY are permitted"})
			return
		}
	}
	tenant, err := h.tenants.GetBySlug(c.Request.Context(), req.TenantSlug)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve tenant"})
		return
	}
	if tenant == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "unknown tenant"})
		return
	}

	sub, storedHash, err := h.provider.SignUp(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		log.Printf("REGISTER IDENTITY ERROR: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create identity"})
		return
	}

	user := &domain.User{
		ID: uuid.New(), TenantID: tenant.ID, CognitoSub: sub,
		Email: req.Email, PasswordHash: storedHash, Role: assignedRole,
	}
	if err := h.users.Create(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": user.ID, "tenant_id": user.TenantID, "email": user.Email, "role": user.Role})
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	token, err := h.provider.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

type promoteUserRequest struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"required"`
}

// PromoteUser allows bootstrapping/promoting users to ADMIN or TECHNICIAN using the provisioning key.
func (h *AuthHandler) PromoteUser(c *gin.Context) {
	expectedKey := "campushub-bootstrap-secret"
	if envKey := os.Getenv("TENANT_PROVISIONING_KEY"); envKey != "" {
		expectedKey = envKey
	}

	providedKey := c.GetHeader("X-Tenant-Provisioning-Key")
	if providedKey != expectedKey {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized provisioning request"})
		return
	}

	var req promoteUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var targetRole domain.Role
	switch domain.Role(req.Role) {
	case domain.RoleAdmin:
		targetRole = domain.RoleAdmin
	case domain.RoleTechnician:
		targetRole = domain.RoleTechnician
	case domain.RoleFaculty:
		targetRole = domain.RoleFaculty
	case domain.RoleStudent:
		targetRole = domain.RoleStudent
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role: must be ADMIN, TECHNICIAN, FACULTY, or STUDENT"})
		return
	}

	if err := h.users.UpdateRole(c.Request.Context(), req.Email, targetRole); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user role: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "user role updated successfully",
		"email":   req.Email,
		"role":    targetRole,
	})
}

