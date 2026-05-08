//go:build integration

package integration

import (
	"context"
	"go-marketplace/internal/testutil"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type DBTestSuite struct {
	testutil.IntegrationSuite
}

func TestDBTestSuite(t *testing.T) {
	suite.Run(t, new(DBTestSuite))
}

func (s *DBTestSuite) TestConnection() {
	var val int
	err := s.DB.QueryRow(context.Background(), "SELECT 1").Scan(&val)
	s.Assert().NoError(err)
	s.Assert().Equal(1, val)
}

func (s *DBTestSuite) TestTruncationIsolation_Part1() {
	// Insert a user
	id := uuid.New()
	_, err := s.DB.Exec(context.Background(), 
		"INSERT INTO users (id, username, email) VALUES ($1, $2, $3)", 
		id, "testuser", "test@example.com")
	s.Require().NoError(err)

	var count int
	s.DB.QueryRow(context.Background(), "SELECT count(*) FROM users").Scan(&count)
	s.Assert().Equal(1, count)
}

func (s *DBTestSuite) TestTruncationIsolation_Part2() {
	// Count should be 0 because of SetupTest truncation
	var count int
	s.DB.QueryRow(context.Background(), "SELECT count(*) FROM users").Scan(&count)
	s.Assert().Equal(0, count)
}
