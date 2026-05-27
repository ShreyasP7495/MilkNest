package offers

import (
	"context"
	"database/sql"
	"errors"

	"github.com/milknest/backend/pkg/utils"
)

type Service struct{ repo *Repository }

func NewService(repo *Repository) *Service { return &Service{repo: repo} }

func (s *Service) List(ctx context.Context) ([]Offer, error) {
	out, err := s.repo.ListActive(ctx)
	if err != nil {
		return nil, utils.NewInternal(err)
	}
	return out, nil
}

// CalculateFree implements LLD §12. Returns 0 when no offer applies.
func (s *Service) CalculateFree(ctx context.Context, packets int) (int, error) {
	o, err := s.repo.HighestApplicable(ctx, packets)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, utils.NewInternal(err)
	}
	return o.FreeQuantity, nil
}
