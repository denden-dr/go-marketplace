package auth

import (
	"context"
	"fmt"
	"time"

	"go-marketplace/internal/core/user"
	"go-marketplace/internal/domain"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func ptr(s string) *string {
	return &s
}

type AuthService interface {
	Register(ctx context.Context, fullName, email, password, username string) (*AuthResponse, error)
	VerifyEmail(ctx context.Context, userID uuid.UUID, code string) error
	Login(ctx context.Context, email, password, ipAddress, userAgent string) (*AuthResponse, error)
	RefreshTokens(ctx context.Context, rawToken, ipAddress, userAgent string) (*AuthResponse, error)
	Logout(ctx context.Context, rawToken string) error
	HandleGoogleLogin(ctx context.Context, code, state, expectedState, ipAddress, userAgent string) (*AuthResponse, error)
}

type authService struct {
	userRepo         user.UserRepository
	sessionRepo      SessionRepository
	verificationRepo VerificationRepository
	mailService      MailService
	googleClient     GoogleClient
	jwtSecret        string
}

func NewAuthService(
	userRepo user.UserRepository,
	sessionRepo SessionRepository,
	verificationRepo VerificationRepository,
	mailService MailService,
	googleClient GoogleClient,
	jwtSecret string,
) AuthService {
	return &authService{
		userRepo:         userRepo,
		sessionRepo:      sessionRepo,
		verificationRepo: verificationRepo,
		mailService:      mailService,
		googleClient:     googleClient,
		jwtSecret:        jwtSecret,
	}
}

func (s *authService) Register(ctx context.Context, fullName, email, password, username string) (*AuthResponse, error) {
	existingUser, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, domain.ErrUserAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	hashedPasswordStr := string(hashedPassword)

	user := &domain.User{
		ID:           uuid.New(),
		FullName:     fullName,
		Username:     username,
		Email:        email,
		Password:     &hashedPasswordStr,
		AuthProvider: domain.AuthProviderLocal,
		IsVerified:   false,
		CreatedAt:    time.Now(),
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	// Generate and send verification code
	code, err := GenerateVerificationCode() // 6 digit
	if err != nil {
		return nil, err
	}

	vc := &domain.VerificationCode{
		ID:        uuid.New(),
		UserID:    user.ID,
		CodeHash:  HashToken(code),
		ExpiresAt: time.Now().Add(time.Minute * 15),
		CreatedAt: time.Now(),
	}

	if err := s.verificationRepo.Create(ctx, vc); err != nil {
		return nil, err
	}

	if err := s.mailService.SendVerificationCode(ctx, user.Email, code); err != nil {
		// Log error but don't fail registration? Or fail?
		// For now, let's return error to inform user.
		return nil, fmt.Errorf("failed to send verification email: %w", err)
	}

	return &AuthResponse{
		ID:        user.ID,
		FullName:  user.FullName,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}, nil
}

func (s *authService) VerifyEmail(ctx context.Context, userID uuid.UUID, code string) error {
	vc, err := s.verificationRepo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if vc == nil {
		return domain.ErrInvalidVerificationCode
	}

	entity := NewVerificationCode(vc)

	if entity.IsExpired() {
		return domain.ErrVerificationCodeExpired
	}

	if !entity.IsValid(code) {
		return domain.ErrInvalidVerificationCode
	}

	if err := s.userRepo.UpdateVerifiedStatus(ctx, userID, true); err != nil {
		return err
	}

	return s.verificationRepo.DeleteByUserID(ctx, userID)
}

func (s *authService) Login(ctx context.Context, email, password, ipAddress, userAgent string) (*AuthResponse, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.ErrInvalidCredentials
	}

	if !user.IsVerified {
		return nil, domain.ErrEmailNotVerified
	}

	if user.Password == nil {
		return nil, domain.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	return s.generateAuthResponse(ctx, user, ipAddress, userAgent)
}

func (s *authService) generateAuthResponse(ctx context.Context, u *domain.User, ipAddress, userAgent string) (*AuthResponse, error) {
	accessToken, err := GenerateAccessToken(u.ID, s.jwtSecret)
	if err != nil {
		return nil, err
	}

	refreshToken, err := GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	familyID := uuid.New()
	tokenHash := HashToken(refreshToken)
	session := &domain.Session{
		ID:        uuid.New(),
		UserID:    u.ID,
		TokenHash: tokenHash,
		FamilyID:  familyID,
		IsRevoked: false,
		IPAddress: ptr(ipAddress),
		UserAgent: ptr(userAgent),
		ExpiresAt: time.Now().Add(time.Hour * 24 * 7),
		CreatedAt: time.Now(),
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	return &AuthResponse{
		ID:           u.ID,
		FullName:     u.FullName,
		Username:     u.Username,
		Email:        u.Email,
		CreatedAt:    u.CreatedAt,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *authService) RefreshTokens(ctx context.Context, rawToken, ipAddress, userAgent string) (*AuthResponse, error) {
	tokenHash := HashToken(rawToken)
	session, err := s.sessionRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}

	if session == nil {
		return nil, domain.ErrInvalidRefreshToken
	}

	entity := NewSession(session)

	// Reuse and expiration detection
	if err := entity.CanRefresh(); err != nil {
		if err == domain.ErrRefreshTokenReused {
			// Revoke the entire family for security
			_ = s.sessionRepo.RevokeAllByFamilyID(ctx, session.FamilyID)
		}
		return nil, err
	}

	// Revoke the old token
	if err := s.sessionRepo.RevokeByID(ctx, session.ID); err != nil {
		return nil, err
	}

	// Generate new tokens
	accessToken, err := GenerateAccessToken(session.UserID, s.jwtSecret)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	// Store new refresh token with SAME familyID
	newTokenHash := HashToken(newRefreshToken)
	newSession := &domain.Session{
		ID:        uuid.New(),
		UserID:    session.UserID,
		TokenHash: newTokenHash,
		FamilyID:  session.FamilyID,
		IsRevoked: false,
		IPAddress: ptr(ipAddress),
		UserAgent: ptr(userAgent),
		ExpiresAt: time.Now().Add(time.Hour * 24 * 7),
		CreatedAt: time.Now(),
	}

	if err := s.sessionRepo.Create(ctx, newSession); err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetUserByID(ctx, session.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}

	return &AuthResponse{
		ID:           session.UserID,
		FullName:     user.FullName,
		Username:     user.Username,
		Email:        user.Email,
		CreatedAt:    user.CreatedAt,
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *authService) Logout(ctx context.Context, rawToken string) error {
	tokenHash := HashToken(rawToken)
	session, err := s.sessionRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return err
	}
	if session == nil {
		return domain.ErrInvalidRefreshToken
	}

	// Revoke the entire family for safety on logout
	return s.sessionRepo.RevokeAllByFamilyID(ctx, session.FamilyID)
}

func (s *authService) HandleGoogleLogin(ctx context.Context, code, state, expectedState, ipAddress, userAgent string) (*AuthResponse, error) {
	if expectedState == "" || state != expectedState {
		return nil, domain.ErrInvalidOAuthState
	}

	token, err := s.googleClient.ExchangeCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInvalidSocialToken, err)
	}

	userInfo, err := s.googleClient.GetUserInfo(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInvalidSocialToken, err)
	}

	existingUser, err := s.userRepo.GetUserByEmail(ctx, userInfo.Email)
	if err != nil {
		return nil, err
	}

	if existingUser != nil {
		// Account linking happens automatically — we just issue tokens for the existing user.
		// We don't change the provider if it's already "local" or something else.
		return s.generateAuthResponse(ctx, existingUser, ipAddress, userAgent)
	}

	// Create new user
	baseUsername := s.generateUsername(userInfo.Email)
	username := baseUsername
	maxRetries := 5

	for i := 0; i < maxRetries; i++ {
		existingByUsername, err := s.userRepo.GetUserByUsername(ctx, username)
		if err != nil {
			return nil, err
		}
		if existingByUsername == nil {
			break
		}
		// Append random suffix and retry
		username = fmt.Sprintf("%s_%s", baseUsername, uuid.New().String()[:4])
		if i == maxRetries-1 {
			return nil, fmt.Errorf("failed to generate unique username after %d attempts", maxRetries)
		}
	}

	user := &domain.User{
		ID:           uuid.New(),
		FullName:     userInfo.Name,
		Username:     username,
		Email:        userInfo.Email,
		Password:     nil, // Social users have no password initially
		AuthProvider: domain.AuthProviderGoogle,
		ProviderID:   &userInfo.Sub,
		IsVerified:   true, // Google emails are already verified
		CreatedAt:    time.Now(),
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return s.generateAuthResponse(ctx, user, ipAddress, userAgent)
}

func (s *authService) generateUsername(email string) string {
	// Simple implementation: take prefix before @
	for i := 0; i < len(email); i++ {
		if email[i] == '@' {
			return email[:i]
		}
	}
	return email
}
