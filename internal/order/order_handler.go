package order

import (
	"errors"
	"go-shop-yourself/internal/common"
	"go-shop-yourself/internal/domain"
	"go-shop-yourself/internal/merchant"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type OrderHandler struct {
	orderService OrderServiceInterface
	merchantRepo merchant.MerchantRepository
}

func NewOrderHandler(orderService OrderServiceInterface, merchantRepo merchant.MerchantRepository) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
		merchantRepo: merchantRepo,
	}
}

// Checkout processes the user's shopping cart into an order
// @Summary Create order from cart
// @Description Converts the current items in the authenticated user's cart into a formal order.
// @Tags orders
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body CheckoutRequest true "Checkout Info"
// @Success 201 {object} common.ResponseWrapper{data=OrderResponse}
// @Failure 400 {object} common.ResponseWrapper
// @Failure 500 {object} common.ResponseWrapper
// @Router /users/orders [post]
func (h *OrderHandler) Checkout(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	var req CheckoutRequest
	if err := c.BodyParser(&req); err != nil {
		return common.NewResponse(c, http.StatusBadRequest, "Invalid request body", nil)
	}

	if err := req.Validate(); err != nil {
		return common.NewResponse(c, http.StatusBadRequest, err.Error(), nil)
	}

	res, err := h.orderService.CreateUserCheckout(c.Context(), userID, req)
	if err != nil {
		if errors.Is(err, domain.ErrInsufficientStock) || errors.Is(err, domain.ErrWalletNotFound) || errors.Is(err, domain.ErrInsufficientBalance) {
			return common.NewResponse(c, http.StatusBadRequest, err.Error(), nil)
		}
		log.Printf("Error during checkout for user %s: %v", userID, err)
		return common.NewResponse(c, http.StatusInternalServerError, "Internal Server Error", nil)
	}

	return common.NewResponse(c, http.StatusCreated, "Checkout successful", res)
}

// UserCancelOrder allows a user to cancel their own order
// @Summary Cancel order (User)
// @Description Cancels a pending or processing order and initiates a refund if applicable.
// @Tags orders
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Order ID (UUID)"
// @Success 200 {object} common.ResponseWrapper
// @Failure 400 {object} common.ResponseWrapper
// @Failure 500 {object} common.ResponseWrapper
// @Router /users/orders/{id}/cancel [put]
func (h *OrderHandler) UserCancelOrder(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	orderID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return common.NewResponse(c, http.StatusBadRequest, "Invalid order ID", nil)
	}

	err = h.orderService.CancelUserOrder(c.Context(), userID, orderID)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotCancellable) || errors.Is(err, domain.ErrOrderNotFound) {
			return common.NewResponse(c, http.StatusBadRequest, err.Error(), nil)
		}
		log.Printf("Error cancelling order %s for user %s: %v", orderID, userID, err)
		return common.NewResponse(c, http.StatusInternalServerError, "Internal Server Error", nil)
	}

	return common.NewResponse(c, http.StatusOK, "Order cancelled and refunded", nil)
}

// UserAppealOrder allows a user to appeal an order status
// @Summary Appeal order (User)
// @Description Submits a formal appeal for an order, typically for disputes.
// @Tags orders
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Order ID (UUID)"
// @Param request body AppealOrderRequest true "Appeal Reason"
// @Success 200 {object} common.ResponseWrapper
// @Failure 400 {object} common.ResponseWrapper
// @Failure 500 {object} common.ResponseWrapper
// @Router /users/orders/{id}/appeal [post]
func (h *OrderHandler) UserAppealOrder(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	orderID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return common.NewResponse(c, http.StatusBadRequest, "Invalid order ID", nil)
	}

	var req AppealOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return common.NewResponse(c, http.StatusBadRequest, "Invalid request body", nil)
	}

	if err := req.Validate(); err != nil {
		return common.NewResponse(c, http.StatusBadRequest, err.Error(), nil)
	}

	err = h.orderService.AppealUserOrder(c.Context(), userID, orderID, req.Reason)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotCancellable) || errors.Is(err, domain.ErrOrderNotFound) {
			return common.NewResponse(c, http.StatusBadRequest, err.Error(), nil)
		}
		log.Printf("Error appealing order %s for user %s: %v", orderID, userID, err)
		return common.NewResponse(c, http.StatusInternalServerError, "Internal Server Error", nil)
	}

	return common.NewResponse(c, http.StatusOK, "Appeal submitted successfully", nil)
}

