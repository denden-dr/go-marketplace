package order

import (
	"context"
	"testing"
	"time"

	"go-marketplace/internal/core/cart"
	"go-marketplace/internal/core/merchant"
	"go-marketplace/internal/core/payment"
	"go-marketplace/internal/core/product"
	"go-marketplace/internal/core/user"
	"go-marketplace/internal/core/wallet"
	"go-marketplace/internal/domain"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func getSqlxTx(t *testing.T) (*sqlx.Tx, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	mock.ExpectBegin()
	tx, err := sqlxDB.BeginTxx(context.Background(), nil)
	assert.NoError(t, err)
	return tx, mock
}

func TestOrderService_CreateUserCheckout(t *testing.T) {
	userID := uuid.New()
	merchantID := uuid.New()
	productID := uuid.New()

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

	addr := &domain.UserAddress{
		ID:            uuid.New(),
		UserID:        userID,
		RecipientName: "John Doe",
		StreetAddress: "123 Main St",
		IsDefault:     true,
	}

	tests := []struct {
		name      string
		req       CheckoutRequest
		mockSetup func(mr *MockOrderRepository, mc *cart.MockCartRepository, mp *product.MockProductRepository, mw *wallet.MockWalletService, mu *user.MockUserRepository, mm *merchant.MockMerchantRepository, mps *payment.MockPaymentService, tx *sqlx.Tx, sqlMock sqlmock.Sqlmock)
		wantErr   bool
		errType   error
	}{
		{
			name: "Success - Default Address",
			req:  CheckoutRequest{PaymentMethod: domain.PaymentMethodWallet},
			mockSetup: func(mr *MockOrderRepository, mc *cart.MockCartRepository, mp *product.MockProductRepository, mw *wallet.MockWalletService, mu *user.MockUserRepository, mm *merchant.MockMerchantRepository, mps *payment.MockPaymentService, tx *sqlx.Tx, sqlMock sqlmock.Sqlmock) {
				mr.On("Begin", mock.Anything).Return(tx, nil)
				mc.On("GetCartByUserID", mock.Anything, userID).Return(cartItems, nil)
				mu.On("GetAddressesByUserID", mock.Anything, userID).Return([]domain.UserAddress{*addr}, nil)

				mp.On("GetByIDForUpdateTX", mock.Anything, tx, productID).Return(cartItems[0].Product, nil)
				mm.On("GetByID", mock.Anything, merchantID).Return(&domain.Merchant{ID: merchantID, UserID: uuid.New()}, nil)

				mps.On("CreatePaymentTX", mock.Anything, tx, mock.Anything).Return(&payment.PaymentResponse{PaymentID: uuid.New(), Status: domain.PaymentStatusSuccess}, nil)

				mr.On("CreateOrderTX", mock.Anything, tx, mock.MatchedBy(func(o *domain.Order) bool {
					return o.ShippingRecipientName == "John Doe"
				})).Return(nil)
				mr.On("CreateOrderItemTX", mock.Anything, tx, mock.Anything).Return(nil)
				mp.On("UpdateStockTX", mock.Anything, tx, productID, 8).Return(nil)
				mc.On("ClearCartTX", mock.Anything, tx, userID).Return(nil)

				sqlMock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "Success - Custom Address",
			req: CheckoutRequest{
				PaymentMethod:         domain.PaymentMethodWallet,
				ShippingRecipientName: "Jane Custom",
				ShippingPhoneNumber:   "0811111111",
				ShippingStreetAddress: "789 Custom Rd",
				ShippingCity:          "Custom City",
				ShippingProvince:      "Custom Province",
				ShippingPostalCode:    "99999",
			},
			mockSetup: func(mr *MockOrderRepository, mc *cart.MockCartRepository, mp *product.MockProductRepository, mw *wallet.MockWalletService, mu *user.MockUserRepository, mm *merchant.MockMerchantRepository, mps *payment.MockPaymentService, tx *sqlx.Tx, sqlMock sqlmock.Sqlmock) {
				mr.On("Begin", mock.Anything).Return(tx, nil)
				mc.On("GetCartByUserID", mock.Anything, userID).Return(cartItems, nil)

				mp.On("GetByIDForUpdateTX", mock.Anything, tx, productID).Return(cartItems[0].Product, nil)
				mm.On("GetByID", mock.Anything, merchantID).Return(&domain.Merchant{ID: merchantID, UserID: uuid.New()}, nil)

				mps.On("CreatePaymentTX", mock.Anything, tx, mock.Anything).Return(&payment.PaymentResponse{PaymentID: uuid.New(), Status: domain.PaymentStatusSuccess}, nil)

				mr.On("CreateOrderTX", mock.Anything, tx, mock.MatchedBy(func(o *domain.Order) bool {
					return o.ShippingRecipientName == "Jane Custom"
				})).Return(nil)
				mr.On("CreateOrderItemTX", mock.Anything, tx, mock.Anything).Return(nil)
				mp.On("UpdateStockTX", mock.Anything, tx, productID, 8).Return(nil)
				mc.On("ClearCartTX", mock.Anything, tx, userID).Return(nil)

				sqlMock.ExpectCommit()
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mo := NewMockOrderRepository(t)
			mc := cart.NewMockCartRepository(t)
			mp := product.NewMockProductRepository(t)
			mu := user.NewMockUserRepository(t)
			mm := merchant.NewMockMerchantRepository(t)
			mps := payment.NewMockPaymentService(t)
			mws := wallet.NewMockWalletService(t)

			tx, sqlMock := getSqlxTx(t)
			tt.mockSetup(mo, mc, mp, mws, mu, mm, mps, tx, sqlMock)

			service := NewOrderService(mo, mc, mp, mws, mu, mm, mps)
			res, err := service.CreateUserCheckout(context.Background(), userID, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
				assert.NoError(t, sqlMock.ExpectationsWereMet())
			}
		})
	}
}

func TestOrderService_CancelUserOrder(t *testing.T) {
	userID := uuid.New()
	orderID := uuid.New()
	productID := uuid.New()

	tests := []struct {
		name      string
		mockSetup func(mr *MockOrderRepository, mp *product.MockProductRepository, mw *wallet.MockWalletService, mm *merchant.MockMerchantRepository, tx *sqlx.Tx, sqlMock sqlmock.Sqlmock)
		wantErr   bool
		errType   error
	}{
		{
			name: "Success",
			mockSetup: func(mr *MockOrderRepository, mp *product.MockProductRepository, mw *wallet.MockWalletService, mm *merchant.MockMerchantRepository, tx *sqlx.Tx, sqlMock sqlmock.Sqlmock) {
				order := &domain.Order{
					ID:          orderID,
					UserID:      userID,
					Status:      domain.OrderStatusProcessing,
					TotalAmount: decimal.NewFromInt(100),
					CreatedAt:   time.Now().Add(-10 * time.Minute),
				}
				items := []domain.OrderItem{{ProductID: productID, Quantity: 2}}

				mr.On("Begin", mock.Anything).Return(tx, nil)
				mr.On("GetOrderByIDForUpdateTX", mock.Anything, tx, orderID).Return(order, nil)
				mr.On("UpdateOrderStatusTX", mock.Anything, tx, orderID, domain.OrderStatusCancelled).Return(nil)

				mm.On("GetByID", mock.Anything, order.MerchantID).Return(&domain.Merchant{ID: order.MerchantID, UserID: uuid.New()}, nil)
				mw.On("RefundFromPendingTX", mock.Anything, tx, mock.Anything, order.TotalAmount, mock.Anything).Return(nil)
				mw.On("AddBalanceTX", mock.Anything, tx, userID, order.TotalAmount, mock.Anything).Return(nil)

				mr.On("GetOrderItems", mock.Anything, orderID).Return(items, nil)
				mp.On("RestoreStockBatchTX", mock.Anything, tx, items).Return(nil)

				sqlMock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "Fail - After One Hour",
			mockSetup: func(mr *MockOrderRepository, mp *product.MockProductRepository, mw *wallet.MockWalletService, mm *merchant.MockMerchantRepository, tx *sqlx.Tx, sqlMock sqlmock.Sqlmock) {
				order := &domain.Order{
					ID:        orderID,
					UserID:    userID,
					Status:    domain.OrderStatusProcessing,
					CreatedAt: time.Now().Add(-2 * time.Hour),
				}
				mr.On("Begin", mock.Anything).Return(tx, nil)
				mr.On("GetOrderByIDForUpdateTX", mock.Anything, tx, orderID).Return(order, nil)
			},
			wantErr: true,
			errType: domain.ErrOrderNotCancellable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mo := NewMockOrderRepository(t)
			mp := product.NewMockProductRepository(t)
			mm := merchant.NewMockMerchantRepository(t)
			mps := payment.NewMockPaymentService(t)
			mws := wallet.NewMockWalletService(t)

			tx, sqlMock := getSqlxTx(t)
			tt.mockSetup(mo, mp, mws, mm, tx, sqlMock)

			service := NewOrderService(mo, nil, mp, mws, nil, mm, mps)
			err := service.CancelUserOrder(context.Background(), userID, orderID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.errType)
			} else {
				assert.NoError(t, err)
				assert.NoError(t, sqlMock.ExpectationsWereMet())
			}
		})
	}
}

func TestOrderService_MerchantUpdateStatus(t *testing.T) {
	merchantID := uuid.New()
	orderID := uuid.New()

	tests := []struct {
		name      string
		newStatus domain.OrderStatus
		mockSetup func(mr *MockOrderRepository, tx *sqlx.Tx, sqlMock sqlmock.Sqlmock)
		wantErr   bool
		errType   error
	}{
		{
			name:      "Success",
			newStatus: domain.OrderStatusPackaging,
			mockSetup: func(mr *MockOrderRepository, tx *sqlx.Tx, sqlMock sqlmock.Sqlmock) {
				order := &domain.Order{
					ID:         orderID,
					MerchantID: merchantID,
					Status:     domain.OrderStatusProcessing,
				}
				mr.On("Begin", mock.Anything).Return(tx, nil)
				mr.On("GetOrderByIDForUpdateTX", mock.Anything, tx, orderID).Return(order, nil)
				mr.On("UpdateOrderStatusTX", mock.Anything, tx, orderID, domain.OrderStatusPackaging).Return(nil)
				sqlMock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name:      "Fail - Too Early Shipment",
			newStatus: domain.OrderStatusShipping,
			mockSetup: func(mr *MockOrderRepository, tx *sqlx.Tx, sqlMock sqlmock.Sqlmock) {
				order := &domain.Order{
					ID:         orderID,
					MerchantID: merchantID,
					Status:     domain.OrderStatusPackaging,
					CreatedAt:  time.Now().Add(-10 * time.Minute),
				}
				mr.On("Begin", mock.Anything).Return(tx, nil)
				mr.On("GetOrderByIDForUpdateTX", mock.Anything, tx, orderID).Return(order, nil)
			},
			wantErr: true,
			errType: domain.ErrMerchantShipmentTooEarly,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mm := merchant.NewMockMerchantRepository(t)
			mps := payment.NewMockPaymentService(t)
			mws := wallet.NewMockWalletService(t)

			mo := NewMockOrderRepository(t)
			tx, sqlMock := getSqlxTx(t)
			tt.mockSetup(mo, tx, sqlMock)
			service := NewOrderService(mo, nil, nil, mws, nil, mm, mps)
			err := service.MerchantUpdateStatus(context.Background(), merchantID, orderID, tt.newStatus)

			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.errType)
			} else {
				assert.NoError(t, err)
				assert.NoError(t, sqlMock.ExpectationsWereMet())
			}
		})
	}
}

