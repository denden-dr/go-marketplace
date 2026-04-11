package dtos

import "github.com/google/uuid"

type MerchantRegisterRequest struct {
	Name  string `json:"name"`
	About string `json:"about"`
	TaxID string `json:"tax_id"`
}

type MerchantResponse struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
	About string    `json:"about"`
}
