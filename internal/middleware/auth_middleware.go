package middleware

import (
	"log"
	"net/http"
	"strings"

	"go-marketplace/internal/core/auth"
	"go-marketplace/internal/domain"

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
			return fiber.NewError(http.StatusUnauthorized, "Missing or invalid access token")
		}

		claims, err := auth.ValidateAccessToken(token, jwtSecret)
		if err != nil {
			log.Printf("Token validation failed: %v", err)
			return fiber.NewError(http.StatusUnauthorized, "Unauthorized: "+err.Error())
		}

		c.Locals("userID", claims.UserID)
		c.Locals("role", claims.Role)
		return c.Next()
	}
}

func RequireRole(roles ...domain.UserRole) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole, ok := c.Locals("role").(domain.UserRole)
		if !ok {
			return fiber.NewError(http.StatusForbidden, "Forbidden: role not found")
		}

		for _, r := range roles {
			if userRole == r {
				return c.Next()
			}
		}

		return fiber.NewError(http.StatusForbidden, "Forbidden: insufficient permissions")
	}
}
