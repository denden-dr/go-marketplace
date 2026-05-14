package order

import (
	"context"
	"database/sql"

	"go-marketplace/internal/domain"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"fmt"
)

type OrderRepository interface {
	Begin(ctx context.Context) (*sqlx.Tx, error)
	CreateOrderTX(ctx context.Context, tx *sqlx.Tx, o *domain.Order) error
	CreateOrderItemTX(ctx context.Context, tx *sqlx.Tx, item *domain.OrderItem) error
	GetOrderByID(ctx context.Context, id uuid.UUID) (*domain.Order, error)
	GetOrderByIDForUpdateTX(ctx context.Context, tx *sqlx.Tx, id uuid.UUID) (*domain.Order, error)
	UpdateOrderStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus) error
	UpdateOrderStatusTX(ctx context.Context, tx *sqlx.Tx, id uuid.UUID, status domain.OrderStatus) error
	CreateAppeal(ctx context.Context, appeal *domain.CancellationAppeal) error
	UpdateOrderAppealTX(ctx context.Context, tx *sqlx.Tx, orderID uuid.UUID, isAppealed bool) error
	GetOrderItems(ctx context.Context, orderID uuid.UUID) ([]domain.OrderItem, error)
	GetOrdersByPaymentID(ctx context.Context, paymentID uuid.UUID) ([]domain.Order, error)
	GetOrdersByPaymentIDForUpdateTX(ctx context.Context, tx *sqlx.Tx, paymentID uuid.UUID) ([]domain.Order, error)
	UpdateOrderStatusByPaymentIDTX(ctx context.Context, tx *sqlx.Tx, paymentID uuid.UUID, status domain.OrderStatus) error
	GetOrderItemsByOrderIDsTX(ctx context.Context, tx *sqlx.Tx, orderIDs []uuid.UUID) ([]domain.OrderItem, error)
	CreateOrderItemsBatchTX(ctx context.Context, tx *sqlx.Tx, items []domain.OrderItem) error
}

type orderRepository struct {
	db *sqlx.DB
}

func NewOrderRepository(db *sqlx.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) CreateOrderTX(ctx context.Context, tx *sqlx.Tx, o *domain.Order) error {
	query := `INSERT INTO orders (id, payment_id, merchant_id, user_id, status, total_amount, 
	          shipping_recipient_name, shipping_phone_number, shipping_street_address, 
	          shipping_city, shipping_province, shipping_postal_code, is_appealed, created_at, updated_at) 
	          VALUES (:id, :payment_id, :merchant_id, :user_id, :status, :total_amount, 
	          :shipping_recipient_name, :shipping_phone_number, :shipping_street_address, 
	          :shipping_city, :shipping_province, :shipping_postal_code, :is_appealed, :created_at, :updated_at)`
	_, err := tx.NamedExecContext(ctx, query, o)
	return err
}

func (r *orderRepository) CreateOrderItemTX(ctx context.Context, tx *sqlx.Tx, item *domain.OrderItem) error {
	query := `INSERT INTO order_items (id, order_id, product_id, quantity, price, created_at) 
	          VALUES (:id, :order_id, :product_id, :quantity, :price, :created_at)`
	_, err := tx.NamedExecContext(ctx, query, item)
	return err
}

func (r *orderRepository) GetOrderByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	query := `SELECT id, payment_id, merchant_id, user_id, status, total_amount, 
	          shipping_recipient_name, shipping_phone_number, shipping_street_address, 
	          shipping_city, shipping_province, shipping_postal_code, is_appealed, created_at, updated_at 
	          FROM orders WHERE id = $1`
	var o domain.Order
	err := r.db.GetContext(ctx, &o, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &o, nil
}

func (r *orderRepository) GetOrderByIDForUpdateTX(ctx context.Context, tx *sqlx.Tx, id uuid.UUID) (*domain.Order, error) {
	query := `SELECT id, payment_id, merchant_id, user_id, status, total_amount, 
	          shipping_recipient_name, shipping_phone_number, shipping_street_address, 
	          shipping_city, shipping_province, shipping_postal_code, is_appealed, created_at, updated_at 
	          FROM orders WHERE id = $1 FOR UPDATE`
	var o domain.Order
	err := tx.GetContext(ctx, &o, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &o, nil
}

func (r *orderRepository) UpdateOrderStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus) error {
	query := `UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}

func (r *orderRepository) UpdateOrderStatusTX(ctx context.Context, tx *sqlx.Tx, id uuid.UUID, status domain.OrderStatus) error {
	query := `UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := tx.ExecContext(ctx, query, status, id)
	return err
}

