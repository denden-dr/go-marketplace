package auth

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type RegisterRequest struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Username string `json:"username"`
}

func (r RegisterRequest) Validate() error {
	if r.FullName == "" {
		return errors.New("full name is required")
	}
	if r.Email == "" {
		return errors.New("email is required")
	}
	if r.Password == "" {
		return errors.New("password is required")
	}
	if len(r.Password) < 6 {
		return errors.New("password must be at least 6 characters")
	}
	if r.Username == "" {
		return errors.New("username is required")
	}
	return nil
}

type AuthResponse struct {
	ID           uuid.UUID `json:"id"`
	FullName     string    `json:"full_name"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	CreatedAt    time.Time `json:"created_at"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (r RefreshRequest) Validate() error {
	if r.RefreshToken == "" {
		return errors.New("refresh token is required")
	}
	return nil
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (r LogoutRequest) Validate() error {
	if r.RefreshToken == "" {
		return errors.New("refresh token is required")
	}
	return nil
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r LoginRequest) Validate() error {
	if r.Email == "" {
		return errors.New("email is required")
	}
	if r.Password == "" {
		return errors.New("password is required")
	}
	return nil
}

type SocialLoginRequest struct {
	AccessToken string `json:"access_token"`
}

func (r SocialLoginRequest) Validate() error {
	if r.AccessToken == "" {
		return errors.New("access token is required")
	}
	if len(r.AccessToken) > 5120 {
		return errors.New("access token is too long")
	}
	return nil
}
