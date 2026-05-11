package payment

import (
	"context"
	"fmt"
	"time"

	"go-marketplace/internal/core/wallet"
	"go-marketplace/internal/domain"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
)

type CreatePaymentRequest struct {
	UserID        uuid.UUID
	Amount        decimal.Decimal
	Type          domain.PaymentType
	Method        domain.PaymentMethod
	ReferenceID   uuid.UUID
	Distributions []PaymentDistribution // Used for order payments to merchants
}

type PaymentDistribution struct {
	RecipientID uuid.UUID
	Amount      decimal.Decimal
}

type PaymentResponse struct {
	PaymentID uuid.UUID
	Status    domain.PaymentStatus
	SnapToken *string
}

type PaymentServiceInterface interface {
	CreatePaymentTX(ctx context.Context, tx *sqlx.Tx, req CreatePaymentRequest) (*PaymentResponse, error)
	ProcessWebhook(ctx context.Context, externalID string, status domain.PaymentStatus) error
}

type PaymentProvider interface {
	CreateTransaction(ctx context.Context, p *domain.Payment) (string, error)
}

type OrderManager interface {
	HandlePaymentStatusChangeTX(ctx context.Context, tx *sqlx.Tx, paymentID uuid.UUID, status domain.PaymentStatus) error
}

type PaymentService struct {
	paymentRepo   PaymentRepository
	walletService wallet.WalletServiceInterface
	provider      PaymentProvider
	orderManager  OrderManager
	db            *sqlx.DB
}

func NewPaymentService(paymentRepo PaymentRepository, walletService wallet.WalletServiceInterface, provider PaymentProvider, orderManager OrderManager, db *sqlx.DB) *PaymentService {
	return &PaymentService{
		paymentRepo:   paymentRepo,
		walletService: walletService,
		provider:      provider,
		orderManager:  orderManager,
		db:            db,
	}
}

func (s *PaymentService) SetOrderManager(om OrderManager) {
	s.orderManager = om
}

func (s *PaymentService) CreatePaymentTX(ctx context.Context, tx *sqlx.Tx, req CreatePaymentRequest) (*PaymentResponse, error) {
	if tx == nil {
		newTx, err := s.db.BeginTxx(ctx, nil)
		if err != nil {
			return nil, err
		}
		defer func() { _ = newTx.Rollback() }()

		res, err := s.createPaymentInternal(ctx, newTx, req)
		if err != nil {
			return nil, err
		}

		if err := newTx.Commit(); err != nil {
			return nil, err
		}
		return res, nil
	}

	return s.createPaymentInternal(ctx, tx, req)
}

