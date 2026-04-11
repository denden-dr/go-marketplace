package main

import (
	"log"
	"os"

	"go-shop-yourself/internal/database"
	"go-shop-yourself/internal/handlers"
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
	userRepo := repos.NewUserRepository(db)
	merchantRepo := repos.NewMerchantRepository(db)
	productRepo := repos.NewProductRepository(db)
	walletRepo := repos.NewWalletRepository(db)
	refreshTokenRepo := repos.NewRefreshTokenRepository(db)

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

	// Setup Routes
	handlers.SetupRoutes(
		app,
		authHandler,
		userHandler,
		merchantHandler,
		productHandler,
		walletHandler,
		jwtSecret)

	// Start server
	log.Printf("Server starting on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
