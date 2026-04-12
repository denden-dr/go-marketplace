package merchant

import (
	"errors"
	"go-shop-yourself/internal/domain"
	"go-shop-yourself/internal/common"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type MerchantHandler struct {
	service MerchantServiceInterface
}

func NewMerchantHandler(service MerchantServiceInterface) *MerchantHandler {
	return &MerchantHandler{service: service}
}

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
