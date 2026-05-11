package cart

import (
	"errors"
	"go-marketplace/internal/common"
	"go-marketplace/internal/domain"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type CartHandler struct {
	cartService CartService
}

func NewCartHandler(cartService CartService) *CartHandler {
	return &CartHandler{cartService: cartService}
}

// AddToCart adds a product to the user's shopping cart
// @Summary Add to cart
// @Description Adds a specified quantity of a product to the authenticated user's cart.
// @Tags cart
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body AddToCartRequest true "Add to Cart Info"
// @Success 201 {object} common.ResponseWrapper
// @Failure 400 {object} common.ResponseWrapper
// @Failure 404 {object} common.ResponseWrapper
// @Failure 500 {object} common.ResponseWrapper
// @Router /users/cart [post]
func (h *CartHandler) AddToCart(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	var req AddToCartRequest
	if err := c.BodyParser(&req); err != nil {
		return common.NewResponse(c, http.StatusBadRequest, "Invalid request body", nil)
	}

	if err := req.Validate(); err != nil {
		return common.NewResponse(c, http.StatusBadRequest, err.Error(), nil)
	}

	err := h.cartService.AddToCart(c.Context(), userID, req)
	if err != nil {
		if errors.Is(err, domain.ErrProductNotFound) {
			return common.NewResponse(c, http.StatusNotFound, err.Error(), nil)
		}
		log.Printf("Error adding to cart for user %s: %v", userID, err)
		return common.NewResponse(c, http.StatusInternalServerError, "Internal Server Error", nil)
	}

	return common.NewResponse(c, http.StatusCreated, "Product added to cart", nil)
}

// UpdateCartItem updates the quantity of an item in the cart
// @Summary Update cart item
// @Description Updates the quantity of a specific product already in the user's cart.
// @Tags cart
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param productID path string true "Product ID (UUID)"
// @Param request body UpdateCartItemRequest true "Update Quantity Info"
// @Success 200 {object} common.ResponseWrapper
// @Failure 400 {object} common.ResponseWrapper
// @Failure 500 {object} common.ResponseWrapper
// @Router /users/cart/{productID} [put]
func (h *CartHandler) UpdateCartItem(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	productID, err := uuid.Parse(c.Params("productID"))
	if err != nil {
		return common.NewResponse(c, http.StatusBadRequest, "Invalid product ID", nil)
	}

	var req UpdateCartItemRequest
	if err := c.BodyParser(&req); err != nil {
		return common.NewResponse(c, http.StatusBadRequest, "Invalid request body", nil)
	}

	if err := req.Validate(); err != nil {
		return common.NewResponse(c, http.StatusBadRequest, err.Error(), nil)
	}

	err = h.cartService.UpdateCartItem(c.Context(), userID, productID, req.Quantity)
	if err != nil {
		log.Printf("Error updating cart for user %s: %v", userID, err)
		return common.NewResponse(c, http.StatusInternalServerError, "Internal Server Error", nil)
	}

	return common.NewResponse(c, http.StatusOK, "Cart updated", nil)
}

// RemoveFromCart removes an item from the cart
// @Summary Remove from cart
// @Description Removes a specific product from the authenticated user's cart.
// @Tags cart
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param productID path string true "Product ID (UUID)"
// @Success 200 {object} common.ResponseWrapper
// @Failure 400 {object} common.ResponseWrapper
// @Failure 500 {object} common.ResponseWrapper
// @Router /users/cart/{productID} [delete]
func (h *CartHandler) RemoveFromCart(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	productID, err := uuid.Parse(c.Params("productID"))
	if err != nil {
		return common.NewResponse(c, http.StatusBadRequest, "Invalid product ID", nil)
	}

	err = h.cartService.RemoveFromCart(c.Context(), userID, productID)
	if err != nil {
		log.Printf("Error removing from cart for user %s: %v", userID, err)
		return common.NewResponse(c, http.StatusInternalServerError, "Internal Server Error", nil)
	}

	return common.NewResponse(c, http.StatusOK, "Product removed from cart", nil)
}

// GetCart retrieves the user's shopping cart
// @Summary Get cart
// @Description Fetches all items in the authenticated user's shopping cart along with the total price.
// @Tags cart
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} common.ResponseWrapper{data=CartResponse}
// @Failure 500 {object} common.ResponseWrapper
// @Router /users/cart [get]
func (h *CartHandler) GetCart(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	cart, err := h.cartService.GetCart(c.Context(), userID)
	if err != nil {
		log.Printf("Error getting cart for user %s: %v", userID, err)
		return common.NewResponse(c, http.StatusInternalServerError, "Internal Server Error", nil)
	}

	return common.NewResponse(c, http.StatusOK, "Cart retrieved", cart)
}

// ClearCart removes all items from the user's cart
// @Summary Clear cart
// @Description Deletes all products from the authenticated user's shopping cart.
// @Tags cart
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} common.ResponseWrapper
// @Failure 500 {object} common.ResponseWrapper
// @Router /users/cart [delete]
func (h *CartHandler) ClearCart(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	err := h.cartService.ClearCart(c.Context(), userID)
	if err != nil {
		log.Printf("Error clearing cart for user %s: %v", userID, err)
		return common.NewResponse(c, http.StatusInternalServerError, "Internal Server Error", nil)
	}

	return common.NewResponse(c, http.StatusOK, "Cart cleared", nil)
}
