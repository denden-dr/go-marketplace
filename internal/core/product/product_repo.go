package product

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"

	"go-marketplace/internal/domain"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ProductRepository interface {
	Create(ctx context.Context, p *domain.Product) error
	Update(ctx context.Context, p *domain.Product) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Product, error)
	GetByIDForUpdateTX(ctx context.Context, tx *sqlx.Tx, id uuid.UUID) (*domain.Product, error)
	UpdateStockTX(ctx context.Context, tx *sqlx.Tx, id uuid.UUID, stock int) error
	RestoreStockBatchTX(ctx context.Context, tx *sqlx.Tx, items []domain.OrderItem) error
	DeductStockBatchTX(ctx context.Context, tx *sqlx.Tx, items []domain.OrderItem) error
	Search(ctx context.Context, query string, limit, offset int) ([]domain.Product, error)
	GetByIDsForUpdateTX(ctx context.Context, tx *sqlx.Tx, ids []uuid.UUID) ([]domain.Product, error)
}

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

func (r *productRepository) RestoreStockBatchTX(ctx context.Context, tx *sqlx.Tx, items []domain.OrderItem) error {
	if len(items) == 0 {
		return nil
	}

	quantities := make(map[uuid.UUID]int)
	var productIDs []string
	for _, item := range items {
		if _, exists := quantities[item.ProductID]; !exists {
			productIDs = append(productIDs, item.ProductID.String())
		}
		quantities[item.ProductID] += item.Quantity
	}

	sort.Strings(productIDs)

	var sb strings.Builder
	sb.WriteString("UPDATE products SET stock = stock + v.quantity FROM (VALUES ")

	args := make([]interface{}, 0, len(productIDs)*2)
	for idx, idStr := range productIDs {
		id := uuid.MustParse(idStr)
		qty := quantities[id]
		if idx > 0 {
			sb.WriteString(", ")
		}
		p := idx * 2
		fmt.Fprintf(&sb, "($%d::uuid, $%d::int)", p+1, p+2)
		args = append(args, id, qty)
	}
	sb.WriteString(") AS v(id, quantity) WHERE products.id = v.id")

	_, err := tx.ExecContext(ctx, sb.String(), args...)
	return err
}

func (r *productRepository) DeductStockBatchTX(ctx context.Context, tx *sqlx.Tx, items []domain.OrderItem) error {
	if len(items) == 0 {
		return nil
	}

	quantities := make(map[uuid.UUID]int)
	var productIDs []string
	for _, item := range items {
		if _, exists := quantities[item.ProductID]; !exists {
			productIDs = append(productIDs, item.ProductID.String())
		}
		quantities[item.ProductID] += item.Quantity
	}

	sort.Strings(productIDs)

	var sb strings.Builder
	sb.WriteString("UPDATE products SET stock = stock - v.quantity FROM (VALUES ")

	args := make([]interface{}, 0, len(productIDs)*2)
	for idx, idStr := range productIDs {
		id := uuid.MustParse(idStr)
		qty := quantities[id]
		if idx > 0 {
			sb.WriteString(", ")
		}
		p := idx * 2
		fmt.Fprintf(&sb, "($%d::uuid, $%d::int)", p+1, p+2)
		args = append(args, id, qty)
	}
	sb.WriteString(") AS v(id, quantity) WHERE products.id = v.id")

	_, err := tx.ExecContext(ctx, sb.String(), args...)
	return err
}

func (r *productRepository) GetByIDsForUpdateTX(ctx context.Context, tx *sqlx.Tx, ids []uuid.UUID) ([]domain.Product, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	query := `SELECT id, store_id, name, description, price, stock, height_cm, width_cm, depth_cm, weight_kg, is_onsale, created_at
	           FROM products WHERE id IN (?) ORDER BY id FOR UPDATE`

	query, args, err := sqlx.In(query, ids)
	if err != nil {
		return nil, err
	}
	query = tx.Rebind(query)

	var products []domain.Product
	err = tx.SelectContext(ctx, &products, query, args...)
	return products, err
}
