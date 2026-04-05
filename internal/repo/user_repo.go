package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/denden-dr/go-shop-yourself/internal/domain"
	"github.com/google/uuid"
)

// UserRepository handles database operations for users.
type UserRepository interface {
	CreateUser(ctx context.Context, u *domain.User, auth *domain.UserAuth, profile *domain.UserProfile) error
	GetUserByEmail(ctx context.Context, email string) (*domain.UserAuth, error)
	GetUserProfileByID(ctx context.Context, userID uuid.UUID) (*domain.UserProfile, error)
	UpdateLastLogin(ctx context.Context, userID uuid.UUID) error
}

type userRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new instance of UserRepository.
func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) CreateUser(ctx context.Context, u *domain.User, auth *domain.UserAuth, profile *domain.UserProfile) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert into users table
	_, err = tx.ExecContext(ctx, "INSERT INTO users (id, created_at) VALUES ($1, $2)", u.ID, u.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}

	// Insert into user_auth table
	_, err = tx.ExecContext(ctx, "INSERT INTO users_auth (user_id, email, password_hash, created_at) VALUES ($1, $2, $3, $4)",
		auth.UserID, auth.Email, auth.PasswordHash, auth.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert user_auth: %w", err)
	}

	// Insert into user_profile table
	_, err = tx.ExecContext(ctx, "INSERT INTO users_profile (user_id, username, updated_at) VALUES ($1, $2, $3)",
		profile.UserID, profile.Username, profile.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert user_profile: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*domain.UserAuth, error) {
	var auth domain.UserAuth
	query := "SELECT user_id, email, password_hash, created_at FROM users_auth WHERE email = $1"
	err := r.db.QueryRowContext(ctx, query, email).Scan(&auth.UserID, &auth.Email, &auth.PasswordHash, &auth.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return &auth, nil
}

func (r *userRepository) GetUserProfileByID(ctx context.Context, userID uuid.UUID) (*domain.UserProfile, error) {
	var profile domain.UserProfile
	var savedAddressJSON sql.NullString

	query := "SELECT user_id, username, saved_address, updated_at, last_login_at FROM users_profile WHERE user_id = $1"
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&profile.UserID,
		&profile.Username,
		&savedAddressJSON,
		&profile.UpdatedAt,
		&profile.LastLoginAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	if savedAddressJSON.Valid && savedAddressJSON.String != "" {
		var addr domain.Address
		if err := json.Unmarshal([]byte(savedAddressJSON.String), &addr); err != nil {
			// Log error but don't fail the whole request
			fmt.Printf("failed to unmarshal address: %v\n", err)
		} else {
			profile.SavedAddress = &addr
		}
	}

	return &profile, nil
}

func (r *userRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	query := "UPDATE users_profile SET last_login_at = $1 WHERE user_id = $2"
	_, err := r.db.ExecContext(ctx, query, now, userID)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}
	return nil
}
