package repos

import (
	"context"

	"go-shop-yourself/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductRepository struct {
	db *pgxpool.Pool
}

func NewProductRepository(db *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Create(ctx context.Context, p *domain.Product) error {
	query := `INSERT INTO products (id, store_id, name, description, price, stock, height_cm, width_cm, depth_cm, weight_kg, is_onsale, created_at) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	_, err := r.db.Exec(ctx, query, p.ID, p.StoreID, p.Name, p.Description, p.Price, p.Stock, p.HeightCM, p.WidthCM, p.DepthCM, p.WeightKG, p.IsOnSale, p.CreatedAt)
	return err
}

func (r *ProductRepository) Update(ctx context.Context, p *domain.Product) error {
	query := `UPDATE products SET name = $1, description = $2, price = $3, stock = $4, height_cm = $5, width_cm = $6, depth_cm = $7, weight_kg = $8, is_onsale = $9 WHERE id = $10`
	_, err := r.db.Exec(ctx, query, p.Name, p.Description, p.Price, p.Stock, p.HeightCM, p.WidthCM, p.DepthCM, p.WeightKG, p.IsOnSale, p.ID)
	return err
}

func (r *ProductRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	query := `SELECT id, store_id, name, description, price, stock, height_cm, width_cm, depth_cm, weight_kg, is_onsale, created_at FROM products WHERE id = $1`
	var p domain.Product
	err := r.db.QueryRow(ctx, query, id).Scan(&p.ID, &p.StoreID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.HeightCM, &p.WidthCM, &p.DepthCM, &p.WeightKG, &p.IsOnSale, &p.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}
