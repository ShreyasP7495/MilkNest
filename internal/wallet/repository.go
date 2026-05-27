package wallet

import (
	"context"
	"database/sql"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) DB() *sql.DB { return r.db }

// EnsureForUser inserts a wallet row if it does not yet exist (idempotent).
func (r *Repository) EnsureForUser(ctx context.Context, tx *sql.Tx, userID string) (string, error) {
	const q = `
		INSERT INTO wallets (user_id) VALUES ($1)
		ON CONFLICT (user_id) DO UPDATE SET updated_at = wallets.updated_at
		RETURNING id`
	var id string
	err := tx.QueryRowContext(ctx, q, userID).Scan(&id)
	return id, err
}

// LockForUser obtains a SELECT ... FOR UPDATE lock on the wallet row.
func (r *Repository) LockForUser(ctx context.Context, tx *sql.Tx, userID string) (*Wallet, error) {
	const q = `
		SELECT id, user_id, balance_paise, created_at, updated_at
		  FROM wallets WHERE user_id = $1 FOR UPDATE`
	w := &Wallet{}
	err := tx.QueryRowContext(ctx, q, userID).Scan(&w.ID, &w.UserID, &w.BalancePaise, &w.CreatedAt, &w.UpdatedAt)
	return w, err
}

// GetForUser is read-only (no FOR UPDATE).
func (r *Repository) GetForUser(ctx context.Context, userID string) (*Wallet, error) {
	const q = `
		SELECT id, user_id, balance_paise, created_at, updated_at
		  FROM wallets WHERE user_id = $1`
	w := &Wallet{}
	err := r.db.QueryRowContext(ctx, q, userID).Scan(&w.ID, &w.UserID, &w.BalancePaise, &w.CreatedAt, &w.UpdatedAt)
	return w, err
}

// SetBalance updates the row's balance to the given value.
func (r *Repository) SetBalance(ctx context.Context, tx *sql.Tx, walletID string, newBalance int64) error {
	const q = `UPDATE wallets SET balance_paise = $2, updated_at = now() WHERE id = $1`
	_, err := tx.ExecContext(ctx, q, walletID, newBalance)
	return err
}

// InsertTxn writes a ledger row. UNIQUE(ref_type, ref_id) makes retries safe.
// Returns sql.ErrNoRows if a duplicate row already exists (caller treats as
// idempotent success).
func (r *Repository) InsertTxn(ctx context.Context, tx *sql.Tx, walletID, txnType string, amount int64, refType, refID string, balanceAfter int64) (string, error) {
	const q = `
		INSERT INTO wallet_transactions (wallet_id, type, amount_paise, ref_type, ref_id, balance_after_paise)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (ref_type, ref_id) DO NOTHING
		RETURNING id`
	var id string
	err := tx.QueryRowContext(ctx, q, walletID, txnType, amount, refType, refID, balanceAfter).Scan(&id)
	return id, err
}

// ListTxns returns the most recent transactions for a wallet.
func (r *Repository) ListTxns(ctx context.Context, walletID string, limit int) ([]Transaction, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	const q = `
		SELECT id, wallet_id, type, amount_paise, ref_type, ref_id, balance_after_paise, created_at
		  FROM wallet_transactions WHERE wallet_id = $1 ORDER BY created_at DESC LIMIT $2`
	rows, err := r.db.QueryContext(ctx, q, walletID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Transaction, 0)
	for rows.Next() {
		var t Transaction
		if err := rows.Scan(&t.ID, &t.WalletID, &t.Type, &t.AmountPaise, &t.RefType, &t.RefID, &t.BalanceAfterPaise, &t.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
