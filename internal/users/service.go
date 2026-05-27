package users

import (
	"context"
	"database/sql"
	"errors"

	"github.com/milknest/backend/internal/auth"
	"github.com/milknest/backend/pkg/utils"
)

// Service is a thin wrapper around the auth.Repository helpers, kept separate
// so users-related logic can grow (e.g. KYC) without bloating auth.
type Service struct {
	repo *auth.Repository
}

func NewService(repo *auth.Repository) *Service { return &Service{repo: repo} }

func (s *Service) Get(ctx context.Context, id string) (*auth.User, error) {
	u, err := s.repo.FindByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, utils.NewNotFound("user")
	}
	if err != nil {
		return nil, utils.NewInternal(err)
	}
	return u, nil
}

func (s *Service) Update(ctx context.Context, id string, req UpdateProfileRequest) (*auth.User, error) {
	var email *string
	if req.Email != "" {
		email = &req.Email
	}
	u, err := s.repo.UpdateProfile(ctx, id, req.Name, email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, utils.NewNotFound("user")
	}
	if err != nil {
		return nil, utils.NewInternal(err)
	}
	return u, nil
}
