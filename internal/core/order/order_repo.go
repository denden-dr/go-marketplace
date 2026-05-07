package order

import (
	"context"

	"go-marketplace/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type orderRepository struct {
	db *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) CreateOrderPaymentTX(ctx context.Context, tx pgx.Tx, p *domain.OrderPayment) error {
	query := `INSERT INTO order_payments (id, user_id, amount, payment_method, status, created_at) 
	          VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := tx.Exec(ctx, query, p.ID, p.UserID, p.Amount, p.PaymentMethod, p.Status, p.CreatedAt)
	return err
}

func (r *orderRepository) CreateOrderTX(ctx context.Context, tx pgx.Tx, o *domain.Order) error {
	query := `INSERT INTO orders (id, payment_id, merchant_id, user_id, status, total_amount, 
	          shipping_recipient_name, shipping_phone_number, shipping_street_address, 
	          shipping_city, shipping_province, shipping_postal_code, is_appealed, created_at, updated_at) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`
	_, err := tx.Exec(ctx, query, o.ID, o.PaymentID, o.MerchantID, o.UserID, o.Status, o.TotalAmount,
		o.ShippingRecipientName, o.ShippingPhoneNumber, o.ShippingStreetAddress,
		o.ShippingCity, o.ShippingProvince, o.ShippingPostalCode, o.IsAppealed, o.CreatedAt, o.UpdatedAt)
	return err
}

func (r *orderRepository) CreateOrderItemTX(ctx context.Context, tx pgx.Tx, item *domain.OrderItem) error {
	query := `INSERT INTO order_items (id, order_id, product_id, quantity, price, created_at) 
	          VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := tx.Exec(ctx, query, item.ID, item.OrderID, item.ProductID, item.Quantity, item.Price, item.CreatedAt)
	return err
}

func (r *orderRepository) GetOrderByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	query := `SELECT id, payment_id, merchant_id, user_id, status, total_amount, 
	          shipping_recipient_name, shipping_phone_number, shipping_street_address, 
	          shipping_city, shipping_province, shipping_postal_code, is_appealed, created_at, updated_at 
	          FROM orders WHERE id = $1`
	var o domain.Order
	err := r.db.QueryRow(ctx, query, id).Scan(
		&o.ID, &o.PaymentID, &o.MerchantID, &o.UserID, &o.Status, &o.TotalAmount,
		&o.ShippingRecipientName, &o.ShippingPhoneNumber, &o.ShippingStreetAddress,
		&o.ShippingCity, &o.ShippingProvince, &o.ShippingPostalCode,
		&o.IsAppealed, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &o, nil
}

func (r *orderRepository) UpdateOrderStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus) error {
	query := `UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(ctx, query, status, id)
	return err
}

func (r *orderRepository) UpdateOrderStatusTX(ctx context.Context, tx pgx.Tx, id uuid.UUID, status domain.OrderStatus) error {
	query := `UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := tx.Exec(ctx, query, status, id)
	return err
}

func (r *orderRepository) CreateAppeal(ctx context.Context, appeal *domain.CancellationAppeal) error {
	query := `INSERT INTO cancellation_appeals (id, order_id, reason, status, created_at) 
	          VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(ctx, query, appeal.ID, appeal.OrderID, appeal.Reason, appeal.Status, appeal.CreatedAt)
	return err
}

func (r *orderRepository) UpdateOrderAppealTX(ctx context.Context, tx pgx.Tx, orderID uuid.UUID, isAppealed bool) error {
	query := `UPDATE orders SET is_appealed = $1, updated_at = NOW() WHERE id = $2`
	_, err := tx.Exec(ctx, query, isAppealed, orderID)
	return err
}

func (r *orderRepository) GetOrderItems(ctx context.Context, orderID uuid.UUID) ([]domain.OrderItem, error) {
	query := `SELECT id, order_id, product_id, quantity, price, created_at FROM order_items WHERE order_id = $1`
	rows, err := r.db.Query(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []domain.OrderItem{}
	for rows.Next() {
		var item domain.OrderItem
		err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.Quantity, &item.Price, &item.CreatedAt)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *orderRepository) Begin(ctx context.Context) (pgx.Tx, error) {
	return r.db.Begin(ctx)
}
