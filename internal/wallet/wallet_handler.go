package wallet

import (
	"errors"
	"go-shop-yourself/internal/domain"
	"go-shop-yourself/internal/common"
	"log"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type WalletHandler struct {
	walletService WalletServiceInterface
}

func NewWalletHandler(walletService WalletServiceInterface) *WalletHandler {
	return &WalletHandler{walletService: walletService}
}

func (h *WalletHandler) GetWallet(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	wallet, err := h.walletService.GetWalletByUserID(c.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrWalletNotFound) {
			return common.NewResponse(c, http.StatusNotFound, err.Error(), nil)
		}
		log.Printf("Error getting wallet for user %s: %v", userID, err)
		return common.NewResponse(c, http.StatusInternalServerError, "Internal Server Error", nil)
	}

	return common.NewResponse(c, http.StatusOK, "Wallet details retrieved", wallet)
}

func (h *WalletHandler) GetHistory(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	history, err := h.walletService.GetWalletHistory(c.Context(), userID, page, limit)
	if err != nil {
		if errors.Is(err, domain.ErrWalletNotFound) {
			return common.NewResponse(c, http.StatusNotFound, err.Error(), nil)
		}
		log.Printf("Error getting wallet history for user %s: %v", userID, err)
		return common.NewResponse(c, http.StatusInternalServerError, "Internal Server Error", nil)
	}

	return common.NewResponse(c, http.StatusOK, "Wallet history retrieved", history)
}

func (h *WalletHandler) Withdraw(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	var req WithdrawRequest
	if err := c.BodyParser(&req); err != nil {
		return common.NewResponse(c, http.StatusBadRequest, "Invalid request body", nil)
	}

	err := h.walletService.Withdraw(c.Context(), userID, req)
	if err != nil {
		if errors.Is(err, domain.ErrInsufficientBalance) || errors.Is(err, domain.ErrWalletNotActive) || errors.Is(err, domain.ErrWalletNotFound) {
			return common.NewResponse(c, http.StatusBadRequest, err.Error(), nil)
		}
		log.Printf("Internal error during withdrawal for user %s: %v", userID, err)
		return common.NewResponse(c, http.StatusInternalServerError, "Internal Server Error", nil)
	}

	return common.NewResponse(c, http.StatusOK, "Withdrawal successful", nil)
}

func (h *WalletHandler) CreateWallet(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	wallet, err := h.walletService.CreateWallet(c.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrWalletAlreadyExists) {
			return common.NewResponse(c, http.StatusConflict, err.Error(), nil)
		}
		log.Printf("Error creating wallet for user %s: %v", userID, err)
		return common.NewResponse(c, http.StatusInternalServerError, "Internal Server Error", nil)
	}

	return common.NewResponse(c, http.StatusCreated, "Wallet created successfully", wallet)
}
