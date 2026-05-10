package cart

import (
	"context"
	"time"

	"go-marketplace/internal/core/product"
	"go-marketplace/internal/domain"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
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
	models, err := s.repo.GetCartByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	cart := NewCart(models)

	res := CartResponse{
		Items:      make([]CartItemResponse, 0, len(cart.Items())),
		TotalPrice: cart.TotalPrice(),
	}

	for _, item := range cart.Items() {
		m := item.model
		res.Items = append(res.Items, CartItemResponse{
			ID:        m.ID,
			ProductID: m.ProductID,
			Name:      m.Product.Name,
			Price:     m.Product.Price,
			Quantity:  m.Quantity,
			Subtotal:  item.Subtotal(),
		})
	}

	return &res, nil
}
