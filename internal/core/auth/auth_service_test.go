package auth

import (
	"context"
	"testing"
	"time"

	"go-marketplace/internal/core/user"
	"go-marketplace/internal/domain"

	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
)

func TestAuthService_Register(t *testing.T) {
	tests := []struct {
		name      string
		fullName  string
		email     string
		password  string
		username  string
		mockSetup func(mr *user.MockUserRepository, msr *MockSessionRepository, mvr *MockVerificationRepository, mms *MockMailService)
		wantErr   bool
		errType   error
	}{
		{
			name:     "Success",
			fullName: "Test User",
			email:    "test@example.com",
			password: "password123",
			username: "testuser",
			mockSetup: func(mr *user.MockUserRepository, msr *MockSessionRepository, mvr *MockVerificationRepository, mms *MockMailService) {
				mr.On("GetUserByEmail", mock.Anything, "test@example.com").Return(nil, nil)
				mr.On("CreateUser", mock.Anything, mock.Anything).Return(nil)
				mvr.On("Create", mock.Anything, mock.Anything).Return(nil)
				mms.On("SendVerificationCode", mock.Anything, "test@example.com", mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			name:     "User Already Exists",
			fullName: "Test User",
			email:    "exists@example.com",
			password: "password123",
			username: "testuser",
			mockSetup: func(mr *user.MockUserRepository, msr *MockSessionRepository, mvr *MockVerificationRepository, mms *MockMailService) {
				mr.On("GetUserByEmail", mock.Anything, "exists@example.com").Return(&domain.User{}, nil)
			},
			wantErr: true,
			errType: domain.ErrUserAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := user.NewMockUserRepository(t)
			mockSessionRepo := NewMockSessionRepository(t)
			mockVerifRepo := NewMockVerificationRepository(t)
			mockMailService := NewMockMailService(t)

			tt.mockSetup(mockRepo, mockSessionRepo, mockVerifRepo, mockMailService)

			service := NewAuthService(mockRepo, mockSessionRepo, mockVerifRepo, mockMailService, nil, "secret")
			res, err := service.Register(context.Background(), tt.fullName, tt.email, tt.password, tt.username)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
				assert.Equal(t, tt.fullName, res.FullName)
				assert.NotEqual(t, uuid.Nil, res.ID)
			}
		})
	}
}

