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

		if err := m.orderRepo.UpdateOrderStatusByPaymentIDTX(ctx, tx, paymentID, domain.OrderStatusCancelled); err != nil {
			return err
		}

		orderIDs := make([]uuid.UUID, len(orders))
		for i, o := range orders {
			orderIDs[i] = o.ID
		}

		items, err := m.orderRepo.GetOrderItemsByOrderIDsTX(ctx, tx, orderIDs)
		if err != nil {
			return err
		}

		if len(items) > 0 {
			if err := m.productRepo.RestoreStockBatchTX(ctx, tx, items); err != nil {
				return err
			}
		}
	case domain.PaymentStatusSuccess:
		if _, err := m.orderRepo.GetOrdersByPaymentIDForUpdateTX(ctx, tx, paymentID); err != nil {
			return err
		}

		if err := m.orderRepo.UpdateOrderStatusByPaymentIDTX(ctx, tx, paymentID, domain.OrderStatusProcessing); err != nil {
			return err
		}
	default:
		log.Printf("[orderManager] Unrecognized or unhandled payment status: %s", status)
	}
	return nil
}
