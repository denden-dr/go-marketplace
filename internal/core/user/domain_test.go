package user

import (
	"go-marketplace/internal/domain"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUser_IsVerified(t *testing.T) {
	tests := []struct {
		name     string
		verified bool
		want     bool
	}{
		{"verified user", true, true},
		{"unverified user", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := NewUser(&domain.User{IsVerified: tt.verified})
			assert.Equal(t, tt.want, u.IsVerified())
		})
	}
}

func TestUserAddress_IsOwnedBy(t *testing.T) {
	userID := uuid.New()
	otherUserID := uuid.New()

	tests := []struct {
		name    string
		ownerID uuid.UUID
		checkID uuid.UUID
		want    bool
	}{
		{"correct owner", userID, userID, true},
		{"wrong owner", userID, otherUserID, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewAddress(&domain.UserAddress{UserID: tt.ownerID})
			assert.Equal(t, tt.want, a.IsOwnedBy(tt.checkID))
		})
	}
}

func TestUserAddress_Update(t *testing.T) {
	addr := &domain.UserAddress{
		Tag: "Home",
	}
	entity := NewAddress(addr)

	req := &AddressRequest{
		Tag:           "Office",
		RecipientName: "John Doe",
		PhoneNumber:   "123456",
		StreetAddress: "Main St",
		City:          "City",
		Province:      "Province",
		PostalCode:    "12345",
		IsDefault:     true,
	}

	entity.Update(req)

	assert.Equal(t, req.Tag, addr.Tag)
	assert.Equal(t, req.RecipientName, addr.RecipientName)
	assert.Equal(t, req.IsDefault, addr.IsDefault)
}
