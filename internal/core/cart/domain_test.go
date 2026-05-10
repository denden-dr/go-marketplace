package cart

import (
	"go-marketplace/internal/domain"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestCartItem_Subtotal(t *testing.T) {
	price := decimal.NewFromInt(100)
	quantity := 3
	item := NewCartItem(&domain.CartItem{
		Quantity: quantity,
		Product: &domain.Product{
			Price: price,
		},
	})

	assert.True(t, item.Subtotal().Equal(decimal.NewFromInt(300)))
}

func TestCart_TotalPrice(t *testing.T) {
	models := []domain.CartItem{
		{
			Quantity: 2,
			Product: &domain.Product{
				Price: decimal.NewFromInt(100),
			},
		},
		{
			Quantity: 1,
			Product: &domain.Product{
				Price: decimal.NewFromInt(50),
			},
		},
	}

	cart := NewCart(models)
	assert.True(t, cart.TotalPrice().Equal(decimal.NewFromInt(250)))
}
