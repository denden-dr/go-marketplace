package main

import (
	"context"
	"log"
	"os"
	"time"

	"go-marketplace/internal/core/auth"
	"go-marketplace/internal/core/cart"
	"go-marketplace/internal/core/health"
	"go-marketplace/internal/core/merchant"
	"go-marketplace/internal/core/order"
	"go-marketplace/internal/core/product"
	"go-marketplace/internal/core/user"
	"go-marketplace/internal/core/wallet"
	"go-marketplace/internal/database"

	"go-marketplace/internal/server"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

// @title Go Marketplace API
// @version 1.0
// @description This is a robust e-commerce backend API.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:3000
// @BasePath /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and then your token.

func main() {
	// Load .env configuration
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using default values")
	}

	// Initialize Database
	db, err := database.ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatalf("JWT_SECRET environment variable is not set")
	}

	appEnv := os.Getenv("APP_ENV")

	// Initialize Supabase Auth
	supabaseJWTSecret := os.Getenv("SUPABASE_JWT_SECRET")
	var socialAuthClient auth.SupabaseAuthClient
	if supabaseJWTSecret != "" {
		socialAuthClient = auth.NewSupabaseAuthClient(supabaseJWTSecret)
		log.Println("Supabase Auth initialized successfully")
	} else {
		log.Println("Warning: SUPABASE_JWT_SECRET not set. Social login will be disabled.")
	}

	// Initialize Layers
	userRepo := user.NewUserRepository(db)
	merchantRepo := merchant.NewMerchantRepository(db)
	productRepo := product.NewProductRepository(db)
	walletRepo := wallet.NewWalletRepository(db)
	refreshTokenRepo := auth.NewRefreshTokenRepository(db)
	cartRepo := cart.NewCartRepository(db)
	orderRepo := order.NewOrderRepository(db)

	// Background cleanup for expired refresh tokens
	log.Printf("Running initial background cleanup for expired refresh tokens...")
	if err := refreshTokenRepo.DeleteExpiredTokens(context.Background()); err != nil {
		log.Printf("Error during initial background cleanup: %v", err)
	}

	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			if err := refreshTokenRepo.DeleteExpiredTokens(context.Background()); err != nil {
				log.Printf("Error cleaning up expired refresh tokens: %v", err)
			}
		}
	}()

	authService := auth.NewAuthService(userRepo, refreshTokenRepo, socialAuthClient, jwtSecret)
	userService := user.NewUserService(userRepo)
	merchantService := merchant.NewMerchantService(merchantRepo, userRepo, walletRepo)
	productService := product.NewProductService(productRepo, merchantRepo)
	walletService := wallet.NewWalletService(walletRepo)
	cartService := cart.NewCartService(cartRepo, productRepo)
	orderService := order.NewOrderService(orderRepo, cartRepo, productRepo, walletRepo, userRepo)

	authHandler := auth.NewAuthHandler(authService)
	userHandler := user.NewUserHandler(userService)
	merchantHandler := merchant.NewMerchantHandler(merchantService)
	productHandler := product.NewProductHandler(productService)
	walletHandler := wallet.NewWalletHandler(walletService)
	cartHandler := cart.NewCartHandler(cartService)
	orderHandler := order.NewOrderHandler(orderService, merchantRepo)
	healthHandler := health.NewHealthHandler(db)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	// Initialize Fiber app
	app := fiber.New()

	// Setup Routes
	socialLoginEnabled := socialAuthClient != nil
	server.SetupRoutes(
		app,
		authHandler,
		userHandler,
		merchantHandler, productHandler, walletHandler,
		cartHandler, orderHandler, healthHandler, jwtSecret, appEnv, socialLoginEnabled)

	// Start server
	log.Printf("Server starting on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
