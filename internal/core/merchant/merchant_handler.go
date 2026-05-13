package merchant

import (
	"go-marketplace/internal/common"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type MerchantHandler struct {
	service MerchantService
}

func NewMerchantHandler(service MerchantService) *MerchantHandler {
	return &MerchantHandler{service: service}
}

// RegisterMerchant registers a user as a merchant
// @Summary Register as merchant
// @Description Allows an authenticated user to register their own shop/merchant profile.
// @Tags merchants
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body MerchantRegisterRequest true "Merchant Registration Info"
// @Success 201 {object} common.SuccessResponse{data=MerchantResponse}
// @Failure 400 {object} common.ProblemDetails
// @Failure 404 {object} common.ProblemDetails
// @Failure 409 {object} common.ProblemDetails
// @Failure 500 {object} common.ProblemDetails
// @Router /auth/register-merchant [post]
func (h *MerchantHandler) RegisterMerchant(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	var req MerchantRegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(http.StatusBadRequest, "Invalid request payload")
	}

	if err := req.Validate(); err != nil {
		return err
	}

	res, err := h.service.RegisterMerchant(c.Context(), userID, req)
	if err != nil {
		return err
	}

	return common.NewSuccessResponse(c, http.StatusCreated, "Merchant registered successfully", res)
}
