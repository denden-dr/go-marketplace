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

func TestGetWallet_Success(t *testing.T) {
	mockService := mocks.NewWalletServiceInterface(t)
	handler := NewWalletHandler(mockService)
	app := setupTestApp()

	userID := uuid.New()
	app.Get("/wallets", authTestMiddleware(userID), handler.GetWallet)

	walletRes := &dtos.WalletResponse{
		ID:           uuid.New(),
		UserID:       userID,
		WalletNumber: "WAL-123",
	}

	mockService.On("GetWalletByUserID", mock.Anything, userID).Return(walletRes, nil)

	req := httptest.NewRequest("GET", "/wallets", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestWithdraw_Success(t *testing.T) {
	mockService := mocks.NewWalletServiceInterface(t)
	handler := NewWalletHandler(mockService)
	app := setupTestApp()

	userID := uuid.New()
	app.Post("/wallets/withdraw", authTestMiddleware(userID), handler.Withdraw)

	reqBody := dtos.WithdrawRequest{
		Amount:      decimal.NewFromInt(100),
		Description: "Test",
	}
	body, _ := json.Marshal(reqBody)

	mockService.On("Withdraw", mock.Anything, userID, mock.Anything).Return(nil)

	req := httptest.NewRequest("POST", "/wallets/withdraw", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
