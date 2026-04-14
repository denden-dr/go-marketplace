package user

import (
	"time"

	"go-shop-yourself/internal/domain"

	"github.com/google/uuid"
)

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	FullName  string    `json:"full_name"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type AddressRequest struct {
	Tag            domain.AddressTag `json:"tag"`
	RecipientName  string            `json:"recipient_name"`
	PhoneNumber    string            `json:"phone_number"`
	StreetAddress  string            `json:"street_address"`
	City           string            `json:"city"`
	Province       string            `json:"province"`
	PostalCode     string            `json:"postal_code"`
	IsDefault      bool              `json:"is_default"`
}

type AddressResponse struct {
	ID             uuid.UUID         `json:"id"`
	Tag            domain.AddressTag `json:"tag"`
	RecipientName  string    `json:"recipient_name"`
	PhoneNumber    string    `json:"phone_number"`
	StreetAddress  string    `json:"street_address"`
	City           string    `json:"city"`
	Province       string    `json:"province"`
	PostalCode     string    `json:"postal_code"`
	IsDefault      bool      `json:"is_default"`
	CreatedAt      time.Time `json:"created_at"`
}