// MerchantUpdateStatus allows a merchant to update the status of an order
// @Summary Update order status (Merchant)
// @Description Allows the shop owner to change the lifecycle status of an order (e.g., to SHIPPED).
// @Tags orders
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Order ID (UUID)"
// @Param request body UpdateStatusRequest true "New Status"
// @Success 200 {object} common.ResponseWrapper
// @Failure 400 {object} common.ResponseWrapper
// @Failure 403 {object} common.ResponseWrapper
// @Failure 500 {object} common.ResponseWrapper
// @Router /merchants/orders/{id}/status [put]
func (h *OrderHandler) MerchantUpdateStatus(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	orderID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return common.NewResponse(c, http.StatusBadRequest, "Invalid order ID", nil)
	}

	// Find Merchant associated with this user
	m, err := h.merchantRepo.GetByUserID(c.Context(), userID)
	if err != nil {
		return common.NewResponse(c, http.StatusInternalServerError, "Error retrieving merchant profile", nil)
	}
	if m == nil {
		return common.NewResponse(c, http.StatusForbidden, "Only merchants can update order status", nil)
	}

	var req UpdateStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return common.NewResponse(c, http.StatusBadRequest, "Invalid request body", nil)
	}

	if err := req.Validate(); err != nil {
		return common.NewResponse(c, http.StatusBadRequest, err.Error(), nil)
	}

	err = h.orderService.MerchantUpdateStatus(c.Context(), m.ID, orderID, req.Status)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidStatusTransition) || errors.Is(err, domain.ErrMerchantShipmentTooEarly) || errors.Is(err, domain.ErrOrderNotFound) {
			return common.NewResponse(c, http.StatusBadRequest, err.Error(), nil)
		}
		log.Printf("Error updating order %s by merchant %s: %v", orderID, m.ID, err)
		return common.NewResponse(c, http.StatusInternalServerError, "Internal Server Error", nil)
	}

	return common.NewResponse(c, http.StatusOK, "Order status updated", nil)
}

// MerchantCancelOrder allows a merchant to cancel an order
// @Summary Cancel order (Merchant)
// @Description Allows the shop owner to cancel an order and initiate a refund.
// @Tags orders
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Order ID (UUID)"
// @Success 200 {object} common.ResponseWrapper
// @Failure 400 {object} common.ResponseWrapper
// @Failure 403 {object} common.ResponseWrapper
// @Failure 500 {object} common.ResponseWrapper
// @Router /merchants/orders/{id}/cancel [put]
func (h *OrderHandler) MerchantCancelOrder(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	orderID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return common.NewResponse(c, http.StatusBadRequest, "Invalid order ID", nil)
	}

	m, err := h.merchantRepo.GetByUserID(c.Context(), userID)
	if err != nil {
		return common.NewResponse(c, http.StatusInternalServerError, "Error retrieving merchant profile", nil)
	}
	if m == nil {
		return common.NewResponse(c, http.StatusForbidden, "Only merchants can cancel orders", nil)
	}

	err = h.orderService.MerchantCancelOrder(c.Context(), m.ID, orderID)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotCancellable) || errors.Is(err, domain.ErrOrderNotFound) {
			return common.NewResponse(c, http.StatusBadRequest, err.Error(), nil)
		}
		log.Printf("Error cancelling order %s by merchant %s: %v", orderID, m.ID, err)
		return common.NewResponse(c, http.StatusInternalServerError, "Internal Server Error", nil)
	}

	return common.NewResponse(c, http.StatusOK, "Order cancelled and refunded", nil)
}

// GetOrderDetail retrieves details for a specific order
// @Summary Get order details
// @Description Fetches full information for a specific order by its ID.
// @Tags orders
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Order ID (UUID)"
// @Success 200 {object} common.ResponseWrapper{data=OrderResponse}
// @Failure 400 {object} common.ResponseWrapper
// @Failure 404 {object} common.ResponseWrapper
// @Failure 500 {object} common.ResponseWrapper
// @Router /users/orders/{id} [get]
func (h *OrderHandler) GetOrderDetail(c *fiber.Ctx) error {
	orderID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return common.NewResponse(c, http.StatusBadRequest, "Invalid order ID", nil)
	}

	res, err := h.orderService.GetOrder(c.Context(), orderID)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			return common.NewResponse(c, http.StatusNotFound, err.Error(), nil)
		}
		log.Printf("Error getting order %s: %v", orderID, err)
		return common.NewResponse(c, http.StatusInternalServerError, "Internal Server Error", nil)
	}

	return common.NewResponse(c, http.StatusOK, "Order details retrieved", res)
}
