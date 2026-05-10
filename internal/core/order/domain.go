package order

import (
	"go-marketplace/internal/domain"
	"time"

	"github.com/google/uuid"
)

// Order represents a rich domain entity for an order
type Order struct {
	model *domain.Order
}

// NewOrder creates a new rich Order entity
func NewOrder(m *domain.Order) *Order {
	return &Order{model: m}
}

// IsAuthorized returns true if the user is the owner of the order
func (o *Order) IsAuthorized(userID uuid.UUID) bool {
	return o.model.UserID == userID
}

// IsMerchantAuthorized returns true if the merchant is the owner of the order
func (o *Order) IsMerchantAuthorized(merchantID uuid.UUID) bool {
	return o.model.MerchantID == merchantID
}

// CanCancel returns true if the user can cancel the order
// Permitted ONLY if status is Processing or Packaging AND time_elapsed < 1 hour
func (o *Order) CanCancel() bool {
	if o.model.Status != domain.OrderStatusProcessing && o.model.Status != domain.OrderStatusPackaging {
		return false
	}
	return time.Since(o.model.CreatedAt) <= time.Hour
}

// CanMerchantCancel returns true if the merchant can cancel the order
// Permitted anytime before Shipping
func (o *Order) CanMerchantCancel() bool {
	switch o.model.Status {
	case domain.OrderStatusProcessing, domain.OrderStatusPackaging:
		return true
	default:
		return false
	}
}

// CanAppeal returns true if the user can appeal the order
// Permitted ONLY if status is Packaging AND time_elapsed > 1 hour
func (o *Order) CanAppeal() bool {
	if o.model.Status != domain.OrderStatusPackaging {
		return false
	}
	return time.Since(o.model.CreatedAt) > time.Hour
}

// ValidateStatusTransition checks if a merchant can move an order to the new status
func (o *Order) ValidateStatusTransition(newStatus domain.OrderStatus) error {
	switch newStatus {
	case domain.OrderStatusPackaging:
		if o.model.Status != domain.OrderStatusProcessing {
			return domain.ErrInvalidStatusTransition
		}
	case domain.OrderStatusShipping:
		// Rule: NOT Permitted if time_elapsed < 1 hour since creation/processing
		if time.Since(o.model.CreatedAt) < time.Hour {
			return domain.ErrMerchantShipmentTooEarly
		}
		if o.model.Status != domain.OrderStatusPackaging {
			return domain.ErrInvalidStatusTransition
		}
	case domain.OrderStatusDelivered:
		if o.model.Status != domain.OrderStatusShipping {
			return domain.ErrInvalidStatusTransition
		}
	default:
		return domain.ErrInvalidStatusTransition
	}
	return nil
}
