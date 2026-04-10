package handlers

import (
	"go-shop-yourself/internal/dtos"
	"go-shop-yourself/internal/services"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"log"
	"strconv"
)

type WalletHandler struct {
	walletService *services.WalletService
}

func NewWalletHandler(walletService *services.WalletService) *WalletHandler {
	return &WalletHandler{walletService: walletService}
}

func (h *WalletHandler) parseUserID(c *fiber.Ctx) (uuid.UUID, error) {
	idStr := c.Query("user_id")
	if idStr == "" {
		return uuid.Nil, fiber.NewError(http.StatusBadRequest, "Missing user_id query parameter")
	}

	userID, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, fiber.NewError(http.StatusBadRequest, "Invalid user_id format")
	}

	return userID, nil
}

func (h *WalletHandler) GetWallet(c *fiber.Ctx) error {
	userID, err := h.parseUserID(c)
	if err != nil {
		fe := err.(*fiber.Error)
		return dtos.NewResponse(c, fe.Code, fe.Message, nil)
	}

	wallet, err := h.walletService.GetWalletByUserID(c.Context(), userID)
	if err != nil {
		log.Printf("Error getting wallet for user %s: %v", userID, err)
		return dtos.NewResponse(c, http.StatusInternalServerError, "Internal Server Error", nil)
	}

	return dtos.NewResponse(c, http.StatusOK, "Wallet details retrieved", wallet)
}

func (h *WalletHandler) GetHistory(c *fiber.Ctx) error {
	userID, err := h.parseUserID(c)
	if err != nil {
		fe := err.(*fiber.Error)
		return dtos.NewResponse(c, fe.Code, fe.Message, nil)
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	history, err := h.walletService.GetWalletHistory(c.Context(), userID, page, limit)
	if err != nil {
		log.Printf("Error getting wallet history for user %s: %v", userID, err)
		return dtos.NewResponse(c, http.StatusInternalServerError, "Internal Server Error", nil)
	}

	return dtos.NewResponse(c, http.StatusOK, "Wallet history retrieved", history)
}

func (h *WalletHandler) Withdraw(c *fiber.Ctx) error {
	userID, err := h.parseUserID(c)
	if err != nil {
		fe := err.(*fiber.Error)
		return dtos.NewResponse(c, fe.Code, fe.Message, nil)
	}

	var req dtos.WithdrawRequest
	if err := c.BodyParser(&req); err != nil {
		return dtos.NewResponse(c, http.StatusBadRequest, "Invalid request body", nil)
	}

	if req.Amount.IsZero() || req.Amount.IsNegative() {
		return dtos.NewResponse(c, http.StatusBadRequest, "Amount must be greater than 0", nil)
	}

	err = h.walletService.Withdraw(c.Context(), userID, req)
	if err != nil {
		// Differentiate between user-facing errors and internal errors
		// For simplicity, we check if it's one of our known business errors
		if err.Error() == "insufficient balance" || err.Error() == "wallet is not active" || err.Error() == "wallet not found" {
			return dtos.NewResponse(c, http.StatusBadRequest, err.Error(), nil)
		}
		log.Printf("Internal error during withdrawal for user %s: %v", userID, err)
		return dtos.NewResponse(c, http.StatusInternalServerError, "Internal Server Error", nil)
	}

	return dtos.NewResponse(c, http.StatusOK, "Withdrawal successful", nil)
}
