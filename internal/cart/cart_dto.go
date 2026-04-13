package cart

import (
	"errors"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type AddToCartRequest struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
}

func (r AddToCartRequest) Validate() error {
	if r.ProductID == uuid.Nil {
		return errors.New("product id is required")
	}
	if r.Quantity <= 0 {
		return errors.New("quantity must be greater than 0")
	}
	return nil
}

type UpdateCartItemRequest struct {
	Quantity int `json:"quantity"`
}

func (r UpdateCartItemRequest) Validate() error {
	if r.Quantity <= 0 {
		return errors.New("quantity must be greater than 0")
	}
	return nil
}

type CartItemResponse struct {
	ID        uuid.UUID       `json:"id"`
	ProductID uuid.UUID       `json:"product_id"`
	Name      string          `json:"product_name"`
	Price     decimal.Decimal `json:"price"`
	Quantity  int             `json:"quantity"`
	Subtotal  decimal.Decimal `json:"subtotal"`
}

type CartResponse struct {
	Items      []CartItemResponse `json:"items"`
	TotalPrice decimal.Decimal    `json:"total_price"`
}
