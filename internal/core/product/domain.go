package product

import (
	"go-marketplace/internal/domain"

	"github.com/google/uuid"
)

// Product represents a rich domain entity for a product
type Product struct {
	model *domain.Product
}

// NewProduct creates a new rich Product entity
func NewProduct(m *domain.Product) *Product {
	return &Product{model: m}
}

// HasStock returns true if the product has at least the requested quantity in stock
func (p *Product) HasStock(quantity int) bool {
	return p.model.Stock >= quantity
}

// UpdateStock updates the product stock and returns an error if it would fall below zero
func (p *Product) UpdateStock(quantity int) error {
	newStock := p.model.Stock + quantity
	if newStock < 0 {
		return domain.ErrInsufficientStock
	}
	p.model.Stock = newStock
	return nil
}

// IsOwnedBy returns true if the product belongs to the given merchant
func (p *Product) IsOwnedBy(merchantID uuid.UUID) bool {
	return p.model.StoreID == merchantID
}

// Update updates the product fields from the provided request
func (p *Product) Update(req ProductUpdateRequest) {
	p.model.Name = req.Name
	p.model.Description = req.Description
	p.model.Price = req.Price
	p.model.Stock = req.Stock
	p.model.IsOnSale = req.IsOnSale
}
