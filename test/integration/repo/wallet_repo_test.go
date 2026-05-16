//go:build integration

package repo

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
		Role:         domain.RoleUser,
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

func (s *WalletRepoSuite) TestGetWalletsByUserIDs() {
	u1 := &domain.User{
		ID: uuid.New(), FullName: "U1", Username: "u1_batch", Email: "u1@example.com",
		AuthProvider: domain.AuthProviderLocal, Role: domain.RoleUser, CreatedAt: time.Now().Truncate(time.Microsecond),
	}
	u2 := &domain.User{
		ID: uuid.New(), FullName: "U2", Username: "u2_batch", Email: "u2@example.com",
		AuthProvider: domain.AuthProviderLocal, Role: domain.RoleUser, CreatedAt: time.Now().Truncate(time.Microsecond),
	}
	s.NoError(s.userRepo.CreateUser(context.Background(), u1))
	s.NoError(s.userRepo.CreateUser(context.Background(), u2))

	w1 := &domain.Wallet{
		ID: uuid.New(), UserID: u1.ID, WalletNumber: "W-BATCH-001",
		Balance: decimal.NewFromInt(100), PendingBalance: decimal.Zero,
		Currency: "IDR", Status: domain.WalletStatusActive,
		CreatedAt: time.Now().Truncate(time.Microsecond), UpdatedAt: time.Now().Truncate(time.Microsecond),
	}
	w2 := &domain.Wallet{
		ID: uuid.New(), UserID: u2.ID, WalletNumber: "W-BATCH-002",
		Balance: decimal.NewFromInt(200), PendingBalance: decimal.Zero,
		Currency: "IDR", Status: domain.WalletStatusActive,
		CreatedAt: time.Now().Truncate(time.Microsecond), UpdatedAt: time.Now().Truncate(time.Microsecond),
	}
	s.NoError(s.repo.Create(context.Background(), w1))
	s.NoError(s.repo.Create(context.Background(), w2))

	wallets, err := s.repo.GetWalletsByUserIDs(context.Background(), []uuid.UUID{u1.ID, u2.ID})
	s.NoError(err)
	s.Len(wallets, 2)
}

func (s *WalletRepoSuite) TestAddPendingBalancesBatchTX() {
	u1 := &domain.User{
		ID: uuid.New(), FullName: "U1", Username: "u1_pb", Email: "u1pb@example.com",
		AuthProvider: domain.AuthProviderLocal, Role: domain.RoleUser, CreatedAt: time.Now().Truncate(time.Microsecond),
	}
	u2 := &domain.User{
		ID: uuid.New(), FullName: "U2", Username: "u2_pb", Email: "u2pb@example.com",
		AuthProvider: domain.AuthProviderLocal, Role: domain.RoleUser, CreatedAt: time.Now().Truncate(time.Microsecond),
	}
	s.NoError(s.userRepo.CreateUser(context.Background(), u1))
	s.NoError(s.userRepo.CreateUser(context.Background(), u2))

	w1 := &domain.Wallet{
		ID: uuid.New(), UserID: u1.ID, WalletNumber: "W-PB-001",
		Balance: decimal.Zero, PendingBalance: decimal.Zero,
		Currency: "IDR", Status: domain.WalletStatusActive,
		CreatedAt: time.Now().Truncate(time.Microsecond), UpdatedAt: time.Now().Truncate(time.Microsecond),
	}
	w2 := &domain.Wallet{
		ID: uuid.New(), UserID: u2.ID, WalletNumber: "W-PB-002",
		Balance: decimal.Zero, PendingBalance: decimal.Zero,
		Currency: "IDR", Status: domain.WalletStatusActive,
		CreatedAt: time.Now().Truncate(time.Microsecond), UpdatedAt: time.Now().Truncate(time.Microsecond),
	}
	s.NoError(s.repo.Create(context.Background(), w1))
	s.NoError(s.repo.Create(context.Background(), w2))

	tx, err := s.DB.BeginTxx(context.Background(), nil)
	s.NoError(err)
	defer tx.Rollback()

	updates := []domain.WalletBalanceUpdate{
		{
			WalletID: w1.ID,
			Amount:   decimal.NewFromInt(100),
			Transaction: domain.WalletTransaction{
				ID:        uuid.New(),
				WalletID:  w1.ID,
				Amount:    decimal.NewFromInt(100),
				Direction: domain.TransactionDirectionIn,
				Type:      domain.TransactionTypePayment,
				Status:    domain.TransactionStatusSuccess,
				CreatedAt: time.Now(),
			},
		},
		{
			WalletID: w2.ID,
			Amount:   decimal.NewFromInt(200),
			Transaction: domain.WalletTransaction{
				ID:        uuid.New(),
				WalletID:  w2.ID,
				Amount:    decimal.NewFromInt(200),
				Direction: domain.TransactionDirectionIn,
				Type:      domain.TransactionTypePayment,
				Status:    domain.TransactionStatusSuccess,
				CreatedAt: time.Now(),
			},
		},
	}

	err = s.repo.AddPendingBalancesBatchTX(context.Background(), tx, updates)
	s.NoError(err)
	s.NoError(tx.Commit())

	got1, _ := s.repo.GetWalletByUserID(context.Background(), u1.ID)
	got2, _ := s.repo.GetWalletByUserID(context.Background(), u2.ID)
	s.True(decimal.NewFromInt(100).Equal(got1.PendingBalance))
	s.True(decimal.NewFromInt(200).Equal(got2.PendingBalance))
}

func TestWalletRepoSuite(t *testing.T) {
	suite.Run(t, new(WalletRepoSuite))
}
