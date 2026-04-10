package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type WalletStatus string

const (
	WalletStatusActive WalletStatus = "active"
	WalletStatusFrozen WalletStatus = "frozen"
	WalletStatusClosed WalletStatus = "closed"
)

type TransactionDirection string

const (
	TransactionDirectionIn  TransactionDirection = "in"
	TransactionDirectionOut TransactionDirection = "out"
)

type TransactionType string

const (
	TransactionTypeTopup    TransactionType = "topup"
	TransactionTypeWithdraw TransactionType = "withdraw"
	TransactionTypePayment  TransactionType = "payment"
	TransactionTypeRefund   TransactionType = "refund"
)

type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusSuccess   TransactionStatus = "success"
	TransactionStatusFailed    TransactionStatus = "failed"
	TransactionStatusCancelled TransactionStatus = "cancelled"
)

type Wallet struct {
	ID           uuid.UUID    `json:"id" db:"id"`
	UserID       uuid.UUID    `json:"user_id" db:"user_id"`
	WalletNumber string       `json:"wallet_number" db:"wallet_number"`
	Balance      decimal.Decimal `json:"balance" db:"balance"`
	Currency     string       `json:"currency" db:"currency"`
	Status       WalletStatus `json:"status" db:"status"`
	CreatedAt    time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at" db:"updated_at"`
}

type WalletTransaction struct {
	ID           uuid.UUID            `json:"id" db:"id"`
	WalletID     uuid.UUID            `json:"wallet_id" db:"wallet_id"`
	Amount       decimal.Decimal      `json:"amount" db:"amount"`
	Direction    TransactionDirection `json:"direction" db:"direction"`
	Type         TransactionType      `json:"type" db:"type"`
	Status       TransactionStatus    `json:"status" db:"status"`
	ReferenceID  string               `json:"reference_id" db:"reference_id"`
	BalanceAfter decimal.Decimal      `json:"balance_after" db:"balance_after"`
	Description  string               `json:"description" db:"description"`
	CreatedAt    time.Time            `json:"created_at" db:"created_at"`
}
