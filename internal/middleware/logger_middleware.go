package middleware

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Logger returns a custom middleware for structured logging using slog
func Logger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Handle request
		err := c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get request ID
		requestID := c.GetRespHeader(fiber.HeaderXRequestID)

		// Determine log level and attributes
		status := c.Response().StatusCode()
		method := c.Method()
		path := c.Path()

		attributes := []any{
			slog.Int("status", status),
			slog.String("method", method),
			slog.String("path", path),
			slog.String("ip", c.IP()),
			slog.String("latency", latency.String()),
			slog.String("trace_id", requestID),
		}

		if err != nil {
			attributes = append(attributes, slog.String("error", err.Error()))
			slog.Error("Request failed", attributes...)
		} else if status >= 400 && status < 500 {
			slog.Warn("Request client error", attributes...)
		} else if status >= 500 {
			slog.Error("Request server error", attributes...)
		} else {
			slog.Info("Request completed", attributes...)
		}

		return err
	}
}
