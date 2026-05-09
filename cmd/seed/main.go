//go:build seed

package main

import (
	"log"
	"os"

	"go-marketplace/internal/config"
	"go-marketplace/internal/database"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env configuration
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using default values or environment variables")
	}

	cfg := config.Load()

	// Initialize Database
	db, err := database.ConnectDB(cfg.DB)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Read seed SQL file
	seedFile := "internal/database/seed/seed.sql"
	content, err := os.ReadFile(seedFile)
	if err != nil {
		log.Fatalf("Failed to read seed file: %v", err)
	}

	// Execute seed SQL
	log.Printf("Seeding database with %s...", seedFile)
	_, err = db.Exec(string(content))
	if err != nil {
		log.Fatalf("Failed to execute seed SQL: %v", err)
	}

	log.Println("Database seeded successfully!")
}
