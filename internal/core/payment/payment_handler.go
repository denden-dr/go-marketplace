package payment

import (
	"go-marketplace/internal/common"
	"go-marketplace/internal/domain"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type PaymentHandler struct {
	paymentService PaymentService
}

func NewPaymentHandler(paymentService PaymentService) *PaymentHandler {
	return &PaymentHandler{paymentService: paymentService}
}

// Topup handles wallet top-up requests
func (h *PaymentHandler) Topup(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	var req TopupRequest
	if err := c.BodyParser(&req); err != nil {
		return common.NewResponse(c, http.StatusBadRequest, "Invalid request body", nil)
	}

	if err := req.Validate(); err != nil {
		return common.NewResponse(c, http.StatusBadRequest, err.Error(), nil)
	}

	res, err := h.paymentService.CreatePaymentTX(c.Context(), nil, CreatePaymentRequest{
		UserID:      userID,
		Amount:      req.Amount,
		Type:        domain.PaymentTypeTopup,
		Method:      req.Method,
		ReferenceID: uuid.New(),
	})
	if err != nil {
		return common.NewResponse(c, http.StatusInternalServerError, err.Error(), nil)
	}

	return common.NewResponse(c, http.StatusOK, "Topup initiated", res)
}

// Webhook handles Midtrans status notifications
func (h *PaymentHandler) Webhook(c *fiber.Ctx) error {
	var req MidtransWebhookRequest
	if err := c.BodyParser(&req); err != nil {
		return c.SendStatus(http.StatusBadRequest)
	}

	status := domain.PaymentStatusPending
	switch req.TransactionStatus {
	case "settlement", "capture":
		status = domain.PaymentStatusSuccess
	case "deny", "cancel":
		status = domain.PaymentStatusFailed
	case "expire":
		status = domain.PaymentStatusExpired
	case "pending":
		status = domain.PaymentStatusPending
	}

	if err := h.paymentService.ProcessWebhook(c.Context(), req.OrderID, status); err != nil {
		return c.SendStatus(http.StatusInternalServerError)
	}

	return c.SendStatus(http.StatusOK)
}
