package wallet

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type WalletResponse struct {
	ID             uuid.UUID       `json:"id"`
	UserID         uuid.UUID       `json:"user_id"`
	WalletNumber   string          `json:"wallet_number"`
	Balance        decimal.Decimal `json:"balance"`
	PendingBalance decimal.Decimal `json:"pending_balance"`
	Currency       string          `json:"currency"`
	Status         string          `json:"status"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

type TransactionResponse struct {
	ID                  uuid.UUID       `json:"id"`
	WalletID            uuid.UUID       `json:"wallet_id"`
	Amount              decimal.Decimal `json:"amount"`
	Direction           string          `json:"direction"`
	Type                string          `json:"type"`
	Status              string          `json:"status"`
	ReferenceID         string          `json:"reference_id"`
	BalanceAfter        decimal.Decimal `json:"balance_after"`
	PendingBalanceAfter decimal.Decimal `json:"pending_balance_after"`
	Description         string          `json:"description"`
	CreatedAt           time.Time       `json:"created_at"`
}

type WithdrawRequest struct {
	Amount      decimal.Decimal `json:"amount"`
	Description string          `json:"description"`
}

func (r WithdrawRequest) Validate() error {
	if r.Amount.IsZero() || r.Amount.IsNegative() {
		return errors.New("amount must be greater than 0")
	}
	return nil
}