func (s *PaymentService) createPaymentInternal(ctx context.Context, tx *sqlx.Tx, req CreatePaymentRequest) (*PaymentResponse, error) {
	payment := &domain.Payment{
		ID:          uuid.New(),
		UserID:      req.UserID,
		Amount:      req.Amount,
		Type:        req.Type,
		Method:      req.Method,
		Status:      domain.PaymentStatusPending,
		ReferenceID: req.ReferenceID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if req.Method == domain.PaymentMethodWallet {
		// 1. Deduct from User Balance
		txData := domain.WalletTransaction{
			ID:          uuid.New(),
			Amount:      req.Amount,
			Direction:   domain.TransactionDirectionOut,
			Type:        domain.TransactionTypePayment,
			Status:      domain.TransactionStatusSuccess,
			ReferenceID: payment.ID.String(),
			Description: fmt.Sprintf("Payment for %s %s", req.Type, req.ReferenceID),
			CreatedAt:   time.Now(),
		}

		if err := s.walletService.DeductBalanceTX(ctx, tx, req.UserID, req.Amount, txData); err != nil {
			return nil, err
		}

		// 2. Add to Merchants' Pending Balance (if applicable)
		for _, dist := range req.Distributions {
			distTxData := domain.WalletTransaction{
				ID:          uuid.New(),
				Amount:      dist.Amount,
				Direction:   domain.TransactionDirectionIn,
				Type:        domain.TransactionTypePayment,
				Status:      domain.TransactionStatusSuccess,
				ReferenceID: payment.ID.String(),
				Description: fmt.Sprintf("Incoming payment for order %s", req.ReferenceID),
				CreatedAt:   time.Now(),
			}
			if err := s.walletService.AddPendingBalanceTX(ctx, tx, dist.RecipientID, dist.Amount, distTxData); err != nil {
				return nil, err
			}
		}

		payment.Status = domain.PaymentStatusSuccess

		if err := s.paymentRepo.CreateTX(ctx, tx, payment); err != nil {
			return nil, err
		}

		// Persist distributions
		for _, dist := range req.Distributions {
			d := &domain.PaymentDistribution{
				ID:          uuid.New(),
				PaymentID:   payment.ID,
				RecipientID: dist.RecipientID,
				Amount:      dist.Amount,
				CreatedAt:   time.Now(),
			}
			if err := s.paymentRepo.CreateDistributionTX(ctx, tx, d); err != nil {
				return nil, err
			}
		}

		return &PaymentResponse{
			PaymentID: payment.ID,
			Status:    payment.Status,
			SnapToken: payment.SnapToken,
		}, nil
	}

	// Handle External Provider (Midtrans) - Atomic with caller's tx
	if err := s.paymentRepo.CreateTX(ctx, tx, payment); err != nil {
		return nil, err
	}

	for _, dist := range req.Distributions {
		d := &domain.PaymentDistribution{
			ID:          uuid.New(),
			PaymentID:   payment.ID,
			RecipientID: dist.RecipientID,
			Amount:      dist.Amount,
			CreatedAt:   time.Now(),
		}
		if err := s.paymentRepo.CreateDistributionTX(ctx, tx, d); err != nil {
			return nil, err
		}
	}

	// Call External Provider (Holding the transaction, but ensures atomicity)
	snapToken, err := s.provider.CreateTransaction(ctx, payment)
	if err != nil {
		return nil, err
	}

	if err := s.paymentRepo.UpdateSnapTokenTX(ctx, tx, payment.ID, snapToken); err != nil {
		return nil, err
	}

	payment.SnapToken = &snapToken
	return &PaymentResponse{
		PaymentID: payment.ID,
		Status:    payment.Status,
		SnapToken: payment.SnapToken,
	}, nil
}

func (s *PaymentService) ProcessWebhook(ctx context.Context, externalID string, status domain.PaymentStatus) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var p *domain.Payment
	var repoErr error
	if id, err := uuid.Parse(externalID); err == nil {
		p, repoErr = s.paymentRepo.GetByIDForUpdateTX(ctx, tx, id)
	} else {
		p, repoErr = s.paymentRepo.GetByExternalIDForUpdateTX(ctx, tx, externalID)
	}

	if repoErr != nil {
		return repoErr
	}
	if p == nil {
		return domain.ErrPaymentNotFound
	}

	if p.Status != domain.PaymentStatusPending {
		return nil
	}

	if err := s.paymentRepo.UpdateStatusTX(ctx, tx, p.ID, status, nil); err != nil {
		return err
	}

	if status == domain.PaymentStatusSuccess {
		switch p.Type {
		case domain.PaymentTypeOrder:
			distributions, err := s.paymentRepo.GetDistributionsByPaymentID(ctx, p.ID)
			if err != nil {
				return err
			}
			for _, dist := range distributions {
				txData := domain.WalletTransaction{
					ID:          uuid.New(),
					Amount:      dist.Amount,
					Direction:   domain.TransactionDirectionIn,
					Type:        domain.TransactionTypePayment,
					Status:      domain.TransactionStatusSuccess,
					ReferenceID: p.ID.String(),
					Description: fmt.Sprintf("Incoming payment for order %s", p.ReferenceID),
					CreatedAt:   time.Now(),
				}
				if err := s.walletService.AddPendingBalanceTX(ctx, tx, dist.RecipientID, dist.Amount, txData); err != nil {
					return err
				}
			}
		case domain.PaymentTypeTopup:
			txData := domain.WalletTransaction{
				ID:          uuid.New(),
				Amount:      p.Amount,
				Direction:   domain.TransactionDirectionIn,
				Type:        domain.TransactionTypeTopup,
				Status:      domain.TransactionStatusSuccess,
				ReferenceID: p.ID.String(),
				Description: "Top-up success",
				CreatedAt:   time.Now(),
			}
			if err := s.walletService.AddBalanceTX(ctx, tx, p.UserID, p.Amount, txData); err != nil {
				return err
			}
		}
	}

	if status == domain.PaymentStatusFailed || status == domain.PaymentStatusExpired {
		if p.Type == domain.PaymentTypeOrder {
			if err := s.orderManager.HandlePaymentStatusChangeTX(ctx, tx, p.ID, status); err != nil {
				// Log error but continue? Or fail?
				// Usually we want to ensure stock is recovered.
				return err
			}
		}
	}

	return tx.Commit()
}
