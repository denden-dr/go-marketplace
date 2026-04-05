package domain

import (
	"time"

	"github.com/google/uuid"
)

// User represents the root user entity.
type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

// UserAuth represents the authentication credentials of a user.
type UserAuth struct {
	UserID       uuid.UUID `json:"user_id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

// UserProfile represents the public profile information of a user.
type UserProfile struct {
	UserID       uuid.UUID  `json:"user_id"`
	Username     string     `json:"username"`
	SavedAddress *Address   `json:"saved_address,omitempty"`
	UpdatedAt    time.Time  `json:"updated_at"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
}

// Address represents a structured physical address.
type Address struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	State   string `json:"state"`
	ZipCode string `json:"zip_code"`
	Country string `json:"country"`
}
