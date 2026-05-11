package cart

import (
	"context"
	"time"

	"go-marketplace/internal/core/product"
	"go-marketplace/internal/domain"

	"github.com/google/uuid"
)

type CartService interface {
	AddToCart(ctx context.Context, userID uuid.UUID, req AddToCartRequest) error
	UpdateCartItem(ctx context.Context, userID, productID uuid.UUID, quantity int) error
	RemoveFromCart(ctx context.Context, userID, productID uuid.UUID) error
	GetCart(ctx context.Context, userID uuid.UUID) (*CartResponse, error)
	ClearCart(ctx context.Context, userID uuid.UUID) error
}

type cartService struct {
	repo        CartRepository
	productRepo product.ProductRepository
}

func NewCartService(repo CartRepository, productRepo product.ProductRepository) CartService {
	return &cartService{
		repo:        repo,
		productRepo: productRepo,
	}
}

func (s *cartService) AddToCart(ctx context.Context, userID uuid.UUID, req AddToCartRequest) error {
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

func (s *cartService) UpdateCartItem(ctx context.Context, userID, productID uuid.UUID, quantity int) error {
	return s.repo.UpdateCartItem(ctx, userID, productID, quantity)
}

func (s *cartService) RemoveFromCart(ctx context.Context, userID, productID uuid.UUID) error {
	return s.repo.DeleteCartItem(ctx, userID, productID)
}

func (s *cartService) ClearCart(ctx context.Context, userID uuid.UUID) error {
	return s.repo.ClearCart(ctx, userID)
}

func (s *cartService) GetCart(ctx context.Context, userID uuid.UUID) (*CartResponse, error) {
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
