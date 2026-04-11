package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-shop-yourself/internal/dtos"
	"go-shop-yourself/internal/mocks"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateProduct_Success(t *testing.T) {
	mockService := mocks.NewProductServiceInterface(t)
	handler := NewProductHandler(mockService)
	app := setupTestApp()
	app.Post("/products", handler.CreateProduct)

	reqBody := dtos.ProductCreateRequest{
		StoreID: uuid.New(),
		Name:    "New Product",
		Price:   decimal.NewFromInt(100),
	}
	body, _ := json.Marshal(reqBody)

	productRes := &dtos.ProductResponse{
		ID:   uuid.New(),
		Name: "New Product",
	}

	mockService.On("CreateProduct", mock.Anything, mock.Anything).Return(productRes, nil)

	req := httptest.NewRequest("POST", "/products", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestUpdateProduct_Success(t *testing.T) {
	mockService := mocks.NewProductServiceInterface(t)
	handler := NewProductHandler(mockService)
	app := setupTestApp()
	app.Put("/products/:id", handler.UpdateProduct)

	productID := uuid.New()
	reqBody := dtos.ProductUpdateRequest{
		Name:  "Updated Product",
		Price: decimal.NewFromInt(100),
	}
	body, _ := json.Marshal(reqBody)

	productRes := &dtos.ProductResponse{
		ID:   productID,
		Name: "Updated Product",
	}

	mockService.On("UpdateProduct", mock.Anything, productID, mock.Anything).Return(productRes, nil)

	req := httptest.NewRequest("PUT", "/products/"+productID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
