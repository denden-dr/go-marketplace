package handlers

import (
	"errors"
	"go-shop-yourself/internal/domain"
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

	if err := req.Validate(); err != nil {
		return dtos.NewResponse(c, http.StatusBadRequest, err.Error(), nil)
	}

	userId, err := h.authService.Register(c.Context(), req.Email, req.Password, req.Username)
	if err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			return dtos.NewResponse(c, http.StatusConflict, err.Error(), nil)
		}
		return dtos.NewResponse(c, http.StatusInternalServerError, err.Error(), nil)
	}

	return dtos.NewResponse(c, http.StatusCreated, "User registered successfully", dtos.AuthResponse{ID: userId})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req dtos.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return dtos.NewResponse(c, http.StatusBadRequest, "Invalid request body", nil)
	}

	if err := req.Validate(); err != nil {
		return dtos.NewResponse(c, http.StatusBadRequest, err.Error(), nil)
	}

	res, err := h.authService.Login(c.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			return dtos.NewResponse(c, http.StatusUnauthorized, err.Error(), nil)
		}
		return dtos.NewResponse(c, http.StatusInternalServerError, err.Error(), nil)
	}

	return dtos.NewResponse(c, http.StatusOK, "Login successful", res)
}

func (h *AuthHandler) RefreshTokens(c *fiber.Ctx) error {
	var req dtos.RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return dtos.NewResponse(c, http.StatusBadRequest, "Invalid request body", nil)
	}

	if err := req.Validate(); err != nil {
		return dtos.NewResponse(c, http.StatusBadRequest, err.Error(), nil)
	}

	res, err := h.authService.RefreshTokens(c.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidRefreshToken) || errors.Is(err, domain.ErrRefreshTokenExpired) || errors.Is(err, domain.ErrRefreshTokenReused) {
			return dtos.NewResponse(c, http.StatusUnauthorized, err.Error(), nil)
		}
		return dtos.NewResponse(c, http.StatusInternalServerError, err.Error(), nil)
	}

	return dtos.NewResponse(c, http.StatusOK, "Token refreshed successfully", res)
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	var req dtos.LogoutRequest
	if err := c.BodyParser(&req); err != nil {
		return dtos.NewResponse(c, http.StatusBadRequest, "Invalid request body", nil)
	}

	if err := req.Validate(); err != nil {
		return dtos.NewResponse(c, http.StatusBadRequest, err.Error(), nil)
	}

	err := h.authService.Logout(c.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidRefreshToken) {
			return dtos.NewResponse(c, http.StatusUnauthorized, err.Error(), nil)
		}
		return dtos.NewResponse(c, http.StatusInternalServerError, err.Error(), nil)
	}

	return dtos.NewResponse(c, http.StatusOK, "Logout successful", nil)
}
