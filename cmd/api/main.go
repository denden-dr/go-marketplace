package main

import (
	"log"
	"os"

	"github.com/denden-dr/go-shop-yourself/internal/database"
	"github.com/denden-dr/go-shop-yourself/internal/handlers"
	"github.com/denden-dr/go-shop-yourself/internal/repo"
	"github.com/denden-dr/go-shop-yourself/internal/services"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	// Environment variables loading
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Failed to load environment variables: %v", err)
	}

	// Initialize database
	cfg := database.NewConfig()
	db, err := database.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize Repositories
	userRepo := repo.NewUserRepository(db)

	// Initialize Services
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is not set")
	}
	authService := services.NewAuthService(userRepo, jwtSecret)
	userService := services.NewUserService(userRepo)

	// Initialize Handlers
	healthCheckHandler := handlers.NewHealthCheckHandler(db)
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService, jwtSecret)

	// Initialize Fiber
	app := fiber.New()

	// Register Routes
	app.Get("/", handlers.HelloWorldHandler)
	healthCheckHandler.RegisterRoutes(app)
	authHandler.RegisterRoutes(app)
	userHandler.RegisterRoutes(app)

	log.Fatal(app.Listen(":3000"))
}
