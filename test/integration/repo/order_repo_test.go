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
		Role:         domain.RoleUser,
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

func (s *OrderRepoSuite) TestUpdateOrderStatusByPaymentIDTX() {
	// Setup: user + merchant + product
	u := &domain.User{
		ID: uuid.New(), FullName: "Batch User", Username: "batchuser",
		Email: "batch@example.com", AuthProvider: domain.AuthProviderLocal,
		Role:  domain.RoleUser,
		CreatedAt: time.Now().Truncate(time.Microsecond),
	}
	s.NoError(s.userRepo.CreateUser(context.Background(), u))

	m := &domain.Merchant{
		ID: uuid.New(), UserID: u.ID, Name: "Batch Shop",
		CreatedAt: time.Now().Truncate(time.Microsecond),
	}
	s.NoError(s.merchantRepo.Create(context.Background(), m))

	p := &domain.Product{
		ID: uuid.New(), StoreID: m.ID, Name: "Batch Product",
		Price: decimal.NewFromInt(50), Stock: 20,
		CreatedAt: time.Now().Truncate(time.Microsecond),
	}
	s.NoError(s.productRepo.Create(context.Background(), p))

	// Create payment + two orders
	tx, err := s.repo.Begin(context.Background())
	s.NoError(err)
	defer tx.Rollback()

	pay := &domain.Payment{
		ID: uuid.New(), UserID: u.ID, Amount: decimal.NewFromInt(100),
		Type: domain.PaymentTypeOrder, Method: domain.PaymentMethodWallet,
		Status: domain.PaymentStatusPending,
		CreatedAt: time.Now().Truncate(time.Microsecond),
		UpdatedAt: time.Now().Truncate(time.Microsecond),
	}
	s.NoError(s.paymentRepo.CreateTX(context.Background(), tx, pay))

	order1 := &domain.Order{
		ID: uuid.New(), PaymentID: pay.ID, MerchantID: m.ID, UserID: u.ID,
		Status: domain.OrderStatusPending, TotalAmount: decimal.NewFromInt(50),
		ShippingRecipientName: "John",
		CreatedAt: time.Now().Truncate(time.Microsecond),
		UpdatedAt: time.Now().Truncate(time.Microsecond),
	}
	order2 := &domain.Order{
		ID: uuid.New(), PaymentID: pay.ID, MerchantID: m.ID, UserID: u.ID,
		Status: domain.OrderStatusPending, TotalAmount: decimal.NewFromInt(50),
		ShippingRecipientName: "Jane",
		CreatedAt: time.Now().Truncate(time.Microsecond),
		UpdatedAt: time.Now().Truncate(time.Microsecond),
	}
	s.NoError(s.repo.CreateOrderTX(context.Background(), tx, order1))
	s.NoError(s.repo.CreateOrderTX(context.Background(), tx, order2))
	s.NoError(tx.Commit())

	// Act: batch update
	tx2, err := s.repo.Begin(context.Background())
	s.NoError(err)
	defer tx2.Rollback()

	err = s.repo.UpdateOrderStatusByPaymentIDTX(context.Background(), tx2, pay.ID, domain.OrderStatusCancelled)
	s.NoError(err)
	s.NoError(tx2.Commit())

	// Assert: both orders are cancelled
	db1, _ := s.repo.GetOrderByID(context.Background(), order1.ID)
	db2, _ := s.repo.GetOrderByID(context.Background(), order2.ID)
	s.Equal(domain.OrderStatusCancelled, db1.Status)
	s.Equal(domain.OrderStatusCancelled, db2.Status)
}

