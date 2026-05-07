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

func TestUserService_GetUserByID_Success(t *testing.T) {
	mockRepo := NewMockUserRepository(t)
	service := NewUserService(mockRepo)

	id := uuid.New()
	u := &domain.User{
		ID:    id,
		Email: "test@example.com",
	}

	mockRepo.On("GetUserByID", context.Background(), id).Return(u, nil)

	res, err := service.GetUserByID(context.Background(), id)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, id, res.ID)
}

func TestUserService_GetUserByID_Fail_NotFound(t *testing.T) {
	mockRepo := NewMockUserRepository(t)
	service := NewUserService(mockRepo)

	id := uuid.New()
	mockRepo.On("GetUserByID", context.Background(), id).Return(nil, nil)

	_, err := service.GetUserByID(context.Background(), id)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrUserNotFound))
}

func TestUserService_GetUserByID_Fail_RepoError(t *testing.T) {
	mockRepo := NewMockUserRepository(t)
	service := NewUserService(mockRepo)

	id := uuid.New()
	mockRepo.On("GetUserByID", context.Background(), id).Return(nil, errors.New("db error"))

	_, err := service.GetUserByID(context.Background(), id)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

func TestUserService_AddAddress_Success(t *testing.T) {
	mockRepo := NewMockUserRepository(t)
	service := NewUserService(mockRepo)

	userID := uuid.New()
	req := &AddressRequest{
		Tag:           "custom-tag",
		RecipientName: "Jane Doe",
		PhoneNumber:   "0812345678",
		StreetAddress: "456 Side St",
		City:          "Bandung",
		Province:      "West Java",
		PostalCode:    "40123",
		IsDefault:     true,
	}

	mockRepo.On("UnsetDefaultAddresses", context.Background(), userID).Return(nil)
	mockRepo.On("CreateAddress", context.Background(), mock.MatchedBy(func(addr *domain.UserAddress) bool {
		return addr.UserID == userID && addr.Tag == "custom-tag" && addr.IsDefault == true
	})).Return(nil)

	res, err := service.AddAddress(context.Background(), userID, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, domain.AddressTag("custom-tag"), res.Tag)
	assert.Equal(t, true, res.IsDefault)
}

func TestUserService_UpdateAddress_Success(t *testing.T) {
	mockRepo := NewMockUserRepository(t)
	service := NewUserService(mockRepo)

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

	mockRepo.On("GetAddressByID", context.Background(), addressID).Return(existingAddr, nil)
	mockRepo.On("UnsetDefaultAddresses", context.Background(), userID).Return(nil)
	mockRepo.On("UpdateAddress", context.Background(), mock.Anything).Return(nil)

	res, err := service.UpdateAddress(context.Background(), userID, addressID, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, domain.AddressTag("new"), res.Tag)
	assert.True(t, res.IsDefault)
}