func TestOrderService_AppealUserOrder(t *testing.T) {
	userID := uuid.New()
	orderID := uuid.New()

	tests := []struct {
		name      string
		mockSetup func(mr *MockOrderRepository, tx *sqlx.Tx, sqlMock sqlmock.Sqlmock)
		wantErr   bool
	}{
		{
			name: "Success",
			mockSetup: func(mr *MockOrderRepository, tx *sqlx.Tx, sqlMock sqlmock.Sqlmock) {
				order := &domain.Order{
					ID:        orderID,
					UserID:    userID,
					Status:    domain.OrderStatusPackaging,
					CreatedAt: time.Now().Add(-2 * time.Hour),
				}
				mr.On("Begin", mock.Anything).Return(tx, nil)
				mr.On("GetOrderByIDForUpdateTX", mock.Anything, tx, orderID).Return(order, nil)
				mr.On("CreateAppeal", mock.Anything, mock.Anything).Return(nil)
				mr.On("UpdateOrderAppealTX", mock.Anything, tx, orderID, true).Return(nil)
				sqlMock.ExpectCommit()
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mm := merchant.NewMockMerchantRepository(t)
			mps := payment.NewMockPaymentService(t)
			mws := wallet.NewMockWalletService(t)

			mo := NewMockOrderRepository(t)
			tx, sqlMock := getSqlxTx(t)
			tt.mockSetup(mo, tx, sqlMock)
			service := NewOrderService(mo, nil, nil, mws, nil, mm, mps)
			err := service.AppealUserOrder(context.Background(), userID, orderID, "Packaging took too long")

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NoError(t, sqlMock.ExpectationsWereMet())
			}
		})
	}
}

