//go:build integration

package repo

import (
	"context"
	"testing"
	"time"

	"go-marketplace/internal/core/merchant"
	"go-marketplace/internal/core/user"
	"go-marketplace/internal/domain"
	"go-marketplace/internal/testutil"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type MerchantRepoSuite struct {
	testutil.IntegrationSuite
	repo     merchant.MerchantRepository
	userRepo user.UserRepository
}

func (s *MerchantRepoSuite) SetupSuite() {
	s.IntegrationSuite.SetupSuite()
	s.repo = merchant.NewMerchantRepository(s.DB)
	s.userRepo = user.NewUserRepository(s.DB)
}

func (s *MerchantRepoSuite) TestCreateAndGet() {
	u := &domain.User{
		ID:           uuid.New(),
		FullName:     "Merchant Owner",
		Username:     "owner",
		Email:        "owner@example.com",
		AuthProvider: domain.AuthProviderLocal,
		CreatedAt:    time.Now().Truncate(time.Microsecond),
	}
	s.NoError(s.userRepo.CreateUser(context.Background(), u))

	m := &domain.Merchant{
		ID:        uuid.New(),
		UserID:    u.ID,
		Name:      "Test Shop",
		About:     "Best shop in town",
		TaxID:     "123-456",
		CreatedAt: time.Now().Truncate(time.Microsecond),
	}

	// Create
	err := s.repo.Create(context.Background(), m)
	s.NoError(err)

	// GetByID
	dbMerchant, err := s.repo.GetByID(context.Background(), m.ID)
	s.NoError(err)
	s.NotNil(dbMerchant)
	s.Equal(m.ID, dbMerchant.ID)
	s.Equal(u.ID, dbMerchant.UserID)
	s.Equal(m.Name, dbMerchant.Name)
	s.Equal(m.About, dbMerchant.About)
	s.Equal(m.TaxID, dbMerchant.TaxID)
	s.True(m.CreatedAt.Equal(dbMerchant.CreatedAt))

	// GetByUserID
	dbMerchant, err = s.repo.GetByUserID(context.Background(), u.ID)
	s.NoError(err)
	s.NotNil(dbMerchant)
	s.Equal(m.ID, dbMerchant.ID)
}

func (s *MerchantRepoSuite) TestCreateTx() {
	u := &domain.User{
		ID:           uuid.New(),
		FullName:     "Tx Owner",
		Username:     "txowner",
		Email:        "tx@example.com",
		AuthProvider: domain.AuthProviderLocal,
		CreatedAt:    time.Now().Truncate(time.Microsecond),
	}
	s.NoError(s.userRepo.CreateUser(context.Background(), u))

	m := &domain.Merchant{
		ID:        uuid.New(),
		UserID:    u.ID,
		Name:      "Tx Shop",
		About:     "Tx shop",
		TaxID:     "999",
		CreatedAt: time.Now().Truncate(time.Microsecond),
	}

	tx, err := s.DB.BeginTxx(context.Background(), nil)
	s.NoError(err)
	defer tx.Rollback()

	err = s.repo.CreateTx(context.Background(), tx, m)
	s.NoError(err)

	s.NoError(tx.Commit())

	dbMerchant, err := s.repo.GetByID(context.Background(), m.ID)
	s.NoError(err)
	s.NotNil(dbMerchant)
}

func (s *MerchantRepoSuite) TestGetPool() {
	pool := s.repo.GetPool()
	s.NotNil(pool)
	s.Equal(s.DB, pool)
}

func TestMerchantRepoSuite(t *testing.T) {
	suite.Run(t, new(MerchantRepoSuite))
}
