package merchant

import (
	"errors"
	"go-shop-yourself/internal/common"
	"go-shop-yourself/internal/domain"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type MerchantHandler struct {
	service MerchantServiceInterface
}

func NewMerchantHandler(service MerchantServiceInterface) *MerchantHandler {
	return &MerchantHandler{service: service}
}

// RegisterMerchant registers a user as a merchant
// @Summary Register as merchant
// @Description Allows an authenticated user to register their own shop/merchant profile.
// @Tags merchants
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body MerchantRegisterRequest true "Merchant Registration Info"
// @Success 201 {object} common.ResponseWrapper{data=MerchantResponse}
// @Failure 400 {object} common.ResponseWrapper
// @Failure 404 {object} common.ResponseWrapper
// @Failure 409 {object} common.ResponseWrapper
// @Failure 500 {object} common.ResponseWrapper
// @Router /auth/register-merchant [post]
func (h *MerchantHandler) RegisterMerchant(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	var req MerchantRegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return common.NewResponse(c, fiber.StatusBadRequest, "Invalid request payload", nil)
	}

	if err := req.Validate(); err != nil {
		return common.NewResponse(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	res, err := h.service.RegisterMerchant(c.Context(), userID, req)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return common.NewResponse(c, fiber.StatusNotFound, err.Error(), nil)
		}
		if errors.Is(err, domain.ErrMerchantAlreadyExists) {
			return common.NewResponse(c, fiber.StatusConflict, err.Error(), nil)
		}
		return common.NewResponse(c, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return common.NewResponse(c, fiber.StatusCreated, "Merchant registered successfully", res)
}
