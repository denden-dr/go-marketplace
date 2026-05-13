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
	errs := make(domain.ValidationErrors)
	if r.RecipientName == "" {
		errs["recipient_name"] = "is required"
	}
	if r.PhoneNumber == "" {
		errs["phone_number"] = "is required"
	}
	if r.StreetAddress == "" {
		errs["street_address"] = "is required"
	}
	if r.City == "" {
		errs["city"] = "is required"
	}
	if r.Province == "" {
		errs["province"] = "is required"
	}
	if r.PostalCode == "" {
		errs["postal_code"] = "is required"
	}
	if r.Tag != domain.AddressTagHome && r.Tag != domain.AddressTagWork {
		errs["tag"] = "is invalid"
	}

	if len(errs) > 0 {
		return errs
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
