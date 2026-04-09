package repos

import (
	"context"

	"go-shop-yourself/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MerchantRepository struct {
	db *pgxpool.Pool
}

func NewMerchantRepository(db *pgxpool.Pool) *MerchantRepository {
	return &MerchantRepository{db: db}
}

func (r *MerchantRepository) Create(ctx context.Context, m *domain.Merchant) error {
	query := `INSERT INTO merchants (id, user_id, name, about, tax_id, created_at) 
	          VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(ctx, query, m.ID, m.UserID, m.Name, m.About, m.TaxID, m.CreatedAt)
	return err
}

func (r *MerchantRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Merchant, error) {
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

func (r *MerchantRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Merchant, error) {
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
