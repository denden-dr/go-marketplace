package product

import (
	"context"
	"database/sql"

	"go-marketplace/internal/domain"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type productRepository struct {
	db *sqlx.DB
}

func NewProductRepository(db *sqlx.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) Create(ctx context.Context, p *domain.Product) error {
	query := `INSERT INTO products (id, store_id, name, description, price, stock, height_cm, width_cm, depth_cm, weight_kg, is_onsale, created_at) 
	          VALUES (:id, :store_id, :name, :description, :price, :stock, :height_cm, :width_cm, :depth_cm, :weight_kg, :is_onsale, :created_at)`
	_, err := r.db.NamedExecContext(ctx, query, p)
	return err
}

func (r *productRepository) Update(ctx context.Context, p *domain.Product) error {
	query := `UPDATE products SET name = :name, description = :description, price = :price, stock = :stock, height_cm = :height_cm, width_cm = :width_cm, depth_cm = :depth_cm, weight_kg = :weight_kg, is_onsale = :is_onsale WHERE id = :id`
	_, err := r.db.NamedExecContext(ctx, query, p)
	return err
}

func (r *productRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	query := `SELECT id, store_id, name, description, price, stock, height_cm, width_cm, depth_cm, weight_kg, is_onsale, created_at FROM products WHERE id = $1`
	var p domain.Product
	err := r.db.GetContext(ctx, &p, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *productRepository) GetByIDForUpdateTX(ctx context.Context, tx *sqlx.Tx, id uuid.UUID) (*domain.Product, error) {
	query := `SELECT id, store_id, name, description, price, stock, height_cm, width_cm, depth_cm, weight_kg, is_onsale, created_at FROM products WHERE id = $1 FOR UPDATE`
	var p domain.Product
	err := tx.GetContext(ctx, &p, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *productRepository) UpdateStockTX(ctx context.Context, tx *sqlx.Tx, id uuid.UUID, stock int) error {
	query := `UPDATE products SET stock = $1 WHERE id = $2`
	_, err := tx.ExecContext(ctx, query, stock, id)
	return err
}

func (r *productRepository) Search(ctx context.Context, query string, limit, offset int) ([]domain.Product, error) {
	sqlQuery := `
		SELECT id, store_id, name, description, price, stock, height_cm, width_cm, depth_cm, weight_kg, is_onsale, created_at 
		FROM products 
		WHERE ($1 = '' OR search_vector @@ websearch_to_tsquery('english', $1) OR name % $1)
		ORDER BY 
			(CASE WHEN $1 = '' THEN 0 ELSE ts_rank(search_vector, websearch_to_tsquery('english', $1)) END) DESC,
			(CASE WHEN $1 = '' THEN 0 ELSE similarity(name, $1) END) DESC,
			created_at DESC
		LIMIT $2 OFFSET $3`

	var products []domain.Product
	err := r.db.SelectContext(ctx, &products, sqlQuery, query, limit, offset)
	if err != nil {
		return nil, err
	}

	return products, nil
}