func (s *OrderRepoSuite) TestGetOrderItemsByOrderIDsTX() {
	// Setup: user + merchant + product
	u := &domain.User{
		ID: uuid.New(), FullName: "Items User", Username: "itemsuser",
		Email: "items@example.com", AuthProvider: domain.AuthProviderLocal,
		Role:  domain.RoleUser,
		CreatedAt: time.Now().Truncate(time.Microsecond),
	}
	s.NoError(s.userRepo.CreateUser(context.Background(), u))

	m := &domain.Merchant{
		ID: uuid.New(), UserID: u.ID, Name: "Items Shop",
		CreatedAt: time.Now().Truncate(time.Microsecond),
	}
	s.NoError(s.merchantRepo.Create(context.Background(), m))

	p := &domain.Product{
		ID: uuid.New(), StoreID: m.ID, Name: "Items Product",
		Price: decimal.NewFromInt(25), Stock: 50,
		CreatedAt: time.Now().Truncate(time.Microsecond),
	}
	s.NoError(s.productRepo.Create(context.Background(), p))

	// Create payment + two orders, each with one item
	tx, err := s.repo.Begin(context.Background())
	s.NoError(err)
	defer tx.Rollback()

	pay := &domain.Payment{
		ID: uuid.New(), UserID: u.ID, Amount: decimal.NewFromInt(50),
		Type: domain.PaymentTypeOrder, Method: domain.PaymentMethodWallet,
		Status: domain.PaymentStatusPending,
		CreatedAt: time.Now().Truncate(time.Microsecond),
		UpdatedAt: time.Now().Truncate(time.Microsecond),
	}
	s.NoError(s.paymentRepo.CreateTX(context.Background(), tx, pay))

	order1 := &domain.Order{
		ID: uuid.New(), PaymentID: pay.ID, MerchantID: m.ID, UserID: u.ID,
		Status: domain.OrderStatusPending, TotalAmount: decimal.NewFromInt(25),
		ShippingRecipientName: "John",
		CreatedAt: time.Now().Truncate(time.Microsecond),
		UpdatedAt: time.Now().Truncate(time.Microsecond),
	}
	order2 := &domain.Order{
		ID: uuid.New(), PaymentID: pay.ID, MerchantID: m.ID, UserID: u.ID,
		Status: domain.OrderStatusPending, TotalAmount: decimal.NewFromInt(25),
		ShippingRecipientName: "Jane",
		CreatedAt: time.Now().Truncate(time.Microsecond),
		UpdatedAt: time.Now().Truncate(time.Microsecond),
	}
	s.NoError(s.repo.CreateOrderTX(context.Background(), tx, order1))
	s.NoError(s.repo.CreateOrderTX(context.Background(), tx, order2))

	item1 := &domain.OrderItem{
		ID: uuid.New(), OrderID: order1.ID, ProductID: p.ID,
		Quantity: 1, Price: decimal.NewFromInt(25),
		CreatedAt: time.Now().Truncate(time.Microsecond),
	}
	item2 := &domain.OrderItem{
		ID: uuid.New(), OrderID: order2.ID, ProductID: p.ID,
		Quantity: 2, Price: decimal.NewFromInt(25),
		CreatedAt: time.Now().Truncate(time.Microsecond),
	}
	s.NoError(s.repo.CreateOrderItemTX(context.Background(), tx, item1))
	s.NoError(s.repo.CreateOrderItemTX(context.Background(), tx, item2))
	s.NoError(tx.Commit())

	// Act: batch fetch within a transaction
	tx2, err := s.repo.Begin(context.Background())
	s.NoError(err)
	defer tx2.Rollback()

	items, err := s.repo.GetOrderItemsByOrderIDsTX(context.Background(), tx2, []uuid.UUID{order1.ID, order2.ID})
	s.NoError(err)
	s.Len(items, 2)
	s.NoError(tx2.Commit())

	// Edge case: empty slice returns empty result, no error
	tx3, err := s.repo.Begin(context.Background())
	s.NoError(err)
	defer tx3.Rollback()

	emptyItems, err := s.repo.GetOrderItemsByOrderIDsTX(context.Background(), tx3, []uuid.UUID{})
	s.NoError(err)
	s.Empty(emptyItems)
}

func (s *OrderRepoSuite) TestCreateOrderItemsBatchTX() {
	u := &domain.User{
		ID: uuid.New(), FullName: "Batch Items User", Username: "batchitemsuser",
		Email: "batchitems@example.com", AuthProvider: domain.AuthProviderLocal,
		Role:  domain.RoleUser,
		CreatedAt: time.Now().Truncate(time.Microsecond),
	}
	s.NoError(s.userRepo.CreateUser(context.Background(), u))

	m := &domain.Merchant{
		ID: uuid.New(), UserID: u.ID, Name: "Batch Items Shop",
		CreatedAt: time.Now().Truncate(time.Microsecond),
	}
	s.NoError(s.merchantRepo.Create(context.Background(), m))

	p := &domain.Product{
		ID: uuid.New(), StoreID: m.ID, Name: "Batch Items Product",
		Price: decimal.NewFromInt(50), Stock: 100,
		CreatedAt: time.Now().Truncate(time.Microsecond),
	}
	s.NoError(s.productRepo.Create(context.Background(), p))

	tx, err := s.repo.Begin(context.Background())
	s.NoError(err)
	defer tx.Rollback()

	pay := &domain.Payment{
		ID: uuid.New(), UserID: u.ID, Amount: decimal.NewFromInt(200),
		Type: domain.PaymentTypeOrder, Method: domain.PaymentMethodWallet,
		Status:    domain.PaymentStatusPending,
		CreatedAt: time.Now().Truncate(time.Microsecond),
		UpdatedAt: time.Now().Truncate(time.Microsecond),
	}
	s.NoError(s.paymentRepo.CreateTX(context.Background(), tx, pay))

	o := &domain.Order{
		ID: uuid.New(), PaymentID: pay.ID, MerchantID: m.ID, UserID: u.ID,
		Status: domain.OrderStatusPending, TotalAmount: decimal.NewFromInt(200),
		ShippingRecipientName: "Batch Test",
		CreatedAt:             time.Now().Truncate(time.Microsecond),
		UpdatedAt:             time.Now().Truncate(time.Microsecond),
	}
	s.NoError(s.repo.CreateOrderTX(context.Background(), tx, o))

	items := []domain.OrderItem{
		{ID: uuid.New(), OrderID: o.ID, ProductID: p.ID, Quantity: 1, Price: decimal.NewFromInt(50), CreatedAt: time.Now().Truncate(time.Microsecond)},
		{ID: uuid.New(), OrderID: o.ID, ProductID: p.ID, Quantity: 3, Price: decimal.NewFromInt(50), CreatedAt: time.Now().Truncate(time.Microsecond)},
	}

	err = s.repo.CreateOrderItemsBatchTX(context.Background(), tx, items)
	s.NoError(err)
	s.NoError(tx.Commit())

	dbItems, err := s.repo.GetOrderItems(context.Background(), o.ID)
	s.NoError(err)
	s.Len(dbItems, 2)

	// Edge case: empty slice is a no-op
	tx2, _ := s.repo.Begin(context.Background())
	defer tx2.Rollback()
	err = s.repo.CreateOrderItemsBatchTX(context.Background(), tx2, []domain.OrderItem{})
	s.NoError(err)
}

func TestOrderRepoSuite(t *testing.T) {
	suite.Run(t, new(OrderRepoSuite))
}
