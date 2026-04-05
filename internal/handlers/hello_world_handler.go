package handlers

import (
	"github.com/denden-dr/go-shop-yourself/internal/dto"
	"github.com/denden-dr/go-shop-yourself/internal/services"
	"github.com/gofiber/fiber/v2"
)

func HelloWorldHandler(c *fiber.Ctx) error {
	message := services.HelloWorldService()
	return c.Status(fiber.StatusOK).JSON(dto.NewSuccessResponse("Success", fiber.StatusOK, fiber.Map{
		"message": message,
	}))
}
