package auth

import (
	"bytes"
	"encoding/json"
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

func TestAuthHandler_Register_Success(t *testing.T) {
	mockService := NewMockAuthServiceInterface(t)
	handler := NewAuthHandler(mockService)
	app := testutil.SetupTestApp()
	app.Post("/register", handler.Register)

	reqBody := RegisterRequest{
		FullName: "Test User",
		Email:    "test@example.com",
		Password: "password123",
		Username: "testuser",
	}
	body, _ := json.Marshal(reqBody)

	authRes := &AuthResponse{
		ID:           uuid.New(),
		FullName:     reqBody.FullName,
		Username:     reqBody.Username,
		Email:        reqBody.Email,
		AccessToken:  "access",
		RefreshToken: "refresh",
	}

	mockService.On("Register", mock.Anything, reqBody.FullName, reqBody.Email, reqBody.Password, reqBody.Username).Return(authRes, nil)

	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var result common.ResponseWrapper
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "User registered successfully", result.Message)
}

func TestAuthHandler_Register_Fail_InvalidBody(t *testing.T) {
	mockService := NewMockAuthServiceInterface(t)
	handler := NewAuthHandler(mockService)
	app := testutil.SetupTestApp()
	app.Post("/register", handler.Register)

	req := httptest.NewRequest("POST", "/register", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var result common.ResponseWrapper
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "Invalid request body", result.Message)
}

func TestAuthHandler_Register_Fail_UserAlreadyExists(t *testing.T) {
	mockService := NewMockAuthServiceInterface(t)
	handler := NewAuthHandler(mockService)
	app := testutil.SetupTestApp()
	app.Post("/register", handler.Register)

	reqBody := RegisterRequest{
		FullName: "Existing User",
		Email:    "exists@example.com",
		Password: "password123",
		Username: "testuser",
	}
	body, _ := json.Marshal(reqBody)

	mockService.On("Register", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, domain.ErrUserAlreadyExists)

	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusConflict, resp.StatusCode)

	var result common.ResponseWrapper
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, domain.ErrUserAlreadyExists.Error(), result.Message)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	mockService := NewMockAuthServiceInterface(t)
	handler := NewAuthHandler(mockService)
	app := testutil.SetupTestApp()
	app.Post("/login", handler.Login)

	reqBody := LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	authRes := &AuthResponse{
		ID:           uuid.New(),
		AccessToken:  "access",
		RefreshToken: "refresh",
	}

	mockService.On("Login", mock.Anything, reqBody.Email, reqBody.Password).Return(authRes, nil)

	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result common.ResponseWrapper
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "Login successful", result.Message)
}

func TestAuthHandler_Login_Fail_InvalidCredentials(t *testing.T) {
	mockService := NewMockAuthServiceInterface(t)
	handler := NewAuthHandler(mockService)
	app := testutil.SetupTestApp()
	app.Post("/login", handler.Login)

	reqBody := LoginRequest{
		Email:    "wrong@example.com",
		Password: "wrongpassword",
	}
	body, _ := json.Marshal(reqBody)

	mockService.On("Login", mock.Anything, mock.Anything, mock.Anything).Return(nil, domain.ErrInvalidCredentials)

	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	var result common.ResponseWrapper
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, domain.ErrInvalidCredentials.Error(), result.Message)
}

func TestAuthHandler_Login_Fail_AuthProviderMismatch(t *testing.T) {
	mockService := NewMockAuthServiceInterface(t)
	handler := NewAuthHandler(mockService)
	app := testutil.SetupTestApp()
	app.Post("/login", handler.Login)

	reqBody := LoginRequest{
		Email:    "social@example.com",
		Password: "password",
	}
	body, _ := json.Marshal(reqBody)

	mockService.On("Login", mock.Anything, mock.Anything, mock.Anything).Return(nil, domain.ErrAuthProviderMismatch)

	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusConflict, resp.StatusCode)

	var result common.ResponseWrapper
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, domain.ErrAuthProviderMismatch.Error(), result.Message)
}

