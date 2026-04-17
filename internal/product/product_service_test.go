package product

import (
	"context"
	"errors"
	"testing"

	"go-shop-yourself/internal/domain"
	"go-shop-yourself/internal/merchant"
	"go-shop-yourself/internal/testutil"

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
		Description: testutil.Ptr("Good product"),
		Price:       decimal.NewFromInt(100),
		Stock:       10,
	}

	userID := uuid.New()
	m := &domain.Merchant{ID: storeID, UserID: userID}

	mockMerchantRepo.On("GetByID", mock.Anything, storeID).Return(m, nil)
	mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	res, err := service.CreateProduct(context.Background(), userID, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, req.Name, res.Name)
}

func TestProductService_CreateProduct_Fail_MerchantNotFound(t *testing.T) {
	mockMerchantRepo := merchant.NewMockMerchantRepository(t)
	service := NewProductService(nil, mockMerchantRepo)

	storeID := uuid.New()
	mockMerchantRepo.On("GetByID", mock.Anything, storeID).Return(nil, nil)

	userID := uuid.New()
	_, err := service.CreateProduct(context.Background(), userID, ProductCreateRequest{StoreID: storeID, Name: "Product"})

	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrMerchantNotFound))
}

func TestProductService_UpdateProduct_Success(t *testing.T) {
	mockRepo := NewMockProductRepository(t)
	mockMerchantRepo := merchant.NewMockMerchantRepository(t)
	service := NewProductService(mockRepo, mockMerchantRepo)

	productID := uuid.New()
	userID := uuid.New()
	p := &domain.Product{ID: productID, Name: "Old Name"}
	req := ProductUpdateRequest{
		Name:  "New Name",
		Price: decimal.NewFromInt(150),
	}

	m := &domain.Merchant{ID: p.StoreID, UserID: userID}

	mockRepo.On("GetByID", mock.Anything, productID).Return(p, nil)
	mockMerchantRepo.On("GetByID", mock.Anything, p.StoreID).Return(m, nil)
	mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(p *domain.Product) bool {
		return p.Name == "New Name"
	})).Return(nil)

	res, err := service.UpdateProduct(context.Background(), userID, productID, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "New Name", res.Name)
}

func TestProductService_UpdateProduct_Fail_NotFound(t *testing.T) {
	mockRepo := NewMockProductRepository(t)
	mockMerchantRepo := merchant.NewMockMerchantRepository(t)
	service := NewProductService(mockRepo, mockMerchantRepo)
	productID := uuid.New()
	mockRepo.On("GetByID", mock.Anything, productID).Return(nil, nil)

	userID := uuid.New()
	_, err := service.UpdateProduct(context.Background(), userID, productID, ProductUpdateRequest{Name: "New"})

	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrProductNotFound))
}

func TestProductService_SearchProducts_Success(t *testing.T) {
	mockRepo := NewMockProductRepository(t)
	service := NewProductService(mockRepo, nil)

	req := ProductSearchRequest{
		Query: "laptop",
		Limit: 5,
		Page:  1,
	}

	productID := uuid.New()
	products := []domain.Product{
		{
			ID:          productID,
			Name:        "Gaming Laptop",
			Description: testutil.Ptr("Powerful laptop"),
		},
	}

	mockRepo.On("Search", mock.Anything, "laptop", 5, 0).Return(products, nil)

	res, err := service.SearchProducts(context.Background(), req)

	assert.NoError(t, err)
	assert.Len(t, res, 1)
	assert.Equal(t, productID, res[0].ID)
}

func TestProductService_SearchProducts_RepositoryError(t *testing.T) {
	mockRepo := NewMockProductRepository(t)
	service := NewProductService(mockRepo, nil)

	req := ProductSearchRequest{Query: "error"}
	mockRepo.On("Search", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("db error"))

	res, err := service.SearchProducts(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, "db error", err.Error())
}
