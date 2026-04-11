package services

import (
	"context"
	"testing"

	"go-shop-yourself/internal/domain"
	"go-shop-yourself/internal/dtos"
	"go-shop-yourself/internal/mocks"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetWalletByUserID_Success(t *testing.T) {
	mockRepo := mocks.NewWalletRepository(t)
	service := NewWalletService(mockRepo)

	userID := uuid.New()
	wallet := &domain.Wallet{
		ID:           uuid.New(),
		UserID:       userID,
		WalletNumber: "WAL-123",
		Balance:      decimal.NewFromInt(100),
		Status:       domain.WalletStatusActive,
	}

	mockRepo.On("GetWalletByUserID", mock.Anything, userID).Return(wallet, nil)

	res, err := service.GetWalletByUserID(context.Background(), userID)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "WAL-123", res.WalletNumber)
}

func TestWithdraw_Success(t *testing.T) {
	mockRepo := mocks.NewWalletRepository(t)
	service := NewWalletService(mockRepo)

	userID := uuid.New()
	wallet := &domain.Wallet{
		ID:      uuid.New(),
		Balance: decimal.NewFromInt(1000),
		Status:  domain.WalletStatusActive,
	}

	req := dtos.WithdrawRequest{
		Amount:      decimal.NewFromInt(500),
		Description: "Withdraw test",
	}

	mockRepo.On("GetWalletByUserID", mock.Anything, userID).Return(wallet, nil)
	mockRepo.On("Withdraw", mock.Anything, wallet.ID, req.Amount, mock.Anything).Return(nil)

	err := service.Withdraw(context.Background(), userID, req)

	assert.NoError(t, err)
}

func TestWithdraw_Fail_InsufficientBalance(t *testing.T) {
	mockRepo := mocks.NewWalletRepository(t)
	service := NewWalletService(mockRepo)

	userID := uuid.New()
	wallet := &domain.Wallet{
		Balance: decimal.NewFromInt(100),
		Status:  domain.WalletStatusActive,
	}

	req := dtos.WithdrawRequest{
		Amount: decimal.NewFromInt(500),
	}

	mockRepo.On("GetWalletByUserID", mock.Anything, userID).Return(wallet, nil)

	err := service.Withdraw(context.Background(), userID, req)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrInsufficientBalance, err)
}
