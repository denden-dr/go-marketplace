package services

import (
	"context"
	"errors"
	"testing"

	"go-shop-yourself/internal/domain"
	"go-shop-yourself/internal/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetUserByID_Success(t *testing.T) {
	mockRepo := mocks.NewUserRepository(t)
	service := NewUserService(mockRepo)

	id := uuid.New()
	user := &domain.User{
		ID:    id,
		Email: "test@example.com",
	}

	mockRepo.On("GetUserByID", context.Background(), id).Return(user, nil)

	res, err := service.GetUserByID(context.Background(), id)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, id, res.ID)
}

func TestGetUserByID_Fail_NotFound(t *testing.T) {
	mockRepo := mocks.NewUserRepository(t)
	service := NewUserService(mockRepo)

	id := uuid.New()
	mockRepo.On("GetUserByID", context.Background(), id).Return(nil, nil)

	_, err := service.GetUserByID(context.Background(), id)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrUserNotFound))
}

func TestGetUserByID_Fail_RepoError(t *testing.T) {
	mockRepo := mocks.NewUserRepository(t)
	service := NewUserService(mockRepo)

	id := uuid.New()
	mockRepo.On("GetUserByID", context.Background(), id).Return(nil, errors.New("db error"))

	_, err := service.GetUserByID(context.Background(), id)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}
