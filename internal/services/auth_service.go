package services

import (
	"context"
	"errors"
	"time"

	"go-shop-yourself/internal/auth"
	"go-shop-yourself/internal/domain"
	"go-shop-yourself/internal/dtos"
	"go-shop-yourself/internal/repos"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo         *repos.UserRepository
	refreshTokenRepo *repos.RefreshTokenRepository
	jwtSecret        string
}

func NewAuthService(userRepo *repos.UserRepository, refreshTokenRepo *repos.RefreshTokenRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtSecret:        jwtSecret,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password, username string) (uuid.UUID, error) {
	// ... (Register method remains largely the same, but I'll keeping it for completeness in this view)
	// Actually, I'll just keep the existing implementation and only update Login and add others.
	// But since I'm replacing the whole file's bottom part, I'll rewrite it clearly.
	existingUser, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return uuid.Nil, err
	}
	if existingUser != nil {
		return uuid.Nil, errors.New("user with this email already exists")
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
		return nil, errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid email or password")
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
		return nil, errors.New("invalid refresh token")
	}

	// Reuse detection
	if rt.IsRevoked {
		// Someone is trying to reuse a revoked token - potential attack!
		// Revoke the entire family
		_ = s.refreshTokenRepo.RevokeAllByFamilyID(ctx, rt.FamilyID)
		return nil, errors.New("token reuse detected - all tokens in family revoked")
	}

	// Expiration check
	if time.Now().After(rt.ExpiresAt) {
		return nil, errors.New("refresh token expired")
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
		return errors.New("invalid refresh token")
	}

	// Revoke the entire family for safety on logout
	return s.refreshTokenRepo.RevokeAllByFamilyID(ctx, rt.FamilyID)
}
