package test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/denden-dr/go-shop-yourself/internal/handlers"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestHealthCheckHandler_GetStatus(t *testing.T) {
	t.Run("successful db ping returns 200 OK", func(t *testing.T) {
		// Create mock db
		db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		assert.NoError(t, err)
		defer db.Close()

		// Expect ping to succeed
		mock.ExpectPing().WillReturnError(nil)

		// Create Fiber app and handler
		app := fiber.New()
		healthHandler := handlers.NewHealthCheckHandler(db)
		app.Get("/api/health", healthHandler.GetStatus)

		// Create request
		req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
		resp, err := app.Test(req)

		// Assertions
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify body
		body, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		json.Unmarshal(body, &result)
		assert.Equal(t, "up", result["status"])
		assert.Equal(t, "database connection is healthy", result["message"])

		// Verify mocks
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("failed db ping returns 500 Internal Server Error", func(t *testing.T) {
		// Create mock db
		db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		assert.NoError(t, err)
		defer db.Close()

		// Expect ping to fail
		mock.ExpectPing().WillReturnError(fmt.Errorf("sql: connection is closed"))

		// Create Fiber app and handler
		app := fiber.New()
		healthHandler := handlers.NewHealthCheckHandler(db)
		app.Get("/api/health", healthHandler.GetStatus)

		// Create request
		req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
		resp, err := app.Test(req)

		// Assertions
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		// Verify body
		body, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		json.Unmarshal(body, &result)
		assert.Equal(t, "down", result["status"])
		assert.Equal(t, "database connection failed", result["message"])

		// Verify mocks
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
