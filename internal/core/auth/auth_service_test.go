package auth

import (
	"context"
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
	mockSessionRepo := NewMockSessionRepository(t)
	mockVerifRepo := NewMockVerificationRepository(t)
	mockMailService := NewMockMailServiceInterface(t)
	service := NewAuthService(mockRepo, mockSessionRepo, mockVerifRepo, mockMailService, "secret")

	email := "test@example.com"
	password := "password123"
	username := "testuser"

	mockRepo.On("GetUserByEmail", mock.Anything, email).Return(nil, nil)
	mockRepo.On("CreateUser", mock.Anything, mock.Anything).Return(nil)
	mockVerifRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	mockMailService.On("SendVerificationCode", mock.Anything, email, mock.Anything).Return(nil)

	res, err := service.Register(context.Background(), "Test User", email, password, username)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotEqual(t, uuid.Nil, res.ID)
	assert.Equal(t, "Test User", res.FullName)
}

func TestAuthService_VerifyEmail_Success(t *testing.T) {
	mockRepo := user.NewMockUserRepository(t)
	mockVerifRepo := NewMockVerificationRepository(t)
	service := NewAuthService(mockRepo, nil, mockVerifRepo, nil, "secret")

	userID := uuid.New()
	code := "123456"
	vc := &domain.VerificationCode{
		UserID:    userID,
		CodeHash:  HashToken(code),
		ExpiresAt: time.Now().Add(time.Minute),
	}

	mockVerifRepo.On("GetByUserID", mock.Anything, userID).Return(vc, nil)
	mockRepo.On("UpdateVerifiedStatus", mock.Anything, userID, true).Return(nil)
	mockVerifRepo.On("DeleteByUserID", mock.Anything, userID).Return(nil)

	err := service.VerifyEmail(context.Background(), userID, code)

	assert.NoError(t, err)
}

func TestAuthService_Login_Success(t *testing.T) {
	mockRepo := user.NewMockUserRepository(t)
	mockSessionRepo := NewMockSessionRepository(t)
	service := NewAuthService(mockRepo, mockSessionRepo, nil, nil, "secret")

	email := "test@example.com"
	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	u := &domain.User{
		ID:           uuid.New(),
		FullName:     "Test User",
		Username:     "testuser",
		Email:        email,
		IsVerified:   true,
		AuthProvider: domain.AuthProviderLocal,
	}
	hashedStr := string(hashedPassword)
	u.Password = &hashedStr

	mockRepo.On("GetUserByEmail", mock.Anything, email).Return(u, nil)
	mockSessionRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	res, err := service.Login(context.Background(), email, password, "127.0.0.1", "test-agent")

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, u.ID, res.ID)
}

func TestAuthService_RefreshTokens_Success(t *testing.T) {
	mockRepo := user.NewMockUserRepository(t)
	mockSessionRepo := NewMockSessionRepository(t)
	service := NewAuthService(mockRepo, mockSessionRepo, nil, nil, "secret")

	rawToken := "old-refresh-token"
	session := &domain.Session{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		FamilyID:  uuid.New(),
		IsRevoked: false,
		ExpiresAt: time.Now().Add(time.Hour),
	}

	u := &domain.User{ID: session.UserID, FullName: "Test User", Email: "test@example.com", AuthProvider: domain.AuthProviderLocal}

	mockSessionRepo.On("GetByTokenHash", mock.Anything, mock.Anything).Return(session, nil)
	mockSessionRepo.On("RevokeByID", mock.Anything, session.ID).Return(nil)
	mockSessionRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	mockRepo.On("GetUserByID", mock.Anything, session.UserID).Return(u, nil)

	res, err := service.RefreshTokens(context.Background(), rawToken, "127.0.0.1", "test-agent")

	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func TestAuthService_Logout_Success(t *testing.T) {
	mockSessionRepo := NewMockSessionRepository(t)
	service := NewAuthService(nil, mockSessionRepo, nil, nil, "secret")

	token := "valid-token"
	session := &domain.Session{
		FamilyID: uuid.New(),
	}

	mockSessionRepo.On("GetByTokenHash", mock.Anything, mock.Anything).Return(session, nil)
	mockSessionRepo.On("RevokeAllByFamilyID", mock.Anything, session.FamilyID).Return(nil)

	err := service.Logout(context.Background(), token)

	assert.NoError(t, err)
}
