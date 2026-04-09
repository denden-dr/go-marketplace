package services

import (
	"context"
	"errors"
	"time"

	"go-shop-yourself/internal/domain"
	"go-shop-yourself/internal/dtos"
	"go-shop-yourself/internal/repos"

	"github.com/google/uuid"
)

type MerchantService struct {
	repo     *repos.MerchantRepository
	userRepo *repos.UserRepository
}

func NewMerchantService(repo *repos.MerchantRepository, userRepo *repos.UserRepository) *MerchantService {
	return &MerchantService{repo: repo, userRepo: userRepo}
}

func (s *MerchantService) RegisterMerchant(ctx context.Context, req dtos.MerchantRegisterRequest) (*dtos.MerchantResponse, error) {
	// Check if user exists
	user, err := s.userRepo.GetUserByID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Check if merchant profile already exists for this user
	existing, err := s.repo.GetByUserID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("merchant already exists for this user")
	}

	merchant := &domain.Merchant{
		ID:        uuid.New(),
		UserID:    req.UserID,
		Name:      req.Name,
		About:     req.About,
		TaxID:     req.TaxID,
		CreatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, merchant); err != nil {
		return nil, err
	}

	return &dtos.MerchantResponse{
		ID:    merchant.ID,
		Name:  merchant.Name,
		Email: user.Email,
		About: merchant.About,
	}, nil
}
