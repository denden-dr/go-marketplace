package product

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-shop-yourself/internal/common"
	"go-shop-yourself/internal/domain"
	"go-shop-yourself/internal/testutil"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProductHandler_CreateProduct_Success(t *testing.T) {
	mockService := NewMockProductServiceInterface(t)
	handler := NewProductHandler(mockService)
	app := testutil.SetupTestApp()
	app.Post("/products", handler.CreateProduct)

	reqBody := ProductCreateRequest{
		StoreID: uuid.New(),
		Name:    "New Product",
		Price:   decimal.NewFromInt(100),
	}
	body, _ := json.Marshal(reqBody)

	productRes := &ProductResponse{
		ID:   uuid.New(),
		Name: "New Product",
	}

	mockService.On("CreateProduct", mock.Anything, mock.Anything).Return(productRes, nil)

	req := httptest.NewRequest("POST", "/products", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var result common.ResponseWrapper
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "Product created successfully", result.Message)
}

func TestProductHandler_CreateProduct_Fail_MerchantNotFound(t *testing.T) {
	mockService := NewMockProductServiceInterface(t)
	handler := NewProductHandler(mockService)
	app := testutil.SetupTestApp()
	app.Post("/products", handler.CreateProduct)

	reqBody := ProductCreateRequest{
		StoreID: uuid.New(),
		Name:    "Product",
		Price:   decimal.NewFromInt(100),
		Stock:   10,
	}
	body, _ := json.Marshal(reqBody)

	mockService.On("CreateProduct", mock.Anything, mock.MatchedBy(func(r ProductCreateRequest) bool {
		return r.Name == "Product"
	})).Return(nil, domain.ErrMerchantNotFound)

	req := httptest.NewRequest("POST", "/products", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	var result common.ResponseWrapper
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, domain.ErrMerchantNotFound.Error(), result.Message)
}

func TestProductHandler_UpdateProduct_Success(t *testing.T) {
	mockService := NewMockProductServiceInterface(t)
	handler := NewProductHandler(mockService)
	app := testutil.SetupTestApp()
	app.Put("/products/:id", handler.UpdateProduct)

	productID := uuid.New()
	reqBody := ProductUpdateRequest{
		Name:  "Updated Product",
		Price: decimal.NewFromInt(100),
	}
	body, _ := json.Marshal(reqBody)

	productRes := &ProductResponse{
		ID:   productID,
		Name: "Updated Product",
	}

	mockService.On("UpdateProduct", mock.Anything, productID, mock.Anything).Return(productRes, nil)

	req := httptest.NewRequest("PUT", "/products/"+productID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result common.ResponseWrapper
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "Product updated successfully", result.Message)
}

func TestProductHandler_UpdateProduct_Fail_NotFound(t *testing.T) {
	mockService := NewMockProductServiceInterface(t)
	handler := NewProductHandler(mockService)
	app := testutil.SetupTestApp()
	app.Put("/products/:id", handler.UpdateProduct)

	productID := uuid.New()
	reqBody := ProductUpdateRequest{
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

	var result common.ResponseWrapper
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, domain.ErrProductNotFound.Error(), result.Message)
}

func TestProductHandler_UpdateProduct_Fail_InternalError(t *testing.T) {
	mockService := NewMockProductServiceInterface(t)
	handler := NewProductHandler(mockService)
	app := testutil.SetupTestApp()
	app.Put("/products/:id", handler.UpdateProduct)

	productID := uuid.New()
	mockService.On("UpdateProduct", mock.Anything, productID, mock.Anything).Return(nil, errors.New("err"))

	reqBody := ProductUpdateRequest{
		Name:  "Valid Name",
		Price: decimal.NewFromInt(100),
		Stock: 10,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("PUT", "/products/"+productID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	var result common.ResponseWrapper
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "err", result.Message)
}

func TestProductHandler_SearchProducts_Success(t *testing.T) {
	mockService := NewMockProductServiceInterface(t)
	handler := NewProductHandler(mockService)
	app := testutil.SetupTestApp()
	app.Get("/products/search", handler.SearchProducts)

	productID := uuid.New()
	products := []ProductResponse{
		{
			ID:   productID,
			Name: "Search Result",
		},
	}

	mockService.On("SearchProducts", mock.Anything, mock.MatchedBy(func(r ProductSearchRequest) bool {
		return r.Query == "test"
	})).Return(products, nil)

	req := httptest.NewRequest("GET", "/products/search?q=test", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result common.ResponseWrapper
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "Products retrieved successfully", result.Message)
}

func TestProductHandler_SearchProducts_ValidationError(t *testing.T) {
	mockService := NewMockProductServiceInterface(t)
	handler := NewProductHandler(mockService)
	app := testutil.SetupTestApp()
	app.Get("/products/search", handler.SearchProducts)

	req := httptest.NewRequest("GET", "/products/search?q=a", nil) // Query too short
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestProductHandler_SearchProducts_InternalError(t *testing.T) {
	mockService := NewMockProductServiceInterface(t)
	handler := NewProductHandler(mockService)
	app := testutil.SetupTestApp()
	app.Get("/products/search", handler.SearchProducts)

	mockService.On("SearchProducts", mock.Anything, mock.Anything).Return(nil, errors.New("err"))

	req := httptest.NewRequest("GET", "/products/search?q=error", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}
