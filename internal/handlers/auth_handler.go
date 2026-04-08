package handlers

import (
	"go-shop-yourself/internal/dtos"
	"go-shop-yourself/internal/services"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req dtos.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return dtos.NewResponse(c, http.StatusBadRequest, "Invalid request body", nil)
	}

	userId, err := h.authService.Register(c.Context(), req.Email, req.Password, req.Username)
	if err != nil {
		return dtos.NewResponse(c, http.StatusInternalServerError, err.Error(), nil)
	}

	return dtos.NewResponse(c, http.StatusCreated, "User registered successfully", dtos.AuthResponse{ID: userId})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req dtos.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return dtos.NewResponse(c, http.StatusBadRequest, "Invalid request body", nil)
	}

	userId, err := h.authService.Login(c.Context(), req.Email, req.Password)
	if err != nil {
		return dtos.NewResponse(c, http.StatusUnauthorized, err.Error(), nil)
	}

	return dtos.NewResponse(c, http.StatusOK, "Login successful", dtos.AuthResponse{ID: userId})
}
