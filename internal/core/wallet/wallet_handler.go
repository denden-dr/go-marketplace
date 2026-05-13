package wallet

import (
	"go-marketplace/internal/common"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type WalletHandler struct {
	walletService WalletService
}

func NewWalletHandler(walletService WalletService) *WalletHandler {
	return &WalletHandler{walletService: walletService}
}

// GetWallet retrieves user's wallet details
// @Summary Get wallet
// @Description Fetches the authenticated user's digital wallet information.
// @Tags wallets
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} common.SuccessResponse{data=WalletResponse}
// @Failure 404 {object} common.ProblemDetails
// @Failure 500 {object} common.ProblemDetails
// @Router /wallets/ [get]
func (h *WalletHandler) GetWallet(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	wallet, err := h.walletService.GetWalletByUserID(c.Context(), userID)
	if err != nil {
		return err
	}

	return common.NewSuccessResponse(c, http.StatusOK, "Wallet details retrieved", wallet)
}

// GetHistory retrieves wallet transaction history
// @Summary Get wallet history
// @Description Fetches a paginated list of transactions for the user's wallet.
// @Tags wallets
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} common.SuccessResponse{data=[]TransactionResponse}
// @Failure 404 {object} common.ProblemDetails
// @Failure 500 {object} common.ProblemDetails
// @Router /wallets/history [get]
func (h *WalletHandler) GetHistory(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	history, err := h.walletService.GetWalletHistory(c.Context(), userID, page, limit)
	if err != nil {
		return err
	}

	return common.NewSuccessResponse(c, http.StatusOK, "Wallet history retrieved", history)
}

// Withdraw performs a wallet withdrawal
// @Summary Withdraw from wallet
// @Description Deducts funds from the user's wallet balance.
// @Tags wallets
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body WithdrawRequest true "Withdrawal Info"
// @Success 200 {object} common.SuccessResponse
// @Failure 400 {object} common.ProblemDetails
// @Failure 500 {object} common.ProblemDetails
// @Router /wallets/withdraw [post]
func (h *WalletHandler) Withdraw(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	var req WithdrawRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(http.StatusBadRequest, "Invalid request body")
	}

	if err := req.Validate(); err != nil {
		return err
	}

	err := h.walletService.Withdraw(c.Context(), userID, req)
	if err != nil {
		return err
	}

	return common.NewSuccessResponse(c, http.StatusOK, "Withdrawal successful", nil)
}

// CreateWallet initializes a wallet for the user
// @Summary Create wallet
// @Description Creates a new digital wallet for the authenticated user.
// @Tags wallets
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 201 {object} common.SuccessResponse{data=WalletResponse}
// @Failure 409 {object} common.ProblemDetails
// @Failure 500 {object} common.ProblemDetails
// @Router /wallets/ [post]
func (h *WalletHandler) CreateWallet(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	wallet, err := h.walletService.CreateWallet(c.Context(), userID)
	if err != nil {
		return err
	}

	return common.NewSuccessResponse(c, http.StatusCreated, "Wallet created successfully", wallet)
}
