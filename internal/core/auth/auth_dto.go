package auth

import (
	"go-marketplace/internal/domain"
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
	errs := make(domain.ValidationErrors)
	if r.FullName == "" {
		errs["full_name"] = "is required"
	}
	if r.Email == "" {
		errs["email"] = "is required"
	}
	if r.Password == "" {
		errs["password"] = "is required"
	} else if len(r.Password) < 6 {
		errs["password"] = "must be at least 6 characters"
	}
	if r.Username == "" {
		errs["username"] = "is required"
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

type AuthResponse struct {
	ID           uuid.UUID `json:"id"`
	FullName     string    `json:"full_name"`
	Username     string    `json:"username"`
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
	errs := make(domain.ValidationErrors)
	if r.UserID == uuid.Nil {
		errs["user_id"] = "is required"
	}
	if r.Code == "" {
		errs["code"] = "is required"
	} else if len(r.Code) != 6 {
		errs["code"] = "must be 6 digits"
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r LoginRequest) Validate() error {
	errs := make(domain.ValidationErrors)
	if r.Email == "" {
		errs["email"] = "is required"
	}
	if r.Password == "" {
		errs["password"] = "is required"
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}
