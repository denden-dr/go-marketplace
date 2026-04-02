package handlers

import "github.com/gofiber/fiber/v2"

func SetupHandler(app *fiber.App) error {
	app.Get("/", HelloWorldHandler)
	return nil
}
