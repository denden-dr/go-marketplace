package dtos

import "github.com/google/uuid"

type AuthResponse struct {
	ID uuid.UUID `json:"id"`
}
