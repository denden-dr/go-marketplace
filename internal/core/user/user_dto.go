package user

import (
	"time"

	"go-marketplace/internal/domain"

	"github.com/google/uuid"
)

type UserResponse struct {
	ID           uuid.UUID `json:"id"`
	FullName     string    `json:"full_name"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	AuthProvider string    `json:"auth_provider"`
	ProviderID   *string   `json:"provider_id"`
	CreatedAt    time.Time `json:"created_at"`
}

type AddressRequest struct {
	Tag           domain.AddressTag `json:"tag"`
	RecipientName string            `json:"recipient_name"`
	PhoneNumber   string            `json:"phone_number"`
	StreetAddress string            `json:"street_address"`
	City          string            `json:"city"`
	Province      string            `json:"province"`
	PostalCode    string            `json:"postal_code"`
	IsDefault     bool              `json:"is_default"`
}

func (r *AddressRequest) Validate() error {
	if r.RecipientName == "" {
		return domain.ErrRecipientNameRequired
	}
	if r.PhoneNumber == "" {
		return domain.ErrPhoneNumberRequired
	}
	if r.StreetAddress == "" {
		return domain.ErrStreetAddressRequired
	}
	if r.City == "" {
		return domain.ErrCityRequired
	}
	if r.Province == "" {
		return domain.ErrProvinceRequired
	}
	if r.PostalCode == "" {
		return domain.ErrPostalCodeRequired
	}
	if r.Tag != domain.AddressTagHome && r.Tag != domain.AddressTagWork {
		return domain.ErrInvalidAddressTag
	}
	return nil
}

type AddressResponse struct {
	ID            uuid.UUID         `json:"id"`
	Tag           domain.AddressTag `json:"tag"`
	RecipientName string            `json:"recipient_name"`
	PhoneNumber   string            `json:"phone_number"`
	StreetAddress string            `json:"street_address"`
	City          string            `json:"city"`
	Province      string            `json:"province"`
	PostalCode    string            `json:"postal_code"`
	IsDefault     bool              `json:"is_default"`
	CreatedAt     time.Time         `json:"created_at"`
}
