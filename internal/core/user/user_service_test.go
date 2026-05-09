package user

import (
	"context"
	"errors"
	"testing"

	"go-marketplace/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserService_GetUserByID(t *testing.T) {
	id := uuid.New()
	u := &domain.User{
		ID:    id,
		Email: "test@example.com",
	}

	tests := []struct {
		name      string
		userID    uuid.UUID
		mockSetup func(mr *MockUserRepository)
		wantErr   bool
		errType   error
		errMsg    string
	}{
		{
			name:   "Success",
			userID: id,
			mockSetup: func(mr *MockUserRepository) {
				mr.On("GetUserByID", mock.Anything, id).Return(u, nil)
			},
			wantErr: false,
		},
		{
			name:   "User Not Found",
			userID: id,
			mockSetup: func(mr *MockUserRepository) {
				mr.On("GetUserByID", mock.Anything, id).Return(nil, nil)
			},
			wantErr: true,
			errType: domain.ErrUserNotFound,
		},
		{
			name:   "Database Error",
			userID: id,
			mockSetup: func(mr *MockUserRepository) {
				mr.On("GetUserByID", mock.Anything, id).Return(nil, errors.New("db error"))
			},
			wantErr: true,
			errMsg:  "db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMockUserRepository(t)
			tt.mockSetup(mockRepo)

			service := NewUserService(mockRepo)
			res, err := service.GetUserByID(context.Background(), tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
				assert.Equal(t, tt.userID, res.ID)
			}
		})
	}
}

func TestUserService_AddAddress(t *testing.T) {
	userID := uuid.New()
	req := &AddressRequest{
		Tag:           "home",
		RecipientName: "John Doe",
		IsDefault:     true,
	}

	tests := []struct {
		name      string
		userID    uuid.UUID
		request   *AddressRequest
		mockSetup func(mr *MockUserRepository)
		wantErr   bool
	}{
		{
			name:    "Success",
			userID:  userID,
			request: req,
			mockSetup: func(mr *MockUserRepository) {
				mr.On("UnsetDefaultAddresses", mock.Anything, userID).Return(nil)
				mr.On("CreateAddress", mock.Anything, mock.MatchedBy(func(addr *domain.UserAddress) bool {
					return addr.UserID == userID && addr.Tag == domain.AddressTag(req.Tag) && addr.IsDefault == true
				})).Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMockUserRepository(t)
			tt.mockSetup(mockRepo)

			service := NewUserService(mockRepo)
			res, err := service.AddAddress(context.Background(), tt.userID, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
				assert.Equal(t, domain.AddressTag(tt.request.Tag), res.Tag)
				assert.True(t, res.IsDefault)
			}
		})
	}
}

func TestUserService_UpdateAddress(t *testing.T) {
	userID := uuid.New()
	addressID := uuid.New()
	existingAddr := &domain.UserAddress{
		ID:        addressID,
		UserID:    userID,
		Tag:       "old",
		IsDefault: false,
	}

	req := &AddressRequest{
		Tag:       "new",
		IsDefault: true,
	}

	tests := []struct {
		name      string
		userID    uuid.UUID
		addressID uuid.UUID
		request   *AddressRequest
		mockSetup func(mr *MockUserRepository)
		wantErr   bool
	}{
		{
			name:      "Success",
			userID:    userID,
			addressID: addressID,
			request:   req,
			mockSetup: func(mr *MockUserRepository) {
				mr.On("GetAddressByID", mock.Anything, addressID).Return(existingAddr, nil)
				mr.On("UnsetDefaultAddresses", mock.Anything, userID).Return(nil)
				mr.On("UpdateAddress", mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMockUserRepository(t)
			tt.mockSetup(mockRepo)

			service := NewUserService(mockRepo)
			res, err := service.UpdateAddress(context.Background(), tt.userID, tt.addressID, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
				assert.Equal(t, domain.AddressTag(tt.request.Tag), res.Tag)
				assert.True(t, res.IsDefault)
			}
		})
	}
}
