package user

import (
	"context"
	"errors"
	"testing"

	"go-shop-yourself/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
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
