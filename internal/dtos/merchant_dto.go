package dtos

import (
	"errors"
	"github.com/google/uuid"
)

type MerchantRegisterRequest struct {
	Name  string `json:"name"`
	About string `json:"about"`
	TaxID string `json:"tax_id"`
}

func (r MerchantRegisterRequest) Validate() error {
	if r.Name == "" {
		return errors.New("merchant name is required")
	}
	if r.TaxID == "" {
		return errors.New("tax id is required")
	}
	return nil
}

type MerchantResponse struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
	About string    `json:"about"`
}
