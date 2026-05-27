package admin

import (
	"context"
	"database/sql"

	"github.com/milknest/backend/pkg/utils"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

// Dashboard runs the four aggregate queries needed by API Spec §12.1.
func (r *Repository) Dashboard(ctx context.Context) (*DashboardStats, error) {
	s := &DashboardStats{}
	if err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM users WHERE role = 'customer'`).Scan(&s.TotalUsers); err != nil {
		return nil, err
	}
	if err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM subscriptions WHERE status = 'active'`).Scan(&s.ActiveSubscriptions); err != nil {
		return nil, err
	}
	today := utils.NowIST().Format("2006-01-02")
	if err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM orders WHERE delivery_date = $1`, today).Scan(&s.DailyOrders); err != nil {
		return nil, err
	}
	if err := r.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(amount_paise), 0) FROM payments WHERE status = 'paid'`).Scan(&s.RevenuePaise); err != nil {
		return nil, err
	}
	s.Revenue = float64(s.RevenuePaise) / 100
	return s, nil
}
