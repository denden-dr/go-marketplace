package auth

import (
	"context"
	"go-shop-yourself/internal/domain"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
)

type FirebaseTokenResult struct {
	UID      string
	Email    string
	Name     string
	Provider string
}

type FirebaseAuthClient interface {
	VerifyIDToken(ctx context.Context, idToken string) (*FirebaseTokenResult, error)
}

type firebaseAuthClient struct {
	client *auth.Client
}

func NewFirebaseAuthClient(app *firebase.App) (FirebaseAuthClient, error) {
	client, err := app.Auth(context.Background())
	if err != nil {
		return nil, err
	}
	return &firebaseAuthClient{client: client}, nil
}

func (c *firebaseAuthClient) VerifyIDToken(ctx context.Context, idToken string) (*FirebaseTokenResult, error) {
	token, err := c.client.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, domain.ErrInvalidFirebaseToken
	}

	res := &FirebaseTokenResult{
		UID: token.UID,
	}

	if email, ok := token.Claims["email"].(string); ok {
		res.Email = email
	}

	if verified, ok := token.Claims["email_verified"].(bool); !ok || !verified {
		return nil, domain.ErrEmailNotVerified
	}

	if name, ok := token.Claims["name"].(string); ok {
		res.Name = name
	}

	// Map Firebase provider to our internal constants
	signInProvider := token.Firebase.SignInProvider
	switch signInProvider {
	case "google.com":
		res.Provider = domain.AuthProviderGoogle
	case "facebook.com":
		res.Provider = domain.AuthProviderFacebook
	case "apple.com":
		res.Provider = domain.AuthProviderApple
	case "twitter.com":
		res.Provider = domain.AuthProviderTwitter
	case "password":
		// Email/Password sign-in via Firebase is explicitly disallowed to prefer local auth
		return nil, domain.ErrFirebasePasswordSignInNotAllowed
	default:
		// Any other unknown providers are tagged as generic firebase social login
		res.Provider = domain.AuthProviderFirebase
	}

	return res, nil
}
