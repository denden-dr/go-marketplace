package services

import (
	"context"
	"errors"
	"time"

	"go-shop-yourself/internal/domain"
	"go-shop-yourself/internal/dtos"
	"go-shop-yourself/internal/repos"

	"github.com/google/uuid"
)

type ProductService struct {
	repo         *repos.ProductRepository
	merchantRepo *repos.MerchantRepository
}

func NewProductService(repo *repos.ProductRepository, merchantRepo *repos.MerchantRepository) *ProductService {
	return &ProductService{repo: repo, merchantRepo: merchantRepo}
}

func (s *ProductService) CreateProduct(ctx context.Context, req dtos.ProductCreateRequest) (*dtos.ProductResponse, error) {
	// Verify merchant exists
	merchant, err := s.merchantRepo.GetByID(ctx, req.StoreID)
	if err != nil {
		return nil, err
	}
	if merchant == nil {
		return nil, errors.New("merchant not found")
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

	return &dtos.ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
		IsOnSale:    product.IsOnSale,
		CreatedAt:   product.CreatedAt,
	}, nil
}

func (s *ProductService) UpdateProduct(ctx context.Context, id uuid.UUID, req dtos.ProductUpdateRequest) (*dtos.ProductResponse, error) {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, errors.New("product not found")
	}

	product.Name = req.Name
	product.Description = req.Description
	product.Price = req.Price
	product.Stock = req.Stock
	product.IsOnSale = req.IsOnSale

	if err := s.repo.Update(ctx, product); err != nil {
		return nil, err
	}

	return &dtos.ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
		IsOnSale:    product.IsOnSale,
		CreatedAt:   product.CreatedAt,
	}, nil
}
