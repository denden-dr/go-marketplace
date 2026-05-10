package testutil

import (
	"context"
	"go-marketplace/internal/config"
	"go-marketplace/internal/database"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"
)

type IntegrationSuite struct {
	suite.Suite
	DB *sqlx.DB
}

func (s *IntegrationSuite) SetupSuite() {
	cfg := config.Load()
	db, err := database.ConnectDB(cfg.DB)
	s.Require().NoError(err)
	s.DB = db
}

func (s *IntegrationSuite) TearDownSuite() {
	if s.DB != nil {
		s.DB.Close()
	}
}

func (s *IntegrationSuite) SetupTest() {
	s.TruncateTables()
}

func (s *IntegrationSuite) TruncateTables() {
	tables := []string{
		"cancellation_appeals",
		"cart_items",
		"order_items",
		"orders",
		"order_payments",
		"wallets_transaction",
		"wallets",
		"products",
		"merchants",
		"sessions",
		"verification_codes",
		"user_addresses",
		"users",
	}

	for _, table := range tables {
		_, err := s.DB.ExecContext(context.Background(), "TRUNCATE TABLE "+table+" CASCADE")
		s.Require().NoError(err)
	}
}
