package main

import (
	"github.com/denden-dr/go-shop-yourself/internal/handlers"
	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()
	handlers.SetupHandler(app)

	app.Listen(":3000")
}
