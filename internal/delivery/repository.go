package delivery

import (
	"context"
	"database/sql"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

// Append writes a delivery_tracking row.
func (r *Repository) Append(ctx context.Context, orderID, partnerID string, lat, lng float64, status string) (*Tracking, error) {
	const q = `
		INSERT INTO delivery_tracking (order_id, delivery_partner_id, latitude, longitude, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, order_id, delivery_partner_id, latitude, longitude, status, recorded_at`
	t := &Tracking{}
	err := r.db.QueryRowContext(ctx, q, orderID, partnerID, lat, lng, status).Scan(
		&t.ID, &t.OrderID, &t.DeliveryPartnerID, &t.Latitude, &t.Longitude, &t.Status, &t.RecordedAt,
	)
	return t, err
}

// LatestForOrder returns the most recent tracking row for an order; used as
// a fallback when the Redis TTL expires.
func (r *Repository) LatestForOrder(ctx context.Context, orderID string) (*Tracking, error) {
	const q = `
		SELECT id, order_id, delivery_partner_id, latitude, longitude, status, recorded_at
		  FROM delivery_tracking
		 WHERE order_id = $1
		 ORDER BY recorded_at DESC
		 LIMIT 1`
	t := &Tracking{}
	err := r.db.QueryRowContext(ctx, q, orderID).Scan(
		&t.ID, &t.OrderID, &t.DeliveryPartnerID, &t.Latitude, &t.Longitude, &t.Status, &t.RecordedAt,
	)
	return t, err
}
