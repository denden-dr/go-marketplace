package dtos

import (
	"time"

	"github.com/google/uuid"
)

type ProductResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	IsOnSale    bool      `json:"is_onsale"`
	CreatedAt   time.Time `json:"created_at"`
}

type ProductCreateRequest struct {
	StoreID     uuid.UUID `json:"store_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	HeightCM    float64   `json:"height_cm"`
	WidthCM     float64   `json:"width_cm"`
	DepthCM     float64   `json:"depth_cm"`
	WeightKG    float64   `json:"weight_kg"`
	IsOnSale    bool      `json:"is_onsale"`
}

type ProductUpdateRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
	IsOnSale    bool    `json:"is_onsale"`
}
