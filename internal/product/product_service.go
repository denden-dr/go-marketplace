package product

import (
	"context"
	"time"

	"go-shop-yourself/internal/domain"

	"go-shop-yourself/internal/merchant"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ProductServiceInterface interface {
	CreateProduct(ctx context.Context, req ProductCreateRequest) (*ProductResponse, error)
	UpdateProduct(ctx context.Context, id uuid.UUID, req ProductUpdateRequest) (*ProductResponse, error)
}

type ProductRepository interface {
	Create(ctx context.Context, p *domain.Product) error
	Update(ctx context.Context, p *domain.Product) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Product, error)
	GetByIDForUpdateTX(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*domain.Product, error)
	UpdateStockTX(ctx context.Context, tx pgx.Tx, id uuid.UUID, stock int) error
}

type ProductService struct {
	repo ProductRepository
	merchantRepo merchant.MerchantRepository
}

func NewProductService(repo ProductRepository, merchantRepo merchant.MerchantRepository) *ProductService {
	return &ProductService{repo: repo, merchantRepo: merchantRepo}
}

func (s *ProductService) CreateProduct(ctx context.Context, req ProductCreateRequest) (*ProductResponse, error) {
	// Verify merchant exists
	merchant, err := s.merchantRepo.GetByID(ctx, req.StoreID)
	if err != nil {
		return nil, err
	}
	if merchant == nil {
		return nil, domain.ErrMerchantNotFound
	}

	product := &domain.Product{
		ID: uuid.New(),
		StoreID: req.StoreID,
		Name: req.Name,
		Description: req.Description,
		Price: req.Price,
		Stock: req.Stock,
		HeightCM: req.HeightCM,
		WidthCM: req.WidthCM,
		DepthCM: req.DepthCM,
		WeightKG: req.WeightKG,
		IsOnSale: req.IsOnSale,
		CreatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, product); err != nil {
		return nil, err
	}

	return &ProductResponse{
		ID: product.ID,
		Name: product.Name,
		Description: product.Description,
		Price: product.Price,
		Stock: product.Stock,
		IsOnSale: product.IsOnSale,
		CreatedAt: product.CreatedAt,
	}, nil
}

func (s *ProductService) UpdateProduct(ctx context.Context, id uuid.UUID, req ProductUpdateRequest) (*ProductResponse, error) {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, domain.ErrProductNotFound
	}

	product.Name = req.Name
	product.Description = req.Description
	product.Price = req.Price
	product.Stock = req.Stock
	product.IsOnSale = req.IsOnSale

	if err := s.repo.Update(ctx, product); err != nil {
		return nil, err
	}

	return &ProductResponse{
		ID: product.ID,
		Name: product.Name,
		Description: product.Description,
		Price: product.Price,
		Stock: product.Stock,
		IsOnSale: product.IsOnSale,
		CreatedAt: product.CreatedAt,
	}, nil
}
