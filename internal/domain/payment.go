package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type PaymentType string

const (
	PaymentTypeOrder    PaymentType = "order"
	PaymentTypeTopup    PaymentType = "topup"
	PaymentTypeWithdraw PaymentType = "withdraw"
)

type PaymentMethod string

const (
	PaymentMethodWallet   PaymentMethod = "wallet"
	PaymentMethodMidtrans PaymentMethod = "midtrans"
)

type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusSuccess   PaymentStatus = "success"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusExpired   PaymentStatus = "expired"
	PaymentStatusCancelled PaymentStatus = "cancelled"
)

type Payment struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	UserID      uuid.UUID       `json:"user_id" db:"user_id"`
	ExternalID  *string         `json:"external_id" db:"external_id"` // Midtrans order_id
	Amount      decimal.Decimal `json:"amount" db:"amount"`
	Type        PaymentType     `json:"payment_type" db:"payment_type"`
	Method      PaymentMethod   `json:"payment_method" db:"payment_method"`
	Status      PaymentStatus   `json:"status" db:"status"`
	ReferenceID uuid.UUID       `json:"reference_id" db:"reference_id"` // OrderID or TopupID
	SnapToken   *string         `json:"snap_token" db:"snap_token"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`
}

type PaymentDistribution struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	PaymentID   uuid.UUID       `json:"payment_id" db:"payment_id"`
	RecipientID uuid.UUID       `json:"recipient_id" db:"recipient_id"` // Merchant User ID
	Amount      decimal.Decimal `json:"amount" db:"amount"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
}
