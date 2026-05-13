package product

import (
	"go-marketplace/internal/domain"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestProductCreateRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request ProductCreateRequest
		wantErr bool
		errKeys []string
	}{
		{
			name: "valid request",
			request: ProductCreateRequest{
				StoreID: uuid.New(),
				Name:    "Test Product",
				Price:   decimal.NewFromFloat(100.0),
				Stock:   10,
			},
			wantErr: false,
		},
		{
			name:    "empty request",
			request: ProductCreateRequest{},
			wantErr: true,
			errKeys: []string{"name", "price", "store_id"},
		},
		{
			name: "negative price and stock",
			request: ProductCreateRequest{
				StoreID: uuid.New(),
				Name:    "Test Product",
				Price:   decimal.NewFromFloat(-10.0),
				Stock:   -5,
			},
			wantErr: true,
			errKeys: []string{"price", "stock"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				vErrs, ok := err.(domain.ValidationErrors)
				assert.True(t, ok, "should be ValidationErrors type")
				for _, key := range tt.errKeys {
					assert.Contains(t, vErrs, key)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
