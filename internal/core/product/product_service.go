package product

import (
	"context"
	"time"

	"go-marketplace/internal/domain"

	"go-marketplace/internal/core/merchant"

	"github.com/google/uuid"
)

type ProductService interface {
	CreateProduct(ctx context.Context, userID uuid.UUID, req ProductCreateRequest) (*ProductResponse, error)
	UpdateProduct(ctx context.Context, userID uuid.UUID, id uuid.UUID, req ProductUpdateRequest) (*ProductResponse, error)
	SearchProducts(ctx context.Context, req ProductSearchRequest) ([]ProductResponse, error)
}

type productService struct {
	repo         ProductRepository
	merchantRepo merchant.MerchantRepository
}

func NewProductService(repo ProductRepository, merchantRepo merchant.MerchantRepository) ProductService {
	return &productService{repo: repo, merchantRepo: merchantRepo}
}

func (s *productService) CreateProduct(ctx context.Context, userID uuid.UUID, req ProductCreateRequest) (*ProductResponse, error) {
	// Verify merchant exists and belongs to the user
	merchant, err := s.merchantRepo.GetByID(ctx, req.StoreID)
	if err != nil {
		return nil, err
	}
	if merchant == nil {
		return nil, domain.ErrMerchantNotFound
	}

	if merchant.UserID != userID {
		return nil, domain.ErrForbidden
	}

	product := &domain.Product{
		ID:          uuid.New(),
		StoreID:     req.StoreID,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		HeightCM:    req.HeightCM,
		WidthCM:     req.WidthCM,
		DepthCM:     req.DepthCM,
		WeightKG:    req.WeightKG,
		IsOnSale:    req.IsOnSale,
		CreatedAt:   time.Now(),
	}

	if err := s.repo.Create(ctx, product); err != nil {
		return nil, err
	}

	return &ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
		IsOnSale:    product.IsOnSale,
		CreatedAt:   product.CreatedAt,
	}, nil
}

func (s *productService) UpdateProduct(ctx context.Context, userID uuid.UUID, id uuid.UUID, req ProductUpdateRequest) (*ProductResponse, error) {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, domain.ErrProductNotFound
	}

	// Verify the user owns the merchant that owns the product
	merch, err := s.merchantRepo.GetByID(ctx, m.StoreID)
	if err != nil {
		return nil, err
	}
	if merch == nil || merch.UserID != userID {
		return nil, domain.ErrForbidden
	}

	product := NewProduct(m)
	product.Update(req)

	if err := s.repo.Update(ctx, m); err != nil {
		return nil, err
	}

	return &ProductResponse{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		Price:       m.Price,
		Stock:       m.Stock,
		IsOnSale:    m.IsOnSale,
		CreatedAt:   m.CreatedAt,
	}, nil
}

func (s *productService) SearchProducts(ctx context.Context, req ProductSearchRequest) ([]ProductResponse, error) {
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	offset := (req.Page - 1) * req.Limit

	products, err := s.repo.Search(ctx, req.Query, req.Limit, offset)
	if err != nil {
		return nil, err
	}

	responses := make([]ProductResponse, 0, len(products))
	for _, p := range products {
		responses = append(responses, ProductResponse{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Stock:       p.Stock,
			IsOnSale:    p.IsOnSale,
			CreatedAt:   p.CreatedAt,
		})
	}

	return responses, nil
}
