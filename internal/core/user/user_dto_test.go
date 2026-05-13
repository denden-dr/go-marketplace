package user

import (
	"go-marketplace/internal/domain"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddressRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request AddressRequest
		wantErr bool
		errKeys []string
	}{
		{
			name: "valid request",
			request: AddressRequest{
				Tag:           domain.AddressTagHome,
				RecipientName: "John Doe",
				PhoneNumber:   "08123456789",
				StreetAddress: "Main St 123",
				City:          "Springfield",
				Province:      "Illinois",
				PostalCode:    "62704",
			},
			wantErr: false,
		},
		{
			name:    "empty request",
			request: AddressRequest{},
			wantErr: true,
			errKeys: []string{"recipient_name", "phone_number", "street_address", "city", "province", "postal_code", "tag"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				vErrs, ok := err.(domain.ValidationErrors)
				assert.True(t, ok, "should be ValidationErrors type")
				for _, key := range tt.errKeys {
					assert.Contains(t, vErrs, key)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
