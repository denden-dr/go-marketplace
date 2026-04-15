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

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserHandler_GetProfile_Success(t *testing.T) {
	mockService := NewMockUserServiceInterface(t)
	handler := NewUserHandler(mockService)
	app := testutil.SetupTestApp()

	userID := uuid.New()
	app.Get("/users/me", func(c *fiber.Ctx) error {
		c.Locals("userID", userID)
		return handler.GetProfile(c)
	})

	userRes := &UserResponse{
		ID:    userID,
		Email: "test@example.com",
	}

	mockService.On("GetUserByID", mock.Anything, userID).Return(userRes, nil)

	req := httptest.NewRequest("GET", "/users/me", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result common.ResponseWrapper
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "User profile retrieved", result.Message)
}

func TestUserHandler_GetProfile_Fail_NotFound(t *testing.T) {
	mockService := NewMockUserServiceInterface(t)
	handler := NewUserHandler(mockService)
	app := testutil.SetupTestApp()

	userID := uuid.New()
	app.Get("/users/me", func(c *fiber.Ctx) error {
		c.Locals("userID", userID)
		return handler.GetProfile(c)
	})

	mockService.On("GetUserByID", mock.Anything, userID).Return(nil, domain.ErrUserNotFound)

	req := httptest.NewRequest("GET", "/users/me", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestUserHandler_GetProfile_Fail_InternalError(t *testing.T) {
	mockService := NewMockUserServiceInterface(t)
	handler := NewUserHandler(mockService)
	app := testutil.SetupTestApp()

	userID := uuid.New()
	app.Get("/users/me", func(c *fiber.Ctx) error {
		c.Locals("userID", userID)
		return handler.GetProfile(c)
	})

	mockService.On("GetUserByID", mock.Anything, userID).Return(nil, errors.New("db error"))

	req := httptest.NewRequest("GET", "/users/me", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}
