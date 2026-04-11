package handlers

import (
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

func TestGetUserByID_Success(t *testing.T) {
	mockService := mocks.NewUserServiceInterface(t)
	handler := NewUserHandler(mockService)
	app := setupTestApp()
	app.Get("/users/:id", handler.GetUserByID)

	userID := uuid.New()
	userRes := &dtos.UserResponse{
		ID:    userID,
		Email: "test@example.com",
	}

	mockService.On("GetUserByID", mock.Anything, userID).Return(userRes, nil)

	req := httptest.NewRequest("GET", "/users/"+userID.String(), nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result dtos.ResponseWrapper
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "User profile retrieved", result.Message)
}
