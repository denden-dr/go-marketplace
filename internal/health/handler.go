package health

import (
	"context"
	"go-shop-yourself/internal/common"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opensearch-project/opensearch-go/v3/opensearchapi"

	firebase "firebase.google.com/go/v4"
)

type HealthHandler struct {
	db *pgxpool.Pool
	os *opensearchapi.Client
	fb *firebase.App
}

func NewHealthHandler(db *pgxpool.Pool, os *opensearchapi.Client, fb *firebase.App) *HealthHandler {
	return &HealthHandler{
		db: db,
		os: os,
		fb: fb,
	}
}

// CheckStatus handles the health check request.
// @Summary Check application health
// @Description Checks connectivity to the database, OpenSearch, and Firebase.
// @Tags Health
// @Produce json
// @Success 200 {object} common.ResponseWrapper "Application is healthy"
// @Failure 503 {object} common.ResponseWrapper "Application is unhealthy"
// @Router /health [get]
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
	if h.os != nil {
		if _, err := h.os.Info(context.Background(), nil); err != nil {
			osStatus = "down"
			message = "application is unhealthy"
			statusCode = fiber.StatusServiceUnavailable
		}
	} else {
		osStatus = "down"
		message = "application is unhealthy"
		statusCode = fiber.StatusServiceUnavailable
	}

	// Check Firebase
	firebaseStatus := "up"
	if h.fb != nil {
		// Attempt to initialize Auth client as a health check
		if _, err := h.fb.Auth(context.Background()); err != nil {
			firebaseStatus = "down"
			message = "application is unhealthy"
			statusCode = fiber.StatusServiceUnavailable
		}
	} else {
		firebaseStatus = "unconfigured"
		// Do NOT set 503 or "unhealthy" if firebase is just not configured
	}

	return common.NewResponse(c, statusCode, message, fiber.Map{
		"components": fiber.Map{
			"database":   dbStatus,
			"opensearch": osStatus,
			"firebase":   firebaseStatus,
		},
	})
}
