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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegisterMerchant_Success(t *testing.T) {
	mockService := mocks.NewMerchantServiceInterface(t)
	handler := NewMerchantHandler(mockService)
	app := setupTestApp()

	userID := uuid.New()
	app.Post("/merchants", authTestMiddleware(userID), handler.RegisterMerchant)

	reqBody := dtos.MerchantRegisterRequest{
		Name:  "Test Shop",
		TaxID: "TAX-123",
	}
	body, _ := json.Marshal(reqBody)

	merchantRes := &dtos.MerchantResponse{
		ID:   uuid.New(),
		Name: "Test Shop",
	}

	mockService.On("RegisterMerchant", mock.Anything, userID, mock.Anything).Return(merchantRes, nil)

	req := httptest.NewRequest("POST", "/merchants", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var result dtos.ResponseWrapper
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "Merchant registered successfully", result.Message)
}

func TestRegisterMerchant_Fail_AlreadyExists(t *testing.T) {
	mockService := mocks.NewMerchantServiceInterface(t)
	handler := NewMerchantHandler(mockService)
	app := setupTestApp()

	userID := uuid.New()
	app.Post("/merchants", authTestMiddleware(userID), handler.RegisterMerchant)

	reqBody := dtos.MerchantRegisterRequest{
		Name:  "Existing Shop",
		TaxID: "TAX-789",
	}
	body, _ := json.Marshal(reqBody)

	mockService.On("RegisterMerchant", mock.Anything, userID, mock.Anything).Return(nil, domain.ErrMerchantAlreadyExists)

	req := httptest.NewRequest("POST", "/merchants", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusConflict, resp.StatusCode)

	var result dtos.ResponseWrapper
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, domain.ErrMerchantAlreadyExists.Error(), result.Message)
}

func TestRegisterMerchant_Fail_InternalError(t *testing.T) {
	mockService := mocks.NewMerchantServiceInterface(t)
	handler := NewMerchantHandler(mockService)
	app := setupTestApp()

	userID := uuid.New()
	app.Post("/merchants", authTestMiddleware(userID), handler.RegisterMerchant)

	reqBody := dtos.MerchantRegisterRequest{
		Name:  "Shop",
		TaxID: "TAX-456",
	}
	body, _ := json.Marshal(reqBody)

	mockService.On("RegisterMerchant", mock.Anything, userID, mock.Anything).Return(nil, errors.New("srv error"))

	req := httptest.NewRequest("POST", "/merchants", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	var result dtos.ResponseWrapper
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "srv error", result.Message)
}
