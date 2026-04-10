package handlers

import (
	"go-shop-yourself/internal/dtos"
	"go-shop-yourself/internal/services"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type WalletHandler struct {
	walletService *services.WalletService
}

func NewWalletHandler(walletService *services.WalletService) *WalletHandler {
	return &WalletHandler{walletService: walletService}
}

func (h *WalletHandler) GetWallet(c *fiber.Ctx) error {
	idStr := c.Query("user_id")
	if idStr == "" {
		return dtos.NewResponse(c, http.StatusBadRequest, "Missing user_id query parameter", nil)
	}

	userID, err := uuid.Parse(idStr)
	if err != nil {
		return dtos.NewResponse(c, http.StatusBadRequest, "Invalid user_id format", nil)
	}

	wallet, err := h.walletService.GetWalletByUserID(c.Context(), userID)
	if err != nil {
		return dtos.NewResponse(c, http.StatusInternalServerError, err.Error(), nil)
	}

	return dtos.NewResponse(c, http.StatusOK, "Wallet details retrieved", wallet)
}

func (h *WalletHandler) GetHistory(c *fiber.Ctx) error {
	idStr := c.Query("user_id")
	if idStr == "" {
		return dtos.NewResponse(c, http.StatusBadRequest, "Missing user_id query parameter", nil)
	}

	userID, err := uuid.Parse(idStr)
	if err != nil {
		return dtos.NewResponse(c, http.StatusBadRequest, "Invalid user_id format", nil)
	}

	history, err := h.walletService.GetWalletHistory(c.Context(), userID)
	if err != nil {
		return dtos.NewResponse(c, http.StatusInternalServerError, err.Error(), nil)
	}

	return dtos.NewResponse(c, http.StatusOK, "Wallet history retrieved", history)
}

func (h *WalletHandler) Withdraw(c *fiber.Ctx) error {
	idStr := c.Query("user_id")
	if idStr == "" {
		return dtos.NewResponse(c, http.StatusBadRequest, "Missing user_id query parameter", nil)
	}

	userID, err := uuid.Parse(idStr)
	if err != nil {
		return dtos.NewResponse(c, http.StatusBadRequest, "Invalid user_id format", nil)
	}

	var req dtos.WithdrawRequest
	if err := c.BodyParser(&req); err != nil {
		return dtos.NewResponse(c, http.StatusBadRequest, "Invalid request body", nil)
	}

	if req.Amount <= 0 {
		return dtos.NewResponse(c, http.StatusBadRequest, "Amount must be greater than 0", nil)
	}

	err = h.walletService.Withdraw(c.Context(), userID, req)
	if err != nil {
		return dtos.NewResponse(c, http.StatusBadRequest, err.Error(), nil)
	}

	return dtos.NewResponse(c, http.StatusOK, "Withdrawal successful", nil)
}
