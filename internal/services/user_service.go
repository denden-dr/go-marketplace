package services

import (
	"context"
	"fmt"

	"github.com/denden-dr/go-shop-yourself/internal/dto"
	"github.com/denden-dr/go-shop-yourself/internal/repo"
	"github.com/google/uuid"
)

// UserService handles user-related logic.
type UserService interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (*dto.UserProfileResponse, error)
}

type userService struct {
	repo repo.UserRepository
}

// NewUserService creates a new instance of UserService.
func NewUserService(repo repo.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) GetProfile(ctx context.Context, userID uuid.UUID) (*dto.UserProfileResponse, error) {
	profile, err := s.repo.GetUserProfileByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, fmt.Errorf("user profile not found")
	}

	return &dto.UserProfileResponse{
		UserID:       profile.UserID,
		Username:     profile.Username,
		SavedAddress: profile.SavedAddress,
		UpdatedAt:    profile.UpdatedAt,
		LastLoginAt:  profile.LastLoginAt,
	}, nil
}
