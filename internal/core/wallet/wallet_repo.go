package wallet

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"go-marketplace/internal/domain"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
)

type WalletRepository interface {
	GetWalletByUserID(ctx context.Context, userID uuid.UUID) (*domain.Wallet, error)
	GetWalletsByUserIDs(ctx context.Context, userIDs []uuid.UUID) ([]domain.Wallet, error)
	GetWalletsByUserIDsTX(ctx context.Context, tx *sqlx.Tx, userIDs []uuid.UUID) ([]domain.Wallet, error)
	GetWalletHistory(ctx context.Context, walletID uuid.UUID, limit, offset int) ([]domain.WalletTransaction, error)
	Withdraw(ctx context.Context, walletID uuid.UUID, amount decimal.Decimal, txData domain.WalletTransaction) error
	DeductBalanceTX(ctx context.Context, tx *sqlx.Tx, walletID uuid.UUID, amount decimal.Decimal, txData domain.WalletTransaction) error
	AddBalanceTX(ctx context.Context, tx *sqlx.Tx, walletID uuid.UUID, amount decimal.Decimal, txData domain.WalletTransaction) error
	AddPendingBalanceTX(ctx context.Context, tx *sqlx.Tx, walletID uuid.UUID, amount decimal.Decimal, txData domain.WalletTransaction) error
	AddPendingBalancesBatchTX(ctx context.Context, tx *sqlx.Tx, updates []domain.WalletBalanceUpdate) error
	SettlePendingBalanceTX(ctx context.Context, tx *sqlx.Tx, walletID uuid.UUID, amount decimal.Decimal, txData domain.WalletTransaction) error
	FreezeBalanceTX(ctx context.Context, tx *sqlx.Tx, walletID uuid.UUID, amount decimal.Decimal, txData domain.WalletTransaction) error
	RefundFromPendingTX(ctx context.Context, tx *sqlx.Tx, walletID uuid.UUID, amount decimal.Decimal, txData domain.WalletTransaction) error
	Create(ctx context.Context, w *domain.Wallet) error
	CreateTx(ctx context.Context, tx *sqlx.Tx, w *domain.Wallet) error
	GetPool() domain.Pool
}

type walletRepository struct {
	db *sqlx.DB
}

func NewWalletRepository(db *sqlx.DB) WalletRepository {
	return &walletRepository{db: db}
}

