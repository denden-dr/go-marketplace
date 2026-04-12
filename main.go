package main

import (
	"log"
	"os"

	"go-shop-yourself/internal/auth"
	"go-shop-yourself/internal/database"
	"go-shop-yourself/internal/merchant"
	"go-shop-yourself/internal/product"
	"go-shop-yourself/internal/server"
	"go-shop-yourself/internal/user"
	"go-shop-yourself/internal/wallet"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

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

	// Initialize Layers
	userRepo := user.NewUserRepository(db)
	merchantRepo := merchant.NewMerchantRepository(db)
	productRepo := product.NewProductRepository(db)
	walletRepo := wallet.NewWalletRepository(db)
	refreshTokenRepo := auth.NewRefreshTokenRepository(db)

	authService := auth.NewAuthService(userRepo, refreshTokenRepo, jwtSecret)
	userService := user.NewUserService(userRepo)
	merchantService := merchant.NewMerchantService(merchantRepo, userRepo, walletRepo)
	productService := product.NewProductService(productRepo, merchantRepo)
	walletService := wallet.NewWalletService(walletRepo)

	authHandler := auth.NewAuthHandler(authService)
	userHandler := user.NewUserHandler(userService)
	merchantHandler := merchant.NewMerchantHandler(merchantService)
	productHandler := product.NewProductHandler(productService)
	walletHandler := wallet.NewWalletHandler(walletService)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	// Initialize Fiber app
	app := fiber.New()

	// Setup Routes
	server.SetupRoutes(
		app,
		authHandler,
		userHandler,
		merchantHandler, productHandler, walletHandler, jwtSecret)

	// Start server
	log.Printf("Server starting on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
