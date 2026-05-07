package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"go-marketplace/internal/core/user"
	"go-marketplace/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthService_Register_Success(t *testing.T) {
	mockRepo := user.NewMockUserRepository(t)
	mockRTRepo := NewMockRefreshTokenRepository(t)
	mockSocialAuthClient := NewMockSupabaseAuthClient(t)
	service := NewAuthService(mockRepo, mockRTRepo, mockSocialAuthClient, "secret")

	email := "test@example.com"
	password := "password123"
	username := "testuser"

	mockRepo.On("GetUserByEmail", mock.Anything, email).Return(nil, nil)
	mockRepo.On("CreateUser", mock.Anything, mock.Anything).Return(nil)
	mockRTRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	res, err := service.Register(context.Background(), "Test User", email, password, username)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotEqual(t, uuid.Nil, res.ID)
	assert.Equal(t, "Test User", res.FullName)
}

func TestAuthService_Register_Fail_UserAlreadyExists(t *testing.T) {
	mockRepo := user.NewMockUserRepository(t)
	mockRTRepo := NewMockRefreshTokenRepository(t)
	mockSocialAuthClient := NewMockSupabaseAuthClient(t)
	service := NewAuthService(mockRepo, mockRTRepo, mockSocialAuthClient, "secret")

	email := "test@example.com"

	mockRepo.On("GetUserByEmail", mock.Anything, email).Return(&domain.User{}, nil)

	_, err := service.Register(context.Background(), "Full Name", email, "pass", "user")

	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrUserAlreadyExists))
}

func TestAuthService_Login_Success(t *testing.T) {
	mockRepo := user.NewMockUserRepository(t)
	mockRTRepo := NewMockRefreshTokenRepository(t)
	mockSocialAuthClient := NewMockSupabaseAuthClient(t)
	service := NewAuthService(mockRepo, mockRTRepo, mockSocialAuthClient, "secret")

	email := "test@example.com"
	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	u := &domain.User{
		ID:           uuid.New(),
		FullName:     "Test User",
		Username:     "testuser",
		Email:        email,
		Password:     &password, // Using password as placeholder for hashed password for simplicity in test setup if needed, but bcrypt.Compare uses u.Password
		AuthProvider: domain.AuthProviderLocal,
	}
	u.Password = func() *string { s := string(hashedPassword); return &s }()

	mockRepo.On("GetUserByEmail", mock.Anything, email).Return(u, nil)
	mockRTRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	res, err := service.Login(context.Background(), email, password)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, u.ID, res.ID)
	assert.NotEmpty(t, res.AccessToken)
	assert.NotEmpty(t, res.RefreshToken)
}

func TestAuthService_Login_Fail_UserNotFound(t *testing.T) {
	mockRepo := user.NewMockUserRepository(t)
	mockRTRepo := NewMockRefreshTokenRepository(t)
	mockSocialAuthClient := NewMockSupabaseAuthClient(t)
	service := NewAuthService(mockRepo, mockRTRepo, mockSocialAuthClient, "secret")

	mockRepo.On("GetUserByEmail", mock.Anything, "notfound@email.com").Return(nil, nil)

	_, err := service.Login(context.Background(), "notfound@email.com", "anypass")

	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInvalidCredentials))
}

func TestAuthService_RefreshTokens_Success(t *testing.T) {
	mockRepo := user.NewMockUserRepository(t)
	mockRTRepo := NewMockRefreshTokenRepository(t)
	mockSocialAuthClient := NewMockSupabaseAuthClient(t)
	service := NewAuthService(mockRepo, mockRTRepo, mockSocialAuthClient, "secret")

	rawToken := "old-refresh-token"
	rt := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		FamilyID:  uuid.New(),
		IsRevoked: false,
		ExpiresAt: time.Now().Add(time.Hour),
	}

	u := &domain.User{ID: rt.UserID, FullName: "Test User", Email: "test@example.com", AuthProvider: domain.AuthProviderLocal}

	mockRTRepo.On("GetByTokenHash", mock.Anything, mock.Anything).Return(rt, nil)
	mockRTRepo.On("RevokeByID", mock.Anything, rt.ID).Return(nil)
	mockRTRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	mockRepo.On("GetUserByID", mock.Anything, rt.UserID).Return(u, nil)

	res, err := service.RefreshTokens(context.Background(), rawToken)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotEmpty(t, res.AccessToken)
}

