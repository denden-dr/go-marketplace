package common

import (
	"errors"
	"go-marketplace/internal/domain"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

// SuccessResponse represents a consistent success envelope
type SuccessResponse struct {
	Message string      `json:"message"`
	Status  int         `json:"status"`
	Data    interface{} `json:"data,omitempty"`
}

// ProblemDetails follows RFC 7807 for error responses
type ProblemDetails struct {
	Type     string            `json:"type"`             // Error category URI
	Title    string            `json:"title"`            // Short summary
	Status   int               `json:"status"`           // HTTP status
	Detail   string            `json:"detail"`           // Specific explanation
	Instance string            `json:"instance"`         // Request URI
	Errors   map[string]string `json:"errors,omitempty"` // Field-specific issues
}

// NewSuccessResponse creates a new standardized success response
func NewSuccessResponse(c *fiber.Ctx, status int, message string, data interface{}) error {
	return c.Status(status).JSON(SuccessResponse{
		Message: message,
		Status:  status,
		Data:    data,
	})
}

// NewProblemResponse creates a new RFC 7807 problem response
func NewProblemResponse(c *fiber.Ctx, status int, title, detail, typeURI string, validationErrors map[string]string) error {
	c.Set(fiber.HeaderContentType, "application/problem+json")
	return c.Status(status).JSON(ProblemDetails{
		Type:     typeURI,
		Title:    title,
		Status:   status,
		Detail:   detail,
		Instance: c.Path(),
		Errors:   validationErrors,
	})
}

// ErrorHandler is the global error handler for the Fiber application
func ErrorHandler(c *fiber.Ctx, err error) error {
	// Default status and error details
	status := fiber.StatusInternalServerError
	title := "Internal Server Error"
	detail := "An unexpected error occurred"
	typeURI := "/errors/internal"
	var validationErrors map[string]string

	// Handle fiber.Error (e.g. 404, 405)
	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		status = fiberErr.Code
		title = http.StatusText(status)
		detail = fiberErr.Message
		typeURI = getErrorTypeURI(status)
	} else if vErrs, ok := err.(domain.ValidationErrors); ok {
		// Handle ValidationErrors
		status = fiber.StatusBadRequest
		title = "Validation Failed"
		detail = "One or more fields failed validation"
		typeURI = "/errors/validation-failed"
		validationErrors = vErrs
	} else {
		// Handle Domain Errors
		status, typeURI = mapDomainError(err)
		title = http.StatusText(status)
		detail = err.Error()
	}

	// Log internal errors
	if status == fiber.StatusInternalServerError {
		log.Printf("[ERROR] Internal Server Error: %v", err)
		detail = "An unexpected error occurred. Please contact support."
	}

	return NewProblemResponse(c, status, title, detail, typeURI, validationErrors)
}

// getErrorTypeURI maps HTTP status codes to standard error URIs
func getErrorTypeURI(status int) string {
	switch status {
	case fiber.StatusBadRequest:
		return "/errors/bad-request"
	case fiber.StatusUnauthorized:
		return "/errors/unauthorized"
	case fiber.StatusForbidden:
		return "/errors/forbidden"
	case fiber.StatusNotFound:
		return "/errors/not-found"
	case fiber.StatusConflict:
		return "/errors/conflict"
	default:
		return "/errors/internal"
	}
}

// mapDomainError maps domain sentinel errors to HTTP status codes and URIs
func mapDomainError(err error) (int, string) {
	switch {
	case errors.Is(err, domain.ErrUserNotFound),
		errors.Is(err, domain.ErrMerchantNotFound),
		errors.Is(err, domain.ErrProductNotFound),
		errors.Is(err, domain.ErrWalletNotFound),
		errors.Is(err, domain.ErrOrderNotFound),
		errors.Is(err, domain.ErrCartItemNotFound),
		errors.Is(err, domain.ErrAddressNotFound),
		errors.Is(err, domain.ErrPaymentNotFound):
		return fiber.StatusNotFound, "/errors/not-found"

	case errors.Is(err, domain.ErrUserAlreadyExists),
		errors.Is(err, domain.ErrMerchantAlreadyExists),
		errors.Is(err, domain.ErrWalletAlreadyExists),
		errors.Is(err, domain.ErrDuplicatePayment):
		return fiber.StatusConflict, "/errors/conflict"

	case errors.Is(err, domain.ErrInvalidCredentials),
		errors.Is(err, domain.ErrInvalidRefreshToken),
		errors.Is(err, domain.ErrRefreshTokenExpired),
		errors.Is(err, domain.ErrRefreshTokenReused),
		errors.Is(err, domain.ErrInvalidSocialToken),
		errors.Is(err, domain.ErrInvalidVerificationCode),
		errors.Is(err, domain.ErrVerificationCodeExpired),
		errors.Is(err, domain.ErrInvalidOAuthState):
		return fiber.StatusUnauthorized, "/errors/unauthorized"

	case errors.Is(err, domain.ErrForbidden):
		return fiber.StatusForbidden, "/errors/forbidden"

	case errors.Is(err, domain.ErrInsufficientBalance),
		errors.Is(err, domain.ErrInsufficientStock),
		errors.Is(err, domain.ErrInvalidStatusTransition),
		errors.Is(err, domain.ErrWalletNotActive),
		errors.Is(err, domain.ErrAuthProviderMismatch),
		errors.Is(err, domain.ErrEmailAlreadyUsedByOtherMethod),
		errors.Is(err, domain.ErrEmailNotVerified),
		errors.Is(err, domain.ErrEmailPasswordSignInNotAllowed),
		errors.Is(err, domain.ErrMerchantShipmentTooEarly),
		errors.Is(err, domain.ErrOrderNotCancellable),
		errors.Is(err, domain.ErrRecipientNameRequired),
		errors.Is(err, domain.ErrPhoneNumberRequired),
		errors.Is(err, domain.ErrStreetAddressRequired),
		errors.Is(err, domain.ErrCityRequired),
		errors.Is(err, domain.ErrProvinceRequired),
		errors.Is(err, domain.ErrPostalCodeRequired),
		errors.Is(err, domain.ErrInvalidAddressTag),
		errors.Is(err, domain.ErrInsufficientPendingBalance):
		return fiber.StatusBadRequest, "/errors/bad-request"

	default:
		return fiber.StatusInternalServerError, "/errors/internal"
	}
}

// Deprecated: Use NewSuccessResponse instead
type ResponseWrapper struct {
	Message string      `json:"message"`
	Status  int         `json:"status"`
	Data    interface{} `json:"data"`
}

// Deprecated: Use NewSuccessResponse instead
func NewResponse(c *fiber.Ctx, status int, message string, data interface{}) error {
	return NewSuccessResponse(c, status, message, data)
}
