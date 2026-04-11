package services

import (
	"context"
	"time"

	"go-shop-yourself/internal/auth"
	"go-shop-yourself/internal/domain"
	"go-shop-yourself/internal/dtos"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo         domain.UserRepository
	refreshTokenRepo domain.RefreshTokenRepository
	jwtSecret        string
}

func NewAuthService(userRepo domain.UserRepository, refreshTokenRepo domain.RefreshTokenRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtSecret:        jwtSecret,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password, username string) (uuid.UUID, error) {
	existingUser, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return uuid.Nil, err
	}
	if existingUser != nil {
		return uuid.Nil, domain.ErrUserAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return uuid.Nil, err
	}

	user := &domain.User{
		ID:        uuid.New(),
		Username:  username,
		Email:     email,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return uuid.Nil, err
	}

	return user.ID, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*dtos.AuthResponse, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	// Generate tokens
	accessToken, err := auth.GenerateAccessToken(user.ID, s.jwtSecret)
	if err != nil {
		return nil, err
	}

	refreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	// Store refresh token
	familyID := uuid.New()
	tokenHash := auth.HashToken(refreshToken)
	rt := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: tokenHash,
		FamilyID:  familyID,
		IsRevoked: false,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 7), // 7 days
		CreatedAt: time.Now(),
	}

	if err := s.refreshTokenRepo.Create(ctx, rt); err != nil {
		return nil, err
	}

	return &dtos.AuthResponse{
		ID:           user.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) RefreshTokens(ctx context.Context, rawToken string) (*dtos.AuthResponse, error) {
	tokenHash := auth.HashToken(rawToken)
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
	accessToken, err := auth.GenerateAccessToken(rt.UserID, s.jwtSecret)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	// Store new refresh token with SAME familyID
	newTokenHash := auth.HashToken(newRefreshToken)
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

	return &dtos.AuthResponse{
		ID:           rt.UserID,
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, rawToken string) error {
	tokenHash := auth.HashToken(rawToken)
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
