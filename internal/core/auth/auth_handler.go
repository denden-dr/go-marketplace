package auth

import (
	"go-marketplace/internal/common"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"time"
)

type AuthHandler struct {
	authService      AuthService
	googleClient     GoogleClient
	googleSuccessURL string
}

func NewAuthHandler(authService AuthService, googleClient GoogleClient, googleSuccessURL string) *AuthHandler {
	return &AuthHandler{
		authService:      authService,
		googleClient:     googleClient,
		googleSuccessURL: googleSuccessURL,
	}
}

// Register handles user registration
// @Summary Register a new user
// @Description Creates a new user account with email, password, and username.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration Info"
// @Success 201 {object} common.SuccessResponse{data=AuthResponse}
// @Failure 400 {object} common.ProblemDetails
// @Failure 409 {object} common.ProblemDetails
// @Failure 500 {object} common.ProblemDetails
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(http.StatusBadRequest, "Invalid request body")
	}

	if err := req.Validate(); err != nil {
		return err
	}

	res, err := h.authService.Register(c.Context(), req.FullName, req.Email, req.Password, req.Username)
	if err != nil {
		return err
	}

	return common.NewSuccessResponse(c, http.StatusCreated, "User registered successfully", res)
}

// Login handles user authentication
// @Summary Login user
// @Description Authenticates a user and returns access and refresh tokens.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login Credentials"
// @Success 200 {object} common.SuccessResponse{data=AuthResponse}
// @Failure 400 {object} common.ProblemDetails
// @Failure 409 {object} common.ProblemDetails
// @Failure 500 {object} common.ProblemDetails
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(http.StatusBadRequest, "Invalid request body")
	}

	if err := req.Validate(); err != nil {
		return err
	}

	res, err := h.authService.Login(c.Context(), req.Email, req.Password, c.IP(), c.Get("User-Agent"))
	if err != nil {
		return err
	}

	return h.handleAuthSuccess(c, res, "Login successful")
}

func (h *AuthHandler) VerifyEmail(c *fiber.Ctx) error {
	var req VerifyRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(http.StatusBadRequest, "Invalid request body")
	}

	if err := req.Validate(); err != nil {
		return err
	}

	err := h.authService.VerifyEmail(c.Context(), req.UserID, req.Code)
	if err != nil {
		return err
	}

	return common.NewSuccessResponse(c, http.StatusOK, "Email verified successfully", nil)
}

func (h *AuthHandler) RefreshTokens(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		return fiber.NewError(http.StatusUnauthorized, "Refresh token missing")
	}

	res, err := h.authService.RefreshTokens(c.Context(), refreshToken, c.IP(), c.Get("User-Agent"))
	if err != nil {
		h.clearTokensCookies(c)
		return err
	}

	return h.handleAuthSuccess(c, res, "Token refreshed successfully")
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken != "" {
		_ = h.authService.Logout(c.Context(), refreshToken)
	}

	h.clearTokensCookies(c)
	return common.NewSuccessResponse(c, http.StatusOK, "Logout successful", nil)
}

func (h *AuthHandler) handleAuthSuccess(c *fiber.Ctx, res *AuthResponse, message string) error {
	h.setTokensCookies(c, res.AccessToken, res.RefreshToken)
	return common.NewSuccessResponse(c, http.StatusOK, message, res)
}

func (h *AuthHandler) setTokensCookies(c *fiber.Ctx, accessToken, refreshToken string) {
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Expires:  time.Now().Add(time.Minute * 15),
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
	})
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Expires:  time.Now().Add(time.Hour * 24 * 7),
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
	})
}

func (h *AuthHandler) clearTokensCookies(c *fiber.Ctx) {
	c.ClearCookie("access_token", "refresh_token")
}

// GoogleLogin redirects to Google's OAuth2 consent page
func (h *AuthHandler) GoogleLogin(c *fiber.Ctx) error {
	state := uuid.New().String()
	c.Cookie(&fiber.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Expires:  time.Now().Add(time.Minute * 10),
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Lax",
	})

	url := h.googleClient.GetAuthURL(state)
	return c.Redirect(url)
}

// GoogleCallback handles the callback from Google OAuth2
func (h *AuthHandler) GoogleCallback(c *fiber.Ctx) error {
	code := c.Query("code")
	state := c.Query("state")
	expectedState := c.Cookies("oauth_state")

	c.ClearCookie("oauth_state")

	res, err := h.authService.HandleGoogleLogin(c.Context(), code, state, expectedState, c.IP(), c.Get("User-Agent"))
	if err != nil {
		return err
	}

	h.setTokensCookies(c, res.AccessToken, res.RefreshToken)
	return c.Redirect(h.googleSuccessURL)
}
