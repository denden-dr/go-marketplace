package product

import (
	"go-marketplace/internal/domain"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestProduct_HasStock(t *testing.T) {
	tests := []struct {
		name     string
		stock    int
		quantity int
		want     bool
	}{
		{"sufficient stock", 10, 5, true},
		{"exact stock", 10, 10, true},
		{"insufficient stock", 10, 11, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProduct(&domain.Product{Stock: tt.stock})
			assert.Equal(t, tt.want, p.HasStock(tt.quantity))
		})
	}
}

func TestProduct_UpdateStock(t *testing.T) {
	tests := []struct {
		name     string
		initial  int
		change   int
		wantStock int
		wantErr  error
	}{
		{"add stock", 10, 5, 15, nil},
		{"deduct sufficient", 10, -5, 5, nil},
		{"deduct exact", 10, -10, 0, nil},
		{"deduct insufficient", 10, -11, 10, domain.ErrInsufficientStock},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &domain.Product{Stock: tt.initial}
			p := NewProduct(m)
			err := p.UpdateStock(tt.change)
			assert.Equal(t, tt.wantErr, err)
			if err == nil {
				assert.Equal(t, tt.wantStock, m.Stock)
			}
		})
	}
}

func TestProduct_IsOwnedBy(t *testing.T) {
	merchantID := uuid.New()
	otherID := uuid.New()

	p := NewProduct(&domain.Product{StoreID: merchantID})
	assert.True(t, p.IsOwnedBy(merchantID))
	assert.False(t, p.IsOwnedBy(otherID))
}

func TestProduct_Update(t *testing.T) {
	m := &domain.Product{
		Name: "Old Name",
	}
	p := NewProduct(m)

	description := "New Desc"
	req := ProductUpdateRequest{
		Name:        "New Name",
		Description: &description,
		Price:       decimal.NewFromInt(100),
		Stock:       20,
		IsOnSale:    true,
	}

	p.Update(req)

	assert.Equal(t, req.Name, m.Name)
	assert.Equal(t, req.Description, m.Description)
	assert.Equal(t, req.Price, m.Price)
	assert.Equal(t, req.Stock, m.Stock)
	assert.Equal(t, req.IsOnSale, m.IsOnSale)
}
