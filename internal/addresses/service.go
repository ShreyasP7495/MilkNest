package addresses

import (
	"context"
	"database/sql"
	"errors"

	"github.com/milknest/backend/pkg/utils"
)

type Service struct{ repo *Repository }

func NewService(repo *Repository) *Service { return &Service{repo: repo} }

func (s *Service) Create(ctx context.Context, userID string, req UpsertRequest) (*Address, error) {
	a, err := s.repo.Create(ctx, userID, req)
	if err != nil {
		return nil, utils.NewInternal(err)
	}
	return a, nil
}

func (s *Service) List(ctx context.Context, userID string) ([]Address, error) {
	out, err := s.repo.ListForUser(ctx, userID)
	if err != nil {
		return nil, utils.NewInternal(err)
	}
	return out, nil
}

func (s *Service) Get(ctx context.Context, userID, id string) (*Address, error) {
	a, err := s.repo.GetForUser(ctx, userID, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, utils.NewNotFound("address")
	}
	if err != nil {
		return nil, utils.NewInternal(err)
	}
	return a, nil
}

func (s *Service) Update(ctx context.Context, userID, id string, req UpsertRequest) (*Address, error) {
	a, err := s.repo.Update(ctx, userID, id, req)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, utils.NewNotFound("address")
	}
	if err != nil {
		return nil, utils.NewInternal(err)
	}
	return a, nil
}

func (s *Service) Delete(ctx context.Context, userID, id string) error {
	err := s.repo.Delete(ctx, userID, id)
	if errors.Is(err, sql.ErrNoRows) {
		return utils.NewNotFound("address")
	}
	if err != nil {
		return utils.NewInternal(err)
	}
	return nil
}
