package wallet

import (
	"context"
	"errors"

	"go-shop-yourself/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type walletRepository struct {
	db *pgxpool.Pool
}

func NewWalletRepository(db *pgxpool.Pool) WalletRepository {
	return &walletRepository{db: db}
}

func (r *walletRepository) GetWalletByUserID(ctx context.Context, userID uuid.UUID) (*domain.Wallet, error) {
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

func (r *walletRepository) GetWalletHistory(ctx context.Context, walletID uuid.UUID, limit, offset int) ([]domain.WalletTransaction, error) {
	query := `SELECT id, wallet_id, amount, direction, type, status, reference_id, balance_after, description, created_at 
	          FROM wallets_transaction WHERE wallet_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, query, walletID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	transactions := []domain.WalletTransaction{}
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

func (r *walletRepository) Withdraw(ctx context.Context, walletID uuid.UUID, amount decimal.Decimal, txData domain.WalletTransaction) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Get current balance and lock row
	var currentBalance decimal.Decimal
	var status domain.WalletStatus
	query := `SELECT balance, status FROM wallets WHERE id = $1 FOR UPDATE`
	err = tx.QueryRow(ctx, query, walletID).Scan(&currentBalance, &status)
	if err != nil {
		return err
	}

	// Check status
	if status != domain.WalletStatusActive {
		return errors.New("wallet is not active")
	}

	// Check balance
	if currentBalance.LessThan(amount) {
		return errors.New("insufficient balance")
	}

	// Update balance
	newBalance := currentBalance.Sub(amount)
	updateQuery := `UPDATE wallets SET balance = $1, updated_at = NOW() WHERE id = $2`
	_, err = tx.Exec(ctx, updateQuery, newBalance, walletID)
	if err != nil {
		return err
	}

	// Create transaction record
	txData.BalanceAfter = newBalance // Ensure balance_after is accurate
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

func (r *walletRepository) DeductBalanceTX(ctx context.Context, tx pgx.Tx, walletID uuid.UUID, amount decimal.Decimal, txData domain.WalletTransaction) error {
	var currentBalance decimal.Decimal
	var status domain.WalletStatus
	query := `SELECT balance, status FROM wallets WHERE id = $1 FOR UPDATE`
	err := tx.QueryRow(ctx, query, walletID).Scan(&currentBalance, &status)
	if err != nil {
		return err
	}

	if status != domain.WalletStatusActive {
		return errors.New("wallet is not active")
	}

	if currentBalance.LessThan(amount) {
		return errors.New("insufficient balance")
	}

	newBalance := currentBalance.Sub(amount)
	updateQuery := `UPDATE wallets SET balance = $1, updated_at = NOW() WHERE id = $2`
	_, err = tx.Exec(ctx, updateQuery, newBalance, walletID)
	if err != nil {
		return err
	}

	txData.BalanceAfter = newBalance
	insertQuery := `INSERT INTO wallets_transaction (id, wallet_id, amount, direction, type, status, reference_id, balance_after, description, created_at) 
	                VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err = tx.Exec(ctx, insertQuery,
		txData.ID, txData.WalletID, txData.Amount, txData.Direction, txData.Type, txData.Status,
		txData.ReferenceID, txData.BalanceAfter, txData.Description, txData.CreatedAt,
	)
	return err
}

func (r *walletRepository) AddBalanceTX(ctx context.Context, tx pgx.Tx, walletID uuid.UUID, amount decimal.Decimal, txData domain.WalletTransaction) error {
	var currentBalance decimal.Decimal
	query := `SELECT balance FROM wallets WHERE id = $1 FOR UPDATE`
	err := tx.QueryRow(ctx, query, walletID).Scan(&currentBalance)
	if err != nil {
		return err
	}

	newBalance := currentBalance.Add(amount)
	updateQuery := `UPDATE wallets SET balance = $1, updated_at = NOW() WHERE id = $2`
	_, err = tx.Exec(ctx, updateQuery, newBalance, walletID)
	if err != nil {
		return err
	}

	txData.BalanceAfter = newBalance
	insertQuery := `INSERT INTO wallets_transaction (id, wallet_id, amount, direction, type, status, reference_id, balance_after, description, created_at) 
	                VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err = tx.Exec(ctx, insertQuery,
		txData.ID, txData.WalletID, txData.Amount, txData.Direction, txData.Type, txData.Status,
		txData.ReferenceID, txData.BalanceAfter, txData.Description, txData.CreatedAt,
	)
	return err
}

func (r *walletRepository) Create(ctx context.Context, w *domain.Wallet) error {
	query := `INSERT INTO wallets (id, user_id, wallet_number, balance, currency, status, created_at, updated_at) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.Exec(ctx, query, w.ID, w.UserID, w.WalletNumber, w.Balance, w.Currency, w.Status, w.CreatedAt, w.UpdatedAt)
	return err
}

func (r *walletRepository) CreateTx(ctx context.Context, tx pgx.Tx, w *domain.Wallet) error {
	query := `INSERT INTO wallets (id, user_id, wallet_number, balance, currency, status, created_at, updated_at) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := tx.Exec(ctx, query, w.ID, w.UserID, w.WalletNumber, w.Balance, w.Currency, w.Status, w.CreatedAt, w.UpdatedAt)
	return err
}

func (r *walletRepository) GetPool() domain.Pool {
	return r.db
}
