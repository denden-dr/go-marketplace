package dto

// APIResponse is a generic wrapper for all API responses.
type APIResponse struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
	Data    any    `json:"data,omitempty"`
}

// NewSuccessResponse creates a new success API response.
func NewSuccessResponse(message string, status int, data any) APIResponse {
	return APIResponse{
		Message: message,
		Status:  status,
		Data:    data,
	}
}

// NewErrorResponse creates a new error API response.
func NewErrorResponse(message string, status int) APIResponse {
	return APIResponse{
		Message: message,
		Status:  status,
	}
}
