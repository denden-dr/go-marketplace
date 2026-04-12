package handlers

import (
	"bytes"
	"encoding/json"
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

	var result dtos.ResponseWrapper
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "Wallet details retrieved", result.Message)
}

func TestGetWallet_Fail_NotFound(t *testing.T) {
	mockService := mocks.NewWalletServiceInterface(t)
	handler := NewWalletHandler(mockService)
	app := setupTestApp()

	userID := uuid.New()
	app.Get("/wallets", authTestMiddleware(userID), handler.GetWallet)

	mockService.On("GetWalletByUserID", mock.Anything, userID).Return(nil, domain.ErrWalletNotFound)

	req := httptest.NewRequest("GET", "/wallets", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	var result dtos.ResponseWrapper
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, domain.ErrWalletNotFound.Error(), result.Message)
}

func TestGetHistory_Success(t *testing.T) {
	mockService := mocks.NewWalletServiceInterface(t)
	handler := NewWalletHandler(mockService)
	app := setupTestApp()

	userID := uuid.New()
	app.Get("/wallets/history", authTestMiddleware(userID), handler.GetHistory)

	txs := []dtos.TransactionResponse{
		{ID: uuid.New(), Amount: decimal.NewFromInt(100)},
	}

	mockService.On("GetWalletHistory", mock.Anything, userID, 1, 10).Return(txs, nil)

	req := httptest.NewRequest("GET", "/wallets/history?page=1&limit=10", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result dtos.ResponseWrapper
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "Wallet history retrieved", result.Message)
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

	var result dtos.ResponseWrapper
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "Withdrawal successful", result.Message)
}

func TestWithdraw_Fail_InsufficientBalance(t *testing.T) {
	mockService := mocks.NewWalletServiceInterface(t)
	handler := NewWalletHandler(mockService)
	app := setupTestApp()

	userID := uuid.New()
	app.Post("/wallets/withdraw", authTestMiddleware(userID), handler.Withdraw)

	reqBody := dtos.WithdrawRequest{
		Amount: decimal.NewFromInt(1000),
	}
	body, _ := json.Marshal(reqBody)

	mockService.On("Withdraw", mock.Anything, userID, mock.Anything).Return(domain.ErrInsufficientBalance)

	req := httptest.NewRequest("POST", "/wallets/withdraw", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var result dtos.ResponseWrapper
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, domain.ErrInsufficientBalance.Error(), result.Message)
}

func TestCreateWallet_Success(t *testing.T) {
	mockService := mocks.NewWalletServiceInterface(t)
	handler := NewWalletHandler(mockService)
	app := setupTestApp()

	userID := uuid.New()
	app.Post("/wallets", authTestMiddleware(userID), handler.CreateWallet)

	wallet := &domain.Wallet{ID: uuid.New(), UserID: userID}

	mockService.On("CreateWallet", mock.Anything, userID).Return(wallet, nil)

	req := httptest.NewRequest("POST", "/wallets", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var result dtos.ResponseWrapper
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "Wallet created successfully", result.Message)
}

func TestCreateWallet_Fail_AlreadyExists(t *testing.T) {
	mockService := mocks.NewWalletServiceInterface(t)
	handler := NewWalletHandler(mockService)
	app := setupTestApp()

	userID := uuid.New()
	app.Post("/wallets", authTestMiddleware(userID), handler.CreateWallet)

	mockService.On("CreateWallet", mock.Anything, userID).Return(nil, domain.ErrWalletAlreadyExists)

	req := httptest.NewRequest("POST", "/wallets", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusConflict, resp.StatusCode)

	var result dtos.ResponseWrapper
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, domain.ErrWalletAlreadyExists.Error(), result.Message)
}
