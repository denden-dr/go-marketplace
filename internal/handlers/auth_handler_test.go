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
}
