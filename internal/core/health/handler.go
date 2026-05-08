package health

import (
	"go-marketplace/internal/common"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"

	"os"
)

type HealthHandler struct {
	db *sqlx.DB
}

func NewHealthHandler(db *sqlx.DB) *HealthHandler {
	return &HealthHandler{
		db: db,
	}
}

// CheckStatus handles the health check request.
// @Summary Check application health
// @Description Checks connectivity to the database, OpenSearch, and Supabase configuration.
// @Tags Health
// @Produce json
// @Success 200 {object} common.ResponseWrapper "Application is healthy"
// @Failure 503 {object} common.ResponseWrapper "Application is unhealthy"
// @Router /health [get]
func (h *HealthHandler) CheckStatus(c *fiber.Ctx) error {
	dbStatus := "up"
	message := "application is healthy"
	statusCode := fiber.StatusOK

	// Check Database
	if err := h.db.Ping(); err != nil {
		dbStatus = "down"
		message = "application is unhealthy"
		statusCode = fiber.StatusServiceUnavailable
	}

	supabaseStatus := "configured"
	if os.Getenv("SUPABASE_JWT_SECRET") == "" {
		supabaseStatus = "unconfigured"
	}

	return common.NewResponse(c, statusCode, message, fiber.Map{
		"components": fiber.Map{
			"database": dbStatus,
			"supabase": supabaseStatus,
		},
	})
}
