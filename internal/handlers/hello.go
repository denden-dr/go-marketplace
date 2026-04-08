package handlers

import "github.com/gofiber/fiber/v2"

// HelloHandler returns a simple "Hello World" response.
func HelloHandler(c *fiber.Ctx) error {
	return c.SendString("Hello World")
}
