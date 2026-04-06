package main

import (
	"log"
	"os"

	"github.com/denden-dr/go-shop-yourself/internal/database"
	"github.com/denden-dr/go-shop-yourself/internal/handlers"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize database
	db, err := database.ConnectDB()
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}
	defer db.Close()

	// Initialize Fiber app
	app := fiber.New()
	router := app.Group("/api")

	// Setup health check handler
	healthHandler := handlers.NewHealthCheckHandler(db)
	healthHandler.RegisterRoutes(router)

	// Get port from .env or default to 3000
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	// Start server
	log.Printf("Server starting on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
