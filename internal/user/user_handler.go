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

// GetProfile retrieves user profile
// @Summary Get user profile
// @Description Fetches the authenticated user profile details.
// @Tags users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} common.ResponseWrapper{data=domain.User}
// @Failure 401 {object} common.ResponseWrapper
// @Failure 404 {object} common.ResponseWrapper
// @Failure 500 {object} common.ResponseWrapper
// @Router /users/me [get]
func (h *UserHandler) GetProfile(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return common.NewResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
	}

	user, err := h.userService.GetUserByID(c.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return common.NewResponse(c, http.StatusNotFound, err.Error(), nil)
		}
		return common.NewResponse(c, http.StatusInternalServerError, "Internal Server Error", nil)
	}

	return common.NewResponse(c, http.StatusOK, "User profile retrieved", user)
}

// AddAddress adds a new address for the user
// @Summary Add user address
// @Description Creates a new address template for the authenticated user.
// @Tags users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body AddressRequest true "Address Info"
// @Success 201 {object} common.ResponseWrapper{data=AddressResponse}
// @Failure 400 {object} common.ResponseWrapper
// @Failure 401 {object} common.ResponseWrapper
// @Failure 500 {object} common.ResponseWrapper
// @Router /users/addresses [post]
func (h *UserHandler) AddAddress(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return common.NewResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
	}

	var req AddressRequest
	if err := c.BodyParser(&req); err != nil {
		return common.NewResponse(c, http.StatusBadRequest, "Invalid request body", nil)
	}

	if err := req.Validate(); err != nil {
		return common.NewResponse(c, http.StatusBadRequest, err.Error(), nil)
	}

	res, err := h.userService.AddAddress(c.Context(), userID, &req)
	if err != nil {
		return common.NewResponse(c, http.StatusInternalServerError, "Internal Server Error", nil)
	}

	return common.NewResponse(c, http.StatusCreated, "Address added successfully", res)
}

// ListAddresses lists all addresses for the user
// @Summary List user addresses
// @Description Returns all saved address templates for the authenticated user.
// @Tags users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} common.ResponseWrapper{data=[]AddressResponse}
// @Failure 401 {object} common.ResponseWrapper
// @Failure 500 {object} common.ResponseWrapper
// @Router /users/addresses [get]
func (h *UserHandler) ListAddresses(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return common.NewResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
	}

	addresses, err := h.userService.ListAddresses(c.Context(), userID)
	if err != nil {
		return common.NewResponse(c, http.StatusInternalServerError, "Internal Server Error", nil)
	}

	return common.NewResponse(c, http.StatusOK, "Addresses retrieved", addresses)
}

// UpdateAddress updates an existing address
// @Summary Update user address
// @Description Modifies an existing address template for the authenticated user.
// @Tags users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Address ID (UUID)"
// @Param request body AddressRequest true "Updated Address Info"
// @Success 200 {object} common.ResponseWrapper{data=AddressResponse}
// @Failure 400 {object} common.ResponseWrapper
// @Failure 401 {object} common.ResponseWrapper
// @Failure 403 {object} common.ResponseWrapper
// @Failure 500 {object} common.ResponseWrapper
// @Router /users/addresses/{id} [put]
func (h *UserHandler) UpdateAddress(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return common.NewResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
	}

	idStr := c.Params("id")
	addressID, err := uuid.Parse(idStr)
	if err != nil {
		return common.NewResponse(c, http.StatusBadRequest, "Invalid address ID format", nil)
	}

	var req AddressRequest
	if err := c.BodyParser(&req); err != nil {
		return common.NewResponse(c, http.StatusBadRequest, "Invalid request body", nil)
	}

	if err := req.Validate(); err != nil {
		return common.NewResponse(c, http.StatusBadRequest, err.Error(), nil)
	}

	res, err := h.userService.UpdateAddress(c.Context(), userID, addressID, &req)
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return common.NewResponse(c, http.StatusForbidden, err.Error(), nil)
		}
		return common.NewResponse(c, http.StatusInternalServerError, "Internal Server Error", nil)
	}

	return common.NewResponse(c, http.StatusOK, "Address updated successfully", res)
}

// DeleteAddress removes an address
// @Summary Delete user address
// @Description Removes an address template from the user's saved addresses.
// @Tags users
// @Security BearerAuth
// @Param id path string true "Address ID (UUID)"
// @Success 200 {object} common.ResponseWrapper
// @Failure 400 {object} common.ResponseWrapper
// @Failure 401 {object} common.ResponseWrapper
// @Failure 403 {object} common.ResponseWrapper
// @Failure 500 {object} common.ResponseWrapper
// @Router /users/addresses/{id} [delete]
func (h *UserHandler) DeleteAddress(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return common.NewResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
	}

	idStr := c.Params("id")
	addressID, err := uuid.Parse(idStr)
	if err != nil {
		return common.NewResponse(c, http.StatusBadRequest, "Invalid address ID format", nil)
	}

	err = h.userService.DeleteAddress(c.Context(), userID, addressID)
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return common.NewResponse(c, http.StatusForbidden, err.Error(), nil)
		}
		return common.NewResponse(c, http.StatusInternalServerError, "Internal Server Error", nil)
	}

	return common.NewResponse(c, http.StatusOK, "Address deleted successfully", nil)
}
