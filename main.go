package main

import (
	"log"
	"os"

	"go-shop-yourself/internal/database"
	"go-shop-yourself/internal/handlers"
	"go-shop-yourself/internal/middleware"
	"go-shop-yourself/internal/repos"
	"go-shop-yourself/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env configuration
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using default values")
	}

	// Initialize Database
	database.ConnectDB()

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatalf("JWT_SECRET environment variable is not set")
	}

	// Initialize Layers
	userRepo := repos.NewUserRepository(database.DB)
	merchantRepo := repos.NewMerchantRepository(database.DB)
	productRepo := repos.NewProductRepository(database.DB)
	walletRepo := repos.NewWalletRepository(database.DB)
	refreshTokenRepo := repos.NewRefreshTokenRepository(database.DB)

	authService := services.NewAuthService(userRepo, refreshTokenRepo, jwtSecret)
	userService := services.NewUserService(userRepo)
	merchantService := services.NewMerchantService(merchantRepo, userRepo, walletRepo)
	productService := services.NewProductService(productRepo, merchantRepo)
	walletService := services.NewWalletService(walletRepo)

	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)
	merchantHandler := handlers.NewMerchantHandler(merchantService)
	productHandler := handlers.NewProductHandler(productService)
	walletHandler := handlers.NewWalletHandler(walletService)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	// Initialize Fiber app
	app := fiber.New()

	// Map routes
	app.Get("/", handlers.HelloHandler)

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

	// Start server
	log.Printf("Server starting on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
