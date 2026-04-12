package product

import (
	"errors"
	"go-shop-yourself/internal/domain"
	"go-shop-yourself/internal/common"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ProductHandler struct {
	service ProductServiceInterface
}

func NewProductHandler(service ProductServiceInterface) *ProductHandler {
	return &ProductHandler{service: service}
}

func (h *ProductHandler) CreateProduct(c *fiber.Ctx) error {
	var req ProductCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return common.NewResponse(c, fiber.StatusBadRequest, "Invalid request payload", nil)
	}

	if err := req.Validate(); err != nil {
		return common.NewResponse(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	res, err := h.service.CreateProduct(c.Context(), req)
	if err != nil {
		if errors.Is(err, domain.ErrMerchantNotFound) {
			return common.NewResponse(c, fiber.StatusNotFound, err.Error(), nil)
		}
		return common.NewResponse(c, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return common.NewResponse(c, fiber.StatusCreated, "Product created successfully", res)
}

func (h *ProductHandler) UpdateProduct(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return common.NewResponse(c, fiber.StatusBadRequest, "Invalid product ID", nil)
	}

	var req ProductUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return common.NewResponse(c, fiber.StatusBadRequest, "Invalid request payload", nil)
	}

	if err := req.Validate(); err != nil {
		return common.NewResponse(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	res, err := h.service.UpdateProduct(c.Context(), id, req)
	if err != nil {
		if errors.Is(err, domain.ErrProductNotFound) {
			return common.NewResponse(c, fiber.StatusNotFound, err.Error(), nil)
		}
		return common.NewResponse(c, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return common.NewResponse(c, fiber.StatusOK, "Product updated successfully", res)
}
