package product

import (
	"context"
	"errors"
	"testing"

	"go-marketplace/internal/common"
	"go-marketplace/internal/core/merchant"
	"go-marketplace/internal/domain"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProductService_CreateProduct(t *testing.T) {
	storeID := uuid.New()
	userID := uuid.New()
	req := ProductCreateRequest{
		StoreID:     storeID,
		Name:        "Test Product",
		Description: common.Ptr("Good product"),
		Price:       decimal.NewFromInt(100),
		Stock:       10,
	}

	tests := []struct {
		name      string
		userID    uuid.UUID
		request   ProductCreateRequest
		mockSetup func(mr *MockProductRepository, mmr *merchant.MockMerchantRepository)
		wantErr   bool
		errType   error
	}{
		{
			name:    "Success",
			userID:  userID,
			request: req,
			mockSetup: func(mr *MockProductRepository, mmr *merchant.MockMerchantRepository) {
				m := &domain.Merchant{ID: storeID, UserID: userID}
				mmr.On("GetByID", mock.Anything, storeID).Return(m, nil)
				mr.On("Create", mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			name:    "Merchant Not Found",
			userID:  userID,
			request: req,
			mockSetup: func(mr *MockProductRepository, mmr *merchant.MockMerchantRepository) {
				mmr.On("GetByID", mock.Anything, storeID).Return(nil, nil)
			},
			wantErr: true,
			errType: domain.ErrMerchantNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMockProductRepository(t)
			mockMerchantRepo := merchant.NewMockMerchantRepository(t)
			tt.mockSetup(mockRepo, mockMerchantRepo)

			service := NewProductService(mockRepo, mockMerchantRepo)
			res, err := service.CreateProduct(context.Background(), tt.userID, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
				assert.Equal(t, tt.request.Name, res.Name)
			}
		})
	}
}

func TestProductService_UpdateProduct(t *testing.T) {
	productID := uuid.New()
	userID := uuid.New()
	storeID := uuid.New()
	p := &domain.Product{ID: productID, StoreID: storeID, Name: "Old Name"}
	req := ProductUpdateRequest{
		Name:  "New Name",
		Price: decimal.NewFromInt(150),
	}

	tests := []struct {
		name      string
		userID    uuid.UUID
		productID uuid.UUID
		request   ProductUpdateRequest
		mockSetup func(mr *MockProductRepository, mmr *merchant.MockMerchantRepository)
		wantErr   bool
		errType   error
	}{
		{
			name:      "Success",
			userID:    userID,
			productID: productID,
			request:   req,
			mockSetup: func(mr *MockProductRepository, mmr *merchant.MockMerchantRepository) {
				m := &domain.Merchant{ID: storeID, UserID: userID}
				mr.On("GetByID", mock.Anything, productID).Return(p, nil)
				mmr.On("GetByID", mock.Anything, storeID).Return(m, nil)
				mr.On("Update", mock.Anything, mock.MatchedBy(func(p *domain.Product) bool {
					return p.Name == "New Name"
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			name:      "Product Not Found",
			userID:    userID,
			productID: productID,
			request:   req,
			mockSetup: func(mr *MockProductRepository, mmr *merchant.MockMerchantRepository) {
				mr.On("GetByID", mock.Anything, productID).Return(nil, nil)
			},
			wantErr: true,
			errType: domain.ErrProductNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMockProductRepository(t)
			mockMerchantRepo := merchant.NewMockMerchantRepository(t)
			tt.mockSetup(mockRepo, mockMerchantRepo)

			service := NewProductService(mockRepo, mockMerchantRepo)
			res, err := service.UpdateProduct(context.Background(), tt.userID, tt.productID, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
				assert.Equal(t, tt.request.Name, res.Name)
			}
		})
	}
}

func TestProductService_SearchProducts(t *testing.T) {
	productID := uuid.New()
	products := []domain.Product{
		{
			ID:          productID,
			Name:        "Gaming Laptop",
			Description: common.Ptr("Powerful laptop"),
		},
	}

	tests := []struct {
		name      string
		request   ProductSearchRequest
		mockSetup func(mr *MockProductRepository)
		wantErr   bool
		errMsg    string
	}{
		{
			name: "Success",
			request: ProductSearchRequest{
				Query: "laptop",
				Limit: 5,
				Page:  1,
			},
			mockSetup: func(mr *MockProductRepository) {
				mr.On("Search", mock.Anything, "laptop", 5, 0).Return(products, nil)
			},
			wantErr: false,
		},
		{
			name:    "Database Error",
			request: ProductSearchRequest{Query: "error"},
			mockSetup: func(mr *MockProductRepository) {
				mr.On("Search", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("db error"))
			},
			wantErr: true,
			errMsg:  "db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMockProductRepository(t)
			tt.mockSetup(mockRepo)

			service := NewProductService(mockRepo, nil)
			res, err := service.SearchProducts(context.Background(), tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Equal(t, tt.errMsg, err.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.Len(t, res, 1)
				assert.Equal(t, productID, res[0].ID)
			}
		})
	}
}
