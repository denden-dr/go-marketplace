package main

import (
	"log"
	"os"

	"go-shop-yourself/internal/auth"
	"go-shop-yourself/internal/cart"
	"go-shop-yourself/internal/database"
	"go-shop-yourself/internal/merchant"
	"go-shop-yourself/internal/order"
	"go-shop-yourself/internal/product"
	"go-shop-yourself/internal/server"
	"go-shop-yourself/internal/user"
	"go-shop-yourself/internal/wallet"
	"go-shop-yourself/internal/health"
	"go-shop-yourself/internal/opensearch"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

// @title Go Shop Yourself API
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

	// Initialize OpenSearch
	osClient, err := opensearch.ConnectOpenSearch()
	if err != nil {
		log.Printf("Warning: Failed to connect to OpenSearch: %v", err)
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatalf("JWT_SECRET environment variable is not set")
	}

	// Initialize Layers
	userRepo := user.NewUserRepository(db)
	merchantRepo := merchant.NewMerchantRepository(db)
	productRepo := product.NewProductRepository(db)
	walletRepo := wallet.NewWalletRepository(db)
	refreshTokenRepo := auth.NewRefreshTokenRepository(db)
	cartRepo := cart.NewCartRepository(db)
	orderRepo := order.NewOrderRepository(db)

	authService := auth.NewAuthService(userRepo, refreshTokenRepo, jwtSecret)
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
	healthHandler := health.NewHealthHandler(db, osClient)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	// Initialize Fiber app
	app := fiber.New()

	appEnv := os.Getenv("APP_ENV")

	// Setup Routes
	server.SetupRoutes(
		app,
		authHandler,
		userHandler,
		merchantHandler, productHandler, walletHandler,
		cartHandler, orderHandler, healthHandler, jwtSecret, appEnv)

	// Start server
	log.Printf("Server starting on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
