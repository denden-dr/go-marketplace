package server

import (
	"go-shop-yourself/internal/auth"
	"go-shop-yourself/internal/cart"
	"go-shop-yourself/internal/merchant"
	"go-shop-yourself/internal/middleware"
	"go-shop-yourself/internal/order"
	"go-shop-yourself/internal/product"
	"go-shop-yourself/internal/user"
	"go-shop-yourself/internal/wallet"
	"go-shop-yourself/internal/health"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	_ "go-shop-yourself/docs"
)

func SetupRoutes(
	app *fiber.App,
	authHandler *auth.AuthHandler,
	userHandler *user.UserHandler,
	merchantHandler *merchant.MerchantHandler,
	productHandler *product.ProductHandler,
	walletHandler *wallet.WalletHandler,
	cartHandler *cart.CartHandler,
	orderHandler *order.OrderHandler,
	healthHandler *health.HealthHandler,
	jwtSecret string,
	appEnv string,
) {
	// Swagger UI
	if appEnv == "development" {
		app.Get("/swagger/*", swagger.HandlerDefault)
	}

	// Create API base group
	apiBase := app.Group("/api")

	// Health check (Public)
	apiBase.Get("/health", healthHandler.CheckStatus)

	// Public Auth routes
	authRoutes := apiBase.Group("/auth")
	authRoutes.Post("/register", authHandler.Register)
	authRoutes.Post("/login", authHandler.Login)
	authRoutes.Post("/refresh", authHandler.RefreshTokens)

	// Middleware for protected routes
	authMiddleware := middleware.AuthMiddleware(jwtSecret)

	// Protected routes (under /api)
	api := apiBase.Group("", authMiddleware)

	// Logout and Merchant registration
	api.Post("/auth/logout", authHandler.Logout)
	api.Post("/auth/register-merchant", merchantHandler.RegisterMerchant)

	// User routes
	users := api.Group("/users")
	users.Get("/addresses", userHandler.ListAddresses)
	users.Post("/addresses", userHandler.AddAddress)
	users.Put("/addresses/:id", userHandler.UpdateAddress)
	users.Delete("/addresses/:id", userHandler.DeleteAddress)
	users.Get("/me", userHandler.GetProfile)

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

	// Cart
	cartRoutes := users.Group("/cart")
	cartRoutes.Get("/", cartHandler.GetCart)
	cartRoutes.Post("/", cartHandler.AddToCart)
	cartRoutes.Put("/:productID", cartHandler.UpdateCartItem)
	cartRoutes.Delete("/:productID", cartHandler.RemoveFromCart)
	cartRoutes.Delete("/", cartHandler.ClearCart)

	// Orders
	orderRoutes := users.Group("/orders")
	orderRoutes.Post("/", orderHandler.Checkout)
	orderRoutes.Get("/:id", orderHandler.GetOrderDetail)
	orderRoutes.Put("/:id/cancel", orderHandler.UserCancelOrder)
	orderRoutes.Post("/:id/appeal", orderHandler.UserAppealOrder)

	// Merchant features (Orders)
	merchants := api.Group("/merchants")
	merchantOrders := merchants.Group("/orders")
	merchantOrders.Put("/:id/cancel", orderHandler.MerchantCancelOrder)
	merchantOrders.Put("/:id/status", orderHandler.MerchantUpdateStatus)
}
