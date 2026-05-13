package health

import (
	"go-marketplace/internal/common"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
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
// @Description Checks connectivity to the database and Supabase configuration.
// @Tags Health
// @Produce json
// @Success 200 {object} common.SuccessResponse "Application is healthy"
// @Failure 503 {object} common.ProblemDetails "Application is unhealthy"
// @Router /health [get]
func (h *HealthHandler) CheckStatus(c *fiber.Ctx) error {
	if err := h.db.Ping(); err != nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "Database is unreachable")
	}

	mailersendStatus := "configured"
	if os.Getenv("MAILERSEND_API_KEY") == "" {
		mailersendStatus = "unconfigured"
	}

	return common.NewSuccessResponse(c, fiber.StatusOK, "Application is healthy", fiber.Map{
		"components": fiber.Map{
			"database":   "up",
			"mailersend": mailersendStatus,
		},
	})
}