func TestAuthService_RefreshTokens_Fail_TokenReused(t *testing.T) {
	mockRepo := user.NewMockUserRepository(t)
	mockRTRepo := NewMockRefreshTokenRepository(t)
	mockSocialAuthClient := NewMockSupabaseAuthClient(t)
	service := NewAuthService(mockRepo, mockRTRepo, mockSocialAuthClient, "secret")

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

func TestAuthService_Logout_Success(t *testing.T) {
	mockRTRepo := NewMockRefreshTokenRepository(t)
	mockSocialAuthClient := NewMockSupabaseAuthClient(t)
	service := NewAuthService(nil, mockRTRepo, mockSocialAuthClient, "secret")

	token := "valid-token"
	rt := &domain.RefreshToken{
		FamilyID: uuid.New(),
	}

	mockRTRepo.On("GetByTokenHash", mock.Anything, mock.Anything).Return(rt, nil)
	mockRTRepo.On("RevokeAllByFamilyID", mock.Anything, rt.FamilyID).Return(nil)

	err := service.Logout(context.Background(), token)

	assert.NoError(t, err)
}

func TestAuthService_SocialLogin_Success_NewUser(t *testing.T) {
	mockRepo := user.NewMockUserRepository(t)
	mockRTRepo := NewMockRefreshTokenRepository(t)
	mockSocialAuthClient := NewMockSupabaseAuthClient(t)
	service := NewAuthService(mockRepo, mockRTRepo, mockSocialAuthClient, "secret")

	idToken := "valid-token"
	uid := "sb-uid-123"
	email := "sb-user@example.com"
	name := "Social User"

	tokenResult := &SupabaseTokenResult{
		UserID:   uid,
		Email:    email,
		Name:     name,
		Provider: domain.AuthProviderGoogle,
	}

	mockSocialAuthClient.On("VerifyAccessToken", mock.Anything, idToken).Return(tokenResult, nil)
	mockRepo.On("GetUserByProviderID", mock.Anything, domain.AuthProviderGoogle, uid).Return(nil, nil)
	mockRepo.On("GetUserByEmail", mock.Anything, email).Return(nil, nil)
	mockRepo.On("GetUserByUsername", mock.Anything, "sb-user").Return(nil, nil)
	mockRepo.On("CreateUser", mock.Anything, mock.Anything).Return(nil)
	mockRTRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	res, err := service.SocialLogin(context.Background(), idToken)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, email, res.Email)
	assert.Equal(t, name, res.FullName)
}

func TestAuthService_SocialLogin_Fail_EmailConflict(t *testing.T) {
	mockRepo := user.NewMockUserRepository(t)
	mockRTRepo := NewMockRefreshTokenRepository(t)
	mockSocialAuthClient := NewMockSupabaseAuthClient(t)
	service := NewAuthService(mockRepo, mockRTRepo, mockSocialAuthClient, "secret")

	uid := "sb-uid-123"
	email := "conflict@example.com"

	tokenResult := &SupabaseTokenResult{
		UserID:   uid,
		Email:    email,
		Provider: domain.AuthProviderGoogle,
	}

	existingUser := &domain.User{
		ID:           uuid.New(),
		Email:        email,
		AuthProvider: domain.AuthProviderLocal,
	}

	mockSocialAuthClient.On("VerifyAccessToken", mock.Anything, mock.Anything).Return(tokenResult, nil)
	mockRepo.On("GetUserByProviderID", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
	mockRepo.On("GetUserByEmail", mock.Anything, email).Return(existingUser, nil)

	_, err := service.SocialLogin(context.Background(), "token")

	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrEmailAlreadyUsedByOtherMethod)
}
