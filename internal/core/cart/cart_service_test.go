package cart

import (
	"context"
	"testing"

	"go-marketplace/internal/core/product"
	"go-marketplace/internal/domain"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCartService_AddToCart(t *testing.T) {
	userID := uuid.New()
	productID := uuid.New()
	req := AddToCartRequest{
		ProductID: productID,
		Quantity:  2,
	}

	p := &domain.Product{
		ID:    productID,
		Name:  "Test Product",
		Price: decimal.NewFromInt(100),
	}

	tests := []struct {
		name      string
		userID    uuid.UUID
		request   AddToCartRequest
		mockSetup func(mr *MockCartRepository, mpr *product.MockProductRepository)
		wantErr   bool
		errType   error
	}{
		{
			name:    "Success",
			userID:  userID,
			request: req,
			mockSetup: func(mr *MockCartRepository, mpr *product.MockProductRepository) {
				mpr.On("GetByID", mock.Anything, productID).Return(p, nil)
				mr.On("UpsertCartItem", mock.Anything, mock.MatchedBy(func(item *domain.CartItem) bool {
					return item.UserID == userID && item.ProductID == productID && item.Quantity == 2
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			name:    "Product Not Found",
			userID:  userID,
			request: req,
			mockSetup: func(mr *MockCartRepository, mpr *product.MockProductRepository) {
				mpr.On("GetByID", mock.Anything, productID).Return(nil, nil)
			},
			wantErr: true,
			errType: domain.ErrProductNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMockCartRepository(t)
			mockProductRepo := product.NewMockProductRepository(t)
			tt.mockSetup(mockRepo, mockProductRepo)

			service := NewCartService(mockRepo, mockProductRepo)
			err := service.AddToCart(context.Background(), tt.userID, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCartService_GetCart(t *testing.T) {
	userID := uuid.New()
	items := []domain.CartItem{
		{
			ProductID: uuid.New(),
			Quantity:  2,
			Product: &domain.Product{
				Name:  "Item 1",
				Price: decimal.NewFromInt(100),
			},
		},
		{
			ProductID: uuid.New(),
			Quantity:  1,
			Product: &domain.Product{
				Name:  "Item 2",
				Price: decimal.NewFromInt(50),
			},
		},
	}

	tests := []struct {
		name      string
		userID    uuid.UUID
		mockSetup func(mr *MockCartRepository)
		wantErr   bool
	}{
		{
			name:   "Success",
			userID: userID,
			mockSetup: func(mr *MockCartRepository) {
				mr.On("GetCartByUserID", mock.Anything, userID).Return(items, nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMockCartRepository(t)
			tt.mockSetup(mockRepo)

			service := NewCartService(mockRepo, nil)
			res, err := service.GetCart(context.Background(), tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
				assert.Equal(t, decimal.NewFromInt(250), res.TotalPrice)
				assert.Len(t, res.Items, 2)
			}
		})
	}
}

func TestCartService_SimpleOperations(t *testing.T) {
	userID := uuid.New()
	productID := uuid.New()

	tests := []struct {
		name      string
		operation func(s CartService) error
		mockSetup func(mr *MockCartRepository)
		wantErr   bool
	}{
		{
			name: "UpdateCartItem Success",
			operation: func(s CartService) error {
				return s.UpdateCartItem(context.Background(), userID, productID, 5)
			},
			mockSetup: func(mr *MockCartRepository) {
				mr.On("UpdateCartItem", mock.Anything, userID, productID, 5).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "RemoveFromCart Success",
			operation: func(s CartService) error {
				return s.RemoveFromCart(context.Background(), userID, productID)
			},
			mockSetup: func(mr *MockCartRepository) {
				mr.On("DeleteCartItem", mock.Anything, userID, productID).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "ClearCart Success",
			operation: func(s CartService) error {
				return s.ClearCart(context.Background(), userID)
			},
			mockSetup: func(mr *MockCartRepository) {
				mr.On("ClearCart", mock.Anything, userID).Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMockCartRepository(t)
			tt.mockSetup(mockRepo)

			service := NewCartService(mockRepo, nil)
			err := tt.operation(service)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
