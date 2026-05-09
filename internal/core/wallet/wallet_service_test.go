package wallet

import (
	"context"
	"testing"

	"go-marketplace/internal/domain"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestWalletService_GetWalletByUserID(t *testing.T) {
	userID := uuid.New()
	w := &domain.Wallet{
		ID:           uuid.New(),
		UserID:       userID,
		WalletNumber: "WAL-123",
		Balance:      decimal.NewFromInt(100),
		Status:       domain.WalletStatusActive,
	}

	tests := []struct {
		name      string
		userID    uuid.UUID
		mockSetup func(mr *MockWalletRepository)
		wantErr   bool
	}{
		{
			name:   "Success",
			userID: userID,
			mockSetup: func(mr *MockWalletRepository) {
				mr.On("GetWalletByUserID", mock.Anything, userID).Return(w, nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr := NewMockWalletRepository(t)
			tt.mockSetup(mr)

			service := NewWalletService(mr)
			res, err := service.GetWalletByUserID(context.Background(), tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
				assert.Equal(t, "WAL-123", res.WalletNumber)
			}
		})
	}
}

func TestWalletService_GetWalletHistory(t *testing.T) {
	userID := uuid.New()
	wID := uuid.New()
	w := &domain.Wallet{ID: wID, UserID: userID}
	txs := []domain.WalletTransaction{{ID: uuid.New(), Amount: decimal.NewFromInt(100)}}

	tests := []struct {
		name      string
		userID    uuid.UUID
		page      int
		limit     int
		mockSetup func(mr *MockWalletRepository)
		wantErr   bool
	}{
		{
			name:   "Success",
			userID: userID,
			page:   1,
			limit:  10,
			mockSetup: func(mr *MockWalletRepository) {
				mr.On("GetWalletByUserID", mock.Anything, userID).Return(w, nil)
				mr.On("GetWalletHistory", mock.Anything, wID, 10, 0).Return(txs, nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr := NewMockWalletRepository(t)
			tt.mockSetup(mr)

			service := NewWalletService(mr)
			res, err := service.GetWalletHistory(context.Background(), tt.userID, tt.page, tt.limit)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, res, 1)
			}
		})
	}
}

func TestWalletService_Withdraw(t *testing.T) {
	userID := uuid.New()
	wID := uuid.New()

	tests := []struct {
		name      string
		request   WithdrawRequest
		mockSetup func(mr *MockWalletRepository)
		wantErr   bool
		errType   error
	}{
		{
			name: "Success",
			request: WithdrawRequest{
				Amount:      decimal.NewFromInt(500),
				Description: "Withdraw test",
			},
			mockSetup: func(mr *MockWalletRepository) {
				w := &domain.Wallet{ID: wID, Balance: decimal.NewFromInt(1000), Status: domain.WalletStatusActive}
				mr.On("GetWalletByUserID", mock.Anything, userID).Return(w, nil)
				mr.On("Withdraw", mock.Anything, wID, decimal.NewFromInt(500), mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			name:    "Insufficient Balance",
			request: WithdrawRequest{Amount: decimal.NewFromInt(500)},
			mockSetup: func(mr *MockWalletRepository) {
				w := &domain.Wallet{Balance: decimal.NewFromInt(100), Status: domain.WalletStatusActive}
				mr.On("GetWalletByUserID", mock.Anything, userID).Return(w, nil)
			},
			wantErr: true,
			errType: domain.ErrInsufficientBalance,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr := NewMockWalletRepository(t)
			tt.mockSetup(mr)

			service := NewWalletService(mr)
			err := service.Withdraw(context.Background(), userID, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestWalletService_CreateWallet(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name      string
		mockSetup func(mr *MockWalletRepository)
		wantErr   bool
		errType   error
	}{
		{
			name: "Success",
			mockSetup: func(mr *MockWalletRepository) {
				mr.On("GetWalletByUserID", mock.Anything, userID).Return(nil, nil)
				mr.On("Create", mock.Anything, mock.MatchedBy(func(w *domain.Wallet) bool {
					return w.UserID == userID
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "Already Exists",
			mockSetup: func(mr *MockWalletRepository) {
				existing := &domain.Wallet{ID: uuid.New(), UserID: userID}
				mr.On("GetWalletByUserID", mock.Anything, userID).Return(existing, nil)
			},
			wantErr: true,
			errType: domain.ErrWalletAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr := NewMockWalletRepository(t)
			tt.mockSetup(mr)

			service := NewWalletService(mr)
			res, err := service.CreateWallet(context.Background(), userID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
			}
		})
	}
}
