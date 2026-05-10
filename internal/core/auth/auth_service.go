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

type AuthServiceInterface interface {
	Register(ctx context.Context, fullName, email, password, username string) (*AuthResponse, error)
	VerifyEmail(ctx context.Context, userID uuid.UUID, code string) error
	Login(ctx context.Context, email, password, ipAddress, userAgent string) (*AuthResponse, error)
	RefreshTokens(ctx context.Context, rawToken, ipAddress, userAgent string) (*AuthResponse, error)
	Logout(ctx context.Context, rawToken string) error
}

type SessionRepository interface {
	Create(ctx context.Context, session *domain.Session) error
	GetByTokenHash(ctx context.Context, hash string) (*domain.Session, error)
	RevokeByID(ctx context.Context, id uuid.UUID) error
	RevokeAllByFamilyID(ctx context.Context, familyID uuid.UUID) error
	DeleteExpiredSessions(ctx context.Context) error
}

type VerificationRepository interface {
	Create(ctx context.Context, vc *domain.VerificationCode) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.VerificationCode, error)
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
}

type AuthService struct {
	userRepo         user.UserRepository
	sessionRepo      SessionRepository
	verificationRepo VerificationRepository
	mailService      MailService
	jwtSecret        string
}

func NewAuthService(
	userRepo user.UserRepository,
	sessionRepo SessionRepository,
	verificationRepo VerificationRepository,
	mailService MailService,
	jwtSecret string,
) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		sessionRepo:      sessionRepo,
		verificationRepo: verificationRepo,
		mailService:      mailService,
		jwtSecret:        jwtSecret,
	}
}

func (s *AuthService) Register(ctx context.Context, fullName, email, password, username string) (*AuthResponse, error) {
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

func (s *AuthService) VerifyEmail(ctx context.Context, userID uuid.UUID, code string) error {
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

func (s *AuthService) Login(ctx context.Context, email, password, ipAddress, userAgent string) (*AuthResponse, error) {
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

func (s *AuthService) generateAuthResponse(ctx context.Context, u *domain.User, ipAddress, userAgent string) (*AuthResponse, error) {
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

func (s *AuthService) RefreshTokens(ctx context.Context, rawToken, ipAddress, userAgent string) (*AuthResponse, error) {
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

func (s *AuthService) Logout(ctx context.Context, rawToken string) error {
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
