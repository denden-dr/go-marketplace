package auth

import (
	"context"
	"database/sql"

	"go-marketplace/internal/domain"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type refreshTokenRepository struct {
	db *sqlx.DB
}

func NewRefreshTokenRepository(db *sqlx.DB) RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
	query := `INSERT INTO refresh_tokens (id, user_id, token_hash, family_id, is_revoked, expires_at, created_at) 
	          VALUES (:id, :user_id, :token_hash, :family_id, :is_revoked, :expires_at, :created_at)`
	_, err := r.db.NamedExecContext(ctx, query, token)
	return err
}

func (r *refreshTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	query := `SELECT id, user_id, token_hash, family_id, is_revoked, expires_at, created_at 
	          FROM refresh_tokens WHERE token_hash = $1`
	var token domain.RefreshToken
	err := r.db.GetContext(ctx, &token, query, tokenHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &token, nil
}

func (r *refreshTokenRepository) RevokeByID(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE refresh_tokens SET is_revoked = TRUE WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *refreshTokenRepository) RevokeAllByFamilyID(ctx context.Context, familyID uuid.UUID) error {
	query := `UPDATE refresh_tokens SET is_revoked = TRUE WHERE family_id = $1`
	_, err := r.db.ExecContext(ctx, query, familyID)
	return err
}

func (r *refreshTokenRepository) DeleteExpiredTokens(ctx context.Context) error {
	query := `DELETE FROM refresh_tokens WHERE expires_at < NOW()`
	_, err := r.db.ExecContext(ctx, query)
	return err
}
