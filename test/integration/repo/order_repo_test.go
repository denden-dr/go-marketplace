//go:build integration

package repo

import (
	"context"
	"testing"
	"time"

	"go-marketplace/internal/core/merchant"
	"go-marketplace/internal/core/order"
	"go-marketplace/internal/core/payment"
	"go-marketplace/internal/core/product"
	"go-marketplace/internal/core/user"
	"go-marketplace/internal/domain"
	"go-marketplace/internal/testutil"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
)

type OrderRepoSuite struct {
	testutil.IntegrationSuite
	repo         order.OrderRepository
	productRepo  product.ProductRepository
	merchantRepo merchant.MerchantRepository
	userRepo     user.UserRepository
	paymentRepo  payment.PaymentRepository
}

func (s *OrderRepoSuite) SetupSuite() {
	s.IntegrationSuite.SetupSuite()
	s.repo = order.NewOrderRepository(s.DB)
	s.productRepo = product.NewProductRepository(s.DB)
	s.merchantRepo = merchant.NewMerchantRepository(s.DB)
	s.userRepo = user.NewUserRepository(s.DB)
	s.paymentRepo = payment.NewPaymentRepository(s.DB)
}

func (s *OrderRepoSuite) TestOrderOperations() {
	// Setup Data
	u := &domain.User{
		ID:           uuid.New(),
		FullName:     "Order User",
		Username:     "orderuser",
		Email:        "order@example.com",
		AuthProvider: domain.AuthProviderLocal,
		CreatedAt:    time.Now().Truncate(time.Microsecond),
	}
	s.NoError(s.userRepo.CreateUser(context.Background(), u))

	m := &domain.Merchant{
		ID:        uuid.New(),
		UserID:    u.ID,
		Name:      "Order Shop",
		CreatedAt: time.Now().Truncate(time.Microsecond),
	}
	s.NoError(s.merchantRepo.Create(context.Background(), m))

	p := &domain.Product{
		ID:        uuid.New(),
		StoreID:   m.ID,
		Name:      "Order Product",
		Price:     decimal.NewFromInt(100),
		Stock:     10,
		CreatedAt: time.Now().Truncate(time.Microsecond),
	}
	s.NoError(s.productRepo.Create(context.Background(), p))

	tx, err := s.repo.Begin(context.Background())
	s.NoError(err)
	defer tx.Rollback()

	pay := &domain.Payment{
		ID:            uuid.New(),
		UserID:        u.ID,
		Amount:        decimal.NewFromInt(100),
		Type:          domain.PaymentTypeOrder,
		Method:        domain.PaymentMethodWallet,
		Status:        domain.PaymentStatusSuccess,
		CreatedAt:     time.Now().Truncate(time.Microsecond),
		UpdatedAt:     time.Now().Truncate(time.Microsecond),
	}
	s.NoError(s.paymentRepo.CreateTX(context.Background(), tx, pay))

	o := &domain.Order{
		ID:                    uuid.New(),
		PaymentID:             pay.ID,
		MerchantID:            m.ID,
		UserID:                u.ID,
		Status:                domain.OrderStatusPending,
		TotalAmount:           decimal.NewFromInt(100),
		ShippingRecipientName: "John Doe",
		CreatedAt:             time.Now().Truncate(time.Microsecond),
		UpdatedAt:             time.Now().Truncate(time.Microsecond),
	}
	s.NoError(s.repo.CreateOrderTX(context.Background(), tx, o))

	item := &domain.OrderItem{
		ID:        uuid.New(),
		OrderID:   o.ID,
		ProductID: p.ID,
		Quantity:  1,
		Price:     decimal.NewFromInt(100),
		CreatedAt: time.Now().Truncate(time.Microsecond),
	}
	s.NoError(s.repo.CreateOrderItemTX(context.Background(), tx, item))

	s.NoError(tx.Commit())

	// 2. Verify
	dbOrder, err := s.repo.GetOrderByID(context.Background(), o.ID)
	s.NoError(err)
	s.NotNil(dbOrder)
	s.Equal(o.ID, dbOrder.ID)
	s.Equal(o.Status, dbOrder.Status)

	items, err := s.repo.GetOrderItems(context.Background(), o.ID)
	s.NoError(err)
	s.Len(items, 1)
	s.Equal(item.ID, items[0].ID)

	// 3. Update Status
	err = s.repo.UpdateOrderStatus(context.Background(), o.ID, domain.OrderStatusProcessing)
	s.NoError(err)

	dbOrder, _ = s.repo.GetOrderByID(context.Background(), o.ID)
	s.Equal(domain.OrderStatusProcessing, dbOrder.Status)

	tx2, err := s.repo.Begin(context.Background())
	s.NoError(err)
	defer tx2.Rollback()

	err = s.repo.UpdateOrderAppealTX(context.Background(), tx2, o.ID, true)
	s.NoError(err)

	appeal := &domain.CancellationAppeal{
		ID:        uuid.New(),
		OrderID:   o.ID,
		Reason:    "Change of mind",
		Status:    "pending",
		CreatedAt: time.Now().Truncate(time.Microsecond),
	}
	err = s.repo.CreateAppeal(context.Background(), appeal)
	s.NoError(err)

	s.NoError(tx2.Commit())

	dbOrder, _ = s.repo.GetOrderByID(context.Background(), o.ID)
	s.True(dbOrder.IsAppealed)
}

func TestOrderRepoSuite(t *testing.T) {
	suite.Run(t, new(OrderRepoSuite))
}
