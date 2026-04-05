package dto

import (
	"time"

	"github.com/denden-dr/go-shop-yourself/internal/domain"
	"github.com/google/uuid"
)

// UserProfileResponse is the data transfer object for returning user profile information.
type UserProfileResponse struct {
	UserID       uuid.UUID       `json:"user_id"`
	Username     string          `json:"username"`
	SavedAddress *domain.Address `json:"saved_address,omitempty"`
	UpdatedAt    time.Time       `json:"updated_at"`
	LastLoginAt  *time.Time      `json:"last_login_at,omitempty"`
}
