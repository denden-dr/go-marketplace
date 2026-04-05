package services

import (
	"context"
	"errors"
	"time"

	"github.com/denden-dr/go-shop-yourself/internal/domain"
	"github.com/denden-dr/go-shop-yourself/internal/dto"
	"github.com/denden-dr/go-shop-yourself/internal/repo"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
)

// AuthService handles authentication logic.
type AuthService interface {
	Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error)
	Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error)
}

type authService struct {
	repo      repo.UserRepository
	validate  *validator.Validate
	jwtSecret []byte
}

// NewAuthService creates a new instance of AuthService.
func NewAuthService(repo repo.UserRepository, jwtSecret string) AuthService {
	return &authService{
		repo:      repo,
		validate:  validator.New(),
		jwtSecret: []byte(jwtSecret),
	}
}

func (s *authService) Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error) {
	// Validate input
	if err := s.validate.Struct(req); err != nil {
		return nil, err
	}

	// Check if user already exists
	existingUser, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user entities
	uID := uuid.New()
	now := time.Now()

	user := &domain.User{ID: uID, CreatedAt: now}
	auth := &domain.UserAuth{UserID: uID, Email: req.Email, PasswordHash: string(hashedPassword), CreatedAt: now}
	profile := &domain.UserProfile{UserID: uID, Username: req.Username, UpdatedAt: now}

	// Save to database
	if err := s.repo.CreateUser(ctx, user, auth, profile); err != nil {
		return nil, err
	}

	// Generate JWT
	token, err := s.generateToken(uID)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{Token: token, UserID: uID}, nil
}

func (s *authService) Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
	// Validate input
	if err := s.validate.Struct(req); err != nil {
		return nil, err
	}

	// Get user by email
	auth, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if auth == nil {
		return nil, ErrInvalidCredentials
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(auth.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Update last login
	if err := s.repo.UpdateLastLogin(ctx, auth.UserID); err != nil {
		// Log error but don't fail login
	}

	// Generate JWT
	token, err := s.generateToken(auth.UserID)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{Token: token, UserID: auth.UserID}, nil
}

func (s *authService) generateToken(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID.String(),
		"exp": time.Now().Add(time.Minute * 15).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}
