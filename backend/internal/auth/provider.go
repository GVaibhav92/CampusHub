package auth

import "context"

// Provider separates "how is identity created/verified" from everything
// else. AuthHandler never knows which implementation is behind it.
type Provider interface {
	// SignUp creates the identity itself - not the campushub `users` row,
	// which AuthHandler still creates, tenant-scoped. Returns the subject
	// to store as cognito_sub, and what to store in password_hash (a real
	// bcrypt hash locally, a placeholder once Cognito owns the password).
	SignUp(ctx context.Context, email, password string) (sub string, storedHash string, err error)
	Login(ctx context.Context, email, password string) (token string, err error)
}
