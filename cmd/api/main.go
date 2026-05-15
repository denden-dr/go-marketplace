package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"time"

	"go-marketplace/internal/core/auth"
	"go-marketplace/internal/core/cart"
	"go-marketplace/internal/core/health"
	"go-marketplace/internal/core/merchant"
	"go-marketplace/internal/core/order"
	"go-marketplace/internal/core/payment"
	"go-marketplace/internal/core/product"
	"go-marketplace/internal/core/user"
	"go-marketplace/internal/core/wallet"
	"go-marketplace/internal/database"

	"go-marketplace/internal/common"
	"go-marketplace/internal/config"
	"go-marketplace/internal/middleware"
	"go-marketplace/internal/server"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
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

	cfg := config.Load()

	// Initialize Structured Logging
	var handler slog.Handler
	if cfg.AppEnv == "production" {
		handler = slog.NewJSONHandler(os.Stdout, nil)
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	}
	slog.SetDefault(slog.New(handler))

	// Initialize Database
	db, err := database.ConnectDB(cfg.DB)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if cfg.JWTSecret == "" {
		log.Fatalf("JWT_SECRET environment variable is not set")
	}

	mailersendAPIKey := os.Getenv("MAILERSEND_API_KEY")
	mailersendFromEmail := os.Getenv("MAILERSEND_FROM_EMAIL")
	if mailersendAPIKey == "" || mailersendFromEmail == "" {
		log.Fatalf("MAILERSEND_API_KEY and MAILERSEND_FROM_EMAIL must be set")
	}

	if cfg.Google.ClientID == "" || cfg.Google.ClientSecret == "" {
		log.Fatalf("GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET must be set")
	}

	// Initialize Layers
	userRepo := user.NewUserRepository(db)
	merchantRepo := merchant.NewMerchantRepository(db)
	productRepo := product.NewProductRepository(db)
	walletRepo := wallet.NewWalletRepository(db)
	sessionRepo := auth.NewSessionRepository(db)
	verificationRepo := auth.NewVerificationRepository(db)
	cartRepo := cart.NewCartRepository(db)
	orderRepo := order.NewOrderRepository(db)
	paymentRepo := payment.NewPaymentRepository(db)

	mailService := auth.NewMailService(mailersendAPIKey, mailersendFromEmail)
	googleClient := auth.NewGoogleClient(cfg.Google.ClientID, cfg.Google.ClientSecret, cfg.Google.RedirectURL)

	// Background cleanup for expired sessions
	log.Printf("Running initial background cleanup for expired sessions...")
	if err := sessionRepo.DeleteExpiredSessions(context.Background()); err != nil {
		log.Printf("Error during initial background cleanup: %v", err)
	}

	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			if err := sessionRepo.DeleteExpiredSessions(context.Background()); err != nil {
				log.Printf("Error cleaning up expired sessions: %v", err)
			}
		}
	}()

	authService := auth.NewAuthService(userRepo, sessionRepo, verificationRepo, mailService, googleClient, cfg.JWTSecret)
	userService := user.NewUserService(userRepo)
	merchantService := merchant.NewMerchantService(merchantRepo, userRepo, walletRepo)
	productService := product.NewProductService(productRepo, merchantRepo)
	walletService := wallet.NewWalletService(walletRepo)
	cartService := cart.NewCartService(cartRepo, productRepo)

	// Payment & Order setup — orderManager breaks the circular dependency
	// by depending only on repositories, not on PaymentService.
	mockPaymentProvider := payment.NewMockProvider()
	orderManager := order.NewOrderManager(orderRepo, productRepo)
	paymentService := payment.NewPaymentService(paymentRepo, walletService, mockPaymentProvider, orderManager, db)
	orderService := order.NewOrderService(orderRepo, cartRepo, productRepo, walletService, userRepo, merchantRepo, paymentService)

	authHandler := auth.NewAuthHandler(authService, googleClient, cfg.Google.LoginRedirectURL)
	userHandler := user.NewUserHandler(userService)
	merchantHandler := merchant.NewMerchantHandler(merchantService)
	productHandler := product.NewProductHandler(productService)
	walletHandler := wallet.NewWalletHandler(walletService)
	cartHandler := cart.NewCartHandler(cartService)
	orderHandler := order.NewOrderHandler(orderService, merchantRepo)
	paymentHandler := payment.NewPaymentHandler(paymentService)
	healthHandler := health.NewHealthHandler(db)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		Immutable:    true,
		ErrorHandler: common.ErrorHandler,
	})

	// Global Middlewares
	app.Use(requestid.New())
	app.Use(middleware.Logger())

	// Setup Routes
	server.SetupRoutes(
		app,
		authHandler,
		userHandler,
		merchantHandler, productHandler, walletHandler,
		cartHandler, orderHandler, paymentHandler, healthHandler, cfg.JWTSecret, cfg.AppEnv)

	// Start server
	log.Printf("Server starting on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
