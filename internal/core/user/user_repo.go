package user

import (
	"context"
	"database/sql"

	"go-marketplace/internal/domain"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type userRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) CreateUser(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (id, full_name, username, email, password, auth_provider, provider_id, is_verified, created_at) 
	          VALUES (:id, :full_name, :username, :email, :password, :auth_provider, :provider_id, :is_verified, :created_at)`
	_, err := r.db.NamedExecContext(ctx, query, user)
	return err
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id, full_name, username, email, password, auth_provider, provider_id, is_verified, created_at FROM users WHERE email = $1`
	var user domain.User
	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `SELECT id, full_name, username, email, password, auth_provider, provider_id, is_verified, created_at FROM users WHERE id = $1`
	var user domain.User
	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetUserByProviderID(ctx context.Context, provider string, providerID string) (*domain.User, error) {
	query := `SELECT id, full_name, username, email, password, auth_provider, provider_id, is_verified, created_at 
	          FROM users WHERE auth_provider = $1 AND provider_id = $2`
	var user domain.User
	err := r.db.GetContext(ctx, &user, query, provider, providerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `SELECT id, full_name, username, email, password, auth_provider, provider_id, is_verified, created_at FROM users WHERE username = $1`
	var user domain.User
	err := r.db.GetContext(ctx, &user, query, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) UpdateVerifiedStatus(ctx context.Context, id uuid.UUID, status bool) error {
	query := `UPDATE users SET is_verified = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}

func (r *userRepository) CreateAddress(ctx context.Context, addr *domain.UserAddress) error {
	query := `INSERT INTO user_addresses (id, user_id, tag, recipient_name, phone_number, street_address, city, province, postal_code, is_default, created_at, updated_at)
	          VALUES (:id, :user_id, :tag, :recipient_name, :phone_number, :street_address, :city, :province, :postal_code, :is_default, :created_at, :updated_at)`
	_, err := r.db.NamedExecContext(ctx, query, addr)
	return err
}

func (r *userRepository) GetAddressesByUserID(ctx context.Context, userID uuid.UUID) ([]domain.UserAddress, error) {
	query := `SELECT id, user_id, tag, recipient_name, phone_number, street_address, city, province, postal_code, is_default, created_at, updated_at 
	          FROM user_addresses WHERE user_id = $1 ORDER BY created_at DESC`
	var addresses []domain.UserAddress
	err := r.db.SelectContext(ctx, &addresses, query, userID)
	if err != nil {
		return nil, err
	}
	return addresses, nil
}

func (r *userRepository) GetAddressByID(ctx context.Context, addressID uuid.UUID) (*domain.UserAddress, error) {
	query := `SELECT id, user_id, tag, recipient_name, phone_number, street_address, city, province, postal_code, is_default, created_at, updated_at 
	          FROM user_addresses WHERE id = $1`
	var addr domain.UserAddress
	err := r.db.GetContext(ctx, &addr, query, addressID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &addr, nil
}

func (r *userRepository) UpdateAddress(ctx context.Context, addr *domain.UserAddress) error {
	query := `UPDATE user_addresses SET tag = :tag, recipient_name = :recipient_name, phone_number = :phone_number, street_address = :street_address, city = :city, province = :province, postal_code = :postal_code, is_default = :is_default, updated_at = :updated_at
	          WHERE id = :id`
	_, err := r.db.NamedExecContext(ctx, query, addr)
	return err
}

func (r *userRepository) DeleteAddress(ctx context.Context, addressID uuid.UUID) error {
	query := `DELETE FROM user_addresses WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, addressID)
	return err
}

func (r *userRepository) UnsetDefaultAddresses(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE user_addresses SET is_default = FALSE WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}
