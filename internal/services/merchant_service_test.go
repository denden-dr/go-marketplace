package services

import (
	"context"
	"errors"
	"testing"

	"go-shop-yourself/internal/domain"
	"go-shop-yourself/internal/dtos"
	"go-shop-yourself/internal/mocks"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegisterMerchant_Success(t *testing.T) {
	mockRepo := mocks.NewMerchantRepository(t)
	mockUserRepo := mocks.NewUserRepository(t)
	mockWalletRepo := mocks.NewWalletRepository(t)
	mockPool := mocks.NewPool(t)

	service := NewMerchantService(mockRepo, mockUserRepo, mockWalletRepo)

	userID := uuid.New()
	req := dtos.MerchantRegisterRequest{
		Name:  "Test Shop",
		About: "About us",
		TaxID: "123456",
	}

	user := &domain.User{ID: userID, Email: "test@example.com"}
	
	// Create pgxmock to simulate the transaction
	mockTx, err := pgxmock.NewConn()
	assert.NoError(t, err)

	mockUserRepo.On("GetUserByID", mock.Anything, userID).Return(user, nil)
	mockRepo.On("GetByUserID", mock.Anything, userID).Return(nil, nil)
	mockRepo.On("GetPool").Return(mockPool)
	
	mockPool.On("Begin", mock.Anything).Return(mockTx, nil)
	mockRepo.On("CreateTx", mock.Anything, mockTx, mock.Anything).Return(nil)
	mockWalletRepo.On("CreateTx", mock.Anything, mockTx, mock.Anything).Return(nil)
	
	// We need to handle the commit since mockTx is a pgxmock
	mockTx.ExpectCommit()

	res, err := service.RegisterMerchant(context.Background(), userID, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "Test Shop", res.Name)
	assert.NoError(t, mockTx.ExpectationsWereMet())
}

func TestRegisterMerchant_Fail_UserNotFound(t *testing.T) {
	mockUserRepo := mocks.NewUserRepository(t)
	service := NewMerchantService(nil, mockUserRepo, nil)

	userID := uuid.New()
	mockUserRepo.On("GetUserByID", mock.Anything, userID).Return(nil, nil)

	_, err := service.RegisterMerchant(context.Background(), userID, dtos.MerchantRegisterRequest{})

	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrUserNotFound))
}

func TestRegisterMerchant_Fail_AlreadyExists(t *testing.T) {
	mockUserRepo := mocks.NewUserRepository(t)
	mockRepo := mocks.NewMerchantRepository(t)
	service := NewMerchantService(mockRepo, mockUserRepo, nil)

	userID := uuid.New()
	user := &domain.User{ID: userID}
	existing := &domain.Merchant{ID: uuid.New(), UserID: userID}

	mockUserRepo.On("GetUserByID", mock.Anything, userID).Return(user, nil)
	mockRepo.On("GetByUserID", mock.Anything, userID).Return(existing, nil)

	_, err := service.RegisterMerchant(context.Background(), userID, dtos.MerchantRegisterRequest{Name: "New"})

	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrMerchantAlreadyExists))
}

func TestRegisterMerchant_Fail_BeginTxError(t *testing.T) {
	mockRepo := mocks.NewMerchantRepository(t)
	mockUserRepo := mocks.NewUserRepository(t)
	mockPool := mocks.NewPool(t)
	service := NewMerchantService(mockRepo, mockUserRepo, nil)

	userID := uuid.New()
	user := &domain.User{ID: userID}

	mockUserRepo.On("GetUserByID", mock.Anything, userID).Return(user, nil)
	mockRepo.On("GetByUserID", mock.Anything, userID).Return(nil, nil)
	mockRepo.On("GetPool").Return(mockPool)
	mockPool.On("Begin", mock.Anything).Return(nil, errors.New("begin fail"))

	_, err := service.RegisterMerchant(context.Background(), userID, dtos.MerchantRegisterRequest{Name: "New"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "begin fail")
}
