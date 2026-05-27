package admin

import (
	"context"

	"github.com/milknest/backend/pkg/utils"
)

type Service struct{ repo *Repository }

func NewService(repo *Repository) *Service { return &Service{repo: repo} }

func (s *Service) Dashboard(ctx context.Context) (*DashboardStats, error) {
	stats, err := s.repo.Dashboard(ctx)
	if err != nil {
		return nil, utils.NewInternal(err)
	}
	return stats, nil
}
