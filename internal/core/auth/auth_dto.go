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
	ID        uuid.UUID `json:"id"`
	FullName  string    `json:"full_name"`
	Username  string    `json:"username"`
	Email        string    `json:"email"`
	CreatedAt    time.Time `json:"created_at"`
	AccessToken  string    `json:"-"`
	RefreshToken string    `json:"-"`
}

type RefreshRequest struct{}

func (r RefreshRequest) Validate() error {
	return nil
}

type LogoutRequest struct{}

func (r LogoutRequest) Validate() error {
	return nil
}

type VerifyRequest struct {
	UserID uuid.UUID `json:"user_id"`
	Code   string    `json:"code"`
}

func (r VerifyRequest) Validate() error {
	if r.UserID == uuid.Nil {
		return errors.New("user_id is required")
	}
	if r.Code == "" {
		return errors.New("code is required")
	}
	if len(r.Code) != 6 {
		return errors.New("code must be 6 digits")
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

