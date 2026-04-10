package dtos

import (
	"time"

	"github.com/google/uuid"
)

type WalletResponse struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	WalletNumber string    `json:"wallet_number"`
	Balance      float64   `json:"balance"`
	Currency     string    `json:"currency"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type TransactionResponse struct {
	ID           uuid.UUID `json:"id"`
	WalletID     uuid.UUID `json:"wallet_id"`
	Amount       float64   `json:"amount"`
	Direction    string    `json:"direction"`
	Type         string    `json:"type"`
	Status       string    `json:"status"`
	ReferenceID  string    `json:"reference_id"`
	BalanceAfter float64   `json:"balance_after"`
	Description  string    `json:"description"`
	CreatedAt    time.Time `json:"created_at"`
}

type WithdrawRequest struct {
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
}
