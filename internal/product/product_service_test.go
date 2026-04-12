package product

import (
	"context"
	"errors"
	"testing"

	"go-shop-yourself/internal/domain"
	"go-shop-yourself/internal/merchant"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProductService_CreateProduct_Success(t *testing.T) {
	mockRepo := NewMockProductRepository(t)
	mockMerchantRepo := merchant.NewMockMerchantRepository(t)
	service := NewProductService(mockRepo, mockMerchantRepo)

	storeID := uuid.New()
	req := ProductCreateRequest{
		StoreID:     storeID,
		Name:        "Test Product",
		Description: "Good product",
		Price:       decimal.NewFromInt(100),
		Stock:       10,
	}

	m := &domain.Merchant{ID: storeID}

	mockMerchantRepo.On("GetByID", mock.Anything, storeID).Return(m, nil)
	mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	res, err := service.CreateProduct(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, req.Name, res.Name)
}

func TestProductService_CreateProduct_Fail_MerchantNotFound(t *testing.T) {
	mockMerchantRepo := merchant.NewMockMerchantRepository(t)
	service := NewProductService(nil, mockMerchantRepo)

	storeID := uuid.New()
	mockMerchantRepo.On("GetByID", mock.Anything, storeID).Return(nil, nil)

	_, err := service.CreateProduct(context.Background(), ProductCreateRequest{StoreID: storeID, Name: "Product"})

	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrMerchantNotFound))
}

func TestProductService_UpdateProduct_Success(t *testing.T) {
	mockRepo := NewMockProductRepository(t)
	service := NewProductService(mockRepo, nil)

	productID := uuid.New()
	p := &domain.Product{ID: productID, Name: "Old Name"}
	req := ProductUpdateRequest{
		Name:  "New Name",
		Price: decimal.NewFromInt(150),
	}

	mockRepo.On("GetByID", mock.Anything, productID).Return(p, nil)
	mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(p *domain.Product) bool {
		return p.Name == "New Name"
	})).Return(nil)

	res, err := service.UpdateProduct(context.Background(), productID, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "New Name", res.Name)
}

func TestProductService_UpdateProduct_Fail_NotFound(t *testing.T) {
	mockRepo := NewMockProductRepository(t)
	service := NewProductService(mockRepo, nil)

	productID := uuid.New()
	mockRepo.On("GetByID", mock.Anything, productID).Return(nil, nil)

	_, err := service.UpdateProduct(context.Background(), productID, ProductUpdateRequest{Name: "New"})

	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrProductNotFound))
}
