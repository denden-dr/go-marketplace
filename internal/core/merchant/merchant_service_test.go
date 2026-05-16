package merchant

import (
	"context"
	"errors"
	"testing"

	"go-marketplace/internal/core/user"
	"go-marketplace/internal/core/wallet"
	"go-marketplace/internal/domain"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func getSqlxTx(t *testing.T) (*sqlx.Tx, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	mock.ExpectBegin()
	tx, err := sqlxDB.BeginTxx(context.Background(), nil)
	assert.NoError(t, err)
	return tx, mock
}

func TestMerchantService_RegisterMerchant(t *testing.T) {
	userID := uuid.New()
	req := MerchantRegisterRequest{
		Name:  "Test Shop",
		About: "About us",
		TaxID: "123456",
	}

	tests := []struct {
		name      string
		mockSetup func(mr *MockMerchantRepository, mur *user.MockUserRepository, mwr *wallet.MockWalletRepository, mp *domain.MockPool, tx *sqlx.Tx, sqlMock sqlmock.Sqlmock)
		wantErr   bool
		errType   error
		errMsg    string
	}{
		{
			name: "Success",
			mockSetup: func(mr *MockMerchantRepository, mur *user.MockUserRepository, mwr *wallet.MockWalletRepository, mp *domain.MockPool, tx *sqlx.Tx, sqlMock sqlmock.Sqlmock) {
				u := &domain.User{ID: userID, Email: "test@example.com"}
				mur.On("GetUserByID", mock.Anything, userID).Return(u, nil)
				mr.On("GetByUserID", mock.Anything, userID).Return(nil, nil)
				mr.On("GetPool").Return(mp)
				mp.On("BeginTxx", mock.Anything, mock.Anything).Return(tx, nil)
				mr.On("CreateTx", mock.Anything, tx, mock.Anything).Return(nil)
				mwr.On("CreateTx", mock.Anything, tx, mock.Anything).Return(nil)
				mur.On("UpdateRoleTx", mock.Anything, tx, userID, domain.RoleMerchant).Return(nil)
				sqlMock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "User Not Found",
			mockSetup: func(mr *MockMerchantRepository, mur *user.MockUserRepository, mwr *wallet.MockWalletRepository, mp *domain.MockPool, tx *sqlx.Tx, sqlMock sqlmock.Sqlmock) {
				mur.On("GetUserByID", mock.Anything, userID).Return(nil, nil)
			},
			wantErr: true,
			errType: domain.ErrUserNotFound,
		},
		{
			name: "Merchant Already Exists",
			mockSetup: func(mr *MockMerchantRepository, mur *user.MockUserRepository, mwr *wallet.MockWalletRepository, mp *domain.MockPool, tx *sqlx.Tx, sqlMock sqlmock.Sqlmock) {
				u := &domain.User{ID: userID}
				mur.On("GetUserByID", mock.Anything, userID).Return(u, nil)
				mr.On("GetByUserID", mock.Anything, userID).Return(&domain.Merchant{}, nil)
			},
			wantErr: true,
			errType: domain.ErrMerchantAlreadyExists,
		},
		{
			name: "Begin Transaction Error",
			mockSetup: func(mr *MockMerchantRepository, mur *user.MockUserRepository, mwr *wallet.MockWalletRepository, mp *domain.MockPool, tx *sqlx.Tx, sqlMock sqlmock.Sqlmock) {
				u := &domain.User{ID: userID}
				mur.On("GetUserByID", mock.Anything, userID).Return(u, nil)
				mr.On("GetByUserID", mock.Anything, userID).Return(nil, nil)
				mr.On("GetPool").Return(mp)
				mp.On("BeginTxx", mock.Anything, mock.Anything).Return(nil, errors.New("begin fail"))
			},
			wantErr: true,
			errMsg:  "begin fail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr := NewMockMerchantRepository(t)
			mur := user.NewMockUserRepository(t)
			mwr := wallet.NewMockWalletRepository(t)
			mp := domain.NewMockPool(t)

			db, sqlMock, _ := sqlmock.New()
			sqlxDB := sqlx.NewDb(db, "sqlmock")
			var tx *sqlx.Tx
			if tt.name == "Success" {
				sqlMock.ExpectBegin()
				tx, _ = sqlxDB.BeginTxx(context.Background(), nil)
			}

			tt.mockSetup(mr, mur, mwr, mp, tx, sqlMock)

			service := NewMerchantService(mr, mur, mwr)
			res, err := service.RegisterMerchant(context.Background(), userID, req)

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
				assert.Equal(t, req.Name, res.Name)
				assert.NoError(t, sqlMock.ExpectationsWereMet())
			}
		})
	}
}
