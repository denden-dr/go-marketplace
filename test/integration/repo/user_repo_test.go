//go:build integration

package repo

import (
	"context"
	"testing"
	"time"

	"go-marketplace/internal/common"
	"go-marketplace/internal/core/user"
	"go-marketplace/internal/domain"
	"go-marketplace/internal/testutil"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type UserRepoSuite struct {
	testutil.IntegrationSuite
	repo user.UserRepository
}

func (s *UserRepoSuite) SetupSuite() {
	s.IntegrationSuite.SetupSuite()
	s.repo = user.NewUserRepository(s.DB)
}

func (s *UserRepoSuite) TestCreateUser() {
	u := &domain.User{
		ID:           uuid.New(),
		FullName:     "Test User",
		Username:     "testuser",
		Email:        "test@example.com",
		Password:     common.Ptr("hashed_password"),
		AuthProvider: domain.AuthProviderLocal,
		Role:         domain.RoleUser,
		CreatedAt:    time.Now().Truncate(time.Microsecond),
	}

	err := s.repo.CreateUser(context.Background(), u)
	s.NoError(err)

	// Verify
	dbUser, err := s.repo.GetUserByID(context.Background(), u.ID)
	s.NoError(err)
	s.NotNil(dbUser)
	s.Equal(u.ID, dbUser.ID)
	s.Equal(u.FullName, dbUser.FullName)
	s.Equal(u.Username, dbUser.Username)
	s.Equal(u.Email, dbUser.Email)
	s.Equal(*u.Password, *dbUser.Password)
	s.Equal(u.AuthProvider, dbUser.AuthProvider)
	s.True(u.CreatedAt.Equal(dbUser.CreatedAt))
}

func (s *UserRepoSuite) TestGetUserByEmail() {
	u := &domain.User{
		ID:           uuid.New(),
		FullName:     "Email User",
		Username:     "emailuser",
		Email:        "email@example.com",
		AuthProvider: domain.AuthProviderLocal,
		Role:         domain.RoleUser,
		CreatedAt:    time.Now().Truncate(time.Microsecond),
	}
	s.NoError(s.repo.CreateUser(context.Background(), u))

	dbUser, err := s.repo.GetUserByEmail(context.Background(), u.Email)
	s.NoError(err)
	s.NotNil(dbUser)
	s.Equal(u.ID, dbUser.ID)

	// Non-existent
	dbUser, err = s.repo.GetUserByEmail(context.Background(), "nonexistent@example.com")
	s.NoError(err)
	s.Nil(dbUser)
}

func (s *UserRepoSuite) TestGetUserByProviderID() {
	providerID := "provider_123"
	u := &domain.User{
		ID:           uuid.New(),
		FullName:     "OAuth User",
		Username:     "oauthuser",
		Email:        "oauth@example.com",
		AuthProvider: domain.AuthProviderGoogle,
		ProviderID:   &providerID,
		Role:         domain.RoleUser,
		CreatedAt:    time.Now().Truncate(time.Microsecond),
	}
	s.NoError(s.repo.CreateUser(context.Background(), u))

	dbUser, err := s.repo.GetUserByProviderID(context.Background(), domain.AuthProviderGoogle, providerID)
	s.NoError(err)
	s.NotNil(dbUser)
	s.Equal(u.ID, dbUser.ID)
	s.Equal(providerID, *dbUser.ProviderID)
}

func (s *UserRepoSuite) TestGetUserByUsername() {
	u := &domain.User{
		ID:           uuid.New(),
		FullName:     "Username User",
		Username:     "user123",
		Email:        "user123@example.com",
		AuthProvider: domain.AuthProviderLocal,
		Role:         domain.RoleUser,
		CreatedAt:    time.Now().Truncate(time.Microsecond),
	}
	s.NoError(s.repo.CreateUser(context.Background(), u))

	dbUser, err := s.repo.GetUserByUsername(context.Background(), "user123")
	s.NoError(err)
	s.NotNil(dbUser)
	s.Equal(u.ID, dbUser.ID)
}

func (s *UserRepoSuite) TestAddressOperations() {
	userID := uuid.New()
	u := &domain.User{
		ID:           userID,
		FullName:     "Address User",
		Username:     "addruser",
		Email:        "addr@example.com",
		AuthProvider: domain.AuthProviderLocal,
		Role:         domain.RoleUser,
		CreatedAt:    time.Now().Truncate(time.Microsecond),
	}
	s.NoError(s.repo.CreateUser(context.Background(), u))

	addr := &domain.UserAddress{
		ID:            uuid.New(),
		UserID:        userID,
		Tag:           domain.AddressTagHome,
		RecipientName: "John Doe",
		PhoneNumber:   "123456789",
		StreetAddress: "123 Main St",
		City:          "Springfield",
		Province:      "State",
		PostalCode:    "12345",
		IsDefault:     true,
		CreatedAt:     time.Now().Truncate(time.Microsecond),
		UpdatedAt:     time.Now().Truncate(time.Microsecond),
	}

	// 1. CreateAddress
	err := s.repo.CreateAddress(context.Background(), addr)
	s.NoError(err)

	// 2. GetAddressByID
	dbAddr, err := s.repo.GetAddressByID(context.Background(), addr.ID)
	s.NoError(err)
	s.NotNil(dbAddr)
	s.Equal(addr.ID, dbAddr.ID)
	s.Equal(addr.UserID, dbAddr.UserID)
	s.Equal(addr.Tag, dbAddr.Tag)
	s.Equal(addr.RecipientName, dbAddr.RecipientName)
	s.Equal(addr.IsDefault, dbAddr.IsDefault)

	// 3. GetAddressesByUserID
	addresses, err := s.repo.GetAddressesByUserID(context.Background(), userID)
	s.NoError(err)
	s.Len(addresses, 1)
	s.Equal(addr.ID, addresses[0].ID)

	// 4. UpdateAddress
	addr.RecipientName = "Jane Doe"
	addr.Tag = domain.AddressTagWork
	addr.UpdatedAt = time.Now().Truncate(time.Microsecond)
	err = s.repo.UpdateAddress(context.Background(), addr)
	s.NoError(err)

	dbAddr, _ = s.repo.GetAddressByID(context.Background(), addr.ID)
	s.Equal("Jane Doe", dbAddr.RecipientName)
	s.Equal(domain.AddressTagWork, dbAddr.Tag)

	// 5. UnsetDefaultAddresses
	err = s.repo.UnsetDefaultAddresses(context.Background(), userID)
	s.NoError(err)

	dbAddr, _ = s.repo.GetAddressByID(context.Background(), addr.ID)
	s.False(dbAddr.IsDefault)

	// 6. DeleteAddress
	err = s.repo.DeleteAddress(context.Background(), addr.ID)
	s.NoError(err)

	dbAddr, err = s.repo.GetAddressByID(context.Background(), addr.ID)
	s.NoError(err)
	s.Nil(dbAddr)
}

func (s *UserRepoSuite) TestUpdateRoleTx() {
	u := &domain.User{
		ID:           uuid.New(),
		FullName:     "Role User",
		Username:     "roleuser",
		Email:        "role@example.com",
		AuthProvider: domain.AuthProviderLocal,
		Role:         domain.RoleUser,
		CreatedAt:    time.Now().Truncate(time.Microsecond),
	}
	s.NoError(s.repo.CreateUser(context.Background(), u))

	// Start transaction
	tx, err := s.DB.BeginTxx(context.Background(), nil)
	s.NoError(err)

	err = s.repo.UpdateRoleTx(context.Background(), tx, u.ID, domain.RoleMerchant)
	s.NoError(err)

	s.NoError(tx.Commit())

	// Verify
	dbUser, err := s.repo.GetUserByID(context.Background(), u.ID)
	s.NoError(err)
	s.Equal(domain.RoleMerchant, dbUser.Role)
}

func TestUserRepoSuite(t *testing.T) {
	suite.Run(t, new(UserRepoSuite))
}
