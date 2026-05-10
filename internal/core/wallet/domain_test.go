package wallet

import (
	"go-marketplace/internal/domain"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestWallet_CanWithdraw(t *testing.T) {
	tests := []struct {
		name    string
		status  domain.WalletStatus
		balance decimal.Decimal
		amount  decimal.Decimal
		wantErr error
	}{
		{"active with enough balance", domain.WalletStatusActive, decimal.NewFromInt(100), decimal.NewFromInt(50), nil},
		{"active with exact balance", domain.WalletStatusActive, decimal.NewFromInt(100), decimal.NewFromInt(100), nil},
		{"inactive wallet", domain.WalletStatusFrozen, decimal.NewFromInt(100), decimal.NewFromInt(50), domain.ErrWalletNotActive},
		{"insufficient balance", domain.WalletStatusActive, decimal.NewFromInt(100), decimal.NewFromInt(101), domain.ErrInsufficientBalance},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := NewWallet(&domain.Wallet{Status: tt.status, Balance: tt.balance})
			assert.Equal(t, tt.wantErr, w.CanWithdraw(tt.amount))
		})
	}
}

func TestGenerateWalletNumber(t *testing.T) {
	n := GenerateWalletNumber()
	assert.Contains(t, n, "WAL-")
	assert.Len(t, n, 12) // WAL- + 8 chars
}
