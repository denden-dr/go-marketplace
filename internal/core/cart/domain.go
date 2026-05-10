package cart

import (
	"go-marketplace/internal/domain"

	"github.com/shopspring/decimal"
)

// CartItem represents a rich domain entity for a cart item
type CartItem struct {
	model *domain.CartItem
}

// NewCartItem creates a new rich CartItem entity
func NewCartItem(m *domain.CartItem) *CartItem {
	return &CartItem{model: m}
}

// Subtotal returns the subtotal for this cart item
func (i *CartItem) Subtotal() decimal.Decimal {
	if i.model.Product == nil {
		return decimal.Zero
	}
	return i.model.Product.Price.Mul(decimal.NewFromInt(int64(i.model.Quantity)))
}

// Cart represents a rich domain entity for a collection of cart items
type Cart struct {
	items []*CartItem
}

// NewCart creates a new rich Cart entity from a slice of persistence models
func NewCart(models []domain.CartItem) *Cart {
	items := make([]*CartItem, len(models))
	for i := range models {
		items[i] = NewCartItem(&models[i])
	}
	return &Cart{items: items}
}

// TotalPrice returns the sum of all item subtotals in the cart
func (c *Cart) TotalPrice() decimal.Decimal {
	total := decimal.Zero
	for _, item := range c.items {
		total = total.Add(item.Subtotal())
	}
	return total
}

// Items returns the rich items in the cart
func (c *Cart) Items() []*CartItem {
	return c.items
}
