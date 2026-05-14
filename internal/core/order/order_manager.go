package order

import (
	"context"
	"log"

	"go-marketplace/internal/core/payment"
	"go-marketplace/internal/core/product"
	"go-marketplace/internal/domain"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type orderManager struct {
	orderRepo   OrderRepository
	productRepo product.ProductRepository
}

var _ payment.OrderManager = (*orderManager)(nil)

// NewOrderManager creates an OrderManager that handles payment status
// changes for orders. The returned value satisfies payment.OrderManager.
func NewOrderManager(orderRepo OrderRepository, productRepo product.ProductRepository) payment.OrderManager {
	return &orderManager{
		orderRepo:   orderRepo,
		productRepo: productRepo,
	}
}

// HandlePaymentStatusChangeTX updates order statuses when a payment
// status changes (e.g., webhook callback). On success it moves orders
// to processing; on failure/expiry it cancels orders and restores stock.
func (m *orderManager) HandlePaymentStatusChangeTX(ctx context.Context, tx *sqlx.Tx, paymentID uuid.UUID, status domain.PaymentStatus) error {
	switch status {
	case domain.PaymentStatusFailed, domain.PaymentStatusExpired:
		orders, err := m.orderRepo.GetOrdersByPaymentIDForUpdateTX(ctx, tx, paymentID)
		if err != nil {
			return err
		}

		var allItems []domain.OrderItem
		for _, o := range orders {
			if err := m.orderRepo.UpdateOrderStatusTX(ctx, tx, o.ID, domain.OrderStatusCancelled); err != nil {
				return err
			}

			items, err := m.orderRepo.GetOrderItems(ctx, o.ID)
			if err != nil {
				return err
			}
			allItems = append(allItems, items...)
		}

		if len(allItems) > 0 {
			if err := m.productRepo.RestoreStockBatchTX(ctx, tx, allItems); err != nil {
				return err
			}
		}
	case domain.PaymentStatusSuccess:
		orders, err := m.orderRepo.GetOrdersByPaymentIDForUpdateTX(ctx, tx, paymentID)
		if err != nil {
			return err
		}
		for _, o := range orders {
			if err := m.orderRepo.UpdateOrderStatusTX(ctx, tx, o.ID, domain.OrderStatusProcessing); err != nil {
				return err
			}
		}
	default:
		log.Printf("[orderManager] Unrecognized or unhandled payment status: %s", status)
	}
	return nil
}
