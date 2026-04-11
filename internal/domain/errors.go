package domain

import "errors"

var (
	ErrUserNotFound           = errors.New("user not found")
	ErrUserAlreadyExists      = errors.New("user already exists")
	ErrInvalidCredentials     = errors.New("invalid email or password")
	ErrMerchantNotFound       = errors.New("merchant not found")
	ErrMerchantAlreadyExists  = errors.New("merchant already exists for this user")
	ErrProductNotFound        = errors.New("product not found")
	ErrWalletNotFound         = errors.New("wallet not found")
	ErrWalletAlreadyExists    = errors.New("wallet already exists for this user")
	ErrWalletNotActive        = errors.New("wallet is not active")
	ErrInsufficientBalance    = errors.New("insufficient balance")
	ErrInvalidRefreshToken    = errors.New("invalid refresh token")
	ErrRefreshTokenExpired    = errors.New("refresh token expired")
	ErrRefreshTokenReused     = errors.New("token reuse detected - all tokens in family revoked")
)