func TestAuthHandler_RefreshTokens_Success(t *testing.T) {
	mockService := NewMockAuthServiceInterface(t)
	handler := NewAuthHandler(mockService)
	app := testutil.SetupTestApp()
	app.Post("/refresh", handler.RefreshTokens)

	reqBody := RefreshRequest{
		RefreshToken: "old-refresh-token",
	}
	body, _ := json.Marshal(reqBody)

	authRes := &AuthResponse{
		AccessToken:  "new-access",
		RefreshToken: "new-refresh",
	}

	mockService.On("RefreshTokens", mock.Anything, "old-refresh-token").Return(authRes, nil)

	req := httptest.NewRequest("POST", "/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result common.ResponseWrapper
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "Token refreshed successfully", result.Message)
}

func TestAuthHandler_RefreshTokens_Fail_TokenReused(t *testing.T) {
	mockService := NewMockAuthServiceInterface(t)
	handler := NewAuthHandler(mockService)
	app := testutil.SetupTestApp()
	app.Post("/refresh", handler.RefreshTokens)

	reqBody := RefreshRequest{
		RefreshToken: "reused-token",
	}
	body, _ := json.Marshal(reqBody)

	mockService.On("RefreshTokens", mock.Anything, "reused-token").Return(nil, domain.ErrRefreshTokenReused)

	req := httptest.NewRequest("POST", "/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	var result common.ResponseWrapper
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, domain.ErrRefreshTokenReused.Error(), result.Message)
}

func TestAuthHandler_Logout_Success(t *testing.T) {
	mockService := NewMockAuthServiceInterface(t)
	handler := NewAuthHandler(mockService)
	app := testutil.SetupTestApp()
	app.Post("/logout", handler.Logout)

	reqBody := LogoutRequest{
		RefreshToken: "valid-token",
	}
	body, _ := json.Marshal(reqBody)

	mockService.On("Logout", mock.Anything, "valid-token").Return(nil)

	req := httptest.NewRequest("POST", "/logout", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result common.ResponseWrapper
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "Logout successful", result.Message)
}

func TestAuthHandler_SocialLogin_Success(t *testing.T) {
	mockService := NewMockAuthServiceInterface(t)
	handler := NewAuthHandler(mockService)
	app := testutil.SetupTestApp()
	app.Post("/social", handler.SocialLogin)
 
 	reqBody := SocialLoginRequest{
 		AccessToken: "valid-social-token",
 	}
 	body, _ := json.Marshal(reqBody)
 
 	authRes := &AuthResponse{
 		ID:           uuid.New(),
 		AccessToken:  "access",
 		RefreshToken: "refresh",
 	}
 
 	mockService.On("SocialLogin", mock.Anything, reqBody.AccessToken).Return(authRes, nil)
 
 	req := httptest.NewRequest("POST", "/social", bytes.NewBuffer(body))
 	req.Header.Set("Content-Type", "application/json")
 
 	resp, _ := app.Test(req)
 
 	assert.Equal(t, http.StatusOK, resp.StatusCode)
 
 	var result common.ResponseWrapper
 	json.NewDecoder(resp.Body).Decode(&result)
 	assert.Equal(t, "Social login successful", result.Message)
 }

func TestAuthHandler_SocialLogin_Fail_EmailNotVerified(t *testing.T) {
	mockService := NewMockAuthServiceInterface(t)
	handler := NewAuthHandler(mockService)
	app := testutil.SetupTestApp()
	app.Post("/social", handler.SocialLogin)
 
 	reqBody := SocialLoginRequest{
 		AccessToken: "unverified-token",
 	}
 	body, _ := json.Marshal(reqBody)
 
 	mockService.On("SocialLogin", mock.Anything, reqBody.AccessToken).Return(nil, domain.ErrEmailNotVerified)
 
 	req := httptest.NewRequest("POST", "/social", bytes.NewBuffer(body))
 	req.Header.Set("Content-Type", "application/json")
 
 	resp, _ := app.Test(req)
 
 	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
 
 	var result common.ResponseWrapper
 	json.NewDecoder(resp.Body).Decode(&result)
 	assert.Equal(t, domain.ErrEmailNotVerified.Error(), result.Message)
 }
