package merchant

import (
	"context"
	"errors"
	"testing"

	"go-shop-yourself/internal/domain"
	"go-shop-yourself/internal/user"
	"go-shop-yourself/internal/wallet"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMerchantService_RegisterMerchant_Success(t *testing.T) {
	mockRepo := NewMockMerchantRepository(t)
	mockUserRepo := user.NewMockUserRepository(t)
	mockWalletRepo := wallet.NewMockWalletRepository(t)
	mockPool := domain.NewMockPool(t)

	service := NewMerchantService(mockRepo, mockUserRepo, mockWalletRepo)

	userID := uuid.New()
	req := MerchantRegisterRequest{
		Name:  "Test Shop",
		About: "About us",
		TaxID: "123456",
	}

	u := &domain.User{ID: userID, Email: "test@example.com"}

	// Create pgxmock to simulate the transaction
	mockTx, err := pgxmock.NewConn()
	assert.NoError(t, err)

	mockUserRepo.On("GetUserByID", mock.Anything, userID).Return(u, nil)
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

func TestMerchantService_RegisterMerchant_Fail_UserNotFound(t *testing.T) {
	mockUserRepo := user.NewMockUserRepository(t)
	service := NewMerchantService(nil, mockUserRepo, nil)

	userID := uuid.New()
	mockUserRepo.On("GetUserByID", mock.Anything, userID).Return(nil, nil)

	_, err := service.RegisterMerchant(context.Background(), userID, MerchantRegisterRequest{})

	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrUserNotFound))
}

func TestMerchantService_RegisterMerchant_Fail_AlreadyExists(t *testing.T) {
	mockUserRepo := user.NewMockUserRepository(t)
	mockRepo := NewMockMerchantRepository(t)
	service := NewMerchantService(mockRepo, mockUserRepo, nil)

	userID := uuid.New()
	u := &domain.User{ID: userID}
	existing := &domain.Merchant{ID: uuid.New(), UserID: userID}

	mockUserRepo.On("GetUserByID", mock.Anything, userID).Return(u, nil)
	mockRepo.On("GetByUserID", mock.Anything, userID).Return(existing, nil)

	_, err := service.RegisterMerchant(context.Background(), userID, MerchantRegisterRequest{Name: "New"})

	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrMerchantAlreadyExists))
}

func TestMerchantService_RegisterMerchant_Fail_BeginTxError(t *testing.T) {
	mockRepo := NewMockMerchantRepository(t)
	mockUserRepo := user.NewMockUserRepository(t)
	mockPool := domain.NewMockPool(t)
	service := NewMerchantService(mockRepo, mockUserRepo, nil)

	userID := uuid.New()
	u := &domain.User{ID: userID}

	mockUserRepo.On("GetUserByID", mock.Anything, userID).Return(u, nil)
	mockRepo.On("GetByUserID", mock.Anything, userID).Return(nil, nil)
	mockRepo.On("GetPool").Return(mockPool)
	mockPool.On("Begin", mock.Anything).Return(nil, errors.New("begin fail"))

	_, err := service.RegisterMerchant(context.Background(), userID, MerchantRegisterRequest{Name: "New"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "begin fail")
}