func (r *walletRepository) GetWalletByUserID(ctx context.Context, userID uuid.UUID) (*domain.Wallet, error) {
	query := `SELECT id, user_id, wallet_number, balance, pending_balance, currency, status, created_at, updated_at 
	          FROM wallets WHERE user_id = $1`
	var w domain.Wallet
	err := r.db.GetContext(ctx, &w, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &w, nil
}

func (r *walletRepository) GetWalletHistory(ctx context.Context, walletID uuid.UUID, limit, offset int) ([]domain.WalletTransaction, error) {
	query := `SELECT id, wallet_id, amount, direction, type, status, reference_id, balance_after, pending_balance_after, description, created_at 
	          FROM wallets_transaction WHERE wallet_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	var transactions []domain.WalletTransaction
	err := r.db.SelectContext(ctx, &transactions, query, walletID, limit, offset)
	if err != nil {
		return nil, err
	}
	return transactions, nil
}

func (r *walletRepository) Withdraw(ctx context.Context, walletID uuid.UUID, amount decimal.Decimal, txData domain.WalletTransaction) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get current balance and lock row
	var currentBalance, currentPending decimal.Decimal
	var status domain.WalletStatus
	query := `SELECT balance, pending_balance, status FROM wallets WHERE id = $1 FOR UPDATE`
	err = tx.QueryRowContext(ctx, query, walletID).Scan(&currentBalance, &currentPending, &status)
	if err != nil {
		return err
	}

	// Check status
	if status != domain.WalletStatusActive {
		return domain.ErrWalletNotActive
	}

	// Check balance
	if currentBalance.LessThan(amount) {
		return domain.ErrInsufficientBalance
	}

	// Update balance
	newBalance := currentBalance.Sub(amount)
	updateQuery := `UPDATE wallets SET balance = $1, updated_at = NOW() WHERE id = $2`
	_, err = tx.ExecContext(ctx, updateQuery, newBalance, walletID)
	if err != nil {
		return err
	}

	// Create transaction record
	txData.BalanceAfter = newBalance
	txData.PendingBalanceAfter = currentPending
	insertQuery := `INSERT INTO wallets_transaction (id, wallet_id, amount, direction, type, status, reference_id, balance_after, pending_balance_after, description, created_at) 
	                VALUES (:id, :wallet_id, :amount, :direction, :type, :status, :reference_id, :balance_after, :pending_balance_after, :description, :created_at)`
	_, err = tx.NamedExecContext(ctx, insertQuery, txData)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *walletRepository) DeductBalanceTX(ctx context.Context, tx *sqlx.Tx, walletID uuid.UUID, amount decimal.Decimal, txData domain.WalletTransaction) error {
	var currentBalance, currentPending decimal.Decimal
	var status domain.WalletStatus
	query := `SELECT balance, pending_balance, status FROM wallets WHERE id = $1 FOR UPDATE`
	err := tx.QueryRowContext(ctx, query, walletID).Scan(&currentBalance, &currentPending, &status)
	if err != nil {
		return err
	}

	if status != domain.WalletStatusActive {
		return domain.ErrWalletNotActive
	}

	if currentBalance.LessThan(amount) {
		return domain.ErrInsufficientBalance
	}

	newBalance := currentBalance.Sub(amount)
	updateQuery := `UPDATE wallets SET balance = $1, updated_at = NOW() WHERE id = $2`
	_, err = tx.ExecContext(ctx, updateQuery, newBalance, walletID)
	if err != nil {
		return err
	}

	txData.BalanceAfter = newBalance
	txData.PendingBalanceAfter = currentPending
	insertQuery := `INSERT INTO wallets_transaction (id, wallet_id, amount, direction, type, status, reference_id, balance_after, pending_balance_after, description, created_at) 
	                VALUES (:id, :wallet_id, :amount, :direction, :type, :status, :reference_id, :balance_after, :pending_balance_after, :description, :created_at)`
	_, err = tx.NamedExecContext(ctx, insertQuery, txData)
	return err
}

func (r *walletRepository) AddBalanceTX(ctx context.Context, tx *sqlx.Tx, walletID uuid.UUID, amount decimal.Decimal, txData domain.WalletTransaction) error {
	var currentBalance, currentPending decimal.Decimal
	query := `SELECT balance, pending_balance FROM wallets WHERE id = $1 FOR UPDATE`
	err := tx.QueryRowContext(ctx, query, walletID).Scan(&currentBalance, &currentPending)
	if err != nil {
		return err
	}

	newBalance := currentBalance.Add(amount)
	updateQuery := `UPDATE wallets SET balance = $1, updated_at = NOW() WHERE id = $2`
	_, err = tx.ExecContext(ctx, updateQuery, newBalance, walletID)
	if err != nil {
		return err
	}

	txData.BalanceAfter = newBalance
	txData.PendingBalanceAfter = currentPending
	insertQuery := `INSERT INTO wallets_transaction (id, wallet_id, amount, direction, type, status, reference_id, balance_after, pending_balance_after, description, created_at) 
	                VALUES (:id, :wallet_id, :amount, :direction, :type, :status, :reference_id, :balance_after, :pending_balance_after, :description, :created_at)`
	_, err = tx.NamedExecContext(ctx, insertQuery, txData)
	return err
}

func (r *walletRepository) AddPendingBalanceTX(ctx context.Context, tx *sqlx.Tx, walletID uuid.UUID, amount decimal.Decimal, txData domain.WalletTransaction) error {
	var currentBalance, currentPending decimal.Decimal
	query := `SELECT balance, pending_balance FROM wallets WHERE id = $1 FOR UPDATE`
	err := tx.QueryRowContext(ctx, query, walletID).Scan(&currentBalance, &currentPending)
	if err != nil {
		return err
	}

	newPending := currentPending.Add(amount)
	updateQuery := `UPDATE wallets SET pending_balance = $1, updated_at = NOW() WHERE id = $2`
	_, err = tx.ExecContext(ctx, updateQuery, newPending, walletID)
	if err != nil {
		return err
	}

	txData.BalanceAfter = currentBalance
	txData.PendingBalanceAfter = newPending
	insertQuery := `INSERT INTO wallets_transaction (id, wallet_id, amount, direction, type, status, reference_id, balance_after, pending_balance_after, description, created_at) 
	                VALUES (:id, :wallet_id, :amount, :direction, :type, :status, :reference_id, :balance_after, :pending_balance_after, :description, :created_at)`
	_, err = tx.NamedExecContext(ctx, insertQuery, txData)
	return err
}

func (r *walletRepository) SettlePendingBalanceTX(ctx context.Context, tx *sqlx.Tx, walletID uuid.UUID, amount decimal.Decimal, txData domain.WalletTransaction) error {
	var currentBalance, currentPending decimal.Decimal
	query := `SELECT balance, pending_balance FROM wallets WHERE id = $1 FOR UPDATE`
	err := tx.QueryRowContext(ctx, query, walletID).Scan(&currentBalance, &currentPending)
	if err != nil {
		return err
	}

	if currentPending.LessThan(amount) {
		return domain.ErrInsufficientPendingBalance
	}

	newBalance := currentBalance.Add(amount)
	newPending := currentPending.Sub(amount)
	updateQuery := `UPDATE wallets SET balance = $1, pending_balance = $2, updated_at = NOW() WHERE id = $3`
	_, err = tx.ExecContext(ctx, updateQuery, newBalance, newPending, walletID)
	if err != nil {
		return err
	}

	txData.BalanceAfter = newBalance
	txData.PendingBalanceAfter = newPending
	insertQuery := `INSERT INTO wallets_transaction (id, wallet_id, amount, direction, type, status, reference_id, balance_after, pending_balance_after, description, created_at) 
	                VALUES (:id, :wallet_id, :amount, :direction, :type, :status, :reference_id, :balance_after, :pending_balance_after, :description, :created_at)`
	_, err = tx.NamedExecContext(ctx, insertQuery, txData)
	return err
}

func (r *walletRepository) FreezeBalanceTX(ctx context.Context, tx *sqlx.Tx, walletID uuid.UUID, amount decimal.Decimal, txData domain.WalletTransaction) error {
	var currentBalance, currentPending decimal.Decimal
	query := `SELECT balance, pending_balance FROM wallets WHERE id = $1 FOR UPDATE`
	err := tx.QueryRowContext(ctx, query, walletID).Scan(&currentBalance, &currentPending)
	if err != nil {
		return err
	}

	if currentBalance.LessThan(amount) {
		return domain.ErrInsufficientBalance
	}

	newBalance := currentBalance.Sub(amount)
	newPending := currentPending.Add(amount)
	updateQuery := `UPDATE wallets SET balance = $1, pending_balance = $2, updated_at = NOW() WHERE id = $3`
	_, err = tx.ExecContext(ctx, updateQuery, newBalance, newPending, walletID)
	if err != nil {
		return err
	}

	txData.BalanceAfter = newBalance
	txData.PendingBalanceAfter = newPending
	insertQuery := `INSERT INTO wallets_transaction (id, wallet_id, amount, direction, type, status, reference_id, balance_after, pending_balance_after, description, created_at) 
	                VALUES (:id, :wallet_id, :amount, :direction, :type, :status, :reference_id, :balance_after, :pending_balance_after, :description, :created_at)`
	_, err = tx.NamedExecContext(ctx, insertQuery, txData)
	return err
}

func (r *walletRepository) RefundFromPendingTX(ctx context.Context, tx *sqlx.Tx, walletID uuid.UUID, amount decimal.Decimal, txData domain.WalletTransaction) error {
	var currentBalance, currentPending decimal.Decimal
	query := `SELECT balance, pending_balance FROM wallets WHERE id = $1 FOR UPDATE`
	err := tx.QueryRowContext(ctx, query, walletID).Scan(&currentBalance, &currentPending)
	if err != nil {
		return err
	}

	if currentPending.LessThan(amount) {
		return domain.ErrInsufficientPendingBalance
	}

	newPending := currentPending.Sub(amount)
	updateQuery := `UPDATE wallets SET pending_balance = $1, updated_at = NOW() WHERE id = $2`
	_, err = tx.ExecContext(ctx, updateQuery, newPending, walletID)
	if err != nil {
		return err
	}

	txData.BalanceAfter = currentBalance
	txData.PendingBalanceAfter = newPending
	insertQuery := `INSERT INTO wallets_transaction (id, wallet_id, amount, direction, type, status, reference_id, balance_after, pending_balance_after, description, created_at) 
	                VALUES (:id, :wallet_id, :amount, :direction, :type, :status, :reference_id, :balance_after, :pending_balance_after, :description, :created_at)`
	_, err = tx.NamedExecContext(ctx, insertQuery, txData)
	return err
}

func (r *walletRepository) Create(ctx context.Context, w *domain.Wallet) error {
	query := `INSERT INTO wallets (id, user_id, wallet_number, balance, pending_balance, currency, status, created_at, updated_at) 
	          VALUES (:id, :user_id, :wallet_number, :balance, :pending_balance, :currency, :status, :created_at, :updated_at)`
	_, err := r.db.NamedExecContext(ctx, query, w)
	return err
}

func (r *walletRepository) CreateTx(ctx context.Context, tx *sqlx.Tx, w *domain.Wallet) error {
	query := `INSERT INTO wallets (id, user_id, wallet_number, balance, pending_balance, currency, status, created_at, updated_at) 
	          VALUES (:id, :user_id, :wallet_number, :balance, :pending_balance, :currency, :status, :created_at, :updated_at)`
	_, err := tx.NamedExecContext(ctx, query, w)
	return err
}

func (r *walletRepository) GetPool() domain.Pool {
	return r.db
}

func (r *walletRepository) GetWalletsByUserIDs(ctx context.Context, userIDs []uuid.UUID) ([]domain.Wallet, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}

	query := `SELECT id, user_id, wallet_number, balance, pending_balance, currency, status, created_at, updated_at
	           FROM wallets WHERE user_id IN (?)`

	query, args, err := sqlx.In(query, userIDs)
	if err != nil {
		return nil, err
	}
	query = r.db.Rebind(query)

	var wallets []domain.Wallet
	err = r.db.SelectContext(ctx, &wallets, query, args...)
	return wallets, err
}

func (r *walletRepository) GetWalletsByUserIDsTX(ctx context.Context, tx *sqlx.Tx, userIDs []uuid.UUID) ([]domain.Wallet, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}

	query := `SELECT id, user_id, wallet_number, balance, pending_balance, currency, status, created_at, updated_at
	           FROM wallets WHERE user_id IN (?)`

	query, args, err := sqlx.In(query, userIDs)
	if err != nil {
		return nil, err
	}
	query = tx.Rebind(query)

	var wallets []domain.Wallet
	err = tx.SelectContext(ctx, &wallets, query, args...)
	return wallets, err
}

func (r *walletRepository) AddPendingBalancesBatchTX(ctx context.Context, tx *sqlx.Tx, updates []domain.WalletBalanceUpdate) error {
	if len(updates) == 0 {
		return nil
	}

	var sb strings.Builder
	sb.WriteString(`WITH updated AS (
		UPDATE wallets SET pending_balance = wallets.pending_balance + v.amount, updated_at = NOW()
		FROM (VALUES `)

	args := make([]interface{}, 0, len(updates)*2+len(updates)*9)
	for idx, u := range updates {
		if idx > 0 {
			sb.WriteString(", ")
		}
		p := idx * 2
		fmt.Fprintf(&sb, "($%d::uuid, $%d::numeric)", p+1, p+2)
		args = append(args, u.WalletID, u.Amount)
	}

	sb.WriteString(`) AS v(wallet_id, amount) 
		WHERE wallets.id = v.wallet_id
		RETURNING wallets.id, wallets.balance, wallets.pending_balance
	)
	INSERT INTO wallets_transaction (id, wallet_id, amount, direction, type, status, reference_id, balance_after, pending_balance_after, description, created_at)
	SELECT t.id, t.wallet_id, t.amount, t.direction, t.type, t.status, t.reference_id, updated.balance, updated.pending_balance, t.description, t.created_at
	FROM (VALUES `)

	j := len(updates) * 2
	for idx, u := range updates {
		if idx > 0 {
			sb.WriteString(", ")
		}
		p := j + idx*9
		fmt.Fprintf(&sb, "($%d::uuid, $%d::uuid, $%d::numeric, $%d, $%d, $%d, $%d, $%d, $%d::timestamp)",
			p+1, p+2, p+3, p+4, p+5, p+6, p+7, p+8, p+9)
		args = append(args,
			u.Transaction.ID,
			u.Transaction.WalletID,
			u.Transaction.Amount,
			u.Transaction.Direction,
			u.Transaction.Type,
			u.Transaction.Status,
			u.Transaction.ReferenceID,
			u.Transaction.Description,
			u.Transaction.CreatedAt,
		)
	}

	sb.WriteString(`) AS t(id, wallet_id, amount, direction, type, status, reference_id, description, created_at)
	JOIN updated ON updated.id = t.wallet_id`)

	_, err := tx.ExecContext(ctx, sb.String(), args...)
	return err
}
