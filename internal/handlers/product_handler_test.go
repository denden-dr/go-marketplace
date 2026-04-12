package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
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

	var result dtos.ResponseWrapper
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "Product created successfully", result.Message)
}

func TestCreateProduct_Fail_MerchantNotFound(t *testing.T) {
	mockService := mocks.NewProductServiceInterface(t)
	handler := NewProductHandler(mockService)
	app := setupTestApp()
	app.Post("/products", handler.CreateProduct)

	reqBody := dtos.ProductCreateRequest{
		StoreID: uuid.New(),
		Name:    "Product",
		Price:   decimal.NewFromInt(100),
		Stock:   10,
	}
	body, _ := json.Marshal(reqBody)

	mockService.On("CreateProduct", mock.Anything, mock.MatchedBy(func(r dtos.ProductCreateRequest) bool {
		return r.Name == "Product"
	})).Return(nil, domain.ErrMerchantNotFound)

	req := httptest.NewRequest("POST", "/products", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	var result dtos.ResponseWrapper
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, domain.ErrMerchantNotFound.Error(), result.Message)
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

	var result dtos.ResponseWrapper
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "Product updated successfully", result.Message)
}

func TestUpdateProduct_Fail_NotFound(t *testing.T) {
	mockService := mocks.NewProductServiceInterface(t)
	handler := NewProductHandler(mockService)
	app := setupTestApp()
	app.Put("/products/:id", handler.UpdateProduct)

	productID := uuid.New()
	reqBody := dtos.ProductUpdateRequest{
		Name:  "Product",
		Price: decimal.NewFromInt(100),
		Stock: 5,
	}
	body, _ := json.Marshal(reqBody)

	mockService.On("UpdateProduct", mock.Anything, productID, mock.Anything).Return(nil, domain.ErrProductNotFound)

	req := httptest.NewRequest("PUT", "/products/"+productID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	var result dtos.ResponseWrapper
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, domain.ErrProductNotFound.Error(), result.Message)
}

func TestUpdateProduct_Fail_InternalError(t *testing.T) {
	mockService := mocks.NewProductServiceInterface(t)
	handler := NewProductHandler(mockService)
	app := setupTestApp()
	app.Put("/products/:id", handler.UpdateProduct)

	productID := uuid.New()
	mockService.On("UpdateProduct", mock.Anything, productID, mock.Anything).Return(nil, errors.New("err"))

	reqBody := dtos.ProductUpdateRequest{
		Name:  "Valid Name",
		Price: decimal.NewFromInt(100),
		Stock: 10,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("PUT", "/products/"+productID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	var result dtos.ResponseWrapper
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "err", result.Message)
}
