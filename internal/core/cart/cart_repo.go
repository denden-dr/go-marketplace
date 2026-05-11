package cart

import (
	"context"

	"go-marketplace/internal/domain"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type CartRepository interface {
	UpsertCartItem(ctx context.Context, item *domain.CartItem) error
	UpdateCartItem(ctx context.Context, userID, productID uuid.UUID, quantity int) error
	DeleteCartItem(ctx context.Context, userID, productID uuid.UUID) error
	ClearCart(ctx context.Context, userID uuid.UUID) error
	ClearCartTX(ctx context.Context, tx *sqlx.Tx, userID uuid.UUID) error
	GetCartByUserID(ctx context.Context, userID uuid.UUID) ([]domain.CartItem, error)
}

type cartRepository struct {
	db *sqlx.DB
}

func NewCartRepository(db *sqlx.DB) CartRepository {
	return &cartRepository{db: db}
}

func (r *cartRepository) UpsertCartItem(ctx context.Context, item *domain.CartItem) error {
	query := `
		INSERT INTO cart_items (id, user_id, product_id, quantity, created_at, updated_at)
		VALUES (:id, :user_id, :product_id, :quantity, :created_at, :updated_at)
		ON CONFLICT (user_id, product_id) DO UPDATE SET
			quantity = cart_items.quantity + EXCLUDED.quantity,
			updated_at = EXCLUDED.updated_at
	`
	_, err := r.db.NamedExecContext(ctx, query, item)
	return err
}

func (r *cartRepository) UpdateCartItem(ctx context.Context, userID, productID uuid.UUID, quantity int) error {
	query := `UPDATE cart_items SET quantity = $1, updated_at = NOW() WHERE user_id = $2 AND product_id = $3`
	_, err := r.db.ExecContext(ctx, query, quantity, userID, productID)
	return err
}

func (r *cartRepository) DeleteCartItem(ctx context.Context, userID, productID uuid.UUID) error {
	query := `DELETE FROM cart_items WHERE user_id = $1 AND product_id = $2`
	_, err := r.db.ExecContext(ctx, query, userID, productID)
	return err
}

func (r *cartRepository) ClearCart(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM cart_items WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

func (r *cartRepository) ClearCartTX(ctx context.Context, tx *sqlx.Tx, userID uuid.UUID) error {
	query := `DELETE FROM cart_items WHERE user_id = $1`
	_, err := tx.ExecContext(ctx, query, userID)
	return err
}

func (r *cartRepository) GetCartByUserID(ctx context.Context, userID uuid.UUID) ([]domain.CartItem, error) {
	query := `
		SELECT ci.id, ci.user_id, ci.product_id, ci.quantity, ci.created_at, ci.updated_at,
		       p.name, p.price, p.stock
		FROM cart_items ci
		JOIN products p ON ci.product_id = p.id
		WHERE ci.user_id = $1
	`
	rows, err := r.db.QueryxContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.CartItem
	for rows.Next() {
		var ci domain.CartItem
		var p domain.Product
		err := rows.Scan(
			&ci.ID, &ci.UserID, &ci.ProductID, &ci.Quantity, &ci.CreatedAt, &ci.UpdatedAt,
			&p.Name, &p.Price, &p.Stock,
		)
		if err != nil {
			return nil, err
		}
		ci.Product = &p
		items = append(items, ci)
	}
	return items, nil
}
