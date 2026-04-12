package middleware

import (
	"log"
	"net/http"
	"strings"

	"go-shop-yourself/internal/auth"
	"go-shop-yourself/internal/common"

	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return common.NewResponse(c, http.StatusUnauthorized, "Missing authorization header", nil)
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return common.NewResponse(c, http.StatusUnauthorized, "Invalid authorization header format", nil)
		}

		token := parts[1]
		userID, err := auth.ValidateAccessToken(token, jwtSecret)
		if err != nil {
			log.Printf("Token validation failed: %v", err)
			return common.NewResponse(c, http.StatusUnauthorized, "Unauthorized: "+err.Error(), nil)
		}

		c.Locals("userID", userID)
		return c.Next()
	}
}
