package handlers

import (
	"log"

	"github.com/denden-dr/go-shop-yourself/internal/dto"
	"github.com/denden-dr/go-shop-yourself/internal/services"
	"github.com/gofiber/fiber/v2"
)

type AuthHandler interface {
	RegisterRoutes(app fiber.Router)
}

// AuthHandler handles requests for authentication.
type authHandler struct {
	authService services.AuthService
}

// NewAuthHandler creates a new instance of AuthHandler.
func NewAuthHandler(authService services.AuthService) AuthHandler {
	return &authHandler{authService: authService}
}

// RegisterRoutes registers the authentication routes.
func (h *authHandler) RegisterRoutes(app fiber.Router) {
	auth := app.Group("/auth")
	auth.Post("/register", h.register)
	auth.Post("/login", h.login)
}

// Register handles user registration request.
func (h *authHandler) register(c *fiber.Ctx) error {
	var req dto.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewErrorResponse("Invalid request body", fiber.StatusBadRequest))
	}

	res, err := h.authService.Register(c.Context(), req)
	if err != nil {
		log.Printf("Registration error: %v", err)
		if err == services.ErrUserAlreadyExists {
			return c.Status(fiber.StatusConflict).JSON(dto.NewErrorResponse(err.Error(), fiber.StatusConflict))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewErrorResponse("Something went wrong during registration", fiber.StatusInternalServerError))
	}

	return c.Status(fiber.StatusCreated).JSON(dto.NewSuccessResponse("User registered successfully", fiber.StatusCreated, res))
}

// Login handles user login request.
func (h *authHandler) login(c *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewErrorResponse("Invalid request body", fiber.StatusBadRequest))
	}

	res, err := h.authService.Login(c.Context(), req)
	if err != nil {
		log.Printf("Login error: %v", err)
		if err == services.ErrInvalidCredentials {
			return c.Status(fiber.StatusUnauthorized).JSON(dto.NewErrorResponse(err.Error(), fiber.StatusUnauthorized))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewErrorResponse("Something went wrong during login", fiber.StatusInternalServerError))
	}

	return c.Status(fiber.StatusOK).JSON(dto.NewSuccessResponse("Login successful", fiber.StatusOK, res))
}
