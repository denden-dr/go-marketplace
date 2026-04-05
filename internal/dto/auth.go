package dto

import "github.com/google/uuid"

// RegisterRequest is the data transfer object for registering a new user.
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Username string `json:"username" validate:"required,min=3,max=30"`
	Password string `json:"password" validate:"required,min=8"`
}

// LoginRequest is the data transfer object for logging in an existing user.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// AuthResponse is the data transfer object for the authentication response.
type AuthResponse struct {
	Token  string    `json:"token"`
	UserID uuid.UUID `json:"user_id"`
}

