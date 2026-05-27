package products

import (
	"context"
	"database/sql"
	"errors"

	"github.com/milknest/backend/pkg/utils"
)

type Service struct{ repo *Repository }

func NewService(repo *Repository) *Service { return &Service{repo: repo} }

func (s *Service) List(ctx context.Context) ([]Product, error) {
	out, err := s.repo.List(ctx)
	if err != nil {
		return nil, utils.NewInternal(err)
	}
	return out, nil
}

func (s *Service) Get(ctx context.Context, id string) (*Product, error) {
	p, err := s.repo.Get(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, utils.NewNotFound("product")
	}
	if err != nil {
		return nil, utils.NewInternal(err)
	}
	return p, nil
}
