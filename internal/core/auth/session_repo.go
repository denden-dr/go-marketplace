package auth

import (
	"context"
	"database/sql"

	"go-marketplace/internal/domain"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type sessionRepository struct {
	db *sqlx.DB
}

func NewSessionRepository(db *sqlx.DB) SessionRepository {
	return &sessionRepository{db: db}
}

func (r *sessionRepository) Create(ctx context.Context, session *domain.Session) error {
	query := `INSERT INTO sessions (id, user_id, token_hash, family_id, is_revoked, ip_address, user_agent, expires_at, created_at) 
	          VALUES (:id, :user_id, :token_hash, :family_id, :is_revoked, :ip_address, :user_agent, :expires_at, :created_at)`
	_, err := r.db.NamedExecContext(ctx, query, session)
	return err
}

func (r *sessionRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error) {
	query := `SELECT id, user_id, token_hash, family_id, is_revoked, ip_address, user_agent, expires_at, created_at 
	          FROM sessions WHERE token_hash = $1`
	var session domain.Session
	err := r.db.GetContext(ctx, &session, query, tokenHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &session, nil
}

func (r *sessionRepository) RevokeByID(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE sessions SET is_revoked = TRUE WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *sessionRepository) RevokeAllByFamilyID(ctx context.Context, familyID uuid.UUID) error {
	query := `UPDATE sessions SET is_revoked = TRUE WHERE family_id = $1`
	_, err := r.db.ExecContext(ctx, query, familyID)
	return err
}

func (r *sessionRepository) DeleteExpiredSessions(ctx context.Context) error {
	query := `DELETE FROM sessions WHERE expires_at < NOW()`
	_, err := r.db.ExecContext(ctx, query)
	return err
}
