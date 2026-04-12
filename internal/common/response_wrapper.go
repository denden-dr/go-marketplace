package common

import "github.com/gofiber/fiber/v2"

type ResponseWrapper struct {
	Message string      `json:"message"`
	Status  int         `json:"status"`
	Data    interface{} `json:"data"`
}

func NewResponse(c *fiber.Ctx, status int, message string, data interface{}) error {
	return c.Status(status).JSON(ResponseWrapper{
		Message: message,
		Status:  status,
		Data:    data,
	})
}
