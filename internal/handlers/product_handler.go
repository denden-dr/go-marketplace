package handlers

import (
	"go-shop-yourself/internal/dtos"
	"go-shop-yourself/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ProductHandler struct {
	service *services.ProductService
}

func NewProductHandler(service *services.ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

func (h *ProductHandler) CreateProduct(c *fiber.Ctx) error {
	var req dtos.ProductCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return dtos.NewResponse(c, fiber.StatusBadRequest, "Invalid request payload", nil)
	}

	res, err := h.service.CreateProduct(c.Context(), req)
	if err != nil {
		return dtos.NewResponse(c, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return dtos.NewResponse(c, fiber.StatusCreated, "Product created successfully", res)
}

func (h *ProductHandler) UpdateProduct(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return dtos.NewResponse(c, fiber.StatusBadRequest, "Invalid product ID", nil)
	}

	var req dtos.ProductUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return dtos.NewResponse(c, fiber.StatusBadRequest, "Invalid request payload", nil)
	}

	res, err := h.service.UpdateProduct(c.Context(), id, req)
	if err != nil {
		return dtos.NewResponse(c, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return dtos.NewResponse(c, fiber.StatusOK, "Product updated successfully", res)
}
