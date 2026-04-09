package domain

import (
	"time"

	"github.com/google/uuid"
)

type Merchant struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Name      string    `json:"name" db:"name"`
	About     string    `json:"about" db:"about"`
	TaxID     string    `json:"tax_id" db:"tax_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
