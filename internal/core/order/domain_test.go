package order

import (
	"go-marketplace/internal/domain"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestOrder_CanCancel(t *testing.T) {
	tests := []struct {
		name      string
		status    domain.OrderStatus
		createdAt time.Time
		want      bool
	}{
		{"within 1 hour processing", domain.OrderStatusProcessing, time.Now().Add(-30 * time.Minute), true},
		{"within 1 hour packaging", domain.OrderStatusPackaging, time.Now().Add(-30 * time.Minute), true},
		{"after 1 hour processing", domain.OrderStatusProcessing, time.Now().Add(-70 * time.Minute), false},
		{"wrong status shipping", domain.OrderStatusShipping, time.Now().Add(-30 * time.Minute), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := NewOrder(&domain.Order{Status: tt.status, CreatedAt: tt.createdAt})
			assert.Equal(t, tt.want, o.CanCancel())
		})
	}
}

func TestOrder_CanAppeal(t *testing.T) {
	tests := []struct {
		name      string
		status    domain.OrderStatus
		createdAt time.Time
		want      bool
	}{
		{"after 1 hour packaging", domain.OrderStatusPackaging, time.Now().Add(-70 * time.Minute), true},
		{"within 1 hour packaging", domain.OrderStatusPackaging, time.Now().Add(-30 * time.Minute), false},
		{"wrong status processing", domain.OrderStatusProcessing, time.Now().Add(-70 * time.Minute), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := NewOrder(&domain.Order{Status: tt.status, CreatedAt: tt.createdAt})
			assert.Equal(t, tt.want, o.CanAppeal())
		})
	}
}

func TestOrder_ValidateStatusTransition(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		current   domain.OrderStatus
		newStatus domain.OrderStatus
		createdAt time.Time
		wantErr   bool
	}{
		{"processing to packaging", domain.OrderStatusProcessing, domain.OrderStatusPackaging, now, false},
		{"packaging to shipping after 1 hour", domain.OrderStatusPackaging, domain.OrderStatusShipping, now.Add(-70 * time.Minute), false},
		{"packaging to shipping before 1 hour", domain.OrderStatusPackaging, domain.OrderStatusShipping, now.Add(-30 * time.Minute), true},
		{"shipping to delivered", domain.OrderStatusShipping, domain.OrderStatusDelivered, now, false},
		{"processing to shipping", domain.OrderStatusProcessing, domain.OrderStatusShipping, now, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := NewOrder(&domain.Order{Status: tt.current, CreatedAt: tt.createdAt})
			err := o.ValidateStatusTransition(tt.newStatus)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOrder_Authorization(t *testing.T) {
	userID := uuid.New()
	merchantID := uuid.New()

	o := NewOrder(&domain.Order{UserID: userID, MerchantID: merchantID})

	assert.True(t, o.IsAuthorized(userID))
	assert.False(t, o.IsAuthorized(uuid.New()))

	assert.True(t, o.IsMerchantAuthorized(merchantID))
	assert.False(t, o.IsMerchantAuthorized(uuid.New()))
}
