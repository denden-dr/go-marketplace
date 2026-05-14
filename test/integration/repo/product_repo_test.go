//go:build integration

package repo

import (
	"context"
	"testing"
	"time"

	"go-marketplace/internal/common"
	"go-marketplace/internal/core/merchant"
	"go-marketplace/internal/core/product"
	"go-marketplace/internal/core/user"
	"go-marketplace/internal/domain"
	"go-marketplace/internal/testutil"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
)

type ProductRepoSuite struct {
	testutil.IntegrationSuite
	repo         product.ProductRepository
	merchantRepo merchant.MerchantRepository
	userRepo     user.UserRepository
}

func (s *ProductRepoSuite) SetupSuite() {
	s.IntegrationSuite.SetupSuite()
	s.repo = product.NewProductRepository(s.DB)
	s.merchantRepo = merchant.NewMerchantRepository(s.DB)
	s.userRepo = user.NewUserRepository(s.DB)
}

func (s *ProductRepoSuite) TestProductOperations() {
	// Setup User & Merchant
	u := &domain.User{
		ID:           uuid.New(),
		FullName:     "Merchant Owner",
		Username:     "owner_prod",
		Email:        "owner_prod@example.com",
		AuthProvider: domain.AuthProviderLocal,
		CreatedAt:    time.Now().Truncate(time.Microsecond),
	}
	s.NoError(s.userRepo.CreateUser(context.Background(), u))

	m := &domain.Merchant{
		ID:        uuid.New(),
		UserID:    u.ID,
		Name:      "Product Shop",
		About:     "Best shop in town",
		TaxID:     "123-456",
		CreatedAt: time.Now().Truncate(time.Microsecond),
	}
	s.NoError(s.merchantRepo.Create(context.Background(), m))

	p := &domain.Product{
		ID:          uuid.New(),
		StoreID:     m.ID,
		Name:        "Test Product",
		Description: common.Ptr("Best product ever"),
		Price:       decimal.NewFromInt(100),
		Stock:       10,
		HeightCM:    10.5,
		WidthCM:     20.0,
		DepthCM:     5.0,
		WeightKG:    0.5,
		IsOnSale:    true,
		CreatedAt:   time.Now().Truncate(time.Microsecond),
	}

	// 1. Create
	err := s.repo.Create(context.Background(), p)
	s.NoError(err)

	// 2. GetByID
	dbProduct, err := s.repo.GetByID(context.Background(), p.ID)
	s.NoError(err)
	s.NotNil(dbProduct)
	s.Equal(p.ID, dbProduct.ID)
	s.Equal(p.Name, dbProduct.Name)
	s.True(p.Price.Equal(dbProduct.Price))

	// 3. Update
	p.Name = "Updated Product"
	p.Price = decimal.NewFromInt(150)
	err = s.repo.Update(context.Background(), p)
	s.NoError(err)

	dbProduct, _ = s.repo.GetByID(context.Background(), p.ID)
	s.Equal("Updated Product", dbProduct.Name)
	s.True(decimal.NewFromInt(150).Equal(dbProduct.Price))

	// 4. TX Operations
	tx, err := s.DB.BeginTxx(context.Background(), nil)
	s.NoError(err)
	defer tx.Rollback()

	pTX, err := s.repo.GetByIDForUpdateTX(context.Background(), tx, p.ID)
	s.NoError(err)
	s.NotNil(pTX)

	err = s.repo.UpdateStockTX(context.Background(), tx, p.ID, 5)
	s.NoError(err)

	s.NoError(tx.Commit())

	dbProduct, _ = s.repo.GetByID(context.Background(), p.ID)
	s.Equal(5, dbProduct.Stock)

	// 5. RestoreStockBatchTX
	p2 := &domain.Product{
		ID:        uuid.New(),
		StoreID:   m.ID,
		Name:      "Second Product",
		Price:     decimal.NewFromInt(200),
		Stock:     20,
		CreatedAt: time.Now().Truncate(time.Microsecond),
	}
	s.NoError(s.repo.Create(context.Background(), p2))

	tx2, err := s.DB.BeginTxx(context.Background(), nil)
	s.NoError(err)
	defer tx2.Rollback()

	batchItems := []domain.OrderItem{
		{ProductID: p.ID, Quantity: 2},
		{ProductID: p2.ID, Quantity: 3},
		{ProductID: p.ID, Quantity: 1}, // Test duplicate productID in batch
	}

	err = s.repo.RestoreStockBatchTX(context.Background(), tx2, batchItems)
	s.NoError(err)
	s.NoError(tx2.Commit())

	// p stock: 5 + 2 + 1 = 8
	// p2 stock: 20 + 3 = 23
	dbP1, _ := s.repo.GetByID(context.Background(), p.ID)
	dbP2, _ := s.repo.GetByID(context.Background(), p2.ID)
	s.Equal(8, dbP1.Stock)
	s.Equal(23, dbP2.Stock)

	// 6. Search
	// Wait a tiny bit for search indexes if necessary, but usually synchronous in Postgres
	products, err := s.repo.Search(context.Background(), "Updated", 10, 0)
	s.NoError(err)
	s.Len(products, 1)
	s.Equal(p.ID, products[0].ID)

	products, err = s.repo.Search(context.Background(), "", 10, 0)
	s.NoError(err)
	s.GreaterOrEqual(len(products), 1)
}

