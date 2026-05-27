package complaints

import (
	"context"
	"database/sql"
	"errors"

	"github.com/milknest/backend/pkg/utils"
)

type Service struct{ repo *Repository }

func NewService(repo *Repository) *Service { return &Service{repo: repo} }

func (s *Service) Create(ctx context.Context, userID string, req CreateRequest) (*Complaint, error) {
	c, err := s.repo.Create(ctx, userID, req)
	if err != nil {
		return nil, utils.NewInternal(err)
	}
	return c, nil
}

func (s *Service) ListForUser(ctx context.Context, userID string) ([]Complaint, error) {
	out, err := s.repo.ListForUser(ctx, userID)
	if err != nil {
		return nil, utils.NewInternal(err)
	}
	return out, nil
}

func (s *Service) ListAll(ctx context.Context) ([]Complaint, error) {
	out, err := s.repo.ListAll(ctx)
	if err != nil {
		return nil, utils.NewInternal(err)
	}
	return out, nil
}

func (s *Service) Resolve(ctx context.Context, id string, req ResolveRequest) (*Complaint, error) {
	c, err := s.repo.Resolve(ctx, id, req.Note)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, utils.NewNotFound("complaint")
	}
	if err != nil {
		return nil, utils.NewInternal(err)
	}
	return c, nil
}
