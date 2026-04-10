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

type WalletService struct {
	walletRepo *repos.WalletRepository
}

func NewWalletService(walletRepo *repos.WalletRepository) *WalletService {
	return &WalletService{walletRepo: walletRepo}
}

func (s *WalletService) GetWalletByUserID(ctx context.Context, userID uuid.UUID) (*dtos.WalletResponse, error) {
	wallet, err := s.walletRepo.GetWalletByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if wallet == nil {
		return nil, errors.New("wallet not found")
	}

	return &dtos.WalletResponse{
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

func (s *WalletService) GetWalletHistory(ctx context.Context, userID uuid.UUID) ([]dtos.TransactionResponse, error) {
	wallet, err := s.walletRepo.GetWalletByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if wallet == nil {
		return nil, errors.New("wallet not found")
	}

	transactions, err := s.walletRepo.GetWalletHistory(ctx, wallet.ID)
	if err != nil {
		return nil, err
	}

	var res []dtos.TransactionResponse
	for _, t := range transactions {
		res = append(res, dtos.TransactionResponse{
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

func (s *WalletService) Withdraw(ctx context.Context, userID uuid.UUID, req dtos.WithdrawRequest) error {
	wallet, err := s.walletRepo.GetWalletByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if wallet == nil {
		return errors.New("wallet not found")
	}

	if wallet.Balance < req.Amount {
		return errors.New("insufficient balance")
	}

	txData := domain.WalletTransaction{
		ID:           uuid.New(),
		WalletID:     wallet.ID,
		Amount:       req.Amount,
		Direction:    domain.TransactionDirectionOut,
		Type:         domain.TransactionTypeWithdraw,
		Status:       domain.TransactionStatusSuccess,
		ReferenceID:  "WD-" + time.Now().Format("20060102150405"),
		BalanceAfter: wallet.Balance - req.Amount,
		Description:  req.Description,
		CreatedAt:    time.Now(),
	}

	return s.walletRepo.Withdraw(ctx, wallet.ID, req.Amount, txData)
}
