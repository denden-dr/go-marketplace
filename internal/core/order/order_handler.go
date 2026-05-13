package order

import (
	"go-marketplace/internal/common"
	"go-marketplace/internal/core/merchant"
	"go-marketplace/internal/domain"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type OrderHandler struct {
	orderService OrderService
	MerchantRepo merchant.MerchantRepository
}

func NewOrderHandler(orderService OrderService, merchantRepo merchant.MerchantRepository) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
		MerchantRepo: merchantRepo,
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
// @Success 201 {object} common.SuccessResponse{data=OrderResponse}
// @Failure 400 {object} common.ProblemDetails
// @Failure 500 {object} common.ProblemDetails
// @Router /users/orders [post]
func (h *OrderHandler) Checkout(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	var req CheckoutRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(http.StatusBadRequest, "Invalid request body")
	}

	if err := req.Validate(); err != nil {
		return err
	}

	res, err := h.orderService.CreateUserCheckout(c.Context(), userID, req)
	if err != nil {
		return err
	}

	return common.NewSuccessResponse(c, http.StatusCreated, "Checkout successful", res)
}

// UserCancelOrder allows a user to cancel their own order
// @Summary Cancel order (User)
// @Description Cancels a pending or processing order and initiates a refund if applicable.
// @Tags orders
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Order ID (UUID)"
// @Success 200 {object} common.SuccessResponse
// @Failure 400 {object} common.ProblemDetails
// @Failure 500 {object} common.ProblemDetails
// @Router /users/orders/{id}/cancel [put]
func (h *OrderHandler) UserCancelOrder(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	orderID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "Invalid order ID")
	}

	err = h.orderService.CancelUserOrder(c.Context(), userID, orderID)
	if err != nil {
		return err
	}

	return common.NewSuccessResponse(c, http.StatusOK, "Order cancelled and refunded", nil)
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
// @Success 200 {object} common.SuccessResponse
// @Failure 400 {object} common.ProblemDetails
// @Failure 500 {object} common.ProblemDetails
// @Router /users/orders/{id}/appeal [post]
func (h *OrderHandler) UserAppealOrder(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	orderID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "Invalid order ID")
	}

	var req AppealOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(http.StatusBadRequest, "Invalid request body")
	}

	if err := req.Validate(); err != nil {
		return err
	}

	err = h.orderService.AppealUserOrder(c.Context(), userID, orderID, req.Reason)
	if err != nil {
		return err
	}

	return common.NewSuccessResponse(c, http.StatusOK, "Appeal submitted successfully", nil)
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
// @Success 200 {object} common.SuccessResponse
// @Failure 400 {object} common.ProblemDetails
// @Failure 403 {object} common.ProblemDetails
// @Failure 500 {object} common.ProblemDetails
// @Router /merchants/orders/{id}/status [put]
func (h *OrderHandler) MerchantUpdateStatus(c *fiber.Ctx) error {
	merchant := c.Locals("merchant").(*domain.Merchant)
	orderID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "Invalid order ID")
	}

	var req UpdateStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(http.StatusBadRequest, "Invalid request body")
	}

	if err := req.Validate(); err != nil {
		return err
	}

	err = h.orderService.MerchantUpdateStatus(c.Context(), merchant.ID, orderID, req.Status)
	if err != nil {
		return err
	}

	return common.NewSuccessResponse(c, http.StatusOK, "Order status updated", nil)
}

// MerchantCancelOrder allows a merchant to cancel an order
// @Summary Cancel order (Merchant)
// @Description Allows the shop owner to cancel an order and initiate a refund.
// @Tags orders
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Order ID (UUID)"
// @Success 200 {object} common.SuccessResponse
// @Failure 400 {object} common.ProblemDetails
// @Failure 403 {object} common.ProblemDetails
// @Failure 500 {object} common.ProblemDetails
// @Router /merchants/orders/{id}/cancel [put]
func (h *OrderHandler) MerchantCancelOrder(c *fiber.Ctx) error {
	merchant := c.Locals("merchant").(*domain.Merchant)
	orderID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "Invalid order ID")
	}

	err = h.orderService.MerchantCancelOrder(c.Context(), merchant.ID, orderID)
	if err != nil {
		return err
	}

	return common.NewSuccessResponse(c, http.StatusOK, "Order cancelled and refunded", nil)
}

// GetOrderDetail retrieves details for a specific order
// @Summary Get order details
// @Description Fetches full information for a specific order by its ID.
// @Tags orders
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Order ID (UUID)"
// @Success 200 {object} common.SuccessResponse{data=OrderResponse}
// @Failure 400 {object} common.ProblemDetails
// @Failure 404 {object} common.ProblemDetails
// @Failure 500 {object} common.ProblemDetails
// @Router /users/orders/{id} [get]
func (h *OrderHandler) GetOrderDetail(c *fiber.Ctx) error {
	orderID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "Invalid order ID")
	}

	res, err := h.orderService.GetOrder(c.Context(), orderID)
	if err != nil {
		return err
	}

	return common.NewSuccessResponse(c, http.StatusOK, "Order details retrieved", res)
}
