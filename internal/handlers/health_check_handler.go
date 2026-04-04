package handlers

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"
)

type HealthCheckHandler struct {
	db *sql.DB
}

func NewHealthCheckHandler(db *sql.DB) *HealthCheckHandler {
	return &HealthCheckHandler{
		db: db,
	}
}

func (h *HealthCheckHandler) RegisterRoutes(app *fiber.App) {
	app.Get("/health", h.healthCheckHandler(h.db))
}

// HealthCheck verifies the database connection is alive.
func (h *HealthCheckHandler) healthCheckHandler(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if err := db.Ping(); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status":  "unhealthy",
				"message": "database connection lost",
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "healthy",
			"message": "database connection active",
		})
	}
}
