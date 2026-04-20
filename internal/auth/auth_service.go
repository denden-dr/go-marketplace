package auth

import (
	"context"
	"fmt"
	"time"

	"go-shop-yourself/internal/domain"

	"go-shop-yourself/internal/user"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthServiceInterface interface {
	Register(ctx context.Context, fullName, email, password, username string) (*AuthResponse, error)
	Login(ctx context.Context, email, password string) (*AuthResponse, error)
	SocialLogin(ctx context.Context, accessToken string) (*AuthResponse, error)
	RefreshTokens(ctx context.Context, rawToken string) (*AuthResponse, error)
	Logout(ctx context.Context, rawToken string) error
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, rt *domain.RefreshToken) error
	GetByTokenHash(ctx context.Context, hash string) (*domain.RefreshToken, error)
	RevokeByID(ctx context.Context, id uuid.UUID) error
	RevokeAllByFamilyID(ctx context.Context, familyID uuid.UUID) error
	DeleteExpiredTokens(ctx context.Context) error
}

type AuthService struct {
	userRepo         user.UserRepository
	refreshTokenRepo RefreshTokenRepository
	socialAuthClient SupabaseAuthClient
	jwtSecret        string
}

func NewAuthService(userRepo user.UserRepository, refreshTokenRepo RefreshTokenRepository, socialAuthClient SupabaseAuthClient, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		socialAuthClient: socialAuthClient,
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
		ProviderID:   nil,
		CreatedAt:    time.Now(),
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return s.generateAuthResponse(ctx, user)
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*AuthResponse, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.ErrInvalidCredentials
	}

	if user.AuthProvider != domain.AuthProviderLocal {
		return nil, domain.ErrAuthProviderMismatch
	}

	if user.Password == nil {
		return nil, domain.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	return s.generateAuthResponse(ctx, user)
}

func (s *AuthService) SocialLogin(ctx context.Context, accessToken string) (*AuthResponse, error) {
	if s.socialAuthClient == nil {
		return nil, domain.ErrSocialLoginNotAvailable
	}

	// 1. Verify token
	tokenResult, err := s.socialAuthClient.VerifyAccessToken(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	// 2. Lookup user by ProviderID
	user, err := s.userRepo.GetUserByProviderID(ctx, tokenResult.Provider, tokenResult.UserID)
	if err != nil {
		return nil, err
	}

	// 3. If user exists, generate tokens
	if user != nil {
		return s.generateAuthResponse(ctx, user)
	}

	// 4. New User - Check if email is taken
	existingUser, err := s.userRepo.GetUserByEmail(ctx, tokenResult.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, domain.ErrEmailAlreadyUsedByOtherMethod
	}

	// 5. Create new user
	// Generate unique username from email prefix
	baseUsername := tokenResult.Email
	for i, char := range tokenResult.Email {
		if char == '@' {
			baseUsername = tokenResult.Email[:i]
			break
		}
	}

	username := baseUsername
	for {
		u, err := s.userRepo.GetUserByUsername(ctx, username)
		if err != nil {
			return nil, err
		}
		if u == nil {
			break
		}
		// Collision! Append random suffix
		username = fmt.Sprintf("%s_%s", baseUsername, uuid.New().String()[:4])
	}

	newUser := &domain.User{
		ID:           uuid.New(),
		FullName:     tokenResult.Name,
		Username:     username,
		Email:        tokenResult.Email,
		Password:     nil,
		AuthProvider: tokenResult.Provider,
		ProviderID:   &tokenResult.UserID,
		CreatedAt:    time.Now(),
	}

	if err := s.userRepo.CreateUser(ctx, newUser); err != nil {
		return nil, err
	}

	return s.generateAuthResponse(ctx, newUser)
}

func (s *AuthService) generateAuthResponse(ctx context.Context, u *domain.User) (*AuthResponse, error) {
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
	rt := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    u.ID,
		TokenHash: tokenHash,
		FamilyID:  familyID,
		IsRevoked: false,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 7),
		CreatedAt: time.Now(),
	}

	if err := s.refreshTokenRepo.Create(ctx, rt); err != nil {
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

func (s *AuthService) RefreshTokens(ctx context.Context, rawToken string) (*AuthResponse, error) {
	tokenHash := HashToken(rawToken)
	rt, err := s.refreshTokenRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}

	if rt == nil {
		return nil, domain.ErrInvalidRefreshToken
	}

	// Reuse detection
	if rt.IsRevoked {
		// Someone is trying to reuse a revoked token - potential attack!
		// Revoke the entire family
		_ = s.refreshTokenRepo.RevokeAllByFamilyID(ctx, rt.FamilyID)
		return nil, domain.ErrRefreshTokenReused
	}

	// Expiration check
	if time.Now().After(rt.ExpiresAt) {
		return nil, domain.ErrRefreshTokenExpired
	}

	// Revoke the old token
	if err := s.refreshTokenRepo.RevokeByID(ctx, rt.ID); err != nil {
		return nil, err
	}

	// Generate new tokens
	accessToken, err := GenerateAccessToken(rt.UserID, s.jwtSecret)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	// Store new refresh token with SAME familyID
	newTokenHash := HashToken(newRefreshToken)
	newRt := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    rt.UserID,
		TokenHash: newTokenHash,
		FamilyID:  rt.FamilyID,
		IsRevoked: false,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 7),
		CreatedAt: time.Now(),
	}

	if err := s.refreshTokenRepo.Create(ctx, newRt); err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetUserByID(ctx, rt.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}

	return &AuthResponse{
		ID:           rt.UserID,
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
	rt, err := s.refreshTokenRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return err
	}
	if rt == nil {
		return domain.ErrInvalidRefreshToken
	}

	// Revoke the entire family for safety on logout
	return s.refreshTokenRepo.RevokeAllByFamilyID(ctx, rt.FamilyID)
}
