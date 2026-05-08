package cart

import (
	"context"
	"time"

	"go-marketplace/internal/core/product"
	"go-marketplace/internal/domain"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
)

type CartServiceInterface interface {
	AddToCart(ctx context.Context, userID uuid.UUID, req AddToCartRequest) error
	UpdateCartItem(ctx context.Context, userID, productID uuid.UUID, quantity int) error
	RemoveFromCart(ctx context.Context, userID, productID uuid.UUID) error
	GetCart(ctx context.Context, userID uuid.UUID) (*CartResponse, error)
	ClearCart(ctx context.Context, userID uuid.UUID) error
}

type CartRepository interface {
	UpsertCartItem(ctx context.Context, item *domain.CartItem) error
	UpdateCartItem(ctx context.Context, userID, productID uuid.UUID, quantity int) error
	DeleteCartItem(ctx context.Context, userID, productID uuid.UUID) error
	ClearCart(ctx context.Context, userID uuid.UUID) error
	ClearCartTX(ctx context.Context, tx *sqlx.Tx, userID uuid.UUID) error
	GetCartByUserID(ctx context.Context, userID uuid.UUID) ([]domain.CartItem, error)
}

type CartService struct {
	repo        CartRepository
	productRepo product.ProductRepository
}

func NewCartService(repo CartRepository, productRepo product.ProductRepository) *CartService {
	return &CartService{
		repo:        repo,
		productRepo: productRepo,
	}
}

func (s *CartService) AddToCart(ctx context.Context, userID uuid.UUID, req AddToCartRequest) error {
	// Check if product exists
	p, err := s.productRepo.GetByID(ctx, req.ProductID)
	if err != nil {
		return err
	}
	if p == nil {
		return domain.ErrProductNotFound
	}

	item := &domain.CartItem{
		ID:        uuid.New(),
		UserID:    userID,
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return s.repo.UpsertCartItem(ctx, item)
}

func (s *CartService) UpdateCartItem(ctx context.Context, userID, productID uuid.UUID, quantity int) error {
	return s.repo.UpdateCartItem(ctx, userID, productID, quantity)
}

func (s *CartService) RemoveFromCart(ctx context.Context, userID, productID uuid.UUID) error {
	return s.repo.DeleteCartItem(ctx, userID, productID)
}

func (s *CartService) ClearCart(ctx context.Context, userID uuid.UUID) error {
	return s.repo.ClearCart(ctx, userID)
}

func (s *CartService) GetCart(ctx context.Context, userID uuid.UUID) (*CartResponse, error) {
	items, err := s.repo.GetCartByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var res CartResponse
	res.Items = make([]CartItemResponse, 0)
	total := decimal.NewFromInt(0)

	for _, item := range items {
		subtotal := item.Product.Price.Mul(decimal.NewFromInt(int64(item.Quantity)))
		total = total.Add(subtotal)

		res.Items = append(res.Items, CartItemResponse{
			ID:        item.ID,
			ProductID: item.ProductID,
			Name:      item.Product.Name,
			Price:     item.Product.Price,
			Quantity:  item.Quantity,
			Subtotal:  subtotal,
		})
	}

	res.TotalPrice = total
	return &res, nil
}
