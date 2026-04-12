package domain

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

type Pool interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

type UserRepository interface {
	CreateUser(ctx context.Context, u *User) error
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*User, error)
}

type MerchantRepository interface {
	Create(ctx context.Context, m *Merchant) error
	GetByID(ctx context.Context, id uuid.UUID) (*Merchant, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (*Merchant, error)
	CreateTx(ctx context.Context, tx pgx.Tx, m *Merchant) error
	GetPool() Pool // Keep for now as per refactor.md step 1.1
}

type ProductRepository interface {
	Create(ctx context.Context, p *Product) error
	Update(ctx context.Context, p *Product) error
	GetByID(ctx context.Context, id uuid.UUID) (*Product, error)
}

type WalletRepository interface {
	GetWalletByUserID(ctx context.Context, userID uuid.UUID) (*Wallet, error)
	GetWalletHistory(ctx context.Context, walletID uuid.UUID, limit, offset int) ([]WalletTransaction, error)
	Withdraw(ctx context.Context, walletID uuid.UUID, amount decimal.Decimal, txData WalletTransaction) error
	Create(ctx context.Context, w *Wallet) error
	CreateTx(ctx context.Context, tx pgx.Tx, w *Wallet) error
	GetPool() Pool // Keep for now as per refactor.md step 1.1
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, rt *RefreshToken) error
	GetByTokenHash(ctx context.Context, hash string) (*RefreshToken, error)
	RevokeByID(ctx context.Context, id uuid.UUID) error
	RevokeAllByFamilyID(ctx context.Context, familyID uuid.UUID) error
}
