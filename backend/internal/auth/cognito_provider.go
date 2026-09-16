package auth

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
)

const cognitoManagedPasswordPlaceholder = "MANAGED_BY_COGNITO"

// CognitoProvider delegates identity entirely to Cognito. Our own `users`
// row still gets created (for the tenant_id/role mapping the whole
// multi-tenancy model depends on) but never stores or checks a real
// password.
type CognitoProvider struct {
	client   *cognitoidentityprovider.Client
	poolID   string
	clientID string
}

func NewCognitoProvider(client *cognitoidentityprovider.Client, poolID, clientID string) *CognitoProvider {
	return &CognitoProvider{client: client, poolID: poolID, clientID: clientID}
}

func (p *CognitoProvider) SignUp(ctx context.Context, email, password string) (string, string, error) {
	out, err := p.client.SignUp(ctx, &cognitoidentityprovider.SignUpInput{
		ClientId: aws.String(p.clientID),
		Username: aws.String(email),
		Password: aws.String(password),
		UserAttributes: []types.AttributeType{
			{Name: aws.String("email"), Value: aws.String(email)},
		},
	})
	if err != nil {
		return "", "", fmt.Errorf("cognito sign up: %w", err)
	}

	// Skips the email-verification loop entirely - matches the Day 12
	// decision not to depend on SES/email delivery for testing.
	_, err = p.client.AdminConfirmSignUp(ctx, &cognitoidentityprovider.AdminConfirmSignUpInput{
		UserPoolId: aws.String(p.poolID),
		Username:   aws.String(email),
	})
	if err != nil {
		return "", "", fmt.Errorf("cognito admin confirm: %w", err)
	}
	return aws.ToString(out.UserSub), cognitoManagedPasswordPlaceholder, nil
}

func (p *CognitoProvider) Login(ctx context.Context, email, password string) (string, error) {
	out, err := p.client.InitiateAuth(ctx, &cognitoidentityprovider.InitiateAuthInput{
		AuthFlow:       types.AuthFlowTypeUserPasswordAuth,
		ClientId:       aws.String(p.clientID),
		AuthParameters: map[string]string{"USERNAME": email, "PASSWORD": password},
	})
	if err != nil {
		return "", fmt.Errorf("cognito login: %w", err)
	}
	if out.AuthenticationResult == nil {
		return "", fmt.Errorf("cognito login: unexpected challenge response")
	}
	return aws.ToString(out.AuthenticationResult.AccessToken), nil
}
