package wallet

import (
	"context"
	"database/sql"
	"errors"

	"github.com/milknest/backend/internal/payments"
	"github.com/milknest/backend/pkg/utils"
)

type Service struct {
	repo *Repository
	pay  *payments.Service
}

func NewService(repo *Repository, pay *payments.Service) *Service {
	return &Service{repo: repo, pay: pay}
}

// Balance returns the wallet balance, ensuring a row exists.
func (s *Service) Balance(ctx context.Context, userID string) (*Wallet, error) {
	w, err := s.repo.GetForUser(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		// auto-create on first read
		tx, txErr := s.repo.DB().BeginTx(ctx, nil)
		if txErr != nil {
			return nil, utils.NewInternal(txErr)
		}
		defer func() { _ = tx.Rollback() }()
		if _, err := s.repo.EnsureForUser(ctx, tx, userID); err != nil {
			return nil, utils.NewInternal(err)
		}
		if err := tx.Commit(); err != nil {
			return nil, utils.NewInternal(err)
		}
		w, err = s.repo.GetForUser(ctx, userID)
		if err != nil {
			return nil, utils.NewInternal(err)
		}
	} else if err != nil {
		return nil, utils.NewInternal(err)
	}
	return w, nil
}

// Topup creates a Razorpay order for the topup amount. The wallet is only
// credited when the webhook arrives (see CreditFromTopup).
func (s *Service) Topup(ctx context.Context, userID string, req TopupRequest) (*payments.CreateResponse, error) {
	// Ensure the wallet exists upfront so CreditFromTopup never has to.
	if _, err := s.Balance(ctx, userID); err != nil {
		return nil, err
	}
	return s.pay.CreateForWalletTopup(ctx, userID, req.AmountPaise)
}

// CreditFromTopup is called by the payments webhook handler. It credits the
// wallet inside a tx, locking the row, and writes an idempotent ledger entry
// keyed by the payment id.
func (s *Service) CreditFromTopup(ctx context.Context, paymentID, userID string, amountPaise int64) error {
	tx, err := s.repo.DB().BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := s.repo.EnsureForUser(ctx, tx, userID); err != nil {
		return err
	}
	w, err := s.repo.LockForUser(ctx, tx, userID)
	if err != nil {
		return err
	}

	newBalance := w.BalancePaise + amountPaise
	if _, err := s.repo.InsertTxn(ctx, tx, w.ID, "credit", amountPaise, "topup", paymentID, newBalance); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Duplicate webhook: ledger row already exists; treat as success.
			return tx.Commit()
		}
		return err
	}
	if err := s.repo.SetBalance(ctx, tx, w.ID, newBalance); err != nil {
		return err
	}
	return tx.Commit()
}

// Debit deducts amountPaise from the user's wallet for the given ref. Returns
// utils.NewConflict("insufficient_balance", ...) when balance is too low.
// Idempotent on (refType, refID).
func (s *Service) Debit(ctx context.Context, userID, refType, refID string, amountPaise int64) (*Transaction, error) {
	if amountPaise <= 0 {
		return nil, utils.NewBadRequest("invalid_amount", "amount must be positive")
	}

	tx, err := s.repo.DB().BeginTx(ctx, nil)
	if err != nil {
		return nil, utils.NewInternal(err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := s.repo.EnsureForUser(ctx, tx, userID); err != nil {
		return nil, utils.NewInternal(err)
	}
	w, err := s.repo.LockForUser(ctx, tx, userID)
	if err != nil {
		return nil, utils.NewInternal(err)
	}
	if w.BalancePaise < amountPaise {
		return nil, utils.NewConflict("insufficient_balance", "wallet balance is insufficient")
	}
	newBalance := w.BalancePaise - amountPaise

	id, err := s.repo.InsertTxn(ctx, tx, w.ID, "debit", amountPaise, refType, refID, newBalance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Duplicate request, already processed.
			if err := tx.Commit(); err != nil {
				return nil, utils.NewInternal(err)
			}
			return nil, nil
		}
		return nil, utils.NewInternal(err)
	}
	if err := s.repo.SetBalance(ctx, tx, w.ID, newBalance); err != nil {
		return nil, utils.NewInternal(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, utils.NewInternal(err)
	}
	return &Transaction{
		ID:                id,
		WalletID:          w.ID,
		Type:              "debit",
		AmountPaise:       amountPaise,
		RefType:           refType,
		RefID:             refID,
		BalanceAfterPaise: newBalance,
	}, nil
}

// ListTxns returns the most recent ledger entries.
func (s *Service) ListTxns(ctx context.Context, userID string) ([]Transaction, error) {
	w, err := s.Balance(ctx, userID)
	if err != nil {
		return nil, err
	}
	out, err := s.repo.ListTxns(ctx, w.ID, 50)
	if err != nil {
		return nil, utils.NewInternal(err)
	}
	return out, nil
}
