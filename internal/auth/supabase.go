package auth

import (
	"context"
	"fmt"
	"go-shop-yourself/internal/domain"
	"log"

	"github.com/golang-jwt/jwt/v5"
)

type SupabaseTokenResult struct {
	UserID   string
	Email    string
	Name     string
	Provider string
}

type SupabaseAuthClient interface {
	VerifyAccessToken(ctx context.Context, accessToken string) (*SupabaseTokenResult, error)
}

type supabaseAuthClient struct {
	jwtSecret string
}

type SupabaseClaims struct {
	Email       string `json:"email"`
	AppMetadata struct {
		Provider  string   `json:"provider"`
		Providers []string `json:"providers"`
	} `json:"app_metadata"`
	UserMetadata struct {
		FullName string `json:"full_name"`
		Name     string `json:"name"`
	} `json:"user_metadata"`
	jwt.RegisteredClaims
}

func NewSupabaseAuthClient(jwtSecret string) SupabaseAuthClient {
	if jwtSecret == "" {
		return nil
	}
	return &supabaseAuthClient{jwtSecret: jwtSecret}
}

func (c *supabaseAuthClient) VerifyAccessToken(ctx context.Context, accessToken string) (*SupabaseTokenResult, error) {
	token, err := jwt.ParseWithClaims(accessToken, &SupabaseClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(c.jwtSecret), nil
	})

	if err != nil {
		log.Printf("Supabase Auth: failed to parse token: %v", err)
		return nil, domain.ErrInvalidSocialToken
	}

	claims, ok := token.Claims.(*SupabaseClaims)
	if !ok || !token.Valid {
		return nil, domain.ErrInvalidSocialToken
	}

	// Supabase user ID is the 'sub' claim
	userID := claims.Subject
	if userID == "" {
		return nil, domain.ErrInvalidSocialToken
	}

	res := &SupabaseTokenResult{
		UserID: userID,
		Email:  claims.Email,
	}

	// Name can be in full_name or name depending on the provider
	res.Name = claims.UserMetadata.FullName
	if res.Name == "" {
		res.Name = claims.UserMetadata.Name
	}

	// Map Supabase provider to our internal constants
	provider := claims.AppMetadata.Provider
	switch provider {
	case "google":
		res.Provider = domain.AuthProviderGoogle
	case "facebook":
		res.Provider = domain.AuthProviderFacebook
	case "apple":
		res.Provider = domain.AuthProviderApple
	case "twitter":
		res.Provider = domain.AuthProviderTwitter
	case "email":
		// Email/Password sign-in via Supabase is disallowed to prefer local auth
		return nil, domain.ErrEmailPasswordSignInNotAllowed
	default:
		// Any other unknown providers are tagged as generic supabase social login
		res.Provider = domain.AuthProviderSupabase
	}

	return res, nil
}
