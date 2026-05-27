package subscriptions

import (
	"context"
	"database/sql"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

// DB returns the underlying handle so the service can open transactions that
// span subscriptions + inventory updates.
func (r *Repository) DB() *sql.DB { return r.db }

// SumActiveQuantity returns the sum of active subscription quantities for
// this user - used to enforce LLD §6's "max 10 packets per household per day".
func (r *Repository) SumActiveQuantity(ctx context.Context, tx *sql.Tx, userID string) (int, error) {
	const q = `
		SELECT COALESCE(SUM(quantity), 0)
		  FROM subscriptions WHERE user_id = $1 AND status = 'active'`
	var total int
	if err := tx.QueryRowContext(ctx, q, userID).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *Repository) Create(ctx context.Context, tx *sql.Tx, userID string, req CreateRequest, freeQty int) (*Subscription, error) {
	const q = `
		INSERT INTO subscriptions (user_id, product_id, hub_id, address_id, quantity, free_quantity, frequency, delivery_time, start_date, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 'active')
		RETURNING id, user_id, product_id, hub_id, address_id, quantity, free_quantity, frequency, delivery_time, start_date, status, created_at, updated_at`
	s := &Subscription{}
	err := tx.QueryRowContext(ctx, q,
		userID, req.ProductID, req.HubID, req.AddressID, req.Quantity, freeQty, req.Frequency, req.DeliveryTime, req.StartDate,
	).Scan(
		&s.ID, &s.UserID, &s.ProductID, &s.HubID, &s.AddressID, &s.Quantity, &s.FreeQuantity,
		&s.Frequency, &s.DeliveryTime, &s.StartDate, &s.Status, &s.CreatedAt, &s.UpdatedAt,
	)
	return s, err
}

func (r *Repository) ListForUser(ctx context.Context, userID string) ([]Subscription, error) {
	const q = `
		SELECT id, user_id, product_id, hub_id, address_id, quantity, free_quantity, frequency, delivery_time, start_date, status, created_at, updated_at
		  FROM subscriptions WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Subscription, 0)
	for rows.Next() {
		var s Subscription
		if err := rows.Scan(&s.ID, &s.UserID, &s.ProductID, &s.HubID, &s.AddressID, &s.Quantity, &s.FreeQuantity, &s.Frequency, &s.DeliveryTime, &s.StartDate, &s.Status, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *Repository) GetForUser(ctx context.Context, userID, id string) (*Subscription, error) {
	const q = `
		SELECT id, user_id, product_id, hub_id, address_id, quantity, free_quantity, frequency, delivery_time, start_date, status, created_at, updated_at
		  FROM subscriptions WHERE id = $1 AND user_id = $2`
	s := &Subscription{}
	err := r.db.QueryRowContext(ctx, q, id, userID).Scan(
		&s.ID, &s.UserID, &s.ProductID, &s.HubID, &s.AddressID, &s.Quantity, &s.FreeQuantity,
		&s.Frequency, &s.DeliveryTime, &s.StartDate, &s.Status, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// Update changes quantity / delivery_time / address_id. Quantity delta is
// returned so the service can adjust the inventory reservation.
func (r *Repository) Update(ctx context.Context, tx *sql.Tx, userID, id string, req UpdateRequest) (*Subscription, int, error) {
	current, err := r.GetForUser(ctx, userID, id)
	if err != nil {
		return nil, 0, err
	}
	newQty := current.Quantity
	if req.Quantity != nil {
		newQty = *req.Quantity
	}
	deliveryTime := current.DeliveryTime
	if req.DeliveryTime != nil {
		deliveryTime = *req.DeliveryTime
	}
	addressID := current.AddressID
	if req.AddressID != nil {
		addressID = *req.AddressID
	}

	const q = `
		UPDATE subscriptions
		   SET quantity = $3, delivery_time = $4, address_id = $5, updated_at = now()
		 WHERE id = $1 AND user_id = $2
		RETURNING id, user_id, product_id, hub_id, address_id, quantity, free_quantity, frequency, delivery_time, start_date, status, created_at, updated_at`
	s := &Subscription{}
	err = tx.QueryRowContext(ctx, q, id, userID, newQty, deliveryTime, addressID).Scan(
		&s.ID, &s.UserID, &s.ProductID, &s.HubID, &s.AddressID, &s.Quantity, &s.FreeQuantity,
		&s.Frequency, &s.DeliveryTime, &s.StartDate, &s.Status, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, 0, err
	}
	return s, newQty - current.Quantity, nil
}

// SetStatus is used by pause/cancel handlers.
func (r *Repository) SetStatus(ctx context.Context, tx *sql.Tx, userID, id, status string) (*Subscription, error) {
	const q = `
		UPDATE subscriptions
		   SET status = $3, updated_at = now()
		 WHERE id = $1 AND user_id = $2
		RETURNING id, user_id, product_id, hub_id, address_id, quantity, free_quantity, frequency, delivery_time, start_date, status, created_at, updated_at`
	s := &Subscription{}
	err := tx.QueryRowContext(ctx, q, id, userID, status).Scan(
		&s.ID, &s.UserID, &s.ProductID, &s.HubID, &s.AddressID, &s.Quantity, &s.FreeQuantity,
		&s.Frequency, &s.DeliveryTime, &s.StartDate, &s.Status, &s.CreatedAt, &s.UpdatedAt,
	)
	return s, err
}

// ListActiveAll is used by the order-generation cron.
func (r *Repository) ListActiveAll(ctx context.Context) ([]Subscription, error) {
	const q = `
		SELECT id, user_id, product_id, hub_id, address_id, quantity, free_quantity, frequency, delivery_time, start_date, status, created_at, updated_at
		  FROM subscriptions WHERE status = 'active'`
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Subscription, 0)
	for rows.Next() {
		var s Subscription
		if err := rows.Scan(&s.ID, &s.UserID, &s.ProductID, &s.HubID, &s.AddressID, &s.Quantity, &s.FreeQuantity, &s.Frequency, &s.DeliveryTime, &s.StartDate, &s.Status, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}
