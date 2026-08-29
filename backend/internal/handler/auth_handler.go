package handler

import (
	"net/http"

	"campushub/backend/internal/domain"
	"campushub/backend/internal/middleware"
	"campushub/backend/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	users   repository.UserRepository
	tenants repository.TenantRepository
	issuer  *middleware.TokenIssuer
}

func NewAuthHandler(users repository.UserRepository, tenants repository.TenantRepository, issuer *middleware.TokenIssuer) *AuthHandler {
	return &AuthHandler{users: users, tenants: tenants, issuer: issuer}
}

type registerRequest struct {
	TenantSlug string `json:"tenant_slug" binding:"required"`
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=8"`
	Role       string `json:"role" binding:"required"`
}

// creates a user under an existing tenant.
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

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process password"})
		return
	}

	user := &domain.User{
		ID:           uuid.New(),
		TenantID:     tenant.ID,
		CognitoSub:   uuid.New().String(),
		Email:        req.Email,
		PasswordHash: string(hash),
		Role:         domain.Role(req.Role),
	}
	if err := h.users.Create(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id": user.ID, "tenant_id": user.TenantID, "email": user.Email, "role": user.Role,
	})
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

	user, err := h.users.GetByEmail(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "login failed"})
		return
	}

	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	token, err := h.issuer.Issue(user.CognitoSub)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}
