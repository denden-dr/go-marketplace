package payment

import (
	"go-marketplace/internal/domain"

	"github.com/shopspring/decimal"
)

type TopupRequest struct {
	Amount decimal.Decimal      `json:"amount"`
	Method domain.PaymentMethod `json:"method"`
}

func (r TopupRequest) Validate() error {
	errs := make(domain.ValidationErrors)
	if r.Amount.IsZero() || r.Amount.IsNegative() {
		errs["amount"] = "must be greater than 0"
	}
	switch r.Method {
	case "":
		errs["method"] = "is required"
	case domain.PaymentMethodWallet:
		errs["method"] = "cannot top-up using wallet balance"
	}

	if len(errs) > 0 {
		return errs
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
