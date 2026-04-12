package domain

import (
	"context"

	"go-shop-yourself/internal/dtos"

	"github.com/google/uuid"
)

type AuthServiceInterface interface {
	Register(ctx context.Context, email, password, username string) (uuid.UUID, error)
	Login(ctx context.Context, email, password string) (*dtos.AuthResponse, error)
	RefreshTokens(ctx context.Context, rawToken string) (*dtos.AuthResponse, error)
	Logout(ctx context.Context, rawToken string) error
}

type UserServiceInterface interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (*dtos.UserResponse, error)
}

type MerchantServiceInterface interface {
	RegisterMerchant(ctx context.Context, userID uuid.UUID, req dtos.MerchantRegisterRequest) (*dtos.MerchantResponse, error)
}

type ProductServiceInterface interface {
	CreateProduct(ctx context.Context, req dtos.ProductCreateRequest) (*dtos.ProductResponse, error)
	UpdateProduct(ctx context.Context, id uuid.UUID, req dtos.ProductUpdateRequest) (*dtos.ProductResponse, error)
}

type WalletServiceInterface interface {
	GetWalletByUserID(ctx context.Context, userID uuid.UUID) (*dtos.WalletResponse, error)
	GetWalletHistory(ctx context.Context, userID uuid.UUID, page, limit int) ([]dtos.TransactionResponse, error)
	Withdraw(ctx context.Context, userID uuid.UUID, req dtos.WithdrawRequest) error
	CreateWallet(ctx context.Context, userID uuid.UUID) (*Wallet, error)
}
