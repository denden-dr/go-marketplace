package user

import (
	"errors"
	"go-shop-yourself/internal/domain"
	"go-shop-yourself/internal/common"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type UserHandler struct {
	userService UserServiceInterface
}

func NewUserHandler(userService UserServiceInterface) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) GetUserByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return common.NewResponse(c, http.StatusBadRequest, "Invalid user ID format", nil)
	}

	user, err := h.userService.GetUserByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return common.NewResponse(c, http.StatusNotFound, err.Error(), nil)
		}
		return common.NewResponse(c, http.StatusInternalServerError, err.Error(), nil)
	}

	return common.NewResponse(c, http.StatusOK, "User profile retrieved", user)
}
