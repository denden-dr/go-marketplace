package auth

import (
	"errors"
	"go-marketplace/internal/common"
	"go-marketplace/internal/domain"
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

	res, err := h.authService.Login(c.Context(), req.Email, req.Password, c.IP(), c.Get("User-Agent"))
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			return common.NewResponse(c, http.StatusUnauthorized, err.Error(), nil)
		}
		if errors.Is(err, domain.ErrEmailNotVerified) {
			return common.NewResponse(c, http.StatusForbidden, err.Error(), nil)
		}
		return common.NewResponse(c, http.StatusInternalServerError, "Internal Server Error", nil)
	}

	// In a real app, these would come from the service response
	// But since I updated AuthResponse to NOT have them, I should probably
	// get them from the service response which I DIDNT update yet in the handler logic.
	// Wait, I updated AuthService.Login to return *AuthResponse.
	// But AuthService.Login returns a struct that I just changed to NOT have tokens.
	// That's a problem. The Service SHOULD return the tokens, but the Handler should
	// decide HOW to return them (JSON vs Cookie).

	// I'll update AuthResponse in DTO to still have them but with `json:"-"`.
	// No, I'll keep them in a separate internal struct or just keep them in AuthResponse but with `json:"-"`.

	return h.handleAuthSuccess(c, res, "Login successful")
}

func (h *AuthHandler) VerifyEmail(c *fiber.Ctx) error {
	var req VerifyRequest
	if err := c.BodyParser(&req); err != nil {
		return common.NewResponse(c, http.StatusBadRequest, "Invalid request body", nil)
	}

	if err := req.Validate(); err != nil {
		return common.NewResponse(c, http.StatusBadRequest, err.Error(), nil)
	}

	err := h.authService.VerifyEmail(c.Context(), req.UserID, req.Code)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidVerificationCode) || errors.Is(err, domain.ErrVerificationCodeExpired) {
			return common.NewResponse(c, http.StatusUnauthorized, err.Error(), nil)
		}
		return common.NewResponse(c, http.StatusInternalServerError, "Internal Server Error", nil)
	}

	return common.NewResponse(c, http.StatusOK, "Email verified successfully", nil)
}

func (h *AuthHandler) RefreshTokens(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		return common.NewResponse(c, http.StatusUnauthorized, "Refresh token missing", nil)
	}

	res, err := h.authService.RefreshTokens(c.Context(), refreshToken, c.IP(), c.Get("User-Agent"))
	if err != nil {
		if errors.Is(err, domain.ErrInvalidRefreshToken) || errors.Is(err, domain.ErrRefreshTokenExpired) || errors.Is(err, domain.ErrRefreshTokenReused) {
			h.clearTokensCookies(c)
			return common.NewResponse(c, http.StatusUnauthorized, err.Error(), nil)
		}
		return common.NewResponse(c, http.StatusInternalServerError, "Internal Server Error", nil)
	}

	return h.handleAuthSuccess(c, res, "Token refreshed successfully")
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken != "" {
		_ = h.authService.Logout(c.Context(), refreshToken)
	}

	h.clearTokensCookies(c)
	return common.NewResponse(c, http.StatusOK, "Logout successful", nil)
}

func (h *AuthHandler) handleAuthSuccess(c *fiber.Ctx, res *AuthResponse, message string) error {
	// Access tokens and refresh tokens are in the response from service
	// but hidden from JSON. We need to access them here.
	// Since I updated AuthResponse to NOT have them, I need to fix that first.
	// I'll add them back with `json:"-"`.

	h.setTokensCookies(c, res.AccessToken, res.RefreshToken)

	return common.NewResponse(c, http.StatusOK, message, res)
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
// @Summary Google Login
// @Description Redirects the user to Google's OAuth2 consent page.
// @Tags auth
// @Success 302
// @Router /auth/google/login [get]
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
// @Summary Google Callback
// @Description Handles the callback from Google, exchanges the code for tokens, and creates a session.
// @Tags auth
// @Param code query string true "OAuth2 Code"
// @Param state query string true "OAuth2 State"
// @Success 302
// @Failure 401 {object} common.ResponseWrapper
// @Router /auth/google/callback [get]
func (h *AuthHandler) GoogleCallback(c *fiber.Ctx) error {
	code := c.Query("code")
	state := c.Query("state")
	expectedState := c.Cookies("oauth_state")

	c.ClearCookie("oauth_state")

	res, err := h.authService.HandleGoogleLogin(c.Context(), code, state, expectedState, c.IP(), c.Get("User-Agent"))
	if err != nil {
		if errors.Is(err, domain.ErrInvalidOAuthState) || errors.Is(err, domain.ErrInvalidSocialToken) {
			return common.NewResponse(c, http.StatusUnauthorized, err.Error(), nil)
		}
		return common.NewResponse(c, http.StatusInternalServerError, "Internal Server Error", nil)
	}

	h.setTokensCookies(c, res.AccessToken, res.RefreshToken)

	return c.Redirect(h.googleSuccessURL)
}
