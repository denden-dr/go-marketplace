package testutil

import (
	"context"
	"go-marketplace/internal/database"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"
)

type IntegrationSuite struct {
	suite.Suite
	DB *pgxpool.Pool
}

func (s *IntegrationSuite) SetupSuite() {
	pool, err := database.ConnectDB()
	s.Require().NoError(err)
	s.DB = pool
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
		"refresh_tokens",
		"user_addresses",
		"users",
	}

	for _, table := range tables {
		_, err := s.DB.Exec(context.Background(), "TRUNCATE TABLE "+table+" CASCADE")
		s.Require().NoError(err)
	}
}
