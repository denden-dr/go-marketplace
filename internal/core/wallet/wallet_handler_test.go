package wallet

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-marketplace/internal/common"
	"go-marketplace/internal/domain"
	"go-marketplace/internal/testutil"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestWalletHandler_GetWallet_Success(t *testing.T) {
	mockService := NewMockWalletServiceInterface(t)
	handler := NewWalletHandler(mockService)
	app := testutil.SetupTestApp()

	userID := uuid.New()
	app.Get("/wallets", testutil.AuthTestMiddleware(userID), handler.GetWallet)

	walletRes := &WalletResponse{
		ID:           uuid.New(),
		UserID:       userID,
		WalletNumber: "WAL-123",
	}

	mockService.On("GetWalletByUserID", mock.Anything, userID).Return(walletRes, nil)

	req := httptest.NewRequest("GET", "/wallets", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result common.ResponseWrapper
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "Wallet details retrieved", result.Message)
}

func TestWalletHandler_GetWallet_Fail_NotFound(t *testing.T) {
	mockService := NewMockWalletServiceInterface(t)
	handler := NewWalletHandler(mockService)
	app := testutil.SetupTestApp()

	userID := uuid.New()
	app.Get("/wallets", testutil.AuthTestMiddleware(userID), handler.GetWallet)

	mockService.On("GetWalletByUserID", mock.Anything, userID).Return(nil, domain.ErrWalletNotFound)

	req := httptest.NewRequest("GET", "/wallets", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	var result common.ResponseWrapper
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, domain.ErrWalletNotFound.Error(), result.Message)
}

func TestWalletHandler_GetHistory_Success(t *testing.T) {
	mockService := NewMockWalletServiceInterface(t)
	handler := NewWalletHandler(mockService)
	app := testutil.SetupTestApp()

	userID := uuid.New()
	app.Get("/wallets/history", testutil.AuthTestMiddleware(userID), handler.GetHistory)

	txs := []TransactionResponse{
		{ID: uuid.New(), Amount: decimal.NewFromInt(100)},
	}

	mockService.On("GetWalletHistory", mock.Anything, userID, 1, 10).Return(txs, nil)

	req := httptest.NewRequest("GET", "/wallets/history?page=1&limit=10", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result common.ResponseWrapper
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "Wallet history retrieved", result.Message)
}

func TestWalletHandler_Withdraw_Success(t *testing.T) {
	mockService := NewMockWalletServiceInterface(t)
	handler := NewWalletHandler(mockService)
	app := testutil.SetupTestApp()

	userID := uuid.New()
	app.Post("/wallets/withdraw", testutil.AuthTestMiddleware(userID), handler.Withdraw)

	reqBody := WithdrawRequest{
		Amount:      decimal.NewFromInt(100),
		Description: "Test",
	}
	body, _ := json.Marshal(reqBody)

	mockService.On("Withdraw", mock.Anything, userID, mock.Anything).Return(nil)

	req := httptest.NewRequest("POST", "/wallets/withdraw", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result common.ResponseWrapper
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "Withdrawal successful", result.Message)
}

func TestWalletHandler_Withdraw_Fail_InsufficientBalance(t *testing.T) {
	mockService := NewMockWalletServiceInterface(t)
	handler := NewWalletHandler(mockService)
	app := testutil.SetupTestApp()

	userID := uuid.New()
	app.Post("/wallets/withdraw", testutil.AuthTestMiddleware(userID), handler.Withdraw)

	reqBody := WithdrawRequest{
		Amount: decimal.NewFromInt(1000),
	}
	body, _ := json.Marshal(reqBody)

	mockService.On("Withdraw", mock.Anything, userID, mock.Anything).Return(domain.ErrInsufficientBalance)

	req := httptest.NewRequest("POST", "/wallets/withdraw", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var result common.ResponseWrapper
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, domain.ErrInsufficientBalance.Error(), result.Message)
}

func TestWalletHandler_CreateWallet_Success(t *testing.T) {
	mockService := NewMockWalletServiceInterface(t)
	handler := NewWalletHandler(mockService)
	app := testutil.SetupTestApp()

	userID := uuid.New()
	app.Post("/wallets", testutil.AuthTestMiddleware(userID), handler.CreateWallet)

	w := &WalletResponse{ID: uuid.New(), UserID: userID}

	mockService.On("CreateWallet", mock.Anything, userID).Return(w, nil)

	req := httptest.NewRequest("POST", "/wallets", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var result common.ResponseWrapper
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "Wallet created successfully", result.Message)
}

func TestWalletHandler_CreateWallet_Fail_AlreadyExists(t *testing.T) {
	mockService := NewMockWalletServiceInterface(t)
	handler := NewWalletHandler(mockService)
	app := testutil.SetupTestApp()

	userID := uuid.New()
	app.Post("/wallets", testutil.AuthTestMiddleware(userID), handler.CreateWallet)

	mockService.On("CreateWallet", mock.Anything, userID).Return(nil, domain.ErrWalletAlreadyExists)

	req := httptest.NewRequest("POST", "/wallets", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusConflict, resp.StatusCode)

	var result common.ResponseWrapper
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, domain.ErrWalletAlreadyExists.Error(), result.Message)
}
