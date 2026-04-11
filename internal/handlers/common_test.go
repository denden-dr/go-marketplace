package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func setupTestApp() *fiber.App {
	return fiber.New()
}

func authTestMiddleware(userID uuid.UUID) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Locals("userID", userID)
		return c.Next()
	}
}
