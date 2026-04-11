package services

import (
	"context"
	"errors"
	"time"

	"go-shop-yourself/internal/domain"
	"go-shop-yourself/internal/dtos"
	"go-shop-yourself/internal/repos"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type MerchantService struct {
	repo       *repos.MerchantRepository
	userRepo   *repos.UserRepository
	walletRepo *repos.WalletRepository
}

func NewMerchantService(repo *repos.MerchantRepository, userRepo *repos.UserRepository, walletRepo *repos.WalletRepository) *MerchantService {
	return &MerchantService{repo: repo, userRepo: userRepo, walletRepo: walletRepo}
}

func (s *MerchantService) RegisterMerchant(ctx context.Context, userID uuid.UUID, req dtos.MerchantRegisterRequest) (*dtos.MerchantResponse, error) {
	// Check if user exists
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Check if merchant profile already exists for this user
	existing, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("merchant already exists for this user")
	}

	merchant := &domain.Merchant{
		ID:        uuid.New(),
		UserID:    userID,
		Name:      req.Name,
		About:     req.About,
		TaxID:     req.TaxID,
		CreatedAt: time.Now(),
	}

	// Prepare wallet data
	wallet := &domain.Wallet{
		ID:           uuid.New(),
		UserID:       userID,
		WalletNumber: "WAL-" + uuid.New().String()[:8],
		Balance:      decimal.NewFromInt(0),
		Currency:     "IDR",
		Status:       domain.WalletStatusActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Start transaction
	tx, err := s.repo.GetPool().Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Create Merchant
	if err := s.repo.CreateTx(ctx, tx, merchant); err != nil {
		return nil, err
	}

	// Create Wallet
	if err := s.walletRepo.CreateTx(ctx, tx, wallet); err != nil {
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &dtos.MerchantResponse{
		ID:    merchant.ID,
		Name:  merchant.Name,
		Email: user.Email,
		About: merchant.About,
	}, nil
}
