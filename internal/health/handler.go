package health

import (
	"context"
	"go-shop-yourself/internal/common"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opensearch-project/opensearch-go/v3/opensearchapi"
)

type HealthHandler struct {
	db *pgxpool.Pool
	os *opensearchapi.Client
}

func NewHealthHandler(db *pgxpool.Pool, os *opensearchapi.Client) *HealthHandler {
	return &HealthHandler{
		db: db,
		os: os,
	}
}

// CheckStatus handles the health check request.
// It checks both the database and OpenSearch connectivity.
func (h *HealthHandler) CheckStatus(c *fiber.Ctx) error {
	dbStatus := "up"
	osStatus := "up"
	message := "application is healthy"
	statusCode := fiber.StatusOK

	// Check Database
	if err := h.db.Ping(context.Background()); err != nil {
		dbStatus = "down"
		message = "application is unhealthy"
		statusCode = fiber.StatusServiceUnavailable
	}

	// Check OpenSearch
	if _, err := h.os.Info(context.Background(), nil); err != nil {
		osStatus = "down"
		message = "application is unhealthy"
		statusCode = fiber.StatusServiceUnavailable
	}

	return common.NewResponse(c, statusCode, message, fiber.Map{
		"components": fiber.Map{
			"database":   dbStatus,
			"opensearch": osStatus,
		},
	})
}
