package product

import (
	"go-marketplace/internal/domain"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ProductResponse struct {
	ID          uuid.UUID       `json:"id"`
	Name        string          `json:"name"`
	Description *string         `json:"description"`
	Price       decimal.Decimal `json:"price"`
	Stock       int             `json:"stock"`
	IsOnSale    bool            `json:"is_onsale"`
	CreatedAt   time.Time       `json:"created_at"`
}

type ProductCreateRequest struct {
	StoreID     uuid.UUID       `json:"store_id"`
	Name        string          `json:"name"`
	Description *string         `json:"description"`
	Price       decimal.Decimal `json:"price"`
	Stock       int             `json:"stock"`
	HeightCM    float64         `json:"height_cm"`
	WidthCM     float64         `json:"width_cm"`
	DepthCM     float64         `json:"depth_cm"`
	WeightKG    float64         `json:"weight_kg"`
	IsOnSale    bool            `json:"is_onsale"`
}

func (r ProductCreateRequest) Validate() error {
	errs := make(domain.ValidationErrors)
	if r.Name == "" {
		errs["name"] = "is required"
	}
	if r.Price.IsZero() || r.Price.IsNegative() {
		errs["price"] = "must be greater than 0"
	}
	if r.Stock < 0 {
		errs["stock"] = "cannot be negative"
	}
	if r.StoreID == uuid.Nil {
		errs["store_id"] = "is required"
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

type ProductUpdateRequest struct {
	Name        string          `json:"name"`
	Description *string         `json:"description"`
	Price       decimal.Decimal `json:"price"`
	Stock       int             `json:"stock"`
	IsOnSale    bool            `json:"is_onsale"`
}

func (r ProductUpdateRequest) Validate() error {
	errs := make(domain.ValidationErrors)
	if r.Name == "" {
		errs["name"] = "is required"
	}
	if r.Price.IsZero() || r.Price.IsNegative() {
		errs["price"] = "must be greater than 0"
	}
	if r.Stock < 0 {
		errs["stock"] = "cannot be negative"
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

type ProductSearchRequest struct {
	Query string `query:"q"`
	Limit int    `query:"limit"`
	Page  int    `query:"page"`
}

func (r ProductSearchRequest) Validate() error {
	errs := make(domain.ValidationErrors)
	if r.Query != "" && len(r.Query) < 2 {
		errs["q"] = "must be at least 2 characters"
	}
	if r.Limit < 0 {
		errs["limit"] = "cannot be negative"
	}
	if r.Page < 0 {
		errs["page"] = "cannot be negative"
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}
