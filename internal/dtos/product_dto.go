package dtos

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ProductResponse struct {
	ID          uuid.UUID       `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Price       decimal.Decimal `json:"price"`
	Stock       int             `json:"stock"`
	IsOnSale    bool            `json:"is_onsale"`
	CreatedAt   time.Time       `json:"created_at"`
}

type ProductCreateRequest struct {
	StoreID     uuid.UUID       `json:"store_id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Price       decimal.Decimal `json:"price"`
	Stock       int             `json:"stock"`
	HeightCM    float64         `json:"height_cm"`
	WidthCM     float64         `json:"width_cm"`
	DepthCM     float64         `json:"depth_cm"`
	WeightKG    float64         `json:"weight_kg"`
	IsOnSale    bool            `json:"is_onsale"`
}

func (r ProductCreateRequest) Validate() error {
	if r.Name == "" {
		return errors.New("product name is required")
	}
	if r.Price.IsZero() || r.Price.IsNegative() {
		return errors.New("price must be greater than 0")
	}
	if r.Stock < 0 {
		return errors.New("stock cannot be negative")
	}
	if r.StoreID == uuid.Nil {
		return errors.New("store id is required")
	}
	return nil
}

type ProductUpdateRequest struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Price       decimal.Decimal `json:"price"`
	Stock       int             `json:"stock"`
	IsOnSale    bool            `json:"is_onsale"`
}

func (r ProductUpdateRequest) Validate() error {
	if r.Name == "" {
		return errors.New("product name is required")
	}
	if r.Price.IsZero() || r.Price.IsNegative() {
		return errors.New("price must be greater than 0")
	}
	if r.Stock < 0 {
		return errors.New("stock cannot be negative")
	}
	return nil
}
