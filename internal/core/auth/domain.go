package auth

import (
	"go-marketplace/internal/domain"
	"time"
)

// VerificationCode represents a rich domain entity for a verification code
type VerificationCode struct {
	model *domain.VerificationCode
}

// NewVerificationCode creates a new rich VerificationCode entity
func NewVerificationCode(m *domain.VerificationCode) *VerificationCode {
	return &VerificationCode{model: m}
}

// IsValid checks if the provided code matches the stored hash
func (v *VerificationCode) IsValid(code string) bool {
	return HashToken(code) == v.model.CodeHash
}

// IsExpired returns true if the verification code has expired
func (v *VerificationCode) IsExpired() bool {
	return time.Now().After(v.model.ExpiresAt)
}

// Session represents a rich domain entity for a user session
type Session struct {
	model *domain.Session
}

// NewSession creates a new rich Session entity
func NewSession(m *domain.Session) *Session {
	return &Session{model: m}
}

// CanRefresh checks if the session is eligible for token rotation
func (s *Session) CanRefresh() error {
	if s.model.IsRevoked {
		return domain.ErrRefreshTokenReused
	}
	if time.Now().After(s.model.ExpiresAt) {
		return domain.ErrRefreshTokenExpired
	}
	return nil
}
