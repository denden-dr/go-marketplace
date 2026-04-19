package auth

import (
	"context"
	"testing"

	"go-shop-yourself/internal/domain"
	"go-shop-yourself/internal/user"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthService_FirebaseLogin_Success_NewUser(t *testing.T) {
	mockRepo := user.NewMockUserRepository(t)
	mockRTRepo := NewMockRefreshTokenRepository(t)
	mockFirebaseClient := NewMockFirebaseAuthClient(t)
	service := NewAuthService(mockRepo, mockRTRepo, mockFirebaseClient, "secret")

	idToken := "valid-token"
	uid := "fb-uid-123"
	email := "fb-user@example.com"
	name := "Firebase User"

	tokenResult := &FirebaseTokenResult{
		UID:      uid,
		Email:    email,
		Name:     name,
		Provider: domain.AuthProviderGoogle,
	}

	mockFirebaseClient.On("VerifyIDToken", mock.Anything, idToken).Return(tokenResult, nil)
	mockRepo.On("GetUserByProviderID", mock.Anything, domain.AuthProviderGoogle, uid).Return(nil, nil)
	mockRepo.On("GetUserByEmail", mock.Anything, email).Return(nil, nil)
	mockRepo.On("GetUserByUsername", mock.Anything, "fb-user").Return(nil, nil)
	mockRepo.On("CreateUser", mock.Anything, mock.Anything).Return(nil)
	mockRTRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	res, err := service.FirebaseLogin(context.Background(), idToken)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, email, res.Email)
	assert.Equal(t, name, res.FullName)
}

func TestAuthService_FirebaseLogin_Success_ReturningUser(t *testing.T) {
	mockRepo := user.NewMockUserRepository(t)
	mockRTRepo := NewMockRefreshTokenRepository(t)
	mockFirebaseClient := NewMockFirebaseAuthClient(t)
	service := NewAuthService(mockRepo, mockRTRepo, mockFirebaseClient, "secret")

	idToken := "valid-token"
	uid := "fb-uid-123"
	email := "fb-user@example.com"

	tokenResult := &FirebaseTokenResult{
		UID:      uid,
		Email:    email,
		Provider: domain.AuthProviderGoogle,
	}

	u := &domain.User{
		ID:           uuid.New(),
		FullName:     "Existing User",
		Email:        email,
		AuthProvider: domain.AuthProviderGoogle,
		ProviderID:   &uid,
	}

	mockFirebaseClient.On("VerifyIDToken", mock.Anything, idToken).Return(tokenResult, nil)
	mockRepo.On("GetUserByProviderID", mock.Anything, domain.AuthProviderGoogle, uid).Return(u, nil)
	mockRTRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	res, err := service.FirebaseLogin(context.Background(), idToken)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, u.ID, res.ID)
}

func TestAuthService_FirebaseLogin_Fail_EmailConflict(t *testing.T) {
	mockRepo := user.NewMockUserRepository(t)
	mockRTRepo := NewMockRefreshTokenRepository(t)
	mockFirebaseClient := NewMockFirebaseAuthClient(t)
	service := NewAuthService(mockRepo, mockRTRepo, mockFirebaseClient, "secret")

	uid := "fb-uid-123"
	email := "conflict@example.com"

	tokenResult := &FirebaseTokenResult{
		UID:      uid,
		Email:    email,
		Provider: domain.AuthProviderGoogle,
	}

	existingUser := &domain.User{
		ID:           uuid.New(),
		Email:        email,
		AuthProvider: domain.AuthProviderLocal,
	}

	mockFirebaseClient.On("VerifyIDToken", mock.Anything, mock.Anything).Return(tokenResult, nil)
	mockRepo.On("GetUserByProviderID", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
	mockRepo.On("GetUserByEmail", mock.Anything, email).Return(existingUser, nil)

	_, err := service.FirebaseLogin(context.Background(), "token")

	assert.Error(t, err)
	assert.Equal(t, domain.ErrEmailAlreadyUsedByOtherMethod, err)
}

func TestAuthService_Login_Fail_SocialUserLocalLogin(t *testing.T) {
	mockRepo := user.NewMockUserRepository(t)
	mockRTRepo := NewMockRefreshTokenRepository(t)
	mockFirebaseClient := NewMockFirebaseAuthClient(t)
	service := NewAuthService(mockRepo, mockRTRepo, mockFirebaseClient, "secret")

	email := "social@example.com"
	u := &domain.User{
		Email:        email,
		AuthProvider: domain.AuthProviderGoogle,
	}

	mockRepo.On("GetUserByEmail", mock.Anything, email).Return(u, nil)

	_, err := service.Login(context.Background(), email, "password")

	assert.Error(t, err)
	assert.Equal(t, domain.ErrAuthProviderMismatch, err)
}

func TestAuthService_FirebaseLogin_Fail_EmailNotVerified(t *testing.T) {
	mockRepo := user.NewMockUserRepository(t)
	mockRTRepo := NewMockRefreshTokenRepository(t)
	mockFirebaseClient := NewMockFirebaseAuthClient(t)
	service := NewAuthService(mockRepo, mockRTRepo, mockFirebaseClient, "secret")

	mockFirebaseClient.On("VerifyIDToken", mock.Anything, mock.Anything).Return(nil, domain.ErrEmailNotVerified)

	_, err := service.FirebaseLogin(context.Background(), "token")

	assert.Error(t, err)
	assert.Equal(t, domain.ErrEmailNotVerified, err)
}

func TestAuthService_FirebaseLogin_Fail_NilClient(t *testing.T) {
	mockRepo := user.NewMockUserRepository(t)
	mockRTRepo := NewMockRefreshTokenRepository(t)
	service := NewAuthService(mockRepo, mockRTRepo, nil, "secret")

	_, err := service.FirebaseLogin(context.Background(), "token")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "social login is not available")
}

func TestAuthService_FirebaseLogin_Success_UsernameCollision(t *testing.T) {
	mockRepo := user.NewMockUserRepository(t)
	mockRTRepo := NewMockRefreshTokenRepository(t)
	mockFirebaseClient := NewMockFirebaseAuthClient(t)
	service := NewAuthService(mockRepo, mockRTRepo, mockFirebaseClient, "secret")

	idToken := "valid-token"
	uid := "fb-uid-123"
	email := "collision@example.com"

	tokenResult := &FirebaseTokenResult{
		UID:      uid,
		Email:    email,
		Name:     "Collision User",
		Provider: domain.AuthProviderGoogle,
	}

	existingUser := &domain.User{Username: "collision"}

	mockFirebaseClient.On("VerifyIDToken", mock.Anything, idToken).Return(tokenResult, nil)
	mockRepo.On("GetUserByProviderID", mock.Anything, domain.AuthProviderGoogle, uid).Return(nil, nil)
	mockRepo.On("GetUserByEmail", mock.Anything, email).Return(nil, nil)
	// First username "collision" is taken
	mockRepo.On("GetUserByUsername", mock.Anything, "collision").Return(existingUser, nil).Once()
	// Second username "collision_xxxx" is free
	mockRepo.On("GetUserByUsername", mock.Anything, mock.MatchedBy(func(s string) bool {
		return len(s) > 10 // collision_ + 4 chars
	})).Return(nil, nil).Once()
	mockRepo.On("CreateUser", mock.Anything, mock.Anything).Return(nil)
	mockRTRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	res, err := service.FirebaseLogin(context.Background(), idToken)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotEqual(t, "collision", res.Username)
	assert.Contains(t, res.Username, "collision_")
}
