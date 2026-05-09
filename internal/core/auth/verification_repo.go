package auth

import (
	"context"
	"database/sql"

	"go-marketplace/internal/domain"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type verificationRepository struct {
	db *sqlx.DB
}

func NewVerificationRepository(db *sqlx.DB) VerificationRepository {
	return &verificationRepository{db: db}
}

func (r *verificationRepository) Create(ctx context.Context, vc *domain.VerificationCode) error {
	query := `INSERT INTO verification_codes (id, user_id, code_hash, expires_at, created_at) 
	          VALUES (:id, :user_id, :code_hash, :expires_at, :created_at)`
	_, err := r.db.NamedExecContext(ctx, query, vc)
	return err
}

func (r *verificationRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.VerificationCode, error) {
	query := `SELECT id, user_id, code_hash, expires_at, created_at 
	          FROM verification_codes WHERE user_id = $1`
	var vc domain.VerificationCode
	err := r.db.GetContext(ctx, &vc, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &vc, nil
}

func (r *verificationRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM verification_codes WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}
