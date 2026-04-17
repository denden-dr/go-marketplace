package merchant

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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMerchantHandler_RegisterMerchant_Success(t *testing.T) {
	mockService := NewMockMerchantServiceInterface(t)
	handler := NewMerchantHandler(mockService)
	app := testutil.SetupTestApp()

	userID := uuid.New()
	app.Post("/merchants", testutil.AuthTestMiddleware(userID), handler.RegisterMerchant)

	reqBody := MerchantRegisterRequest{
		Name:  "Test Shop",
		TaxID: "TAX-123",
	}
	body, _ := json.Marshal(reqBody)

	merchantRes := &MerchantResponse{
		ID:   uuid.New(),
		Name: "Test Shop",
	}

	mockService.On("RegisterMerchant", mock.Anything, userID, mock.Anything).Return(merchantRes, nil)

	req := httptest.NewRequest("POST", "/merchants", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var result common.ResponseWrapper
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "Merchant registered successfully", result.Message)
}

func TestMerchantHandler_RegisterMerchant_Fail_AlreadyExists(t *testing.T) {
	mockService := NewMockMerchantServiceInterface(t)
	handler := NewMerchantHandler(mockService)
	app := testutil.SetupTestApp()

	userID := uuid.New()
	app.Post("/merchants", testutil.AuthTestMiddleware(userID), handler.RegisterMerchant)

	reqBody := MerchantRegisterRequest{
		Name:  "Existing Shop",
		TaxID: "TAX-789",
	}
	body, _ := json.Marshal(reqBody)

	mockService.On("RegisterMerchant", mock.Anything, userID, mock.Anything).Return(nil, domain.ErrMerchantAlreadyExists)

	req := httptest.NewRequest("POST", "/merchants", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusConflict, resp.StatusCode)

	var result common.ResponseWrapper
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, domain.ErrMerchantAlreadyExists.Error(), result.Message)
}

func TestMerchantHandler_RegisterMerchant_Fail_InternalError(t *testing.T) {
	mockService := NewMockMerchantServiceInterface(t)
	handler := NewMerchantHandler(mockService)
	app := testutil.SetupTestApp()

	userID := uuid.New()
	app.Post("/merchants", testutil.AuthTestMiddleware(userID), handler.RegisterMerchant)

	reqBody := MerchantRegisterRequest{
		Name:  "Shop",
		TaxID: "TAX-456",
	}
	body, _ := json.Marshal(reqBody)

	mockService.On("RegisterMerchant", mock.Anything, userID, mock.Anything).Return(nil, errors.New("srv error"))

	req := httptest.NewRequest("POST", "/merchants", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	var result common.ResponseWrapper
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "Internal Server Error", result.Message)
}
