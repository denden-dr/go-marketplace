package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"go-shop-yourself/internal/domain"
	"go-shop-yourself/internal/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestRegister_Success(t *testing.T) {
	mockRepo := mocks.NewUserRepository(t)
	mockRTRepo := mocks.NewRefreshTokenRepository(t)
	service := NewAuthService(mockRepo, mockRTRepo, "secret")

	email := "test@example.com"
	password := "password123"
	username := "testuser"

	mockRepo.On("GetUserByEmail", mock.Anything, email).Return(nil, nil)
	mockRepo.On("CreateUser", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
		return u.Email == email && u.Username == username
	})).Return(nil)

	userID, err := service.Register(context.Background(), email, password, username)

	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, userID)
}

func TestRegister_Fail_UserAlreadyExists(t *testing.T) {
	mockRepo := mocks.NewUserRepository(t)
	mockRTRepo := mocks.NewRefreshTokenRepository(t)
	service := NewAuthService(mockRepo, mockRTRepo, "secret")

	email := "test@example.com"

	mockRepo.On("GetUserByEmail", mock.Anything, email).Return(&domain.User{}, nil)

	_, err := service.Register(context.Background(), email, "pass", "user")

	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrUserAlreadyExists))
}

func TestLogin_Success(t *testing.T) {
	mockRepo := mocks.NewUserRepository(t)
	mockRTRepo := mocks.NewRefreshTokenRepository(t)
	service := NewAuthService(mockRepo, mockRTRepo, "secret")

	email := "test@example.com"
	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := &domain.User{
		ID:       uuid.New(),
		Email:    email,
		Password: string(hashedPassword),
	}

	mockRepo.On("GetUserByEmail", mock.Anything, email).Return(user, nil)
	mockRTRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	res, err := service.Login(context.Background(), email, password)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, user.ID, res.ID)
	assert.NotEmpty(t, res.AccessToken)
	assert.NotEmpty(t, res.RefreshToken)
}

func TestLogin_Fail_UserNotFound(t *testing.T) {
	mockRepo := mocks.NewUserRepository(t)
	mockRTRepo := mocks.NewRefreshTokenRepository(t)
	service := NewAuthService(mockRepo, mockRTRepo, "secret")

	mockRepo.On("GetUserByEmail", mock.Anything, "notfound@email.com").Return(nil, nil)

	_, err := service.Login(context.Background(), "notfound@email.com", "anypass")

	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInvalidCredentials))
}

func TestRefreshTokens_Success(t *testing.T) {
	mockRepo := mocks.NewUserRepository(t)
	mockRTRepo := mocks.NewRefreshTokenRepository(t)
	service := NewAuthService(mockRepo, mockRTRepo, "secret")

	rawToken := "old-refresh-token"
	rt := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		FamilyID:  uuid.New(),
		IsRevoked: false,
		ExpiresAt: time.Now().Add(time.Hour),
	}

	mockRTRepo.On("GetByTokenHash", mock.Anything, mock.Anything).Return(rt, nil)
	mockRTRepo.On("RevokeByID", mock.Anything, rt.ID).Return(nil)
	mockRTRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	res, err := service.RefreshTokens(context.Background(), rawToken)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotEmpty(t, res.AccessToken)
}

func TestRefreshTokens_Fail_TokenReused(t *testing.T) {
	mockRepo := mocks.NewUserRepository(t)
	mockRTRepo := mocks.NewRefreshTokenRepository(t)
	service := NewAuthService(mockRepo, mockRTRepo, "secret")

	rt := &domain.RefreshToken{
		FamilyID:  uuid.New(),
		IsRevoked: true,
	}

	mockRTRepo.On("GetByTokenHash", mock.Anything, mock.Anything).Return(rt, nil)
	mockRTRepo.On("RevokeAllByFamilyID", mock.Anything, rt.FamilyID).Return(nil)

	_, err := service.RefreshTokens(context.Background(), "reused-token")

	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrRefreshTokenReused))
}

func TestLogout_Success(t *testing.T) {
	mockRTRepo := mocks.NewRefreshTokenRepository(t)
	service := NewAuthService(nil, mockRTRepo, "secret")

	token := "valid-token"
	rt := &domain.RefreshToken{
		FamilyID: uuid.New(),
	}

	mockRTRepo.On("GetByTokenHash", mock.Anything, mock.Anything).Return(rt, nil)
	mockRTRepo.On("RevokeAllByFamilyID", mock.Anything, rt.FamilyID).Return(nil)

	err := service.Logout(context.Background(), token)

	assert.NoError(t, err)
}
