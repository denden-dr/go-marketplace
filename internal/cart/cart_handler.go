package cart

import (
	"errors"
	"go-shop-yourself/internal/common"
	"go-shop-yourself/internal/domain"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type CartHandler struct {
	cartService CartServiceInterface
}

func NewCartHandler(cartService CartServiceInterface) *CartHandler {
	return &CartHandler{cartService: cartService}
}

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

func (h *CartHandler) UpdateCartItem(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	productID, err := uuid.Parse(c.Params("id"))
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

func (h *CartHandler) RemoveFromCart(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	productID, err := uuid.Parse(c.Params("id"))
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

func (h *CartHandler) GetCart(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	cart, err := h.cartService.GetCart(c.Context(), userID)
	if err != nil {
		log.Printf("Error getting cart for user %s: %v", userID, err)
		return common.NewResponse(c, http.StatusInternalServerError, "Internal Server Error", nil)
	}

	return common.NewResponse(c, http.StatusOK, "Cart retrieved", cart)
}

func (h *CartHandler) ClearCart(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	err := h.cartService.ClearCart(c.Context(), userID)
	if err != nil {
		log.Printf("Error clearing cart for user %s: %v", userID, err)
		return common.NewResponse(c, http.StatusInternalServerError, "Internal Server Error", nil)
	}

	return common.NewResponse(c, http.StatusOK, "Cart cleared", nil)
}
