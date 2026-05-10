package user

import (
	"go-marketplace/internal/domain"
	"time"

	"github.com/google/uuid"
)

// User represents a rich domain entity for a user
type User struct {
	model *domain.User
}

// NewUser creates a new rich User entity
func NewUser(m *domain.User) *User {
	return &User{model: m}
}

// IsVerified returns true if the user has verified their email
func (u *User) IsVerified() bool {
	return u.model.IsVerified
}

// UserAddress represents a rich domain entity for a user's address
type UserAddress struct {
	model *domain.UserAddress
}

// NewAddress creates a new rich UserAddress entity
func NewAddress(m *domain.UserAddress) *UserAddress {
	return &UserAddress{model: m}
}

// IsOwnedBy returns true if the address belongs to the given userID
func (a *UserAddress) IsOwnedBy(userID uuid.UUID) bool {
	return a.model.UserID == userID
}

// Update updates the address fields from the provided request
func (a *UserAddress) Update(req *AddressRequest) {
	a.model.Tag = req.Tag
	a.model.RecipientName = req.RecipientName
	a.model.PhoneNumber = req.PhoneNumber
	a.model.StreetAddress = req.StreetAddress
	a.model.City = req.City
	a.model.Province = req.Province
	a.model.PostalCode = req.PostalCode
	a.model.IsDefault = req.IsDefault
	a.model.UpdatedAt = time.Now()
}
