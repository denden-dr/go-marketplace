package server

import (
	"go-marketplace/internal/common"
	"go-marketplace/internal/core/auth"
	"go-marketplace/internal/core/cart"
	"go-marketplace/internal/core/health"
	"go-marketplace/internal/core/merchant"
	"go-marketplace/internal/core/order"
	"go-marketplace/internal/core/product"
	"go-marketplace/internal/core/user"
	"go-marketplace/internal/core/wallet"
	"go-marketplace/internal/middleware"

	_ "go-marketplace/docs"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/swagger"
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
	authRoutes.Use(limiter.New(limiter.Config{
		Max:        10,
		Expiration: 1 * time.Minute,
		LimitReached: func(c *fiber.Ctx) error {
			return common.NewResponse(c, fiber.StatusTooManyRequests, "Too many requests, please try again later", nil)
		},
	}))

	authRoutes.Post("/register", authHandler.Register)
	authRoutes.Post("/login", authHandler.Login)
	authRoutes.Post("/verify-email", authHandler.VerifyEmail)
	authRoutes.Post("/refresh", authHandler.RefreshTokens)

	// Public products
	publicProducts := apiBase.Group("/products")
	publicProducts.Get("/search", productHandler.SearchProducts)

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
	merchantOrders := merchants.Group("/orders", order.MerchantMiddleware(orderHandler.MerchantRepo))
	merchantOrders.Put("/:id/cancel", orderHandler.MerchantCancelOrder)
	merchantOrders.Put("/:id/status", orderHandler.MerchantUpdateStatus)
}
