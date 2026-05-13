package order

import (
	"time"

	"go-marketplace/internal/domain"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CheckoutRequest struct {
	PaymentMethod         domain.PaymentMethod `json:"payment_method"`
	AddressID             *uuid.UUID           `json:"address_id,omitempty"`
	ShippingRecipientName string               `json:"shipping_recipient_name,omitempty"`
	ShippingPhoneNumber   string               `json:"shipping_phone_number,omitempty"`
	ShippingStreetAddress string               `json:"shipping_street_address,omitempty"`
	ShippingCity          string               `json:"shipping_city,omitempty"`
	ShippingProvince      string               `json:"shipping_province,omitempty"`
	ShippingPostalCode    string               `json:"shipping_postal_code,omitempty"`
}

func (r CheckoutRequest) Validate() error {
	errs := make(domain.ValidationErrors)
	if r.PaymentMethod == "" {
		errs["payment_method"] = "is required"
	} else if r.PaymentMethod != domain.PaymentMethodWallet && r.PaymentMethod != domain.PaymentMethodMidtrans {
		errs["payment_method"] = "is invalid"
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

type OrderResponse struct {
	ID                    uuid.UUID           `json:"id"`
	PaymentID             uuid.UUID           `json:"payment_id"`
	MerchantID            uuid.UUID           `json:"merchant_id"`
	Status                domain.OrderStatus  `json:"status"`
	TotalAmount           decimal.Decimal     `json:"total_amount"`
	ShippingRecipientName string              `json:"shipping_recipient_name"`
	ShippingPhoneNumber   string              `json:"shipping_phone_number"`
	ShippingStreetAddress string              `json:"shipping_street_address"`
	ShippingCity          string              `json:"shipping_city"`
	ShippingProvince      string              `json:"shipping_province"`
	ShippingPostalCode    string              `json:"shipping_postal_code"`
	IsAppealed            bool                `json:"is_appealed"`
	Items                 []OrderItemResponse `json:"items,omitempty"`
	CreatedAt             time.Time           `json:"created_at"`
	UpdatedAt             time.Time           `json:"updated_at"`
}

type OrderItemResponse struct {
	ID        uuid.UUID       `json:"id"`
	ProductID uuid.UUID       `json:"product_id"`
	Quantity  int             `json:"quantity"`
	Price     decimal.Decimal `json:"price"`
}

type UpdateStatusRequest struct {
	Status domain.OrderStatus `json:"status"`
}

func (r UpdateStatusRequest) Validate() error {
	errs := make(domain.ValidationErrors)
	if r.Status == "" {
		errs["status"] = "is required"
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

type AppealOrderRequest struct {
	Reason string `json:"reason"`
}

func (r AppealOrderRequest) Validate() error {
	errs := make(domain.ValidationErrors)
	if r.Reason == "" {
		errs["reason"] = "is required"
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}
