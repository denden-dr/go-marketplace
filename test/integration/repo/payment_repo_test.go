//go:build integration

package repo

import (
	"context"
	"go-marketplace/internal/core/payment"
	"go-marketplace/internal/domain"
	"testing"
	"time"

	"go-marketplace/internal/testutil"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
)

type PaymentRepoTestSuite struct {
	testutil.IntegrationSuite
	repo payment.PaymentRepository
}

func TestPaymentRepoTestSuite(t *testing.T) {
	suite.Run(t, new(PaymentRepoTestSuite))
}

func (s *PaymentRepoTestSuite) SetupTest() {
	s.IntegrationSuite.SetupTest()
	s.repo = payment.NewPaymentRepository(s.DB)
}

func (s *PaymentRepoTestSuite) TestCreateAndGetPayment() {
	ctx := context.Background()
	userID := uuid.New()
	
	// Create a dummy user first due to FK
	_, err := s.DB.ExecContext(ctx, "INSERT INTO users (id, username, email) VALUES ($1, $2, $3)", userID, "payuser", "pay@example.com")
	s.Require().NoError(err)

	p := &domain.Payment{
		ID:            uuid.New(),
		UserID:        userID,
		Amount:        decimal.NewFromInt(1000),
		Type:          domain.PaymentTypeOrder,
		Method:        domain.PaymentMethodWallet,
		Status:        domain.PaymentStatusPending,
		ReferenceID:   uuid.New(),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	tx, err := s.DB.BeginTxx(ctx, nil)
	s.Require().NoError(err)
	
	err = s.repo.CreateTX(ctx, tx, p)
	s.Assert().NoError(err)
	
	err = tx.Commit()
	s.Require().NoError(err)

	// Get by ID
	res, err := s.repo.GetByID(ctx, p.ID)
	s.Assert().NoError(err)
	s.Assert().NotNil(res)
	s.Assert().Equal(p.ID, res.ID)
	s.Assert().Equal(p.Amount.String(), res.Amount.String())
}

func (s *PaymentRepoTestSuite) TestUpdateStatus() {
	ctx := context.Background()
	userID := uuid.New()
	_, _ = s.DB.ExecContext(ctx, "INSERT INTO users (id, username, email) VALUES ($1, $2, $3)", userID, "payuser2", "pay2@example.com")

	p := &domain.Payment{
		ID:            uuid.New(),
		UserID:        userID,
		Amount:        decimal.NewFromInt(500),
		Type:          domain.PaymentTypeTopup,
		Method:        domain.PaymentMethodMidtrans,
		Status:        domain.PaymentStatusPending,
		ReferenceID:   uuid.New(),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	tx, _ := s.DB.BeginTxx(ctx, nil)
	_ = s.repo.CreateTX(ctx, tx, p)
	_ = tx.Commit()

	// Update Status
	tx2, _ := s.DB.BeginTxx(ctx, nil)
	extID := "EXT-123"
	err := s.repo.UpdateStatusTX(ctx, tx2, p.ID, domain.PaymentStatusSuccess, &extID)
	s.Assert().NoError(err)
	_ = tx2.Commit()

	res, _ := s.repo.GetByID(ctx, p.ID)
	s.Assert().Equal(domain.PaymentStatusSuccess, res.Status)
	s.Assert().NotNil(res.ExternalID)
	s.Assert().Equal(extID, *res.ExternalID)
}

func (s *PaymentRepoTestSuite) TestDistributions() {
	ctx := context.Background()
	userID := uuid.New()
	merchantID := uuid.New()
	_, _ = s.DB.ExecContext(ctx, "INSERT INTO users (id, username, email) VALUES ($1, $2, $3)", userID, "payuser3", "pay3@example.com")
	_, _ = s.DB.ExecContext(ctx, "INSERT INTO users (id, username, email) VALUES ($1, $2, $3)", merchantID, "merchuser", "merch@example.com")

	pID := uuid.New()
	p := &domain.Payment{
		ID:            pID,
		UserID:        userID,
		Amount:        decimal.NewFromInt(100),
		Type:          domain.PaymentTypeOrder,
		Method:        domain.PaymentMethodWallet,
		Status:        domain.PaymentStatusSuccess,
		ReferenceID:   uuid.New(),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	tx, _ := s.DB.BeginTxx(ctx, nil)
	_ = s.repo.CreateTX(ctx, tx, p)

	dist := &domain.PaymentDistribution{
		ID:          uuid.New(),
		PaymentID:   pID,
		RecipientID: merchantID,
		Amount:      decimal.NewFromInt(100),
		CreatedAt:   time.Now(),
	}
	err := s.repo.CreateDistributionTX(ctx, tx, dist)
	s.Assert().NoError(err)
	_ = tx.Commit()

	dists, err := s.repo.GetDistributionsByPaymentID(ctx, pID)
	s.Assert().NoError(err)
	s.Assert().Len(dists, 1)
	s.Assert().Equal(dist.Amount.String(), dists[0].Amount.String())
}
