package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Product struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	StoreID     uuid.UUID       `json:"store_id" db:"store_id"`
	Name        string          `json:"name" db:"name"`
	Description string          `json:"description" db:"description"`
	Price       decimal.Decimal `json:"price" db:"price"`
	Stock       int       `json:"stock" db:"stock"`
	HeightCM    float64   `json:"height_cm" db:"height_cm"`
	WidthCM     float64   `json:"width_cm" db:"width_cm"`
	DepthCM     float64   `json:"depth_cm" db:"depth_cm"`
	WeightKG    float64   `json:"weight_kg" db:"weight_kg"`
	IsOnSale    bool      `json:"is_onsale" db:"is_onsale"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}
