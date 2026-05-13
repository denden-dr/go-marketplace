package product

import (
	"go-marketplace/internal/common"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ProductHandler struct {
	service ProductService
}

func NewProductHandler(service ProductService) *ProductHandler {
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
// @Success 201 {object} common.SuccessResponse{data=ProductResponse}
// @Failure 400 {object} common.ProblemDetails
// @Failure 404 {object} common.ProblemDetails
// @Failure 500 {object} common.ProblemDetails
// @Router /products/ [post]
func (h *ProductHandler) CreateProduct(c *fiber.Ctx) error {
	var req ProductCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(http.StatusBadRequest, "Invalid request payload")
	}

	if err := req.Validate(); err != nil {
		return err
	}

	userID := c.Locals("userID").(uuid.UUID)
	res, err := h.service.CreateProduct(c.Context(), userID, req)
	if err != nil {
		return err
	}

	return common.NewSuccessResponse(c, http.StatusCreated, "Product created successfully", res)
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
// @Success 200 {object} common.SuccessResponse{data=ProductResponse}
// @Failure 400 {object} common.ProblemDetails
// @Failure 404 {object} common.ProblemDetails
// @Failure 500 {object} common.ProblemDetails
// @Router /products/{id} [put]
func (h *ProductHandler) UpdateProduct(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "Invalid product ID")
	}

	var req ProductUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(http.StatusBadRequest, "Invalid request payload")
	}

	if err := req.Validate(); err != nil {
		return err
	}

	userID := c.Locals("userID").(uuid.UUID)
	res, err := h.service.UpdateProduct(c.Context(), userID, id, req)
	if err != nil {
		return err
	}

	return common.NewSuccessResponse(c, http.StatusOK, "Product updated successfully", res)
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
// @Success 200 {object} common.SuccessResponse{data=[]ProductResponse}
// @Failure 400 {object} common.ProblemDetails
// @Failure 500 {object} common.ProblemDetails
// @Router /products/search [get]
func (h *ProductHandler) SearchProducts(c *fiber.Ctx) error {
	var req ProductSearchRequest
	if err := c.QueryParser(&req); err != nil {
		return fiber.NewError(http.StatusBadRequest, "Invalid query parameters")
	}

	if err := req.Validate(); err != nil {
		return err
	}

	res, err := h.service.SearchProducts(c.Context(), req)
	if err != nil {
		return err
	}

	return common.NewSuccessResponse(c, http.StatusOK, "Products retrieved successfully", res)
}
