package auth

import (
	"context"
	"errors"

	"campushub/backend/internal/middleware"
	"campushub/backend/internal/repository"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type LocalProvider struct {
	users  repository.UserRepository
	issuer *middleware.LocalTokenIssuer
}

func NewLocalProvider(users repository.UserRepository, issuer *middleware.LocalTokenIssuer) *LocalProvider {
	return &LocalProvider{users: users, issuer: issuer}
}

func (p *LocalProvider) SignUp(ctx context.Context, email, password string) (string, string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", "", err
	}
	return uuid.New().String(), string(hash), nil
}

func (p *LocalProvider) Login(ctx context.Context, email, password string) (string, error) {
	user, err := p.users.GetByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.New("invalid email or password")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid email or password")
	}
	return p.issuer.Issue(user.CognitoSub)
}
