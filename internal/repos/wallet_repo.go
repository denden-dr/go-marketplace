package repos

import (
	"context"

	"go-shop-yourself/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WalletRepository struct {
	db *pgxpool.Pool
}

func NewWalletRepository(db *pgxpool.Pool) *WalletRepository {
	return &WalletRepository{db: db}
}

func (r *WalletRepository) GetWalletByUserID(ctx context.Context, userID uuid.UUID) (*domain.Wallet, error) {
	query := `SELECT id, user_id, wallet_number, balance, currency, status, created_at, updated_at 
	          FROM wallets WHERE user_id = $1`
	var w domain.Wallet
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&w.ID, &w.UserID, &w.WalletNumber, &w.Balance, &w.Currency, &w.Status, &w.CreatedAt, &w.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &w, nil
}

func (r *WalletRepository) GetWalletHistory(ctx context.Context, walletID uuid.UUID) ([]domain.WalletTransaction, error) {
	query := `SELECT id, wallet_id, amount, direction, type, status, reference_id, balance_after, description, created_at 
	          FROM wallets_transaction WHERE wallet_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, walletID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []domain.WalletTransaction
	for rows.Next() {
		var t domain.WalletTransaction
		err := rows.Scan(
			&t.ID, &t.WalletID, &t.Amount, &t.Direction, &t.Type, &t.Status, &t.ReferenceID, &t.BalanceAfter, &t.Description, &t.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, t)
	}
	return transactions, nil
}

func (r *WalletRepository) Withdraw(ctx context.Context, walletID uuid.UUID, amount float64, txData domain.WalletTransaction) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Update balance
	updateQuery := `UPDATE wallets SET balance = balance - $1, updated_at = NOW() WHERE id = $2 AND balance >= $1`
	tag, err := tx.Exec(ctx, updateQuery, amount, walletID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows // Or a custom insufficient balance error
	}

	// Create transaction record
	insertQuery := `INSERT INTO wallets_transaction (id, wallet_id, amount, direction, type, status, reference_id, balance_after, description, created_at) 
	                VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err = tx.Exec(ctx, insertQuery, 
		txData.ID, txData.WalletID, txData.Amount, txData.Direction, txData.Type, txData.Status, 
		txData.ReferenceID, txData.BalanceAfter, txData.Description, txData.CreatedAt,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
