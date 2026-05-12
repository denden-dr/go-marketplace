package testutil

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func SetupTestApp() *fiber.App {
	return fiber.New()
}

func AuthTestMiddleware(userID uuid.UUID) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Locals("userID", userID)
		return c.Next()
	}
}
