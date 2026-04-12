package user

import (
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

func TestUserHandler_GetUserByID_Success(t *testing.T) {
	mockService := NewMockUserServiceInterface(t)
	handler := NewUserHandler(mockService)
	app := testutil.SetupTestApp()
	app.Get("/users/:id", handler.GetUserByID)

	userID := uuid.New()
	userRes := &UserResponse{
		ID:    userID,
		Email: "test@example.com",
	}

	mockService.On("GetUserByID", mock.Anything, userID).Return(userRes, nil)

	req := httptest.NewRequest("GET", "/users/"+userID.String(), nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result common.ResponseWrapper
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "User profile retrieved", result.Message)
}

func TestUserHandler_GetUserByID_Fail_NotFound(t *testing.T) {
	mockService := NewMockUserServiceInterface(t)
	handler := NewUserHandler(mockService)
	app := testutil.SetupTestApp()
	app.Get("/users/:id", handler.GetUserByID)

	userID := uuid.New()
	mockService.On("GetUserByID", mock.Anything, userID).Return(nil, domain.ErrUserNotFound)

	req := httptest.NewRequest("GET", "/users/"+userID.String(), nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestUserHandler_GetUserByID_Fail_InternalError(t *testing.T) {
	mockService := NewMockUserServiceInterface(t)
	handler := NewUserHandler(mockService)
	app := testutil.SetupTestApp()
	app.Get("/users/:id", handler.GetUserByID)

	userID := uuid.New()
	mockService.On("GetUserByID", mock.Anything, userID).Return(nil, errors.New("db error"))

	req := httptest.NewRequest("GET", "/users/"+userID.String(), nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}
