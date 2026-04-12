package merchant

import (
	"context"
	"time"

	"go-shop-yourself/internal/domain"

	"go-shop-yourself/internal/user"
	"go-shop-yourself/internal/wallet"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

type MerchantServiceInterface interface {
	RegisterMerchant(ctx context.Context, userID uuid.UUID, req MerchantRegisterRequest) (*MerchantResponse, error)
}

type MerchantRepository interface {
	Create(ctx context.Context, m *domain.Merchant) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Merchant, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Merchant, error)
	CreateTx(ctx context.Context, tx pgx.Tx, m *domain.Merchant) error
	GetPool() domain.Pool
}

type MerchantService struct {
	repo MerchantRepository
	userRepo user.UserRepository
	walletRepo wallet.WalletRepository
}

func NewMerchantService(repo MerchantRepository, userRepo user.UserRepository, walletRepo wallet.WalletRepository) *MerchantService {
	return &MerchantService{repo: repo, userRepo: userRepo, walletRepo: walletRepo}
}

func (s *MerchantService) RegisterMerchant(ctx context.Context, userID uuid.UUID, req MerchantRegisterRequest) (*MerchantResponse, error) {
	// Check if user exists
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}

	// Check if merchant profile already exists for this user
	existing, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, domain.ErrMerchantAlreadyExists
	}

	merchant := &domain.Merchant{
		ID: uuid.New(),
		UserID: userID,
		Name: req.Name,
		About: req.About,
		TaxID: req.TaxID,
		CreatedAt: time.Now(),
	}

	// Prepare wallet data
	wallet := &domain.Wallet{
		ID: uuid.New(),
		UserID: userID,
		WalletNumber: "WAL-" + uuid.New().String()[:8],
		Balance: decimal.NewFromInt(0),
		Currency: "IDR",
		Status: domain.WalletStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
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

	return &MerchantResponse{
		ID: merchant.ID,
		Name: merchant.Name,
		Email: user.Email,
		About: merchant.About,
	}, nil
}
