package handlers

import (
	"go-shop-yourself/internal/dtos"
	"go-shop-yourself/internal/services"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) GetUserByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return dtos.NewResponse(c, http.StatusBadRequest, "Invalid user ID format", nil)
	}

	user, err := h.userService.GetUserByID(c.Context(), id)
	if err != nil {
		return dtos.NewResponse(c, http.StatusNotFound, err.Error(), nil)
	}

	return dtos.NewResponse(c, http.StatusOK, "User profile retrieved", user)
}
