package handlers

import (
	"go-shop-yourself/internal/dtos"
	"go-shop-yourself/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type MerchantHandler struct {
	service *services.MerchantService
}

func NewMerchantHandler(service *services.MerchantService) *MerchantHandler {
	return &MerchantHandler{service: service}
}

func (h *MerchantHandler) RegisterMerchant(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	var req dtos.MerchantRegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return dtos.NewResponse(c, fiber.StatusBadRequest, "Invalid request payload", nil)
	}

	res, err := h.service.RegisterMerchant(c.Context(), userID, req)
	if err != nil {
		return dtos.NewResponse(c, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return dtos.NewResponse(c, fiber.StatusCreated, "Merchant registered successfully", res)
}