func (r *orderRepository) CreateAppeal(ctx context.Context, appeal *domain.CancellationAppeal) error {
	query := `INSERT INTO cancellation_appeals (id, order_id, reason, status, created_at) 
	          VALUES (:id, :order_id, :reason, :status, :created_at)`
	_, err := r.db.NamedExecContext(ctx, query, appeal)
	return err
}

func (r *orderRepository) UpdateOrderAppealTX(ctx context.Context, tx *sqlx.Tx, orderID uuid.UUID, isAppealed bool) error {
	query := `UPDATE orders SET is_appealed = $1, updated_at = NOW() WHERE id = $2`
	_, err := tx.ExecContext(ctx, query, isAppealed, orderID)
	return err
}

func (r *orderRepository) GetOrderItems(ctx context.Context, orderID uuid.UUID) ([]domain.OrderItem, error) {
	query := `SELECT id, order_id, product_id, quantity, price, created_at FROM order_items WHERE order_id = $1`
	var items []domain.OrderItem
	err := r.db.SelectContext(ctx, &items, query, orderID)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (r *orderRepository) Begin(ctx context.Context) (*sqlx.Tx, error) {
	return r.db.BeginTxx(ctx, nil)
}

func (r *orderRepository) GetOrdersByPaymentID(ctx context.Context, paymentID uuid.UUID) ([]domain.Order, error) {
	query := `SELECT id, payment_id, merchant_id, user_id, status, total_amount, 
	          shipping_recipient_name, shipping_phone_number, shipping_street_address, 
	          shipping_city, shipping_province, shipping_postal_code, is_appealed, created_at, updated_at 
	          FROM orders WHERE payment_id = $1`
	var orders []domain.Order
	err := r.db.SelectContext(ctx, &orders, query, paymentID)
	return orders, err
}
func (r *orderRepository) GetOrdersByPaymentIDForUpdateTX(ctx context.Context, tx *sqlx.Tx, paymentID uuid.UUID) ([]domain.Order, error) {
	query := `SELECT id, payment_id, merchant_id, user_id, status, total_amount, 
	          shipping_recipient_name, shipping_phone_number, shipping_street_address, 
	          shipping_city, shipping_province, shipping_postal_code, is_appealed, created_at, updated_at 
	          FROM orders WHERE payment_id = $1 ORDER BY id FOR UPDATE`
	var orders []domain.Order
	err := tx.SelectContext(ctx, &orders, query, paymentID)
	return orders, err
}

func (r *orderRepository) UpdateOrderStatusByPaymentIDTX(ctx context.Context, tx *sqlx.Tx, paymentID uuid.UUID, status domain.OrderStatus) error {
	query := `UPDATE orders SET status = $1, updated_at = NOW() WHERE payment_id = $2`
	_, err := tx.ExecContext(ctx, query, status, paymentID)
	return err
}

func (r *orderRepository) GetOrderItemsByOrderIDsTX(ctx context.Context, tx *sqlx.Tx, orderIDs []uuid.UUID) ([]domain.OrderItem, error) {
	if len(orderIDs) == 0 {
		return nil, nil
	}

	query := `SELECT id, order_id, product_id, quantity, price, created_at
	           FROM order_items WHERE order_id IN (?)`

	// sqlx.In expands the slice into individual placeholders
	query, args, err := sqlx.In(query, orderIDs)
	if err != nil {
		return nil, err
	}
	// Rebind to PostgreSQL placeholder style ($1, $2, ...)
	query = tx.Rebind(query)

	var items []domain.OrderItem
	err = tx.SelectContext(ctx, &items, query, args...)
	return items, err
}

func (r *orderRepository) CreateOrderItemsBatchTX(ctx context.Context, tx *sqlx.Tx, items []domain.OrderItem) error {
	if len(items) == 0 {
		return nil
	}

	query := `INSERT INTO order_items (id, order_id, product_id, quantity, price, created_at) VALUES `
	args := []interface{}{}
	i := 1
	for idx, item := range items {
		if idx > 0 {
			query += ", "
		}
		query += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d)", i, i+1, i+2, i+3, i+4, i+5)
		args = append(args, item.ID, item.OrderID, item.ProductID, item.Quantity, item.Price, item.CreatedAt)
		i += 6
	}

	_, err := tx.ExecContext(ctx, query, args...)
	return err
}
