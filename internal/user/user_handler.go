package user

import (
	"errors"
	"go-shop-yourself/internal/common"
	"go-shop-yourself/internal/domain"
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

// GetUserByID retrieves user profile
// @Summary Get user profile
// @Description Fetches the user profile details by their unique ID.
// @Tags users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} common.ResponseWrapper{data=domain.User}
// @Failure 400 {object} common.ResponseWrapper
// @Failure 404 {object} common.ResponseWrapper
// @Failure 500 {object} common.ResponseWrapper
// @Router /users/{id} [get]
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
