//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"go-marketplace/internal/core/user"
	"go-marketplace/internal/core/wallet"
	"go-marketplace/internal/domain"
	"go-marketplace/internal/testutil"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
)

type WalletRepoSuite struct {
	testutil.IntegrationSuite
	repo     wallet.WalletRepository
	userRepo user.UserRepository
}

func (s *WalletRepoSuite) SetupSuite() {
	s.IntegrationSuite.SetupSuite()
	s.repo = wallet.NewWalletRepository(s.DB)
	s.userRepo = user.NewUserRepository(s.DB)
}

func (s *WalletRepoSuite) TestWalletOperations() {
	// Setup User
	u := &domain.User{
		ID:           uuid.New(),
		FullName:     "Wallet User",
		Username:     "walletuser",
		Email:        "wallet@example.com",
		AuthProvider: domain.AuthProviderLocal,
		CreatedAt:    time.Now().Truncate(time.Microsecond),
	}
	s.NoError(s.userRepo.CreateUser(context.Background(), u))

	w := &domain.Wallet{
		ID:           uuid.New(),
		UserID:       u.ID,
		WalletNumber: "1234567890",
		Balance:      decimal.NewFromInt(1000),
		Currency:     "IDR",
		Status:       domain.WalletStatusActive,
		CreatedAt:    time.Now().Truncate(time.Microsecond),
		UpdatedAt:    time.Now().Truncate(time.Microsecond),
	}

	// 1. Create
	err := s.repo.Create(context.Background(), w)
	s.NoError(err)

	// 2. GetWalletByUserID
	dbWallet, err := s.repo.GetWalletByUserID(context.Background(), u.ID)
	s.NoError(err)
	s.NotNil(dbWallet)
	s.Equal(w.ID, dbWallet.ID)
	s.True(w.Balance.Equal(dbWallet.Balance))

	// 3. Withdraw
	txData := domain.WalletTransaction{
		ID:        uuid.New(),
		WalletID:  w.ID,
		Amount:    decimal.NewFromInt(100),
		Direction: domain.TransactionDirectionOut,
		Type:      domain.TransactionTypeWithdraw,
		Status:    domain.TransactionStatusSuccess,
		CreatedAt: time.Now().Truncate(time.Microsecond),
	}
	err = s.repo.Withdraw(context.Background(), w.ID, decimal.NewFromInt(100), txData)
	s.NoError(err)

	dbWallet, _ = s.repo.GetWalletByUserID(context.Background(), u.ID)
	s.True(decimal.NewFromInt(900).Equal(dbWallet.Balance))

	// 4. GetWalletHistory
	history, err := s.repo.GetWalletHistory(context.Background(), w.ID, 10, 0)
	s.NoError(err)
	s.Len(history, 1)
	s.Equal(txData.ID, history[0].ID)

	// 5. TX Operations
	tx, err := s.DB.BeginTxx(context.Background(), nil)
	s.NoError(err)
	defer tx.Rollback()

	txData2 := domain.WalletTransaction{
		ID:        uuid.New(),
		WalletID:  w.ID,
		Amount:    decimal.NewFromInt(50),
		Direction: domain.TransactionDirectionOut,
		Type:      domain.TransactionTypePayment,
		Status:    domain.TransactionStatusSuccess,
		CreatedAt: time.Now().Truncate(time.Microsecond),
	}
	err = s.repo.DeductBalanceTX(context.Background(), tx, w.ID, decimal.NewFromInt(50), txData2)
	s.NoError(err)

	txData3 := domain.WalletTransaction{
		ID:        uuid.New(),
		WalletID:  w.ID,
		Amount:    decimal.NewFromInt(200),
		Direction: domain.TransactionDirectionIn,
		Type:      domain.TransactionTypeTopup,
		Status:    domain.TransactionStatusSuccess,
		CreatedAt: time.Now().Truncate(time.Microsecond),
	}
	err = s.repo.AddBalanceTX(context.Background(), tx, w.ID, decimal.NewFromInt(200), txData3)
	s.NoError(err)

	s.NoError(tx.Commit())

	dbWallet, _ = s.repo.GetWalletByUserID(context.Background(), u.ID)
	// 900 - 50 + 200 = 1050
	s.True(decimal.NewFromInt(1050).Equal(dbWallet.Balance))
}

func TestWalletRepoSuite(t *testing.T) {
	suite.Run(t, new(WalletRepoSuite))
}
