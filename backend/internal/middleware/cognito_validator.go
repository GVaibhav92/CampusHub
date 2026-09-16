package middleware

import (
	"errors"
	"fmt"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/golang-jwt/jwt/v5"
)

// validates ACCESS tokens (not ID tokens) issued by
// a Cognito user pool. Access tokens are what Cognito intends for
// authorizing API calls.
type CognitoTokenValidator struct {
	jwks     *keyfunc.JWKS
	issuer   string
	clientID string
}

func NewCognitoTokenValidator(region, userPoolID, clientID string) (*CognitoTokenValidator, error) {
	issuer := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s", region, userPoolID)
	jwks, err := keyfunc.Get(issuer+"/.well-known/jwks.json", keyfunc.Options{})
	if err != nil {
		return nil, fmt.Errorf("fetching Cognito JWKS: %w", err)
	}
	return &CognitoTokenValidator{jwks: jwks, issuer: issuer, clientID: clientID}, nil
}

type cognitoAccessClaims struct {
	Sub      string `json:"sub"`
	ClientID string `json:"client_id"`
	TokenUse string `json:"token_use"`
	Iss      string `json:"iss"`
	jwt.RegisteredClaims
}

func (v *CognitoTokenValidator) Validate(tokenString string) (string, error) {
	claims := &cognitoAccessClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, v.jwks.Keyfunc)
	if err != nil || !token.Valid {
		return "", fmt.Errorf("invalid token: %w", err)
	}
	if claims.TokenUse != "access" {
		return "", errors.New("not an access token")
	}
	if claims.ClientID != v.clientID {
		return "", errors.New("token issued for a different app client")
	}
	if claims.Iss != v.issuer {
		return "", errors.New("unexpected token issuer")
	}
	return claims.Sub, nil
}
