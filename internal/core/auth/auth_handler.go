package auth

import (
	"errors"
	"go-marketplace/internal/common"
	"go-marketplace/internal/domain"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authService AuthServiceInterface
}

func NewAuthHandler(authService AuthServiceInterface) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register handles user registration
// @Summary Register a new user
// @Description Creates a new user account with email, password, and username.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration Info"
// @Success 201 {object} common.ResponseWrapper{data=AuthResponse}
// @Failure 400 {object} common.ResponseWrapper
// @Failure 409 {object} common.ResponseWrapper
// @Failure 500 {object} common.ResponseWrapper
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return common.NewResponse(c, http.StatusBadRequest, "Invalid request body", nil)
	}

	if err := req.Validate(); err != nil {
		return common.NewResponse(c, http.StatusBadRequest, err.Error(), nil)
	}

	res, err := h.authService.Register(c.Context(), req.FullName, req.Email, req.Password, req.Username)
	if err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			return common.NewResponse(c, http.StatusConflict, err.Error(), nil)
		}
		return common.NewResponse(c, http.StatusInternalServerError, "Internal Server Error", nil)
	}

	return common.NewResponse(c, http.StatusCreated, "User registered successfully", res)
}

// Login handles user authentication
// @Summary Login user
// @Description Authenticates a user and returns access and refresh tokens.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login Credentials"
// @Success 200 {object} common.ResponseWrapper{data=AuthResponse}
// @Failure 400 {object} common.ResponseWrapper
// @Failure 409 {object} common.ResponseWrapper
// @Failure 500 {object} common.ResponseWrapper
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return common.NewResponse(c, http.StatusBadRequest, "Invalid request body", nil)
	}

	if err := req.Validate(); err != nil {
		return common.NewResponse(c, http.StatusBadRequest, err.Error(), nil)
	}

	res, err := h.authService.Login(c.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			return common.NewResponse(c, http.StatusUnauthorized, err.Error(), nil)
		}
		if errors.Is(err, domain.ErrAuthProviderMismatch) {
			return common.NewResponse(c, http.StatusConflict, err.Error(), nil)
		}
		return common.NewResponse(c, http.StatusInternalServerError, "Internal Server Error", nil)
	}

	return common.NewResponse(c, http.StatusOK, "Login successful", res)
}

// RefreshTokens handles token refreshing
// @Summary Refresh access tokens
// @Description Rotates refresh tokens and provides a new access token.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshRequest true "Refresh Token"
// @Success 200 {object} common.ResponseWrapper{data=AuthResponse}
// @Failure 400 {object} common.ResponseWrapper
// @Failure 401 {object} common.ResponseWrapper
// @Failure 500 {object} common.ResponseWrapper
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshTokens(c *fiber.Ctx) error {
	var req RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return common.NewResponse(c, http.StatusBadRequest, "Invalid request body", nil)
	}

	if err := req.Validate(); err != nil {
		return common.NewResponse(c, http.StatusBadRequest, err.Error(), nil)
	}

	res, err := h.authService.RefreshTokens(c.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidRefreshToken) || errors.Is(err, domain.ErrRefreshTokenExpired) || errors.Is(err, domain.ErrRefreshTokenReused) {
			return common.NewResponse(c, http.StatusUnauthorized, err.Error(), nil)
		}
		return common.NewResponse(c, http.StatusInternalServerError, "Internal Server Error", nil)
	}

	return common.NewResponse(c, http.StatusOK, "Token refreshed successfully", res)
}

// Logout handles user logout
// @Summary Logout user
// @Description Invalidates the current refresh token.
// @Tags auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body LogoutRequest true "Refresh Token to invalidate"
// @Success 200 {object} common.ResponseWrapper
// @Failure 400 {object} common.ResponseWrapper
// @Failure 401 {object} common.ResponseWrapper
// @Failure 500 {object} common.ResponseWrapper
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	var req LogoutRequest
	if err := c.BodyParser(&req); err != nil {
		return common.NewResponse(c, http.StatusBadRequest, "Invalid request body", nil)
	}

	if err := req.Validate(); err != nil {
		return common.NewResponse(c, http.StatusBadRequest, err.Error(), nil)
	}

	err := h.authService.Logout(c.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidRefreshToken) {
			return common.NewResponse(c, http.StatusUnauthorized, err.Error(), nil)
		}
		return common.NewResponse(c, http.StatusInternalServerError, "Internal Server Error", nil)
	}

	return common.NewResponse(c, http.StatusOK, "Logout successful", nil)
}

// SocialLogin handles social sign-in via Supabase
// @Summary Social login
// @Description Authenticates a user using a Supabase access token and returns access and refresh tokens.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body SocialLoginRequest true "Supabase Access Token"
// @Success 200 {object} common.ResponseWrapper{data=AuthResponse}
// @Failure 400 {object} common.ResponseWrapper
// @Failure 401 {object} common.ResponseWrapper
// @Failure 409 {object} common.ResponseWrapper
// @Failure 500 {object} common.ResponseWrapper
// @Router /auth/social [post]
func (h *AuthHandler) SocialLogin(c *fiber.Ctx) error {
	var req SocialLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return common.NewResponse(c, http.StatusBadRequest, "Invalid request body", nil)
	}

	if err := req.Validate(); err != nil {
		return common.NewResponse(c, http.StatusBadRequest, err.Error(), nil)
	}

	res, err := h.authService.SocialLogin(c.Context(), req.AccessToken)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidSocialToken) || errors.Is(err, domain.ErrEmailNotVerified) {
			return common.NewResponse(c, http.StatusUnauthorized, err.Error(), nil)
		}
		if errors.Is(err, domain.ErrEmailAlreadyUsedByOtherMethod) || errors.Is(err, domain.ErrAuthProviderMismatch) {
			return common.NewResponse(c, http.StatusConflict, err.Error(), nil)
		}
		if errors.Is(err, domain.ErrEmailPasswordSignInNotAllowed) {
			return common.NewResponse(c, http.StatusForbidden, err.Error(), nil)
		}
		if errors.Is(err, domain.ErrSocialLoginNotAvailable) {
			return common.NewResponse(c, http.StatusServiceUnavailable, err.Error(), nil)
		}
		return common.NewResponse(c, http.StatusInternalServerError, "Internal Server Error", nil)
	}

	return common.NewResponse(c, http.StatusOK, "Social login successful", res)
}
