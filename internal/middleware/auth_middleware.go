package middleware

import (
	"log"
	"net/http"
	"strings"

	"go-marketplace/internal/common"
	"go-marketplace/internal/core/auth"

	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Cookies("access_token")
		if token == "" {
			// Fallback to Header for flexibility (optional, but let's stick to cookies as primary)
			authHeader := c.Get("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == "Bearer" {
					token = parts[1]
				}
			}
		}

		if token == "" {
			return common.NewResponse(c, http.StatusUnauthorized, "Missing or invalid access token", nil)
		}

		userID, err := auth.ValidateAccessToken(token, jwtSecret)
		if err != nil {
			log.Printf("Token validation failed: %v", err)
			return common.NewResponse(c, http.StatusUnauthorized, "Unauthorized: "+err.Error(), nil)
		}

		c.Locals("userID", userID)
		return c.Next()
	}
}
