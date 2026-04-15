package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusPackaging  OrderStatus = "packaging"
	OrderStatusShipping   OrderStatus = "shipping"
	OrderStatusDelivered  OrderStatus = "delivered"
	OrderStatusCancelled  OrderStatus = "cancelled"
)

type CartItem struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	ProductID uuid.UUID `json:"product_id" db:"product_id"`
	Quantity  int       `json:"quantity" db:"quantity"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	// Join fields
	Product *Product `json:"product,omitempty"`
}

type OrderPayment struct {
	ID            uuid.UUID       `json:"id" db:"id"`
	UserID        uuid.UUID       `json:"user_id" db:"user_id"`
	Amount        decimal.Decimal `json:"amount" db:"amount"`
	PaymentMethod string          `json:"payment_method" db:"payment_method"`
	Status        string          `json:"status" db:"status"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
}

type Order struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	PaymentID   uuid.UUID       `json:"payment_id" db:"payment_id"`
	MerchantID  uuid.UUID       `json:"merchant_id" db:"merchant_id"` // StoreID from Product
	UserID          uuid.UUID       `json:"user_id" db:"user_id"`
	Status          OrderStatus     `json:"status" db:"status"`
	TotalAmount           decimal.Decimal `json:"total_amount" db:"total_amount"`
	ShippingRecipientName string          `json:"shipping_recipient_name" db:"shipping_recipient_name"`
	ShippingPhoneNumber   string          `json:"shipping_phone_number" db:"shipping_phone_number"`
	ShippingStreetAddress string          `json:"shipping_street_address" db:"shipping_street_address"`
	ShippingCity          string          `json:"shipping_city" db:"shipping_city"`
	ShippingProvince      string          `json:"shipping_province" db:"shipping_province"`
	ShippingPostalCode    string          `json:"shipping_postal_code" db:"shipping_postal_code"`
	IsAppealed            bool            `json:"is_appealed" db:"is_appealed"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
}

type OrderItem struct {
	ID        uuid.UUID       `json:"id" db:"id"`
	OrderID   uuid.UUID       `json:"order_id" db:"order_id"`
	ProductID uuid.UUID       `json:"product_id" db:"product_id"`
	Quantity  int             `json:"quantity" db:"quantity"`
	Price     decimal.Decimal `json:"price" db:"price"`
	CreatedAt time.Time       `json:"created_at" db:"created_at"`
}

type CancellationAppeal struct {
	ID        uuid.UUID `json:"id" db:"id"`
	OrderID   uuid.UUID `json:"order_id" db:"order_id"`
	Reason    string    `json:"reason" db:"reason"`
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
