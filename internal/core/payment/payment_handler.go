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
// @Summary Top up wallet
// @Description Initiates a wallet top-up payment.
// @Tags payment
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body TopupRequest true "Topup Info"
// @Success 200 {object} common.SuccessResponse{data=domain.Payment}
// @Failure 400 {object} common.ProblemDetails
// @Failure 500 {object} common.ProblemDetails
// @Router /wallets/topup [post]
func (h *PaymentHandler) Topup(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	var req TopupRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(http.StatusBadRequest, "Invalid request body")
	}

	if err := req.Validate(); err != nil {
		return err
	}

	res, err := h.paymentService.CreatePaymentTX(c.Context(), nil, CreatePaymentRequest{
		UserID:      userID,
		Amount:      req.Amount,
		Type:        domain.PaymentTypeTopup,
		Method:      req.Method,
		ReferenceID: uuid.New(),
	})
	if err != nil {
		return err
	}

	return common.NewSuccessResponse(c, http.StatusOK, "Topup initiated", res)
}

// Webhook handles Midtrans status notifications
// @Summary Payment Webhook
// @Description Handles payment status webhooks from Midtrans.
// @Tags payment
// @Accept json
// @Produce json
// @Param request body MidtransWebhookRequest true "Webhook Payload"
// @Success 200
// @Failure 400
// @Failure 500 {object} common.ProblemDetails
// @Router /payments/webhook [post]
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
		return err
	}

	return c.SendStatus(http.StatusOK)
}
