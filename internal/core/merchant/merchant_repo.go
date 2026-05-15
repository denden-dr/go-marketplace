package merchant

import (
	"context"
	"database/sql"

	"go-marketplace/internal/domain"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type MerchantRepository interface {
	Create(ctx context.Context, m *domain.Merchant) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Merchant, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Merchant, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]domain.Merchant, error)
	CreateTx(ctx context.Context, tx *sqlx.Tx, m *domain.Merchant) error
	GetPool() domain.Pool
}

type merchantRepository struct {
	db *sqlx.DB
}

func NewMerchantRepository(db *sqlx.DB) MerchantRepository {
	return &merchantRepository{db: db}
}

func (r *merchantRepository) Create(ctx context.Context, m *domain.Merchant) error {
	query := `INSERT INTO merchants (id, user_id, name, about, tax_id, created_at) 
	          VALUES (:id, :user_id, :name, :about, :tax_id, :created_at)`
	_, err := r.db.NamedExecContext(ctx, query, m)
	return err
}

func (r *merchantRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Merchant, error) {
	query := `SELECT id, user_id, name, about, tax_id, created_at FROM merchants WHERE id = $1`
	var m domain.Merchant
	err := r.db.GetContext(ctx, &m, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (r *merchantRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Merchant, error) {
	query := `SELECT id, user_id, name, about, tax_id, created_at FROM merchants WHERE user_id = $1`
	var m domain.Merchant
	err := r.db.GetContext(ctx, &m, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (r *merchantRepository) CreateTx(ctx context.Context, tx *sqlx.Tx, m *domain.Merchant) error {
	query := `INSERT INTO merchants (id, user_id, name, about, tax_id, created_at) 
	          VALUES (:id, :user_id, :name, :about, :tax_id, :created_at)`
	_, err := tx.NamedExecContext(ctx, query, m)
	return err
}

func (r *merchantRepository) GetPool() domain.Pool {
	return r.db
}

func (r *merchantRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]domain.Merchant, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	query := `SELECT id, user_id, name, about, tax_id, created_at FROM merchants WHERE id IN (?)`
	query, args, err := sqlx.In(query, ids)
	if err != nil {
		return nil, err
	}
	query = r.db.Rebind(query)

	var merchants []domain.Merchant
	err = r.db.SelectContext(ctx, &merchants, query, args...)
	return merchants, err
}
