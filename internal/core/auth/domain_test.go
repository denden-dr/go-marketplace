package auth

import (
	"go-marketplace/internal/domain"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestVerificationCode_IsValid(t *testing.T) {
	code := "123456"
	hash := HashToken(code)

	vc := NewVerificationCode(&domain.VerificationCode{
		CodeHash: hash,
	})

	assert.True(t, vc.IsValid(code))
	assert.False(t, vc.IsValid("wrong"))
}

func TestVerificationCode_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		want      bool
	}{
		{"not expired", time.Now().Add(time.Hour), false},
		{"expired", time.Now().Add(-time.Hour), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vc := NewVerificationCode(&domain.VerificationCode{
				ExpiresAt: tt.expiresAt,
			})
			assert.Equal(t, tt.want, vc.IsExpired())
		})
	}
}

func TestSession_CanRefresh(t *testing.T) {
	tests := []struct {
		name      string
		isRevoked bool
		expiresAt time.Time
		wantErr   error
	}{
		{"valid session", false, time.Now().Add(time.Hour), nil},
		{"revoked session", true, time.Now().Add(time.Hour), domain.ErrRefreshTokenReused},
		{"expired session", false, time.Now().Add(-time.Hour), domain.ErrRefreshTokenExpired},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSession(&domain.Session{
				IsRevoked: tt.isRevoked,
				ExpiresAt: tt.expiresAt,
			})
			assert.Equal(t, tt.wantErr, s.CanRefresh())
		})
	}
}
