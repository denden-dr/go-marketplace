package main

import (
	"context"
	"log"
	"os"
	"time"

	"go-shop-yourself/internal/auth"
	"go-shop-yourself/internal/cart"
	"go-shop-yourself/internal/database"
	"go-shop-yourself/internal/health"
	"go-shop-yourself/internal/merchant"
	"go-shop-yourself/internal/opensearch"
	"go-shop-yourself/internal/order"
	"go-shop-yourself/internal/product"
	"go-shop-yourself/internal/server"
	"go-shop-yourself/internal/user"
	"go-shop-yourself/internal/wallet"

	firebase "firebase.google.com/go/v4"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
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

	appEnv := os.Getenv("APP_ENV")

	// Initialize Firebase
	var firebaseAuthClient auth.FirebaseAuthClient
	var fbApp *firebase.App
	var fbConfig *firebase.Config
	var opts []option.ClientOption

	if appEnv == "development" {
		emulatorHost := os.Getenv("FIREBASE_AUTH_EMULATOR_HOST")
		log.Printf("Firebase Auth using emulator at %s", emulatorHost)

		projectID := os.Getenv("FIREBASE_PROJECT_ID")
		fbConfig = &firebase.Config{ProjectID: projectID}

		// In development with emulator, we skip real credentials to avoid strict signature verification
		// of "alg: none" tokens which are common in emulator usage.
		if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") != "" {
			log.Println("Development mode: Unsetting GOOGLE_APPLICATION_CREDENTIALS for Firebase Emulator compatibility")
			os.Unsetenv("GOOGLE_APPLICATION_CREDENTIALS")
		}
		opts = append(opts, option.WithoutAuthentication())
	}

	fbApp, err = firebase.NewApp(context.Background(), fbConfig, opts...)
	if err != nil {
		log.Printf("Warning: Error initializing firebase app: %v. Social login will be disabled.", err)
	} else {
		firebaseAuthClient, err = auth.NewFirebaseAuthClient(fbApp)
		if err != nil {
			log.Printf("Warning: Error initializing firebase auth client: %v. Social login will be disabled.", err)
		}
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

	authService := auth.NewAuthService(userRepo, refreshTokenRepo, firebaseAuthClient, jwtSecret)
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
	healthHandler := health.NewHealthHandler(db, osClient, fbApp)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	// Initialize Fiber app
	app := fiber.New()

	// Setup Routes
	firebaseEnabled := firebaseAuthClient != nil
	server.SetupRoutes(
		app,
		authHandler,
		userHandler,
		merchantHandler, productHandler, walletHandler,
		cartHandler, orderHandler, healthHandler, jwtSecret, appEnv, firebaseEnabled)

	// Start server
	log.Printf("Server starting on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
