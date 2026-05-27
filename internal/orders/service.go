package orders

import (
	"context"
	"database/sql"
	"errors"

	"github.com/milknest/backend/pkg/middleware"
	"github.com/milknest/backend/pkg/utils"
)

type Service struct {
	Repo *Repository
}

func NewService(repo *Repository) *Service { return &Service{Repo: repo} }

func (s *Service) ListForUser(ctx context.Context, userID string) ([]Order, error) {
	out, err := s.Repo.ListForUser(ctx, userID)
	if err != nil {
		return nil, utils.NewInternal(err)
	}
	return out, nil
}

func (s *Service) ListAll(ctx context.Context) ([]Order, error) {
	out, err := s.Repo.ListAll(ctx, 100)
	if err != nil {
		return nil, utils.NewInternal(err)
	}
	return out, nil
}

func (s *Service) Get(ctx context.Context, callerID, callerRole, id string) (*Order, error) {
	o, err := s.Repo.Get(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, utils.NewNotFound("order")
	}
	if err != nil {
		return nil, utils.NewInternal(err)
	}
	// Customers can only view their own orders. Admins/delivery see all.
	if callerRole == middleware.RoleCustomer && o.UserID != callerID {
		return nil, utils.NewForbidden("order belongs to another user")
	}
	return o, nil
}
