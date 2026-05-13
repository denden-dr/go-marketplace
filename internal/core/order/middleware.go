package order

import (
	"go-marketplace/internal/core/merchant"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func MerchantMiddleware(repo merchant.MerchantRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("userID").(uuid.UUID)
		if !ok {
			return fiber.ErrUnauthorized
		}

		m, err := repo.GetByUserID(c.Context(), userID)
		if err != nil {
			return err
		}
		if m == nil {
			return fiber.NewError(http.StatusForbidden, "Only merchants can access this resource")
		}

		c.Locals("merchant", m)
		return c.Next()
	}
}
