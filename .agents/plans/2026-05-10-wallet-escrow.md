# Wallet Escrow System Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement a pending balance system for wallets to support marketplace escrow and wallet-based refunds for cancelled orders.

**Architecture:** Extend existing `wallets` and `wallets_transaction` tables with `pending_balance` and `pending_balance_after` columns. Update Domain, Repository, and Service layers to handle atomic transitions including a cross-wallet refund operation.

**Tech Stack:** Go, Fiber v2, sqlx, PostgreSQL, shopspring/decimal.

---

### Task 1: Database Migration
**Files:**
- Create: `internal/database/migrations/000002_add_pending_balance.up.sql`
- Create: `internal/database/migrations/000002_add_pending_balance.down.sql`

- [ ] **Step 1: Create the 'up' migration file**
```sql
ALTER TABLE wallets ADD COLUMN pending_balance numeric(15,2) DEFAULT 0.00 NOT NULL;
ALTER TABLE wallets_transaction ADD COLUMN pending_balance_after numeric(15,2) DEFAULT 0.00 NOT NULL;
```
- [ ] **Step 2: Create the 'down' migration file**
```sql
ALTER TABLE wallets_transaction DROP COLUMN pending_balance_after;
ALTER TABLE wallets DROP COLUMN pending_balance;
```
- [ ] **Step 3: Run the migration**
Run: `migrate -path internal/database/migrations -database "$DB_URL" up`
- [ ] **Step 4: Commit**
```bash
git add internal/database/migrations/
git commit -m "db: add pending balance columns"
```

---

### Task 2: Update Domain Model
**Files:**
- Modify: `internal/domain/wallet.go`

- [ ] **Step 1: Update Wallet and WalletTransaction structs**
Include `PendingBalance` in `Wallet` and `PendingBalanceAfter` in `WalletTransaction`.
- [ ] **Step 2: Commit**
```bash
git add internal/domain/wallet.go
git commit -m "domain: add pending balance fields"
```

---

### Task 3: Update Repository Layer (Interfaces & SQL)
**Files:**
- Modify: `internal/core/wallet/wallet_service.go` (WalletRepository interface)
- Modify: `internal/core/wallet/wallet_repo.go` (Implementation)

- [ ] **Step 1: Update interface in `wallet_service.go`**
```go
type WalletRepository interface {
    // ...
    AddPendingBalanceTX(ctx context.Context, tx *sqlx.Tx, walletID uuid.UUID, amount decimal.Decimal, txData domain.WalletTransaction) error
    SettlePendingBalanceTX(ctx context.Context, tx *sqlx.Tx, walletID uuid.UUID, amount decimal.Decimal, txData domain.WalletTransaction) error
    RefundFromPendingTX(ctx context.Context, tx *sqlx.Tx, merchantWalletID, userWalletID uuid.UUID, amount decimal.Decimal, merchantTxData, userTxData domain.WalletTransaction) error
}
```
- [ ] **Step 2: Update SQL queries in `wallet_repo.go`**
Include new columns in all SELECT/INSERT/UPDATE statements.
- [ ] **Step 3: Commit**
```bash
git commit -m "repo: update queries and interface for escrow/refund"
```

---

### Task 4: Implement Add and Settle Logic
**Files:**
- Modify: `internal/core/wallet/wallet_repo.go`

- [ ] **Step 1: Implement `AddPendingBalanceTX`**
Increase `pending_balance`, log transaction.
- [ ] **Step 2: Implement `SettlePendingBalanceTX`**
Decrease `pending_balance`, increase `balance`, update transaction to SUCCESS.
- [ ] **Step 3: Commit**
```bash
git commit -m "repo: implement add and settle pending balance"
```

---

### Task 5: Implement RefundFromPendingTX
**Files:**
- Modify: `internal/core/wallet/wallet_repo.go`

- [ ] **Step 1: Implement `RefundFromPendingTX`**
This is a cross-wallet atomic operation.
```go
func (r *walletRepository) RefundFromPendingTX(ctx context.Context, tx *sqlx.Tx, merchantWalletID, userWalletID uuid.UUID, amount decimal.Decimal, merchantTxData, userTxData domain.WalletTransaction) error {
    // 1. Lock both wallets (ensure consistent order to avoid deadlock)
    // 2. Decrease Merchant PendingBalance
    // 3. Increase User Balance
    // 4. Update Merchant original transaction to 'cancelled'
    // 5. Insert User new 'refund' transaction
    return nil // Implementation details in task
}
```
- [ ] **Step 2: Commit**
```bash
git commit -m "repo: implement cross-wallet refund logic"
```

---

### Task 6: Service Layer & Integration Verification
**Files:**
- Modify: `internal/core/wallet/wallet_service.go`
- Modify: `test/integration/wallet_repo_test.go`

- [ ] **Step 1: Add service methods**
Implement `CreditPending`, `SettlePending`, and `RefundFromPending` in `WalletService`.
- [ ] **Step 2: Add integration test cases**
- [ ] **Step 3: Commit**
```bash
git commit -m "feat: complete wallet escrow and refund implementation"
```
