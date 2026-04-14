package order

import (
	"context"
	"errors"
	"testing"
	"time"

	"go-shop-yourself/internal/cart"
	"go-shop-yourself/internal/domain"
	"go-shop-yourself/internal/product"
	"go-shop-yourself/internal/user"
	"go-shop-yourself/internal/wallet"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOrderService_CreateUserCheckout_Success(t *testing.T) {
	mockOrderRepo := NewMockOrderRepository(t)
	mockCartRepo := cart.NewMockCartRepository(t)
	mockProductRepo := product.NewMockProductRepository(t)
	mockWalletRepo := wallet.NewMockWalletRepository(t)
	mockUserRepo := user.NewMockUserRepository(t)

	service := NewOrderService(mockOrderRepo, mockCartRepo, mockProductRepo, mockWalletRepo, mockUserRepo)

	userID := uuid.New()
	req := CheckoutRequest{PaymentMethod: "wallet"}

	// Mocks for transaction
	mockTx, err := pgxmock.NewConn()
	assert.NoError(t, err)

	productID := uuid.New()
	merchantID := uuid.New()
	cartItems := []domain.CartItem{
		{
			ID:        uuid.New(),
			UserID:    userID,
			ProductID: productID,
			Quantity:  2,
			Product: &domain.Product{
				ID:      productID,
				Price:   decimal.NewFromInt(50),
				Stock:   10,
				StoreID: merchantID,
			},
		},
	}

	w := &domain.Wallet{ID: uuid.New(), UserID: userID, Balance: decimal.NewFromInt(200), Status: domain.WalletStatusActive}

	addr := &domain.UserAddress{
		ID:           uuid.New(),
		UserID:       userID,
		RecipientName: "John Doe",
		StreetAddress: "123 Main St",
		IsDefault:    true,
	}

	mockOrderRepo.On("Begin", mock.Anything).Return(mockTx, nil)
	mockCartRepo.On("GetCartByUserID", mock.Anything, userID).Return(cartItems, nil)
	mockUserRepo.On("GetAddressesByUserID", mock.Anything, userID).Return([]domain.UserAddress{*addr}, nil)

	mockProductRepo.On("GetByIDForUpdateTX", mock.Anything, mockTx, productID).Return(cartItems[0].Product, nil)
	mockWalletRepo.On("GetWalletByUserID", mock.Anything, userID).Return(w, nil)
	mockWalletRepo.On("DeductBalanceTX", mock.Anything, mockTx, w.ID, decimal.NewFromInt(100), mock.Anything).Return(nil)

	mockOrderRepo.On("CreateOrderPaymentTX", mock.Anything, mockTx, mock.Anything).Return(nil)
	mockOrderRepo.On("CreateOrderTX", mock.Anything, mockTx, mock.MatchedBy(func(o *domain.Order) bool {
		return o.ShippingRecipientName == "John Doe" && o.ShippingStreetAddress == "123 Main St"
	})).Return(nil)
	mockOrderRepo.On("CreateOrderItemTX", mock.Anything, mockTx, mock.Anything).Return(nil)
	mockProductRepo.On("UpdateStockTX", mock.Anything, mockTx, productID, 8).Return(nil)
	mockCartRepo.On("ClearCartTX", mock.Anything, mockTx, userID).Return(nil)

	mockTx.ExpectCommit()

	res, err := service.CreateUserCheckout(context.Background(), userID, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, decimal.NewFromInt(100), res.Amount)
	assert.Len(t, res.Orders, 1)
	assert.NoError(t, mockTx.ExpectationsWereMet())
}

func TestOrderService_CreateUserCheckout_CustomAddress(t *testing.T) {
	mockOrderRepo := NewMockOrderRepository(t)
	mockCartRepo := cart.NewMockCartRepository(t)
	mockProductRepo := product.NewMockProductRepository(t)
	mockWalletRepo := wallet.NewMockWalletRepository(t)
	mockUserRepo := user.NewMockUserRepository(t)

	service := NewOrderService(mockOrderRepo, mockCartRepo, mockProductRepo, mockWalletRepo, mockUserRepo)

	userID := uuid.New()
	req := CheckoutRequest{
		PaymentMethod:         "wallet",
		ShippingRecipientName: "Jane Custom",
		ShippingPhoneNumber:   "0811111111",
		ShippingStreetAddress: "789 Custom Rd",
		ShippingCity:          "Custom City",
		ShippingProvince:      "Custom Province",
		ShippingPostalCode:    "99999",
	}

	mockTx, err := pgxmock.NewConn()
	assert.NoError(t, err)

	productID := uuid.New()
	merchantID := uuid.New()
	cartItems := []domain.CartItem{
		{
			ID:        uuid.New(),
			UserID:    userID,
			ProductID: productID,
			Quantity:  1,
			Product: &domain.Product{
				ID:      productID,
				Price:   decimal.NewFromInt(100),
				Stock:   10,
				StoreID: merchantID,
			},
		},
	}

	w := &domain.Wallet{ID: uuid.New(), UserID: userID, Balance: decimal.NewFromInt(500), Status: domain.WalletStatusActive}

	mockOrderRepo.On("Begin", mock.Anything).Return(mockTx, nil)
	mockCartRepo.On("GetCartByUserID", mock.Anything, userID).Return(cartItems, nil)

	mockProductRepo.On("GetByIDForUpdateTX", mock.Anything, mockTx, productID).Return(cartItems[0].Product, nil)
	mockWalletRepo.On("GetWalletByUserID", mock.Anything, userID).Return(w, nil)
	mockWalletRepo.On("DeductBalanceTX", mock.Anything, mockTx, w.ID, decimal.NewFromInt(100), mock.Anything).Return(nil)

	mockOrderRepo.On("CreateOrderPaymentTX", mock.Anything, mockTx, mock.Anything).Return(nil)
	mockOrderRepo.On("CreateOrderTX", mock.Anything, mockTx, mock.MatchedBy(func(o *domain.Order) bool {
		return o.ShippingRecipientName == "Jane Custom" && o.ShippingStreetAddress == "789 Custom Rd"
	})).Return(nil)
	mockOrderRepo.On("CreateOrderItemTX", mock.Anything, mockTx, mock.Anything).Return(nil)
	mockProductRepo.On("UpdateStockTX", mock.Anything, mockTx, productID, 9).Return(nil)
	mockCartRepo.On("ClearCartTX", mock.Anything, mockTx, userID).Return(nil)

	mockTx.ExpectCommit()

	res, err := service.CreateUserCheckout(context.Background(), userID, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NoError(t, mockTx.ExpectationsWereMet())
}

func TestOrderService_CancelUserOrder_Success(t *testing.T) {
	mockOrderRepo := NewMockOrderRepository(t)
	mockWalletRepo := wallet.NewMockWalletRepository(t)
	mockProductRepo := product.NewMockProductRepository(t)

	service := NewOrderService(mockOrderRepo, nil, mockProductRepo, mockWalletRepo, nil)

	userID := uuid.New()
	orderID := uuid.New()
	productID := uuid.New()

	order := &domain.Order{
		ID:          orderID,
		UserID:      userID,
		Status:      domain.OrderStatusProcessing,
		TotalAmount: decimal.NewFromInt(100),
		CreatedAt:   time.Now().Add(-10 * time.Minute), // Within 1 hour
	}

	mockTx, err := pgxmock.NewConn()
	assert.NoError(t, err)

	w := &domain.Wallet{ID: uuid.New(), UserID: userID}
	items := []domain.OrderItem{{ProductID: productID, Quantity: 2}}
	p := &domain.Product{ID: productID, Stock: 5}

	mockOrderRepo.On("GetOrderByID", mock.Anything, orderID).Return(order, nil)
	mockOrderRepo.On("Begin", mock.Anything).Return(mockTx, nil)
	mockOrderRepo.On("UpdateOrderStatusTX", mock.Anything, mockTx, orderID, domain.OrderStatusCancelled).Return(nil)
	mockWalletRepo.On("GetWalletByUserID", mock.Anything, userID).Return(w, nil)
	mockWalletRepo.On("AddBalanceTX", mock.Anything, mockTx, w.ID, order.TotalAmount, mock.Anything).Return(nil)
	mockOrderRepo.On("GetOrderItems", mock.Anything, orderID).Return(items, nil)
	mockProductRepo.On("GetByIDForUpdateTX", mock.Anything, mockTx, productID).Return(p, nil)
	mockProductRepo.On("UpdateStockTX", mock.Anything, mockTx, productID, 7).Return(nil)

	mockTx.ExpectCommit()

	err = service.CancelUserOrder(context.Background(), userID, orderID)

	assert.NoError(t, err)
	assert.NoError(t, mockTx.ExpectationsWereMet())
}

func TestOrderService_CancelUserOrder_FailAfterOneHour(t *testing.T) {
	mockOrderRepo := NewMockOrderRepository(t)
	service := NewOrderService(mockOrderRepo, nil, nil, nil, nil)

	userID := uuid.New()
	orderID := uuid.New()
	order := &domain.Order{
		ID:        orderID,
		UserID:    userID,
		Status:    domain.OrderStatusProcessing,
		CreatedAt: time.Now().Add(-2 * time.Hour), // Expired
	}

	mockOrderRepo.On("GetOrderByID", mock.Anything, orderID).Return(order, nil)

	err := service.CancelUserOrder(context.Background(), userID, orderID)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrOrderNotCancellable, err)
}

func TestOrderService_MerchantUpdateStatus_Success(t *testing.T) {
	mockOrderRepo := NewMockOrderRepository(t)
	service := NewOrderService(mockOrderRepo, nil, nil, nil, nil)

	merchantID := uuid.New()
	orderID := uuid.New()
	order := &domain.Order{
		ID:         orderID,
		MerchantID: merchantID,
		Status:     domain.OrderStatusProcessing,
	}

	mockOrderRepo.On("GetOrderByID", mock.Anything, orderID).Return(order, nil)
	mockOrderRepo.On("UpdateOrderStatus", mock.Anything, orderID, domain.OrderStatusPackaging).Return(nil)

	err := service.MerchantUpdateStatus(context.Background(), merchantID, orderID, domain.OrderStatusPackaging)

	assert.NoError(t, err)
}

func TestOrderService_MerchantUpdateStatus_FailTooEarlyShipment(t *testing.T) {
	mockOrderRepo := NewMockOrderRepository(t)
	service := NewOrderService(mockOrderRepo, nil, nil, nil, nil)

	merchantID := uuid.New()
	orderID := uuid.New()

	// Packaging status but only 10 mins since creation
	order := &domain.Order{
		ID:         orderID,
		MerchantID: merchantID,
		Status:     domain.OrderStatusPackaging,
		CreatedAt:  time.Now().Add(-10 * time.Minute),
	}

	mockOrderRepo.On("GetOrderByID", mock.Anything, orderID).Return(order, nil)

	err := service.MerchantUpdateStatus(context.Background(), merchantID, orderID, domain.OrderStatusShipping)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrMerchantShipmentTooEarly))
}

func TestOrderService_AppealUserOrder_Success(t *testing.T) {
	mockOrderRepo := NewMockOrderRepository(t)
	service := NewOrderService(mockOrderRepo, nil, nil, nil, nil)

	userID := uuid.New()
	orderID := uuid.New()
	order := &domain.Order{
		ID:        orderID,
		UserID:    userID,
		Status:    domain.OrderStatusPackaging,
		CreatedAt: time.Now().Add(-2 * time.Hour), // Eligible for appeal
	}

	mockTx, err := pgxmock.NewConn()
	assert.NoError(t, err)

	mockOrderRepo.On("GetOrderByID", mock.Anything, orderID).Return(order, nil)
	mockOrderRepo.On("Begin", mock.Anything).Return(mockTx, nil)
	mockOrderRepo.On("CreateAppeal", mock.Anything, mock.Anything).Return(nil)
	mockOrderRepo.On("UpdateOrderAppealTX", mock.Anything, mockTx, orderID, true).Return(nil)

	mockTx.ExpectCommit()

	err = service.AppealUserOrder(context.Background(), userID, orderID, "Packaging took too long")

	assert.NoError(t, err)
	assert.NoError(t, mockTx.ExpectationsWereMet())
}
