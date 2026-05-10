package payment

import (
	"context"
	"database/sql"
	"go-marketplace/internal/domain"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type PaymentRepository interface {
	CreateTX(ctx context.Context, tx *sqlx.Tx, p *domain.Payment) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error)
	GetByExternalID(ctx context.Context, externalID string) (*domain.Payment, error)
	UpdateStatusTX(ctx context.Context, tx *sqlx.Tx, id uuid.UUID, status domain.PaymentStatus, externalID *string) error
	CreateDistributionTX(ctx context.Context, tx *sqlx.Tx, d *domain.PaymentDistribution) error
	GetDistributionsByPaymentID(ctx context.Context, paymentID uuid.UUID) ([]domain.PaymentDistribution, error)
}

type paymentRepository struct {
	db *sqlx.DB
}

func NewPaymentRepository(db *sqlx.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) CreateTX(ctx context.Context, tx *sqlx.Tx, p *domain.Payment) error {
	query := `INSERT INTO payments (id, user_id, external_id, amount, payment_type, payment_method, status, reference_id, snap_token, created_at, updated_at) 
	          VALUES (:id, :user_id, :external_id, :amount, :payment_type, :payment_method, :status, :reference_id, :snap_token, :created_at, :updated_at)`
	_, err := tx.NamedExecContext(ctx, query, p)
	return err
}

func (r *paymentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error) {
	query := `SELECT id, user_id, external_id, amount, payment_type, payment_method, status, reference_id, snap_token, created_at, updated_at 
	          FROM payments WHERE id = $1`
	var p domain.Payment
	err := r.db.GetContext(ctx, &p, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *paymentRepository) GetByExternalID(ctx context.Context, externalID string) (*domain.Payment, error) {
	query := `SELECT id, user_id, external_id, amount, payment_type, payment_method, status, reference_id, snap_token, created_at, updated_at 
	          FROM payments WHERE external_id = $1`
	var p domain.Payment
	err := r.db.GetContext(ctx, &p, query, externalID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *paymentRepository) UpdateStatusTX(ctx context.Context, tx *sqlx.Tx, id uuid.UUID, status domain.PaymentStatus, externalID *string) error {
	query := `UPDATE payments SET status = $1, external_id = COALESCE($2, external_id), updated_at = NOW() WHERE id = $3`
	_, err := tx.ExecContext(ctx, query, status, externalID, id)
	return err
}

func (r *paymentRepository) CreateDistributionTX(ctx context.Context, tx *sqlx.Tx, d *domain.PaymentDistribution) error {
	query := `INSERT INTO payment_distributions (id, payment_id, recipient_id, amount, created_at) 
	          VALUES (:id, :payment_id, :recipient_id, :amount, :created_at)`
	_, err := tx.NamedExecContext(ctx, query, d)
	return err
}

func (r *paymentRepository) GetDistributionsByPaymentID(ctx context.Context, paymentID uuid.UUID) ([]domain.PaymentDistribution, error) {
	query := `SELECT id, payment_id, recipient_id, amount, created_at FROM payment_distributions WHERE payment_id = $1`
	var distributions []domain.PaymentDistribution
	err := r.db.SelectContext(ctx, &distributions, query, paymentID)
	return distributions, err
}
