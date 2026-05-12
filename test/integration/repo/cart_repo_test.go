//go:build integration

package repo

import (
	"context"
	"testing"
	"time"

	"go-marketplace/internal/core/cart"
	"go-marketplace/internal/core/merchant"
	"go-marketplace/internal/core/product"
	"go-marketplace/internal/core/user"
	"go-marketplace/internal/domain"
	"go-marketplace/internal/testutil"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
)

type CartRepoSuite struct {
	testutil.IntegrationSuite
	repo         cart.CartRepository
	productRepo  product.ProductRepository
	merchantRepo merchant.MerchantRepository
	userRepo     user.UserRepository
}

func (s *CartRepoSuite) SetupSuite() {
	s.IntegrationSuite.SetupSuite()
	s.repo = cart.NewCartRepository(s.DB)
	s.productRepo = product.NewProductRepository(s.DB)
	s.merchantRepo = merchant.NewMerchantRepository(s.DB)
	s.userRepo = user.NewUserRepository(s.DB)
}

func (s *CartRepoSuite) TestCartOperations() {
	// Setup User, Merchant, Product
	u := &domain.User{
		ID:           uuid.New(),
		FullName:     "Cart User",
		Username:     "cartuser",
		Email:        "cart@example.com",
		AuthProvider: domain.AuthProviderLocal,
		CreatedAt:    time.Now().Truncate(time.Microsecond),
	}
	s.NoError(s.userRepo.CreateUser(context.Background(), u))

	m := &domain.Merchant{
		ID:        uuid.New(),
		UserID:    u.ID,
		Name:      "Cart Shop",
		CreatedAt: time.Now().Truncate(time.Microsecond),
	}
	s.NoError(s.merchantRepo.Create(context.Background(), m))

	p := &domain.Product{
		ID:        uuid.New(),
		StoreID:   m.ID,
		Name:      "Cart Product",
		Price:     decimal.NewFromInt(50),
		Stock:     100,
		CreatedAt: time.Now().Truncate(time.Microsecond),
	}
	s.NoError(s.productRepo.Create(context.Background(), p))

	item := &domain.CartItem{
		ID:        uuid.New(),
		UserID:    u.ID,
		ProductID: p.ID,
		Quantity:  2,
		CreatedAt: time.Now().Truncate(time.Microsecond),
		UpdatedAt: time.Now().Truncate(time.Microsecond),
	}

	// 1. Upsert
	err := s.repo.UpsertCartItem(context.Background(), item)
	s.NoError(err)

	// 2. GetCartByUserID
	items, err := s.repo.GetCartByUserID(context.Background(), u.ID)
	s.NoError(err)
	s.Len(items, 1)
	s.Equal(item.ID, items[0].ID)
	s.Equal(2, items[0].Quantity)
	s.NotNil(items[0].Product)
	s.Equal(p.Name, items[0].Product.Name)

	// 3. Update (Upsert again) - ON CONFLICT behavior
	item.Quantity = 3
	err = s.repo.UpsertCartItem(context.Background(), item)
	s.NoError(err)

	items, _ = s.repo.GetCartByUserID(context.Background(), u.ID)
	s.Equal(5, items[0].Quantity) // 2 + 3

	// 4. UpdateCartItem (Set absolute)
	err = s.repo.UpdateCartItem(context.Background(), u.ID, p.ID, 10)
	s.NoError(err)

	items, _ = s.repo.GetCartByUserID(context.Background(), u.ID)
	s.Equal(10, items[0].Quantity)

	// 5. DeleteCartItem
	err = s.repo.DeleteCartItem(context.Background(), u.ID, p.ID)
	s.NoError(err)

	items, _ = s.repo.GetCartByUserID(context.Background(), u.ID)
	s.Len(items, 0)

	// 6. ClearCart
	s.repo.UpsertCartItem(context.Background(), item)
	err = s.repo.ClearCart(context.Background(), u.ID)
	s.NoError(err)

	items, _ = s.repo.GetCartByUserID(context.Background(), u.ID)
	s.Len(items, 0)

	// 7. ClearCartTX
	s.repo.UpsertCartItem(context.Background(), item)
	tx, err := s.DB.BeginTxx(context.Background(), nil)
	s.NoError(err)
	defer tx.Rollback()

	err = s.repo.ClearCartTX(context.Background(), tx, u.ID)
	s.NoError(err)

	s.NoError(tx.Commit())

	items, _ = s.repo.GetCartByUserID(context.Background(), u.ID)
	s.Len(items, 0)
}

func TestCartRepoSuite(t *testing.T) {
	suite.Run(t, new(CartRepoSuite))
}
