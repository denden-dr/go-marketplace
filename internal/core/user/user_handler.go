package user

import (
	"go-marketplace/internal/common"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type UserHandler struct {
	userService UserService
}

func NewUserHandler(userService UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// GetProfile retrieves user profile
// @Summary Get user profile
// @Description Fetches the authenticated user profile details.
// @Tags users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} common.SuccessResponse{data=UserResponse}
// @Failure 401 {object} common.ProblemDetails
// @Failure 404 {object} common.ProblemDetails
// @Failure 500 {object} common.ProblemDetails
// @Router /users/me [get]
func (h *UserHandler) GetProfile(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return fiber.NewError(http.StatusUnauthorized, "Unauthorized")
	}

	user, err := h.userService.GetUserByID(c.Context(), userID)
	if err != nil {
		return err
	}

	return common.NewSuccessResponse(c, http.StatusOK, "User profile retrieved", user)
}

// AddAddress adds a new address for the user
// @Summary Add user address
// @Description Creates a new address template for the authenticated user.
// @Tags users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body AddressRequest true "Address Info"
// @Success 201 {object} common.SuccessResponse{data=AddressResponse}
// @Failure 400 {object} common.ProblemDetails
// @Failure 401 {object} common.ProblemDetails
// @Failure 500 {object} common.ProblemDetails
// @Router /users/addresses [post]
func (h *UserHandler) AddAddress(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return fiber.NewError(http.StatusUnauthorized, "Unauthorized")
	}

	var req AddressRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(http.StatusBadRequest, "Invalid request body")
	}

	if err := req.Validate(); err != nil {
		return err
	}

	res, err := h.userService.AddAddress(c.Context(), userID, &req)
	if err != nil {
		return err
	}

	return common.NewSuccessResponse(c, http.StatusCreated, "Address added successfully", res)
}

// ListAddresses lists all addresses for the user
// @Summary List user addresses
// @Description Returns all saved address templates for the authenticated user.
// @Tags users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} common.SuccessResponse{data=[]AddressResponse}
// @Failure 401 {object} common.ProblemDetails
// @Failure 500 {object} common.ProblemDetails
// @Router /users/addresses [get]
func (h *UserHandler) ListAddresses(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return fiber.NewError(http.StatusUnauthorized, "Unauthorized")
	}

	addresses, err := h.userService.ListAddresses(c.Context(), userID)
	if err != nil {
		return err
	}

	return common.NewSuccessResponse(c, http.StatusOK, "Addresses retrieved", addresses)
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
// @Success 200 {object} common.SuccessResponse{data=AddressResponse}
// @Failure 400 {object} common.ProblemDetails
// @Failure 401 {object} common.ProblemDetails
// @Failure 403 {object} common.ProblemDetails
// @Failure 500 {object} common.ProblemDetails
// @Router /users/addresses/{id} [put]
func (h *UserHandler) UpdateAddress(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return fiber.NewError(http.StatusUnauthorized, "Unauthorized")
	}

	idStr := c.Params("id")
	addressID, err := uuid.Parse(idStr)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "Invalid address ID format")
	}

	var req AddressRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(http.StatusBadRequest, "Invalid request body")
	}

	if err := req.Validate(); err != nil {
		return err
	}

	res, err := h.userService.UpdateAddress(c.Context(), userID, addressID, &req)
	if err != nil {
		return err
	}

	return common.NewSuccessResponse(c, http.StatusOK, "Address updated successfully", res)
}

// DeleteAddress removes an address
// @Summary Delete user address
// @Description Removes an address template from the user's saved addresses.
// @Tags users
// @Security BearerAuth
// @Param id path string true "Address ID (UUID)"
// @Success 200 {object} common.SuccessResponse
// @Failure 400 {object} common.ProblemDetails
// @Failure 401 {object} common.ProblemDetails
// @Failure 403 {object} common.ProblemDetails
// @Failure 500 {object} common.ProblemDetails
// @Router /users/addresses/{id} [delete]
func (h *UserHandler) DeleteAddress(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return fiber.NewError(http.StatusUnauthorized, "Unauthorized")
	}

	idStr := c.Params("id")
	addressID, err := uuid.Parse(idStr)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "Invalid address ID format")
	}

	err = h.userService.DeleteAddress(c.Context(), userID, addressID)
	if err != nil {
		return err
	}

	return common.NewSuccessResponse(c, http.StatusOK, "Address deleted successfully", nil)
}
