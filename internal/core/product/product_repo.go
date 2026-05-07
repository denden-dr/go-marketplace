package product

import (
	"context"

	"go-marketplace/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type productRepository struct {
	db *pgxpool.Pool
}

func NewProductRepository(db *pgxpool.Pool) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) Create(ctx context.Context, p *domain.Product) error {
	query := `INSERT INTO products (id, store_id, name, description, price, stock, height_cm, width_cm, depth_cm, weight_kg, is_onsale, created_at) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	_, err := r.db.Exec(ctx, query, p.ID, p.StoreID, p.Name, p.Description, p.Price, p.Stock, p.HeightCM, p.WidthCM, p.DepthCM, p.WeightKG, p.IsOnSale, p.CreatedAt)
	return err
}

func (r *productRepository) Update(ctx context.Context, p *domain.Product) error {
	query := `UPDATE products SET name = $1, description = $2, price = $3, stock = $4, height_cm = $5, width_cm = $6, depth_cm = $7, weight_kg = $8, is_onsale = $9 WHERE id = $10`
	_, err := r.db.Exec(ctx, query, p.Name, p.Description, p.Price, p.Stock, p.HeightCM, p.WidthCM, p.DepthCM, p.WeightKG, p.IsOnSale, p.ID)
	return err
}

func (r *productRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
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

func (r *productRepository) GetByIDForUpdateTX(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*domain.Product, error) {
	query := `SELECT id, store_id, name, description, price, stock, height_cm, width_cm, depth_cm, weight_kg, is_onsale, created_at FROM products WHERE id = $1 FOR UPDATE`
	var p domain.Product
	err := tx.QueryRow(ctx, query, id).Scan(&p.ID, &p.StoreID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.HeightCM, &p.WidthCM, &p.DepthCM, &p.WeightKG, &p.IsOnSale, &p.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *productRepository) UpdateStockTX(ctx context.Context, tx pgx.Tx, id uuid.UUID, stock int) error {
	query := `UPDATE products SET stock = $1 WHERE id = $2`
	_, err := tx.Exec(ctx, query, stock, id)
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

	rows, err := r.db.Query(ctx, sqlQuery, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []domain.Product
	for rows.Next() {
		var p domain.Product
		err := rows.Scan(&p.ID, &p.StoreID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.HeightCM, &p.WidthCM, &p.DepthCM, &p.WeightKG, &p.IsOnSale, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}
