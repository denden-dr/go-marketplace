package wallet

import (
	"go-marketplace/internal/domain"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Wallet represents a rich domain entity for a wallet
type Wallet struct {
	model *domain.Wallet
}

// NewWallet creates a new rich Wallet entity
func NewWallet(m *domain.Wallet) *Wallet {
	return &Wallet{model: m}
}

// CanWithdraw checks if the wallet is active and has sufficient balance
func (w *Wallet) CanWithdraw(amount decimal.Decimal) error {
	if w.model.Status != domain.WalletStatusActive {
		return domain.ErrWalletNotActive
	}
	if w.model.Balance.LessThan(amount) {
		return domain.ErrInsufficientBalance
	}
	return nil
}

// GenerateWalletNumber generates a new wallet number
func GenerateWalletNumber() string {
	return "WAL-" + uuid.New().String()[:8]
}
