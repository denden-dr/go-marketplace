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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegister_Success(t *testing.T) {
	mockService := mocks.NewAuthServiceInterface(t)
	handler := NewAuthHandler(mockService)
	app := setupTestApp()
	app.Post("/register", handler.Register)

	userID := uuid.New()
	reqBody := dtos.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Username: "testuser",
	}
	body, _ := json.Marshal(reqBody)

	mockService.On("Register", mock.Anything, reqBody.Email, reqBody.Password, reqBody.Username).Return(userID, nil)

	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var result dtos.ResponseWrapper
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "User registered successfully", result.Message)
}

func TestRegister_Fail_InvalidBody(t *testing.T) {
	mockService := mocks.NewAuthServiceInterface(t)
	handler := NewAuthHandler(mockService)
	app := setupTestApp()
	app.Post("/register", handler.Register)

	req := httptest.NewRequest("POST", "/register", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var result dtos.ResponseWrapper
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "Invalid request body", result.Message)
}

func TestRegister_Fail_UserAlreadyExists(t *testing.T) {
	mockService := mocks.NewAuthServiceInterface(t)
	handler := NewAuthHandler(mockService)
	app := setupTestApp()
	app.Post("/register", handler.Register)

	reqBody := dtos.RegisterRequest{
		Email:    "exists@example.com",
		Password: "password123",
		Username: "testuser",
	}
	body, _ := json.Marshal(reqBody)

	mockService.On("Register", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(uuid.Nil, domain.ErrUserAlreadyExists)

	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusConflict, resp.StatusCode)

	var result dtos.ResponseWrapper
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, domain.ErrUserAlreadyExists.Error(), result.Message)
}

func TestLogin_Success(t *testing.T) {
	mockService := mocks.NewAuthServiceInterface(t)
	handler := NewAuthHandler(mockService)
	app := setupTestApp()
	app.Post("/login", handler.Login)

	reqBody := dtos.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	authRes := &dtos.AuthResponse{
		ID:           uuid.New(),
		AccessToken:  "access",
		RefreshToken: "refresh",
	}

	mockService.On("Login", mock.Anything, reqBody.Email, reqBody.Password).Return(authRes, nil)

	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result dtos.ResponseWrapper
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "Login successful", result.Message)
}

func TestLogin_Fail_InvalidCredentials(t *testing.T) {
	mockService := mocks.NewAuthServiceInterface(t)
	handler := NewAuthHandler(mockService)
	app := setupTestApp()
	app.Post("/login", handler.Login)

	reqBody := dtos.LoginRequest{
		Email:    "wrong@example.com",
		Password: "wrongpassword",
	}
	body, _ := json.Marshal(reqBody)

	mockService.On("Login", mock.Anything, mock.Anything, mock.Anything).Return(nil, domain.ErrInvalidCredentials)

	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	var result dtos.ResponseWrapper
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, domain.ErrInvalidCredentials.Error(), result.Message)
}

func TestRefreshTokens_Success(t *testing.T) {
	mockService := mocks.NewAuthServiceInterface(t)
	handler := NewAuthHandler(mockService)
	app := setupTestApp()
	app.Post("/refresh", handler.RefreshTokens)

	reqBody := dtos.RefreshRequest{
		RefreshToken: "old-refresh-token",
	}
	body, _ := json.Marshal(reqBody)

	authRes := &dtos.AuthResponse{
		AccessToken:  "new-access",
		RefreshToken: "new-refresh",
	}

	mockService.On("RefreshTokens", mock.Anything, "old-refresh-token").Return(authRes, nil)

	req := httptest.NewRequest("POST", "/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result dtos.ResponseWrapper
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "Token refreshed successfully", result.Message)
}

func TestRefreshTokens_Fail_TokenReused(t *testing.T) {
	mockService := mocks.NewAuthServiceInterface(t)
	handler := NewAuthHandler(mockService)
	app := setupTestApp()
	app.Post("/refresh", handler.RefreshTokens)

	reqBody := dtos.RefreshRequest{
		RefreshToken: "reused-token",
	}
	body, _ := json.Marshal(reqBody)

	mockService.On("RefreshTokens", mock.Anything, "reused-token").Return(nil, domain.ErrRefreshTokenReused)

	req := httptest.NewRequest("POST", "/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	var result dtos.ResponseWrapper
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, domain.ErrRefreshTokenReused.Error(), result.Message)
}

func TestLogout_Success(t *testing.T) {
	mockService := mocks.NewAuthServiceInterface(t)
	handler := NewAuthHandler(mockService)
	app := setupTestApp()
	app.Post("/logout", handler.Logout)

	reqBody := dtos.LogoutRequest{
		RefreshToken: "valid-token",
	}
	body, _ := json.Marshal(reqBody)

	mockService.On("Logout", mock.Anything, "valid-token").Return(nil)

	req := httptest.NewRequest("POST", "/logout", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result dtos.ResponseWrapper
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "Logout successful", result.Message)
}
