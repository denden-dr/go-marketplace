package handlers

import (
	"go-shop-yourself/internal/domain"
	"go-shop-yourself/internal/dtos"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type UserHandler struct {
	userService domain.UserServiceInterface
}

func NewUserHandler(userService domain.UserServiceInterface) *UserHandler {
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
