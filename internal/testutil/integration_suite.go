package testutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

var (
	sharedDB        *sqlx.DB
	sharedContainer *postgres.PostgresContainer
	setupOnce       sync.Once
	setupErr        error
)

type IntegrationSuite struct {
	suite.Suite
	DB        *sqlx.DB
	container *postgres.PostgresContainer
}

func (s *IntegrationSuite) SetupSuite() {
	setupOnce.Do(func() {
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
			testcontainers.CustomizeRequestOption(func(req *testcontainers.GenericContainerRequest) error {
				req.Reuse = true
				req.Name = "marketplace-integration-db"
				return nil
			}),
		)
		if err != nil {
			setupErr = err
			return
		}
		sharedContainer = postgresContainer

		// Get connection string
		connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
		if err != nil {
			setupErr = err
			return
		}

		// Connect to DB
		db, err := sqlx.Connect("pgx", connStr)
		if err != nil {
			setupErr = err
			return
		}
		sharedDB = db

		// Run migrations
		_, b, _, _ := runtime.Caller(0)
		basepath := filepath.Dir(b)
		migrationPath := filepath.Join(basepath, "..", "database", "migrations")

		m, err := migrate.New(
			fmt.Sprintf("file://%s", migrationPath),
			connStr,
		)
		if err != nil {
			setupErr = err
			return
		}
		defer m.Close()

		err = m.Up()
		if err != nil && err != migrate.ErrNoChange {
			setupErr = err
			return
		}
	})

	s.Require().NoError(setupErr)
	s.DB = sharedDB
	s.container = sharedContainer
}

func (s *IntegrationSuite) TearDownSuite() {
	// We no longer close the DB or terminate the container here to allow reuse across suites within the same process.
	// In a real-world scenario, you might use TestMain to clean up at the very end of the package tests.
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
