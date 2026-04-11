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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegisterMerchant_Success(t *testing.T) {
	mockService := mocks.NewMerchantServiceInterface(t)
	handler := NewMerchantHandler(mockService)
	app := setupTestApp()

	userID := uuid.New()
	// Use auth middleware to set userID
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
}
