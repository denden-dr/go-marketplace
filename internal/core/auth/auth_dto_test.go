package auth

import (
	"go-marketplace/internal/domain"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegisterRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request RegisterRequest
		wantErr bool
		errKeys []string
	}{
		{
			name: "valid request",
			request: RegisterRequest{
				FullName: "John Doe",
				Email:    "john@example.com",
				Password: "password123",
				Username: "johndoe",
			},
			wantErr: false,
		},
		{
			name:    "empty request",
			request: RegisterRequest{},
			wantErr: true,
			errKeys: []string{"full_name", "email", "password", "username"},
		},
		{
			name: "short password",
			request: RegisterRequest{
				FullName: "John Doe",
				Email:    "john@example.com",
				Password: "123",
				Username: "johndoe",
			},
			wantErr: true,
			errKeys: []string{"password"},
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
				assert.Equal(t, len(tt.errKeys), len(vErrs))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLoginRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request LoginRequest
		wantErr bool
		errKeys []string
	}{
		{
			name: "valid request",
			request: LoginRequest{
				Email:    "john@example.com",
				Password: "password123",
			},
			wantErr: false,
		},
		{
			name:    "empty request",
			request: LoginRequest{},
			wantErr: true,
			errKeys: []string{"email", "password"},
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