func TestAuthService_VerifyEmail(t *testing.T) {
	userID := uuid.New()
	code := "123456"
	codeHash := HashToken(code)

	tests := []struct {
		name      string
		userID    uuid.UUID
		code      string
		mockSetup func(mr *user.MockUserRepository, mvr *MockVerificationRepository)
		wantErr   bool
		errType   error
	}{
		{
			name:   "Success",
			userID: userID,
			code:   code,
			mockSetup: func(mr *user.MockUserRepository, mvr *MockVerificationRepository) {
				vc := &domain.VerificationCode{
					UserID:    userID,
					CodeHash:  codeHash,
					ExpiresAt: time.Now().Add(time.Minute),
				}
				mvr.On("GetByUserID", mock.Anything, userID).Return(vc, nil)
				mr.On("UpdateVerifiedStatus", mock.Anything, userID, true).Return(nil)
				mvr.On("DeleteByUserID", mock.Anything, userID).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "Invalid Code",
			userID: userID,
			code:   "wrong",
			mockSetup: func(mr *user.MockUserRepository, mvr *MockVerificationRepository) {
				vc := &domain.VerificationCode{
					UserID:    userID,
					CodeHash:  codeHash,
					ExpiresAt: time.Now().Add(time.Minute),
				}
				mvr.On("GetByUserID", mock.Anything, userID).Return(vc, nil)
			},
			wantErr: true,
			errType: domain.ErrInvalidVerificationCode,
		},
		{
			name:   "Expired Code",
			userID: userID,
			code:   code,
			mockSetup: func(mr *user.MockUserRepository, mvr *MockVerificationRepository) {
				vc := &domain.VerificationCode{
					UserID:    userID,
					CodeHash:  codeHash,
					ExpiresAt: time.Now().Add(-time.Minute),
				}
				mvr.On("GetByUserID", mock.Anything, userID).Return(vc, nil)
			},
			wantErr: true,
			errType: domain.ErrVerificationCodeExpired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := user.NewMockUserRepository(t)
			mockVerifRepo := NewMockVerificationRepository(t)

			tt.mockSetup(mockRepo, mockVerifRepo)

			service := NewAuthService(mockRepo, nil, mockVerifRepo, nil, nil, "secret")
			err := service.VerifyEmail(context.Background(), tt.userID, tt.code)

			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.errType)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	email := "test@example.com"
	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	hashedStr := string(hashedPassword)

	u := &domain.User{
		ID:           uuid.New(),
		FullName:     "Test User",
		Username:     "testuser",
		Email:        email,
		IsVerified:   true,
		Password:     &hashedStr,
		AuthProvider: domain.AuthProviderLocal,
	}

	tests := []struct {
		name      string
		email     string
		password  string
		mockSetup func(mr *user.MockUserRepository, msr *MockSessionRepository)
		wantErr   bool
		errType   error
	}{
		{
			name:     "Success",
			email:    email,
			password: password,
			mockSetup: func(mr *user.MockUserRepository, msr *MockSessionRepository) {
				mr.On("GetUserByEmail", mock.Anything, email).Return(u, nil)
				msr.On("Create", mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			name:     "Invalid Credentials - User Not Found",
			email:    "wrong@example.com",
			password: password,
			mockSetup: func(mr *user.MockUserRepository, msr *MockSessionRepository) {
				mr.On("GetUserByEmail", mock.Anything, "wrong@example.com").Return(nil, nil)
			},
			wantErr: true,
			errType: domain.ErrInvalidCredentials,
		},
		{
			name:     "Email Not Verified",
			email:    email,
			password: password,
			mockSetup: func(mr *user.MockUserRepository, msr *MockSessionRepository) {
				unverified := *u
				unverified.IsVerified = false
				mr.On("GetUserByEmail", mock.Anything, email).Return(&unverified, nil)
			},
			wantErr: true,
			errType: domain.ErrEmailNotVerified,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := user.NewMockUserRepository(t)
			mockSessionRepo := NewMockSessionRepository(t)

			tt.mockSetup(mockRepo, mockSessionRepo)

			service := NewAuthService(mockRepo, mockSessionRepo, nil, nil, nil, "secret")
			res, err := service.Login(context.Background(), tt.email, tt.password, "127.0.0.1", "test-agent")

			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.errType)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
				assert.Equal(t, u.ID, res.ID)
			}
		})
	}
}

func TestAuthService_RefreshTokens(t *testing.T) {
	rawToken := "old-refresh-token"
	userID := uuid.New()
	familyID := uuid.New()
	u := &domain.User{ID: userID, FullName: "Test User", Email: "test@example.com", AuthProvider: domain.AuthProviderLocal}

	tests := []struct {
		name      string
		rawToken  string
		mockSetup func(mr *user.MockUserRepository, msr *MockSessionRepository)
		wantErr   bool
		errType   error
	}{
		{
			name:     "Success",
			rawToken: rawToken,
			mockSetup: func(mr *user.MockUserRepository, msr *MockSessionRepository) {
				session := &domain.Session{
					ID:        uuid.New(),
					UserID:    userID,
					FamilyID:  familyID,
					IsRevoked: false,
					ExpiresAt: time.Now().Add(time.Hour),
				}
				msr.On("GetByTokenHash", mock.Anything, mock.Anything).Return(session, nil)
				msr.On("RevokeByID", mock.Anything, session.ID).Return(nil)
				msr.On("Create", mock.Anything, mock.Anything).Return(nil)
				mr.On("GetUserByID", mock.Anything, userID).Return(u, nil)
			},
			wantErr: false,
		},
		{
			name:     "Token Reused (Revoked)",
			rawToken: rawToken,
			mockSetup: func(mr *user.MockUserRepository, msr *MockSessionRepository) {
				session := &domain.Session{
					ID:        uuid.New(),
					UserID:    userID,
					FamilyID:  familyID,
					IsRevoked: true,
					ExpiresAt: time.Now().Add(time.Hour),
				}
				msr.On("GetByTokenHash", mock.Anything, mock.Anything).Return(session, nil)
				msr.On("RevokeAllByFamilyID", mock.Anything, familyID).Return(nil)
			},
			wantErr: true,
			errType: domain.ErrRefreshTokenReused,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := user.NewMockUserRepository(t)
			mockSessionRepo := NewMockSessionRepository(t)

			tt.mockSetup(mockRepo, mockSessionRepo)

			service := NewAuthService(mockRepo, mockSessionRepo, nil, nil, nil, "secret")
			res, err := service.RefreshTokens(context.Background(), tt.rawToken, "127.0.0.1", "test-agent")

			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.errType)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
			}
		})
	}
}

