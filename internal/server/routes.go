package server

import (
	"go-shop-yourself/internal/auth"
	"go-shop-yourself/internal/merchant"
	"go-shop-yourself/internal/middleware"
	"go-shop-yourself/internal/product"
	"go-shop-yourself/internal/user"
	"go-shop-yourself/internal/wallet"
	"go-shop-yourself/internal/cart"
	"go-shop-yourself/internal/order"

	"github.com/gofiber/fiber/v2"
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
	jwtSecret string,
) {
	// Create API base group
	apiBase := app.Group("/api")

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

	// V1 Routes
	v1 := api.Group("/v1")

	// User features (Cart & Orders)
	userV1 := v1.Group("/user")
	
	// Cart
	cartRoutes := userV1.Group("/cart")
	cartRoutes.Get("/", cartHandler.GetCart)
	cartRoutes.Post("/", cartHandler.AddToCart)
	cartRoutes.Put("/:id", cartHandler.UpdateCartItem)
	cartRoutes.Delete("/:id", cartHandler.RemoveFromCart)
	cartRoutes.Delete("/", cartHandler.ClearCart)

	// Orders
	orderRoutes := userV1.Group("/orders")
	orderRoutes.Post("/", orderHandler.Checkout)
	orderRoutes.Get("/:id", orderHandler.GetOrderDetail)
	orderRoutes.Put("/:id/cancel", orderHandler.UserCancelOrder)
	orderRoutes.Post("/:id/appeal", orderHandler.UserAppealOrder)

	// Merchant features (Orders)
	merchantV1 := v1.Group("/merchant")
	merchantOrders := merchantV1.Group("/orders")
	merchantOrders.Put("/:id/cancel", orderHandler.MerchantCancelOrder)
	merchantOrders.Put("/:id/status", orderHandler.MerchantUpdateStatus)
}
