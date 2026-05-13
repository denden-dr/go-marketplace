package merchant

import (
	"go-marketplace/internal/domain"

	"github.com/google/uuid"
)

type MerchantRegisterRequest struct {
	Name  string `json:"name"`
	About string `json:"about"`
	TaxID string `json:"tax_id"`
}

func (r MerchantRegisterRequest) Validate() error {
	errs := make(domain.ValidationErrors)
	if r.Name == "" {
		errs["name"] = "is required"
	}
	if r.TaxID == "" {
		errs["tax_id"] = "is required"
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

type MerchantResponse struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
	About string    `json:"about"`
}
