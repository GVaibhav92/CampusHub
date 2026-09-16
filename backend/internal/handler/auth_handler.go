package handler

import (
	"net/http"

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
	Role       string `json:"role" binding:"required"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create identity"})
		return
	}

	user := &domain.User{
		ID: uuid.New(), TenantID: tenant.ID, CognitoSub: sub,
		Email: req.Email, PasswordHash: storedHash, Role: domain.Role(req.Role),
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
