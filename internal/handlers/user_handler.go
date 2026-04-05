package handlers

import (
	"github.com/denden-dr/go-shop-yourself/internal/dto"
	"github.com/denden-dr/go-shop-yourself/internal/middleware"
	"github.com/denden-dr/go-shop-yourself/internal/services"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// UserHandler handles user-related HTTP requests.
type UserHandler interface {
	getProfile(c *fiber.Ctx) error
	RegisterRoutes(app fiber.Router)
}

type userHandler struct {
	service   services.UserService
	jwtSecret string
}

// NewUserHandler creates a new instance of UserHandler.
func NewUserHandler(service services.UserService, jwtSecret string) UserHandler {
	return &userHandler{service: service, jwtSecret: jwtSecret}
}

func (h *userHandler) getProfile(c *fiber.Ctx) error {
	idParam := c.Params("id")
	userID, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewErrorResponse("Invalid user ID format", fiber.StatusBadRequest))
	}

	// Optional: Check if the authenticated user is requesting their own profile
	// authUserID := c.Locals("userID").(string)
	// if authUserID != userID.String() {
	// 	return c.Status(fiber.StatusForbidden).JSON(dto.NewErrorResponse("Forbidden", fiber.StatusForbidden))
	// }

	profile, err := h.service.GetProfile(c.Context(), userID)
	if err != nil {
		if err.Error() == "user profile not found" {
			return c.Status(fiber.StatusNotFound).JSON(dto.NewErrorResponse("User not found", fiber.StatusNotFound))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewErrorResponse("Failed to retrieve profile", fiber.StatusInternalServerError))
	}

	return c.Status(fiber.StatusOK).JSON(dto.NewSuccessResponse("User profile retrieved successfully", fiber.StatusOK, profile))
}

func (h *userHandler) RegisterRoutes(app fiber.Router) {

	user := app.Group("/user").Use(middleware.AuthRequired(h.jwtSecret))
	user.Get("/:id", h.getProfile)
}
