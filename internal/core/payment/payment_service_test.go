package payment

import (
	"context"
	"testing"

	"go-marketplace/internal/core/wallet"
	"go-marketplace/internal/domain"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func getSqlxTx(t *testing.T) (*sqlx.Tx, sqlmock.Sqlmock, *sqlx.DB) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	mock.ExpectBegin()
	tx, err := sqlxDB.BeginTxx(context.Background(), nil)
	assert.NoError(t, err)
	return tx, mock, sqlxDB
}

func TestPaymentService_CreatePaymentTX_ExternalAtomicity(t *testing.T) {
	mockRepo := NewMockPaymentRepository(t)
	mockWallet := wallet.NewMockWalletService(t)
	mockProvider := NewMockPaymentProvider(t)
	mockOrder := NewMockOrderManager(t)

	tx, sqlMock, db := getSqlxTx(t)
	s := NewPaymentService(mockRepo, mockWallet, mockProvider, mockOrder, db)

	ctx := context.Background()
	req := CreatePaymentRequest{
		UserID:      uuid.New(),
		Amount:      decimal.NewFromInt(100),
		Type:        domain.PaymentTypeOrder,
		Method:      domain.PaymentMethodMidtrans,
		ReferenceID: uuid.New(),
	}

	// The current implementation starts a new transaction with s.db.BeginTxx()
	// for external payments inside createPaymentInternal.
	// This test expects it NOT to call BeginTxx if a tx is provided.

	mockRepo.On("CreateTX", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	mockProvider.On("CreateTransaction", mock.Anything, mock.Anything).Return("snap-token", nil).Once()
	mockRepo.On("UpdateSnapTokenTX", mock.Anything, mock.Anything, mock.Anything, "snap-token").Return(nil).Once()

	_, err := s.CreatePaymentTX(ctx, tx, req)
	assert.NoError(t, err)

	// If the implementation called BeginTxx, sqlMock would have an unexpected call.
	err = sqlMock.ExpectationsWereMet()
	assert.NoError(t, err, "Should not have started extra transactions")
}
