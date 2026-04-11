package handlers

import (
	"go-shop-yourself/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(
	app *fiber.App,
	authHandler *AuthHandler,
	userHandler *UserHandler,
	merchantHandler *MerchantHandler,
	productHandler *ProductHandler,
	walletHandler *WalletHandler,
	jwtSecret string,
) {
	// Map initial route
	app.Get("/", HelloHandler)

	// Public Auth routes
	authRoutes := app.Group("/auth")
	authRoutes.Post("/register", authHandler.Register)
	authRoutes.Post("/login", authHandler.Login)
	authRoutes.Post("/refresh", authHandler.RefreshTokens)

	// Middleware for protected routes
	authMiddleware := middleware.AuthMiddleware(jwtSecret)

	// Protected routes
	api := app.Group("", authMiddleware)

	// Logout and Merchant registration
	api.Post("/auth/logout", authHandler.Logout)
	api.Post("/auth/register-merchant", merchantHandler.RegisterMerchant)

	// User routes
	users := api.Group("/users")
	users.Get("/:id", userHandler.GetUserByID)

	// Product routes
	products := api.Group("/products")
	products.Post("/", productHandler.CreateProduct)
	products.Put("/:id", productHandler.UpdateProduct)

	// Wallet routes
	wallets := api.Group("/wallets")
	wallets.Post("/", walletHandler.CreateWallet)
	wallets.Get("/", walletHandler.GetWallet)
	wallets.Get("/history", walletHandler.GetHistory)
	wallets.Post("/withdraw", walletHandler.Withdraw)
}
