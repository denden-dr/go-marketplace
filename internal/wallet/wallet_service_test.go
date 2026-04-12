package wallet

import (
	"context"
	"errors"
	"testing"

	"go-shop-yourself/internal/domain"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestWalletService_GetWalletByUserID_Success(t *testing.T) {
	mockRepo := NewMockWalletRepository(t)
	service := NewWalletService(mockRepo)

	userID := uuid.New()
	w := &domain.Wallet{
		ID:           uuid.New(),
		UserID:       userID,
		WalletNumber: "WAL-123",
		Balance:      decimal.NewFromInt(100),
		Status:       domain.WalletStatusActive,
	}

	mockRepo.On("GetWalletByUserID", mock.Anything, userID).Return(w, nil)

	res, err := service.GetWalletByUserID(context.Background(), userID)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "WAL-123", res.WalletNumber)
}

func TestWalletService_GetWalletHistory_Success(t *testing.T) {
	mockRepo := NewMockWalletRepository(t)
	service := NewWalletService(mockRepo)

	userID := uuid.New()
	w := &domain.Wallet{ID: uuid.New(), UserID: userID}
	txs := []domain.WalletTransaction{
		{ID: uuid.New(), Amount: decimal.NewFromInt(100)},
	}

	mockRepo.On("GetWalletByUserID", mock.Anything, userID).Return(w, nil)
	mockRepo.On("GetWalletHistory", mock.Anything, w.ID, 10, 0).Return(txs, nil)

	res, err := service.GetWalletHistory(context.Background(), userID, 1, 10)

	assert.NoError(t, err)
	assert.Len(t, res, 1)
}

func TestWalletService_Withdraw_Success(t *testing.T) {
	mockRepo := NewMockWalletRepository(t)
	service := NewWalletService(mockRepo)

	userID := uuid.New()
	w := &domain.Wallet{
		ID:      uuid.New(),
		Balance: decimal.NewFromInt(1000),
		Status:  domain.WalletStatusActive,
	}

	req := WithdrawRequest{
		Amount:      decimal.NewFromInt(500),
		Description: "Withdraw test",
	}

	mockRepo.On("GetWalletByUserID", mock.Anything, userID).Return(w, nil)
	mockRepo.On("Withdraw", mock.Anything, w.ID, req.Amount, mock.Anything).Return(nil)

	err := service.Withdraw(context.Background(), userID, req)

	assert.NoError(t, err)
}

func TestWalletService_Withdraw_Fail_InsufficientBalance(t *testing.T) {
	mockRepo := NewMockWalletRepository(t)
	service := NewWalletService(mockRepo)

	userID := uuid.New()
	w := &domain.Wallet{
		Balance: decimal.NewFromInt(100),
		Status:  domain.WalletStatusActive,
	}

	req := WithdrawRequest{
		Amount: decimal.NewFromInt(500),
	}

	mockRepo.On("GetWalletByUserID", mock.Anything, userID).Return(w, nil)

	err := service.Withdraw(context.Background(), userID, req)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrInsufficientBalance, err)
}

func TestWalletService_CreateWallet_Success(t *testing.T) {
	mockRepo := NewMockWalletRepository(t)
	service := NewWalletService(mockRepo)

	userID := uuid.New()
	mockRepo.On("GetWalletByUserID", mock.Anything, userID).Return(nil, nil)
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(w *domain.Wallet) bool {
		return w.UserID == userID
	})).Return(nil)

	res, err := service.CreateWallet(context.Background(), userID)

	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func TestWalletService_CreateWallet_Fail_AlreadyExists(t *testing.T) {
	mockRepo := NewMockWalletRepository(t)
	service := NewWalletService(mockRepo)

	userID := uuid.New()
	existing := &domain.Wallet{ID: uuid.New(), UserID: userID}
	mockRepo.On("GetWalletByUserID", mock.Anything, userID).Return(existing, nil)

	_, err := service.CreateWallet(context.Background(), userID)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWalletAlreadyExists))
}
