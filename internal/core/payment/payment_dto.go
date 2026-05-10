package payment

import (
	"errors"
	"go-marketplace/internal/domain"

	"github.com/shopspring/decimal"
)

type TopupRequest struct {
	Amount decimal.Decimal      `json:"amount"`
	Method domain.PaymentMethod `json:"method"`
}

func (r TopupRequest) Validate() error {
	if r.Amount.IsZero() || r.Amount.IsNegative() {
		return errors.New("amount must be greater than 0")
	}
	if r.Method == "" {
		return errors.New("payment method is required")
	}
	if r.Method == domain.PaymentMethodWallet {
		return errors.New("cannot top-up using wallet balance")
	}
	return nil
}

type MidtransWebhookRequest struct {
	TransactionStatus string `json:"transaction_status"`
	OrderID           string `json:"order_id"`
	PaymentType       string `json:"payment_type"`
	GrossAmount       string `json:"gross_amount"`
	SignatureKey      string `json:"signature_key"`
}