func (s *ProductRepoSuite) TestGetByIDsForUpdateTX() {
	m := &domain.Merchant{
		ID: uuid.New(), UserID: uuid.New(), Name: "Batch Lock Shop",
		CreatedAt: time.Now().Truncate(time.Microsecond),
	}
	// Need a real user first
	u := &domain.User{
		ID: m.UserID, FullName: "Owner", Username: "owner_batch", Email: "batch@example.com",
		AuthProvider: domain.AuthProviderLocal, CreatedAt: time.Now().Truncate(time.Microsecond),
	}
	s.NoError(s.userRepo.CreateUser(context.Background(), u))
	s.NoError(s.merchantRepo.Create(context.Background(), m))

	p1 := &domain.Product{
		ID: uuid.New(), StoreID: m.ID, Name: "Product A",
		Price: decimal.NewFromInt(10), Stock: 5,
		CreatedAt: time.Now().Truncate(time.Microsecond),
	}
	p2 := &domain.Product{
		ID: uuid.New(), StoreID: m.ID, Name: "Product B",
		Price: decimal.NewFromInt(20), Stock: 10,
		CreatedAt: time.Now().Truncate(time.Microsecond),
	}
	s.NoError(s.repo.Create(context.Background(), p1))
	s.NoError(s.repo.Create(context.Background(), p2))

	tx, err := s.DB.BeginTxx(context.Background(), nil)
	s.NoError(err)
	defer tx.Rollback()

	products, err := s.repo.GetByIDsForUpdateTX(context.Background(), tx, []uuid.UUID{p1.ID, p2.ID})
	s.NoError(err)
	s.Len(products, 2)

	// Edge case: empty slice
	empty, err := s.repo.GetByIDsForUpdateTX(context.Background(), tx, []uuid.UUID{})
	s.NoError(err)
	s.Empty(empty)
}

func (s *ProductRepoSuite) TestDeductStockBatchTX() {
	u := &domain.User{
		ID: uuid.New(), FullName: "Owner Deduct", Username: "owner_deduct", Email: "deduct@example.com",
		AuthProvider: domain.AuthProviderLocal, CreatedAt: time.Now().Truncate(time.Microsecond),
	}
	s.NoError(s.userRepo.CreateUser(context.Background(), u))

	m := &domain.Merchant{
		ID: uuid.New(), UserID: u.ID, Name: "Stock Deduct Shop",
		CreatedAt: time.Now().Truncate(time.Microsecond),
	}
	s.NoError(s.merchantRepo.Create(context.Background(), m))

	p1 := &domain.Product{
		ID: uuid.New(), StoreID: m.ID, Name: "Deduct A",
		Price: decimal.NewFromInt(10), Stock: 20,
		CreatedAt: time.Now().Truncate(time.Microsecond),
	}
	p2 := &domain.Product{
		ID: uuid.New(), StoreID: m.ID, Name: "Deduct B",
		Price: decimal.NewFromInt(20), Stock: 30,
		CreatedAt: time.Now().Truncate(time.Microsecond),
	}
	s.NoError(s.repo.Create(context.Background(), p1))
	s.NoError(s.repo.Create(context.Background(), p2))

	items := []domain.OrderItem{
		{ProductID: p1.ID, Quantity: 3},
		{ProductID: p2.ID, Quantity: 5},
	}

	tx, err := s.DB.BeginTxx(context.Background(), nil)
	s.NoError(err)
	defer tx.Rollback()

	err = s.repo.DeductStockBatchTX(context.Background(), tx, items)
	s.NoError(err)
	s.NoError(tx.Commit())

	// Verify stock was deducted
	got1, _ := s.repo.GetByID(context.Background(), p1.ID)
	got2, _ := s.repo.GetByID(context.Background(), p2.ID)
	s.Equal(17, got1.Stock) // 20 - 3
	s.Equal(25, got2.Stock) // 30 - 5
}

func TestProductRepoSuite(t *testing.T) {
	suite.Run(t, new(ProductRepoSuite))
}
