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
	database.ConnectDB()

	// Initialize Layers
	userRepo := repos.NewUserRepository(database.DB)
	merchantRepo := repos.NewMerchantRepository(database.DB)
	productRepo := repos.NewProductRepository(database.DB)
	walletRepo := repos.NewWalletRepository(database.DB)

	authService := services.NewAuthService(userRepo)
	userService := services.NewUserService(userRepo)
	merchantService := services.NewMerchantService(merchantRepo, userRepo)
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

	// Auth routes
	auth := app.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/register-merchant", merchantHandler.RegisterMerchant)

	// User routes
	users := app.Group("/users")
	users.Get("/:id", userHandler.GetUserByID)

	// Product routes
	products := app.Group("/products")
	products.Post("/", productHandler.CreateProduct)
	products.Put("/:id", productHandler.UpdateProduct)

	// Wallet routes
	wallets := app.Group("/wallets")
	wallets.Get("/", walletHandler.GetWallet)
	wallets.Get("/history", walletHandler.GetHistory)
	wallets.Post("/withdraw", walletHandler.Withdraw)

	// Start server
	log.Printf("Server starting on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
