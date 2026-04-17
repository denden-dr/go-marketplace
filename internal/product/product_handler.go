package product

import (
	"errors"
	"go-shop-yourself/internal/common"
	"go-shop-yourself/internal/domain"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ProductHandler struct {
	service ProductServiceInterface
}

func NewProductHandler(service ProductServiceInterface) *ProductHandler {
	return &ProductHandler{service: service}
}

// CreateProduct adds a new product to a store
// @Summary Create product
// @Description Creates a new product for a specific merchant store.
// @Tags products
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body ProductCreateRequest true "Product Info"
// @Success 201 {object} common.ResponseWrapper{data=ProductResponse}
// @Failure 400 {object} common.ResponseWrapper
// @Failure 404 {object} common.ResponseWrapper
// @Failure 500 {object} common.ResponseWrapper
// @Router /products/ [post]
func (h *ProductHandler) CreateProduct(c *fiber.Ctx) error {
	var req ProductCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return common.NewResponse(c, fiber.StatusBadRequest, "Invalid request payload", nil)
	}

	if err := req.Validate(); err != nil {
		return common.NewResponse(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	userID := c.Locals("userID").(uuid.UUID)
	res, err := h.service.CreateProduct(c.Context(), userID, req)
	if err != nil {
		if errors.Is(err, domain.ErrMerchantNotFound) {
			return common.NewResponse(c, fiber.StatusNotFound, err.Error(), nil)
		}
		if errors.Is(err, domain.ErrForbidden) {
			return common.NewResponse(c, fiber.StatusForbidden, err.Error(), nil)
		}
		return common.NewResponse(c, fiber.StatusInternalServerError, "Internal Server Error", nil)
	}

	return common.NewResponse(c, fiber.StatusCreated, "Product created successfully", res)
}

// UpdateProduct updates an existing product
// @Summary Update product
// @Description Updates the details of an existing product.
// @Tags products
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Product ID (UUID)"
// @Param request body ProductUpdateRequest true "Product Update Info"
// @Success 200 {object} common.ResponseWrapper{data=ProductResponse}
// @Failure 400 {object} common.ResponseWrapper
// @Failure 404 {object} common.ResponseWrapper
// @Failure 500 {object} common.ResponseWrapper
// @Router /products/{id} [put]
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

	userID := c.Locals("userID").(uuid.UUID)
	res, err := h.service.UpdateProduct(c.Context(), userID, id, req)
	if err != nil {
		if errors.Is(err, domain.ErrProductNotFound) {
			return common.NewResponse(c, fiber.StatusNotFound, err.Error(), nil)
		}
		if errors.Is(err, domain.ErrForbidden) {
			return common.NewResponse(c, fiber.StatusForbidden, err.Error(), nil)
		}
		return common.NewResponse(c, fiber.StatusInternalServerError, "Internal Server Error", nil)
	}

	return common.NewResponse(c, fiber.StatusOK, "Product updated successfully", res)
}

// SearchProducts searches for products
// @Summary Search products
// @Description Searches for products using full-text search and fuzzy matching.
// @Tags products
// @Accept json
// @Produce json
// @Param q query string false "Search query"
// @Param limit query int false "Limit"
// @Param page query int false "Page"
// @Success 200 {object} common.ResponseWrapper{data=[]ProductResponse}
// @Failure 400 {object} common.ResponseWrapper
// @Failure 500 {object} common.ResponseWrapper
// @Router /products/search [get]
func (h *ProductHandler) SearchProducts(c *fiber.Ctx) error {
	var req ProductSearchRequest
	if err := c.QueryParser(&req); err != nil {
		return common.NewResponse(c, fiber.StatusBadRequest, "Invalid query parameters", nil)
	}

	if err := req.Validate(); err != nil {
		return common.NewResponse(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	res, err := h.service.SearchProducts(c.Context(), req)
	if err != nil {
		return common.NewResponse(c, fiber.StatusInternalServerError, "Internal Server Error", nil)
	}

	return common.NewResponse(c, fiber.StatusOK, "Products retrieved successfully", res)
}
