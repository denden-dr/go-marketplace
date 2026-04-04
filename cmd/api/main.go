package main

import (
	"log"

	"github.com/denden-dr/go-shop-yourself/internal/database"
	"github.com/denden-dr/go-shop-yourself/internal/handlers"
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

	// Initialize Handlers
	healthCheckHandler := handlers.NewHealthCheckHandler(db)

	// Initialize Fiber
	app := fiber.New()

	// Register Routes
	app.Get("/", handlers.HelloWorldHandler)
	healthCheckHandler.RegisterRoutes(app)

	log.Fatal(app.Listen(":3000"))
}
