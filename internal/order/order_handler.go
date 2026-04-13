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
