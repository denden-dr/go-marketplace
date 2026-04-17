package order

import (
	"go-shop-yourself/internal/common"
	"go-shop-yourself/internal/merchant"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func MerchantMiddleware(repo merchant.MerchantRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("userID").(uuid.UUID)
		if !ok {
			return common.NewResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
		}

		m, err := repo.GetByUserID(c.Context(), userID)
		if err != nil {
			return common.NewResponse(c, http.StatusInternalServerError, "Error retrieving merchant profile", nil)
		}
		if m == nil {
			return common.NewResponse(c, http.StatusForbidden, "Only merchants can access this resource", nil)
		}

		c.Locals("merchant", m)
		return c.Next()
	}
}
