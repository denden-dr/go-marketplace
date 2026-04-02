package handlers

import (
	"github.com/denden-dr/go-shop-yourself/internal/services"
	"github.com/gofiber/fiber/v2"
)

func HelloWorldHandler(c *fiber.Ctx) error {
	message := services.HelloWorldService()
	return c.Status(200).JSON(fiber.Map{
		"message": message,
	})
}
