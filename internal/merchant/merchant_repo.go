package merchant

import (
	"context"

	"go-shop-yourself/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type merchantRepository struct {
	db *pgxpool.Pool
}

func NewMerchantRepository(db *pgxpool.Pool) MerchantRepository {
	return &merchantRepository{db: db}
}

func (r *merchantRepository) Create(ctx context.Context, m *domain.Merchant) error {
	query := `INSERT INTO merchants (id, user_id, name, about, tax_id, created_at) 
	          VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(ctx, query, m.ID, m.UserID, m.Name, m.About, m.TaxID, m.CreatedAt)
	return err
}

func (r *merchantRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Merchant, error) {
	query := `SELECT id, user_id, name, about, tax_id, created_at FROM merchants WHERE id = $1`
	var m domain.Merchant
	err := r.db.QueryRow(ctx, query, id).Scan(&m.ID, &m.UserID, &m.Name, &m.About, &m.TaxID, &m.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (r *merchantRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Merchant, error) {
	query := `SELECT id, user_id, name, about, tax_id, created_at FROM merchants WHERE user_id = $1`
	var m domain.Merchant
	err := r.db.QueryRow(ctx, query, userID).Scan(&m.ID, &m.UserID, &m.Name, &m.About, &m.TaxID, &m.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (r *merchantRepository) CreateTx(ctx context.Context, tx pgx.Tx, m *domain.Merchant) error {
	query := `INSERT INTO merchants (id, user_id, name, about, tax_id, created_at) 
	          VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := tx.Exec(ctx, query, m.ID, m.UserID, m.Name, m.About, m.TaxID, m.CreatedAt)
	return err
}

func (r *merchantRepository) GetPool() domain.Pool {
	return r.db
}
