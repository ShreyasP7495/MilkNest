package inventory

import (
	"context"
	"database/sql"
	"errors"

	"github.com/milknest/backend/pkg/utils"
)

type Service struct{ Repo *Repository }

func NewService(repo *Repository) *Service { return &Service{Repo: repo} }

// Reserve is the public service-level reservation that opens its own tx. For
// callers that already hold a tx (e.g. subscription create), use Repo.Reserve
// directly.
func (s *Service) Reserve(ctx context.Context, db *sql.DB, productID, hubID string, qty int) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return utils.NewInternal(err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := s.Repo.Reserve(ctx, tx, productID, hubID, qty); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return utils.NewConflict("insufficient_stock", "insufficient stock at hub")
		}
		return utils.NewInternal(err)
	}
	return tx.Commit()
}

// Update sets absolute quantity (admin only).
func (s *Service) Update(ctx context.Context, req UpdateRequest) (*Inventory, error) {
	inv, err := s.Repo.AdminUpsert(ctx, req)
	if err != nil {
		return nil, utils.NewInternal(err)
	}
	return inv, nil
}

func (s *Service) List(ctx context.Context) ([]Inventory, error) {
	out, err := s.Repo.List(ctx)
	if err != nil {
		return nil, utils.NewInternal(err)
	}
	return out, nil
}
