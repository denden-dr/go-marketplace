package domain

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	RoleUser          UserRole = "user"
	RoleMerchant      UserRole = "merchant"
	RoleAdministrator UserRole = "administrator"
)

type User struct {
	ID           uuid.UUID `json:"id" db:"id"`
	FullName     string    `json:"full_name" db:"full_name"`
	Username     string    `json:"username" db:"username"`
	Email        string    `json:"email" db:"email"`
	Password     *string   `json:"-" db:"password"`
	AuthProvider string    `json:"auth_provider" db:"auth_provider"`
	ProviderID   *string   `json:"provider_id" db:"provider_id"`
	IsVerified   bool      `json:"is_verified" db:"is_verified"`
	Role         UserRole  `json:"role" db:"role"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

const (
	AuthProviderLocal    = "local"
	AuthProviderGoogle   = "google"
	AuthProviderFacebook = "facebook"
	AuthProviderApple    = "apple"
	AuthProviderTwitter  = "twitter"
	AuthProviderSupabase = "supabase"
)

type AddressTag string

const (
	AddressTagHome AddressTag = "home"
	AddressTagWork AddressTag = "work"
)

type UserAddress struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	UserID        uuid.UUID  `json:"user_id" db:"user_id"`
	Tag           AddressTag `json:"tag" db:"tag"` // home, work, other, etc.
	RecipientName string     `json:"recipient_name" db:"recipient_name"`
	PhoneNumber   string     `json:"phone_number" db:"phone_number"`
	StreetAddress string     `json:"street_address" db:"street_address"`
	City          string     `json:"city" db:"city"`
	Province      string     `json:"province" db:"province"`
	PostalCode    string     `json:"postal_code" db:"postal_code"`
	IsDefault     bool       `json:"is_default" db:"is_default"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}
