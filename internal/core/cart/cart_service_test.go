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

func TestCartService_AddToCart_Success(t *testing.T) {
	mockRepo := NewMockCartRepository(t)
	mockProductRepo := product.NewMockProductRepository(t)
	service := NewCartService(mockRepo, mockProductRepo)

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

	mockProductRepo.On("GetByID", mock.Anything, productID).Return(p, nil)
	mockRepo.On("UpsertCartItem", mock.Anything, mock.MatchedBy(func(item *domain.CartItem) bool {
		return item.UserID == userID && item.ProductID == productID && item.Quantity == 2
	})).Return(nil)

	err := service.AddToCart(context.Background(), userID, req)

	assert.NoError(t, err)
}

func TestCartService_AddToCart_DuplicateProduct_Success(t *testing.T) {
	// Note: The logic for "increase quantity" is handled by the Repository's ON CONFLICT clause.
	// This unit test verifies that AddToCart correctly calls the UpsertCartItem method.
	mockRepo := NewMockCartRepository(t)
	mockProductRepo := product.NewMockProductRepository(t)
	service := NewCartService(mockRepo, mockProductRepo)

	userID := uuid.New()
	productID := uuid.New()
	req := AddToCartRequest{
		ProductID: productID,
		Quantity:  1,
	}

	p := &domain.Product{
		ID:    productID,
		Name:  "Existing Product",
		Price: decimal.NewFromInt(50),
	}

	// First add
	mockProductRepo.On("GetByID", mock.Anything, productID).Return(p, nil)
	mockRepo.On("UpsertCartItem", mock.Anything, mock.MatchedBy(func(item *domain.CartItem) bool {
		return item.ProductID == productID && item.Quantity == 1
	})).Return(nil)

	err := service.AddToCart(context.Background(), userID, req)
	assert.NoError(t, err)
}

func TestCartService_AddToCart_ProductNotFound(t *testing.T) {
	mockRepo := NewMockCartRepository(t)
	mockProductRepo := product.NewMockProductRepository(t)
	service := NewCartService(mockRepo, mockProductRepo)

	userID := uuid.New()
	productID := uuid.New()
	req := AddToCartRequest{ProductID: productID}

	mockProductRepo.On("GetByID", mock.Anything, productID).Return(nil, nil)

	err := service.AddToCart(context.Background(), userID, req)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrProductNotFound, err)
}

func TestCartService_GetCart_Success(t *testing.T) {
	mockRepo := NewMockCartRepository(t)
	service := NewCartService(mockRepo, nil)

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

	mockRepo.On("GetCartByUserID", mock.Anything, userID).Return(items, nil)

	res, err := service.GetCart(context.Background(), userID)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, decimal.NewFromInt(250), res.TotalPrice)
	assert.Len(t, res.Items, 2)
	assert.Equal(t, decimal.NewFromInt(200), res.Items[0].Subtotal)
}

func TestCartService_UpdateCartItem_Success(t *testing.T) {
	mockRepo := NewMockCartRepository(t)
	service := NewCartService(mockRepo, nil)

	userID := uuid.New()
	productID := uuid.New()

	mockRepo.On("UpdateCartItem", mock.Anything, userID, productID, 5).Return(nil)

	err := service.UpdateCartItem(context.Background(), userID, productID, 5)

	assert.NoError(t, err)
}

func TestCartService_RemoveFromCart_Success(t *testing.T) {
	mockRepo := NewMockCartRepository(t)
	service := NewCartService(mockRepo, nil)

	userID := uuid.New()
	productID := uuid.New()

	mockRepo.On("DeleteCartItem", mock.Anything, userID, productID).Return(nil)

	err := service.RemoveFromCart(context.Background(), userID, productID)

	assert.NoError(t, err)
}

func TestCartService_ClearCart_Success(t *testing.T) {
	mockRepo := NewMockCartRepository(t)
	service := NewCartService(mockRepo, nil)

	userID := uuid.New()

	mockRepo.On("ClearCart", mock.Anything, userID).Return(nil)

	err := service.ClearCart(context.Background(), userID)

	assert.NoError(t, err)
}
