package services

import (
	"context"
	"testing"

	"go-shop-yourself/internal/domain"
	"go-shop-yourself/internal/dtos"
	"go-shop-yourself/internal/mocks"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateProduct_Success(t *testing.T) {
	mockRepo := mocks.NewProductRepository(t)
	mockMerchantRepo := mocks.NewMerchantRepository(t)
	service := NewProductService(mockRepo, mockMerchantRepo)

	storeID := uuid.New()
	req := dtos.ProductCreateRequest{
		StoreID:     storeID,
		Name:        "Test Product",
		Description: "Good product",
		Price:       decimal.NewFromInt(100),
		Stock:       10,
	}

	merchant := &domain.Merchant{ID: storeID}

	mockMerchantRepo.On("GetByID", mock.Anything, storeID).Return(merchant, nil)
	mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	res, err := service.CreateProduct(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, req.Name, res.Name)
}

func TestUpdateProduct_Success(t *testing.T) {
	mockRepo := mocks.NewProductRepository(t)
	service := NewProductService(mockRepo, nil)

	productID := uuid.New()
	product := &domain.Product{ID: productID, Name: "Old Name"}
	req := dtos.ProductUpdateRequest{
		Name:  "New Name",
		Price: decimal.NewFromInt(150),
	}

	mockRepo.On("GetByID", mock.Anything, productID).Return(product, nil)
	mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(p *domain.Product) bool {
		return p.Name == "New Name"
	})).Return(nil)

	res, err := service.UpdateProduct(context.Background(), productID, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "New Name", res.Name)
}
