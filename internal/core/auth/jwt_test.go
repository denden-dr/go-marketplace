package auth

import (
	"testing"

	"go-marketplace/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestJWT(t *testing.T) {
	secret := "test-secret"
	userID := uuid.New()
	role := domain.RoleMerchant

	t.Run("Generate and Validate Access Token", func(t *testing.T) {
		token, err := GenerateAccessToken(userID, role, secret)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		claims, err := ValidateAccessToken(token, secret)
		assert.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, role, claims.Role)
	})

	t.Run("Invalid Token", func(t *testing.T) {
		claims, err := ValidateAccessToken("invalid-token", secret)
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("Wrong Secret", func(t *testing.T) {
		token, err := GenerateAccessToken(userID, role, secret)
		assert.NoError(t, err)

		claims, err := ValidateAccessToken(token, "wrong-secret")
		assert.Error(t, err)
		assert.Nil(t, claims)
	})
}

func TestGenerateRefreshToken(t *testing.T) {
	token, err := GenerateRefreshToken()
	assert.NoError(t, err)
	assert.Len(t, token, 64) // 32 bytes hex encoded
}

func TestGenerateVerificationCode(t *testing.T) {
	code, err := GenerateVerificationCode()
	assert.NoError(t, err)
	assert.Len(t, code, 6)
}
