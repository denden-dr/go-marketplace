package handlers

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"
)

// HealthCheckHandler handles health check requests
type HealthCheckHandler struct {
	DB *sql.DB
}

// NewHealthCheckHandler creates a new instance of HealthCheckHandler
func NewHealthCheckHandler(db *sql.DB) *HealthCheckHandler {
	return &HealthCheckHandler{DB: db}
}

// GetStatus checks the database connection and returns the status
func (h *HealthCheckHandler) GetStatus(c *fiber.Ctx) error {
	if err := h.DB.Ping(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "down",
			"message": "database connection failed",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "up",
		"message": "database connection is healthy",
	})
}

func (h *HealthCheckHandler) RegisterRoutes(router fiber.Router) {
	router.Get("/health", h.GetStatus)
}
