package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type TokenValidator interface {
	Validate(tokenString string) (sub string, err error)
}

const ctxKeySub = "auth_sub"

func RequireAuth(validator TokenValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}
		sub, err := validator.Validate(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}
		c.Set(ctxKeySub, sub)
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

type localClaims struct {
	Sub string `json:"sub"`
	jwt.RegisteredClaims
}

type LocalTokenIssuer struct {
	secret []byte
	ttl    time.Duration
}

func NewLocalTokenIssuer(secret string, ttl time.Duration) *LocalTokenIssuer {
	return &LocalTokenIssuer{secret: []byte(secret), ttl: ttl}
}

func (i *LocalTokenIssuer) Issue(sub string) (string, error) {
	claims := localClaims{
		Sub: sub,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(i.ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(i.secret)
}

type LocalTokenValidator struct {
	secret []byte
}

func NewLocalTokenValidator(secret string) *LocalTokenValidator {
	return &LocalTokenValidator{secret: []byte(secret)}
}

func (v *LocalTokenValidator) Validate(tokenString string) (string, error) {
	claims := &localClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return v.secret, nil
	})
	if err != nil || !token.Valid {
		return "", err
	}
	return claims.Sub, nil
}
