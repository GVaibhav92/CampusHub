package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type localClaims struct {
	Sub string `json:"sub"`
	jwt.RegisteredClaims
}

// TokenIssuer is a stand-in for Cognito during local development. It is
// deleted, not modified, in Week 3 — replaced by JWKS signature validation
// against Cognito's public keys. Nothing downstream changes: every other
// piece of the pipeline only ever consumes `sub`, never trusts a claim it
// issued itself, so the swap is a one-file change.
type TokenIssuer struct {
	secret []byte
	ttl    time.Duration
}

func NewTokenIssuer(secret string, ttl time.Duration) *TokenIssuer {
	return &TokenIssuer{secret: []byte(secret), ttl: ttl}
}

func (i *TokenIssuer) Issue(sub string) (string, error) {
	claims := localClaims{
		Sub: sub,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(i.ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(i.secret)
}

const ctxKeySub = "auth_sub"

// RequireAuth validates the bearer token and stores only `sub` in context.
// It deliberately does NOT resolve tenant_id or role — that's the tenant
// middleware's single responsibility, kept separate on purpose.
func RequireAuth(issuer *TokenIssuer) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}
		tokenString := strings.TrimPrefix(header, "Bearer ")

		claims := &localClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
			return issuer.secret, nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set(ctxKeySub, claims.Sub)
		c.Next()
	}
}

func SubFromContext(c *gin.Context) (string, bool) {
	v, ok := c.Get(ctxKeySub)
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}
