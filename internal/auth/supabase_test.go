package auth

import (
	"context"
	"go-shop-yourself/internal/domain"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestSupabaseAuthClient_VerifyAccessToken_Success(t *testing.T) {
	secret := "test-secret"
	client := NewSupabaseAuthClient(secret)

	claims := &SupabaseClaims{
		Email: "user@example.com",
		AppMetadata: struct {
			Provider  string   `json:"provider"`
			Providers []string `json:"providers"`
		}{
			Provider:  "google",
			Providers: []string{"google"},
		},
		UserMetadata: struct {
			FullName string `json:"full_name"`
			Name     string `json:"name"`
		}{
			FullName: "Test User",
		},
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-123",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))

	res, err := client.VerifyAccessToken(context.Background(), tokenString)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "user-123", res.UserID)
	assert.Equal(t, "user@example.com", res.Email)
	assert.Equal(t, "Test User", res.Name)
	assert.Equal(t, domain.AuthProviderGoogle, res.Provider)
}

func TestSupabaseAuthClient_VerifyAccessToken_Fail_Expired(t *testing.T) {
	secret := "test-secret"
	client := NewSupabaseAuthClient(secret)

	claims := &SupabaseClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-123",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))

	_, err := client.VerifyAccessToken(context.Background(), tokenString)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrInvalidSocialToken, err)
}

func TestSupabaseAuthClient_VerifyAccessToken_Fail_BadSignature(t *testing.T) {
	secret := "test-secret"
	client := NewSupabaseAuthClient(secret)

	claims := &SupabaseClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "user-123",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("wrong-secret"))

	_, err := client.VerifyAccessToken(context.Background(), tokenString)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrInvalidSocialToken, err)
}

func TestSupabaseAuthClient_VerifyAccessToken_Fail_EmailSignInNotAllowed(t *testing.T) {
	secret := "test-secret"
	client := NewSupabaseAuthClient(secret)

	claims := &SupabaseClaims{
		AppMetadata: struct {
			Provider  string   `json:"provider"`
			Providers []string `json:"providers"`
		}{
			Provider: "email",
		},
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "user-123",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))

	_, err := client.VerifyAccessToken(context.Background(), tokenString)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrEmailPasswordSignInNotAllowed, err)
}
