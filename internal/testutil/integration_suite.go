package testutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type IntegrationSuite struct {
	suite.Suite
	DB        *sqlx.DB
	container *postgres.PostgresContainer
}

func (s *IntegrationSuite) SetupSuite() {
	ctx := context.Background()

	// Start Postgres container
	dbName := "marketplace_test"
	dbUser := "postgres"
	dbPassword := "postgres"

	postgresContainer, err := postgres.Run(ctx,
		"postgres:18-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	s.Require().NoError(err)
	s.container = postgresContainer

	// Get connection string
	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	s.Require().NoError(err)

	// Connect to DB
	db, err := sqlx.Connect("pgx", connStr)
	s.Require().NoError(err)
	s.DB = db

	// Run migrations
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	migrationPath := filepath.Join(basepath, "..", "database", "migrations")

	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationPath),
		connStr,
	)
	s.Require().NoError(err)
	defer m.Close()

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		s.Require().NoError(err)
	}
}

func (s *IntegrationSuite) TearDownSuite() {
	ctx := context.Background()
	if s.DB != nil {
		s.DB.Close()
	}
	if s.container != nil {
		err := s.container.Terminate(ctx)
		s.Require().NoError(err)
	}
}

func (s *IntegrationSuite) SetupTest() {
	s.TruncateTables()
}

func (s *IntegrationSuite) Seed() {
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	seedPath := filepath.Join(basepath, "..", "database", "seed", "seed.sql")

	content, err := os.ReadFile(seedPath)
	s.Require().NoError(err)

	_, err = s.DB.ExecContext(context.Background(), string(content))
	s.Require().NoError(err)
}

func (s *IntegrationSuite) TruncateTables() {
	tables := []string{
		"cancellation_appeals",
		"cart_items",
		"order_items",
		"orders",
		"payment_distributions",
		"payments",
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
