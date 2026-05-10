package payment

import (
	"context"
	"go-marketplace/internal/domain"

	"github.com/google/uuid"
)

type MockProvider struct{}

func NewMockProvider() *MockProvider {
	return &MockProvider{}
}

func (m *MockProvider) CreateTransaction(ctx context.Context, p *domain.Payment) (string, error) {
	// Simulate external token generation
	return "snap-token-" + uuid.New().String()[:8], nil
}
