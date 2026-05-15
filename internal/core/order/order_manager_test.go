package order

import (
	"context"
	"testing"

	"go-marketplace/internal/core/product"
	"go-marketplace/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOrderManager_HandlePaymentStatusChangeTX_Success(t *testing.T) {
	mockRepo := NewMockOrderRepository(t)
	mockProductRepo := product.NewMockProductRepository(t)
	m := NewOrderManager(mockRepo, mockProductRepo)

	ctx := context.Background()
	paymentID := uuid.New()

	// Batch: get orders for update, then batch update status
	mockRepo.On("GetOrdersByPaymentIDForUpdateTX", ctx, mock.Anything, paymentID).Return([]domain.Order{
		{ID: uuid.New(), Status: domain.OrderStatusPending},
		{ID: uuid.New(), Status: domain.OrderStatusPending},
	}, nil).Once()
	mockRepo.On("UpdateOrderStatusByPaymentIDTX", ctx, mock.Anything, paymentID, domain.OrderStatusProcessing).Return(nil).Once()

	err := m.HandlePaymentStatusChangeTX(ctx, nil, paymentID, domain.PaymentStatusSuccess)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestOrderManager_HandlePaymentStatusChangeTX_Failed(t *testing.T) {
	mockRepo := NewMockOrderRepository(t)
	mockProductRepo := product.NewMockProductRepository(t)
	m := NewOrderManager(mockRepo, mockProductRepo)

	ctx := context.Background()
	paymentID := uuid.New()
	orderID1 := uuid.New()
	orderID2 := uuid.New()
	productID := uuid.New()

	orders := []domain.Order{
		{ID: orderID1, Status: domain.OrderStatusPending},
		{ID: orderID2, Status: domain.OrderStatusPending},
	}
	items := []domain.OrderItem{
		{OrderID: orderID1, ProductID: productID, Quantity: 2},
		{OrderID: orderID2, ProductID: productID, Quantity: 3},
	}

	// Batch: get orders, batch update status, batch get items, batch restore stock
	mockRepo.On("GetOrdersByPaymentIDForUpdateTX", ctx, mock.Anything, paymentID).Return(orders, nil).Once()
	mockRepo.On("UpdateOrderStatusByPaymentIDTX", ctx, mock.Anything, paymentID, domain.OrderStatusCancelled).Return(nil).Once()
	mockRepo.On("GetOrderItemsByOrderIDsTX", ctx, mock.Anything, []uuid.UUID{orderID1, orderID2}).Return(items, nil).Once()
	mockProductRepo.On("RestoreStockBatchTX", ctx, mock.Anything, items).Return(nil).Once()

	err := m.HandlePaymentStatusChangeTX(ctx, nil, paymentID, domain.PaymentStatusFailed)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockProductRepo.AssertExpectations(t)
}

func TestOrderManager_HandlePaymentStatusChangeTX_Expired(t *testing.T) {
	mockRepo := NewMockOrderRepository(t)
	mockProductRepo := product.NewMockProductRepository(t)
	m := NewOrderManager(mockRepo, mockProductRepo)

	ctx := context.Background()
	paymentID := uuid.New()
	orderID := uuid.New()
	productID := uuid.New()

	orders := []domain.Order{{ID: orderID, Status: domain.OrderStatusPending}}
	items := []domain.OrderItem{{OrderID: orderID, ProductID: productID, Quantity: 1}}

	mockRepo.On("GetOrdersByPaymentIDForUpdateTX", ctx, mock.Anything, paymentID).Return(orders, nil).Once()
	mockRepo.On("UpdateOrderStatusByPaymentIDTX", ctx, mock.Anything, paymentID, domain.OrderStatusCancelled).Return(nil).Once()
	mockRepo.On("GetOrderItemsByOrderIDsTX", ctx, mock.Anything, []uuid.UUID{orderID}).Return(items, nil).Once()
	mockProductRepo.On("RestoreStockBatchTX", ctx, mock.Anything, items).Return(nil).Once()

	err := m.HandlePaymentStatusChangeTX(ctx, nil, paymentID, domain.PaymentStatusExpired)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockProductRepo.AssertExpectations(t)
}
