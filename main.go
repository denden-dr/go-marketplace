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
	authService := services.NewAuthService(userRepo)
	userService := services.NewUserService(userRepo)

	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)

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

	// User routes
	users := app.Group("/users")
	users.Get("/:id", userHandler.GetUserByID)

	// Start server
	log.Printf("Server starting on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