func TestAuthService_Logout(t *testing.T) {
	token := "valid-token"
	familyID := uuid.New()

	tests := []struct {
		name      string
		token     string
		mockSetup func(msr *MockSessionRepository)
		wantErr   bool
		errType   error
	}{
		{
			name:  "Success",
			token: token,
			mockSetup: func(msr *MockSessionRepository) {
				session := &domain.Session{FamilyID: familyID}
				msr.On("GetByTokenHash", mock.Anything, mock.Anything).Return(session, nil)
				msr.On("RevokeAllByFamilyID", mock.Anything, familyID).Return(nil)
			},
			wantErr: false,
		},
		{
			name:  "Session Not Found",
			token: "unknown",
			mockSetup: func(msr *MockSessionRepository) {
				msr.On("GetByTokenHash", mock.Anything, mock.Anything).Return(nil, nil)
			},
			wantErr: true,
			errType: domain.ErrInvalidRefreshToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSessionRepo := NewMockSessionRepository(t)

			tt.mockSetup(mockSessionRepo)

			service := NewAuthService(nil, mockSessionRepo, nil, nil, nil, "secret")
			err := service.Logout(context.Background(), tt.token)

			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.errType)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthService_HandleGoogleLogin(t *testing.T) {
	state := "valid-state"
	code := "valid-code"
	email := "google@example.com"
	sub := "google-sub-123"
	name := "Google User"

	userInfo := &GoogleUserInfo{
		Sub:   sub,
		Email: email,
		Name:  name,
	}

	tests := []struct {
		name      string
		state     string
		code      string
		mockSetup func(mr *user.MockUserRepository, msr *MockSessionRepository, mgc *MockGoogleClient)
		wantErr   bool
		errType   error
	}{
		{
			name:  "Success - New User",
			state: state,
			code:  code,
			mockSetup: func(mr *user.MockUserRepository, msr *MockSessionRepository, mgc *MockGoogleClient) {
				mgc.On("ExchangeCode", mock.Anything, code).Return(&oauth2.Token{}, nil)
				mgc.On("GetUserInfo", mock.Anything, mock.Anything).Return(userInfo, nil)
				mr.On("GetUserByEmail", mock.Anything, email).Return(nil, nil)
				mr.On("GetUserByUsername", mock.Anything, "google").Return(nil, nil)
				mr.On("CreateUser", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
					return u.Email == email && u.AuthProvider == domain.AuthProviderGoogle && *u.ProviderID == sub
				})).Return(nil)
				msr.On("Create", mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			name:  "Success - Existing User (Linking)",
			state: state,
			code:  code,
			mockSetup: func(mr *user.MockUserRepository, msr *MockSessionRepository, mgc *MockGoogleClient) {
				mgc.On("ExchangeCode", mock.Anything, code).Return(&oauth2.Token{}, nil)
				mgc.On("GetUserInfo", mock.Anything, mock.Anything).Return(userInfo, nil)
				mr.On("GetUserByEmail", mock.Anything, email).Return(&domain.User{ID: uuid.New(), Email: email}, nil)
				msr.On("Create", mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			name:  "Invalid State",
			state: "wrong-state",
			code:  code,
			mockSetup: func(mr *user.MockUserRepository, msr *MockSessionRepository, mgc *MockGoogleClient) {
			},
			wantErr: true,
			errType: domain.ErrInvalidOAuthState,
		},
		{
			name:  "Exchange Code Failure",
			state: state,
			code:  code,
			mockSetup: func(mr *user.MockUserRepository, msr *MockSessionRepository, mgc *MockGoogleClient) {
				mgc.On("ExchangeCode", mock.Anything, code).Return(nil, fmt.Errorf("exchange failed"))
			},
			wantErr: true,
			errType: domain.ErrInvalidSocialToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := user.NewMockUserRepository(t)
			mockSessionRepo := NewMockSessionRepository(t)
			mockGoogleClient := NewMockGoogleClient(t)

			tt.mockSetup(mockRepo, mockSessionRepo, mockGoogleClient)

			service := NewAuthService(mockRepo, mockSessionRepo, nil, nil, mockGoogleClient, "secret")
			res, err := service.HandleGoogleLogin(context.Background(), tt.code, tt.state, state, "127.0.0.1", "test-agent")

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
				assert.Equal(t, email, res.Email)
			}
		})
	}
}
