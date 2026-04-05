package services

import (
	"context"
	"testing"

	"github.com/denden-dr/go-shop-yourself/internal/domain"
	"github.com/denden-dr/go-shop-yourself/internal/dto"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// MockUserRepository is a mock implementation of UserRepository.
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(ctx context.Context, u *domain.User, auth *domain.UserAuth, profile *domain.UserProfile) error {
	args := m.Called(ctx, u, auth, profile)
	return args.Error(0)
}

func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (*domain.UserAuth, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserAuth), args.Error(1)
}

func (m *MockUserRepository) GetUserProfileByID(ctx context.Context, userID uuid.UUID) (*domain.UserProfile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserProfile), args.Error(1)
}

func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func TestRegister(t *testing.T) {
	repo := new(MockUserRepository)
	service := NewAuthService(repo, "secret")
	ctx := context.Background()

	req := dto.RegisterRequest{
		Email:    "test@example.com",
		Username: "testuser",
		Password: "password123",
	}

	repo.On("GetUserByEmail", ctx, req.Email).Return(nil, nil)
	repo.On("CreateUser", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	res, err := service.Register(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotEmpty(t, res.Token)
	repo.AssertExpectations(t)
}

func TestLogin(t *testing.T) {
	repo := new(MockUserRepository)
	service := NewAuthService(repo, "secret")
	ctx := context.Background()

	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)

	auth := &domain.UserAuth{
		UserID:       uuid.New(),
		Email:        "test@example.com",
		PasswordHash: string(hashedPassword),
	}

	req := dto.LoginRequest{
		Email:    "test@example.com",
		Password: password,
	}

	repo.On("GetUserByEmail", ctx, req.Email).Return(auth, nil)
	repo.On("UpdateLastLogin", ctx, auth.UserID).Return(nil)

	res, err := service.Login(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, auth.UserID, res.UserID)
	assert.NotEmpty(t, res.Token)
	repo.AssertExpectations(t)
}
