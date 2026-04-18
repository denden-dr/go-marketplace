package user

import (
	"context"

	"go-shop-yourself/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) CreateUser(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (id, full_name, username, email, password, auth_provider, provider_id, created_at) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.Exec(ctx, query, user.ID, user.FullName, user.Username, user.Email, user.Password, user.AuthProvider, user.ProviderID, user.CreatedAt)
	return err
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id, full_name, username, email, password, auth_provider, provider_id, created_at FROM users WHERE email = $1`
	var user domain.User
	err := r.db.QueryRow(ctx, query, email).Scan(&user.ID, &user.FullName, &user.Username, &user.Email, &user.Password, &user.AuthProvider, &user.ProviderID, &user.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `SELECT id, full_name, username, email, password, auth_provider, provider_id, created_at FROM users WHERE id = $1`
	var user domain.User
	err := r.db.QueryRow(ctx, query, id).Scan(&user.ID, &user.FullName, &user.Username, &user.Email, &user.Password, &user.AuthProvider, &user.ProviderID, &user.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetUserByProviderID(ctx context.Context, provider string, providerID string) (*domain.User, error) {
	query := `SELECT id, full_name, username, email, password, auth_provider, provider_id, created_at 
	          FROM users WHERE auth_provider = $1 AND provider_id = $2`
	var user domain.User
	err := r.db.QueryRow(ctx, query, provider, providerID).Scan(&user.ID, &user.FullName, &user.Username, &user.Email, &user.Password, &user.AuthProvider, &user.ProviderID, &user.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `SELECT id, full_name, username, email, password, auth_provider, provider_id, created_at FROM users WHERE username = $1`
	var user domain.User
	err := r.db.QueryRow(ctx, query, username).Scan(&user.ID, &user.FullName, &user.Username, &user.Email, &user.Password, &user.AuthProvider, &user.ProviderID, &user.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) CreateAddress(ctx context.Context, addr *domain.UserAddress) error {
	query := `INSERT INTO user_addresses (id, user_id, tag, recipient_name, phone_number, street_address, city, province, postal_code, is_default, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	_, err := r.db.Exec(ctx, query, addr.ID, addr.UserID, addr.Tag, addr.RecipientName, addr.PhoneNumber, addr.StreetAddress, addr.City, addr.Province, addr.PostalCode, addr.IsDefault, addr.CreatedAt, addr.UpdatedAt)
	return err
}

func (r *userRepository) GetAddressesByUserID(ctx context.Context, userID uuid.UUID) ([]domain.UserAddress, error) {
	query := `SELECT id, user_id, tag, recipient_name, phone_number, street_address, city, province, postal_code, is_default, created_at, updated_at 
	          FROM user_addresses WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var addresses []domain.UserAddress
	for rows.Next() {
		var addr domain.UserAddress
		err := rows.Scan(&addr.ID, &addr.UserID, &addr.Tag, &addr.RecipientName, &addr.PhoneNumber, &addr.StreetAddress, &addr.City, &addr.Province, &addr.PostalCode, &addr.IsDefault, &addr.CreatedAt, &addr.UpdatedAt)
		if err != nil {
			return nil, err
		}
		addresses = append(addresses, addr)
	}
	return addresses, nil
}

func (r *userRepository) GetAddressByID(ctx context.Context, addressID uuid.UUID) (*domain.UserAddress, error) {
	query := `SELECT id, user_id, tag, recipient_name, phone_number, street_address, city, province, postal_code, is_default, created_at, updated_at 
	          FROM user_addresses WHERE id = $1`
	var addr domain.UserAddress
	err := r.db.QueryRow(ctx, query, addressID).Scan(&addr.ID, &addr.UserID, &addr.Tag, &addr.RecipientName, &addr.PhoneNumber, &addr.StreetAddress, &addr.City, &addr.Province, &addr.PostalCode, &addr.IsDefault, &addr.CreatedAt, &addr.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &addr, nil
}

func (r *userRepository) UpdateAddress(ctx context.Context, addr *domain.UserAddress) error {
	query := `UPDATE user_addresses SET tag = $1, recipient_name = $2, phone_number = $3, street_address = $4, city = $5, province = $6, postal_code = $7, is_default = $8, updated_at = $9
	          WHERE id = $10`
	_, err := r.db.Exec(ctx, query, addr.Tag, addr.RecipientName, addr.PhoneNumber, addr.StreetAddress, addr.City, addr.Province, addr.PostalCode, addr.IsDefault, addr.UpdatedAt, addr.ID)
	return err
}

func (r *userRepository) DeleteAddress(ctx context.Context, addressID uuid.UUID) error {
	query := `DELETE FROM user_addresses WHERE id = $1`
	_, err := r.db.Exec(ctx, query, addressID)
	return err
}

func (r *userRepository) UnsetDefaultAddresses(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE user_addresses SET is_default = FALSE WHERE user_id = $1`
	_, err := r.db.Exec(ctx, query, userID)
	return err
}
