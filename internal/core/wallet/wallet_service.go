package wallet

import (
	"context"
	"time"

	"go-marketplace/internal/domain"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
)

type WalletServiceInterface interface {
	GetWalletByUserID(ctx context.Context, userID uuid.UUID) (*WalletResponse, error)
	GetWalletHistory(ctx context.Context, userID uuid.UUID, page, limit int) ([]TransactionResponse, error)
	Withdraw(ctx context.Context, userID uuid.UUID, req WithdrawRequest) error
	CreateWallet(ctx context.Context, userID uuid.UUID) (*WalletResponse, error)
}

type WalletRepository interface {
	GetWalletByUserID(ctx context.Context, userID uuid.UUID) (*domain.Wallet, error)
	GetWalletHistory(ctx context.Context, walletID uuid.UUID, limit, offset int) ([]domain.WalletTransaction, error)
	Withdraw(ctx context.Context, walletID uuid.UUID, amount decimal.Decimal, txData domain.WalletTransaction) error
	DeductBalanceTX(ctx context.Context, tx *sqlx.Tx, walletID uuid.UUID, amount decimal.Decimal, txData domain.WalletTransaction) error
	AddBalanceTX(ctx context.Context, tx *sqlx.Tx, walletID uuid.UUID, amount decimal.Decimal, txData domain.WalletTransaction) error
	Create(ctx context.Context, w *domain.Wallet) error
	CreateTx(ctx context.Context, tx *sqlx.Tx, w *domain.Wallet) error
	GetPool() domain.Pool
}

type WalletService struct {
	walletRepo WalletRepository
}

func NewWalletService(walletRepo WalletRepository) *WalletService {
	return &WalletService{walletRepo: walletRepo}
}

func (s *WalletService) GetWalletByUserID(ctx context.Context, userID uuid.UUID) (*WalletResponse, error) {
	wallet, err := s.walletRepo.GetWalletByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if wallet == nil {
		return nil, domain.ErrWalletNotFound
	}

	return &WalletResponse{
		ID:           wallet.ID,
		UserID:       wallet.UserID,
		WalletNumber: wallet.WalletNumber,
		Balance:      wallet.Balance,
		Currency:     wallet.Currency,
		Status:       string(wallet.Status),
		CreatedAt:    wallet.CreatedAt,
		UpdatedAt:    wallet.UpdatedAt,
	}, nil
}

func (s *WalletService) GetWalletHistory(ctx context.Context, userID uuid.UUID, page, limit int) ([]TransactionResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	wallet, err := s.walletRepo.GetWalletByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if wallet == nil {
		return nil, domain.ErrWalletNotFound
	}

	transactions, err := s.walletRepo.GetWalletHistory(ctx, wallet.ID, limit, offset)
	if err != nil {
		return nil, err
	}

	res := []TransactionResponse{}
	for _, t := range transactions {
		res = append(res, TransactionResponse{
			ID:           t.ID,
			WalletID:     t.WalletID,
			Amount:       t.Amount,
			Direction:    string(t.Direction),
			Type:         string(t.Type),
			Status:       string(t.Status),
			ReferenceID:  t.ReferenceID,
			BalanceAfter: t.BalanceAfter,
			Description:  t.Description,
			CreatedAt:    t.CreatedAt,
		})
	}
	return res, nil
}

func (s *WalletService) Withdraw(ctx context.Context, userID uuid.UUID, req WithdrawRequest) error {
	wallet, err := s.walletRepo.GetWalletByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if wallet == nil {
		return domain.ErrWalletNotFound
	}

	// Double check status before attempting (Repo will also check)
	if wallet.Status != domain.WalletStatusActive {
		return domain.ErrWalletNotActive
	}

	// Fast-fail if balance is clearly insufficient (Repo will also check atomically)
	if wallet.Balance.LessThan(req.Amount) {
		return domain.ErrInsufficientBalance
	}

	txData := domain.WalletTransaction{
		ID:          uuid.New(),
		WalletID:    wallet.ID,
		Amount:      req.Amount,
		Direction:   domain.TransactionDirectionOut,
		Type:        domain.TransactionTypeWithdraw,
		Status:      domain.TransactionStatusSuccess,
		ReferenceID: "WD-" + time.Now().Format("20060102150405") + "-" + uuid.New().String()[:8],
		Description: req.Description,
		CreatedAt:   time.Now(),
	}

	// BalanceAfter will be set by the repo during the atomic transaction
	return s.walletRepo.Withdraw(ctx, wallet.ID, req.Amount, txData)
}

func (s *WalletService) CreateWallet(ctx context.Context, userID uuid.UUID) (*WalletResponse, error) {
	// Check if wallet already exists
	existing, err := s.walletRepo.GetWalletByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, domain.ErrWalletAlreadyExists
	}

	// Generate wallet number: WAL- + first 8 characters of UUID
	walletNumber := "WAL-" + uuid.New().String()[:8]

	wallet := &domain.Wallet{
		ID:           uuid.New(),
		UserID:       userID,
		WalletNumber: walletNumber,
		Balance:      decimal.NewFromInt(0),
		Currency:     "IDR",
		Status:       domain.WalletStatusActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.walletRepo.Create(ctx, wallet); err != nil {
		return nil, err
	}

	return &WalletResponse{
		ID:           wallet.ID,
		UserID:       wallet.UserID,
		WalletNumber: wallet.WalletNumber,
		Balance:      wallet.Balance,
		Currency:     wallet.Currency,
		Status:       string(wallet.Status),
		CreatedAt:    wallet.CreatedAt,
		UpdatedAt:    wallet.UpdatedAt,
	}, nil
}
