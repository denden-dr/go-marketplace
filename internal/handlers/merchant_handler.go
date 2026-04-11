package handlers

import (
	"errors"
	"go-shop-yourself/internal/domain"
	"go-shop-yourself/internal/dtos"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type MerchantHandler struct {
	service domain.MerchantServiceInterface
}

func NewMerchantHandler(service domain.MerchantServiceInterface) *MerchantHandler {
	return &MerchantHandler{service: service}
}

func (h *MerchantHandler) RegisterMerchant(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	var req dtos.MerchantRegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return dtos.NewResponse(c, fiber.StatusBadRequest, "Invalid request payload", nil)
	}

	if err := req.Validate(); err != nil {
		return dtos.NewResponse(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	res, err := h.service.RegisterMerchant(c.Context(), userID, req)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return dtos.NewResponse(c, fiber.StatusNotFound, err.Error(), nil)
		}
		if errors.Is(err, domain.ErrMerchantAlreadyExists) {
			return dtos.NewResponse(c, fiber.StatusConflict, err.Error(), nil)
		}
		return dtos.NewResponse(c, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return dtos.NewResponse(c, fiber.StatusCreated, "Merchant registered successfully", res)
}
