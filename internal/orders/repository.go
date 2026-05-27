package orders

import (
	"context"
	"database/sql"
	"time"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) DB() *sql.DB { return r.db }

const baseSelect = `
	SELECT id, subscription_id, user_id, product_id, hub_id, address_id, delivery_date,
	       quantity, amount_paise, status, payment_id, delivery_partner_id, created_at, updated_at
	  FROM orders`

func scanOne(row *sql.Row) (*Order, error) {
	o := &Order{}
	err := row.Scan(
		&o.ID, &o.SubscriptionID, &o.UserID, &o.ProductID, &o.HubID, &o.AddressID, &o.DeliveryDate,
		&o.Quantity, &o.AmountPaise, &o.Status, &o.PaymentID, &o.DeliveryPartnerID, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return o, nil
}

func scanMany(rows *sql.Rows) ([]Order, error) {
	out := make([]Order, 0)
	for rows.Next() {
		var o Order
		if err := rows.Scan(
			&o.ID, &o.SubscriptionID, &o.UserID, &o.ProductID, &o.HubID, &o.AddressID, &o.DeliveryDate,
			&o.Quantity, &o.AmountPaise, &o.Status, &o.PaymentID, &o.DeliveryPartnerID, &o.CreatedAt, &o.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

func (r *Repository) ListForUser(ctx context.Context, userID string) ([]Order, error) {
	rows, err := r.db.QueryContext(ctx, baseSelect+` WHERE user_id = $1 ORDER BY delivery_date DESC, created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMany(rows)
}

func (r *Repository) ListAll(ctx context.Context, limit int) ([]Order, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := r.db.QueryContext(ctx, baseSelect+` ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMany(rows)
}

func (r *Repository) Get(ctx context.Context, id string) (*Order, error) {
	return scanOne(r.db.QueryRowContext(ctx, baseSelect+` WHERE id = $1`, id))
}

// InsertGeneratedOrder is called by the cron. It uses ON CONFLICT DO NOTHING
// against the (subscription_id, delivery_date) unique index for idempotency.
func (r *Repository) InsertGeneratedOrder(ctx context.Context, tx *sql.Tx, sub generatedOrderInput, deliveryDate time.Time) (string, bool, error) {
	const q = `
		INSERT INTO orders (subscription_id, user_id, product_id, hub_id, address_id, delivery_date, quantity, amount_paise, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'pending')
		ON CONFLICT (subscription_id, delivery_date) DO NOTHING
		RETURNING id`
	var id string
	err := tx.QueryRowContext(ctx, q,
		sub.SubscriptionID, sub.UserID, sub.ProductID, sub.HubID, sub.AddressID, deliveryDate, sub.Quantity, sub.AmountPaise,
	).Scan(&id)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return id, true, nil
}

// SetStatus updates order status (used by delivery service).
func (r *Repository) SetStatus(ctx context.Context, id, status string) (*Order, error) {
	const q = `
		UPDATE orders SET status = $2, updated_at = now() WHERE id = $1
		RETURNING id, subscription_id, user_id, product_id, hub_id, address_id, delivery_date,
		          quantity, amount_paise, status, payment_id, delivery_partner_id, created_at, updated_at`
	return scanOne(r.db.QueryRowContext(ctx, q, id, status))
}

// Assign sets delivery_partner_id and status='assigned'. Used by admin.
func (r *Repository) Assign(ctx context.Context, orderID, partnerID string) (*Order, error) {
	const q = `
		UPDATE orders
		   SET delivery_partner_id = $2, status = 'assigned', updated_at = now()
		 WHERE id = $1
		RETURNING id, subscription_id, user_id, product_id, hub_id, address_id, delivery_date,
		          quantity, amount_paise, status, payment_id, delivery_partner_id, created_at, updated_at`
	return scanOne(r.db.QueryRowContext(ctx, q, orderID, partnerID))
}

// AttachPayment links a payment id to an order and marks it paid.
func (r *Repository) AttachPayment(ctx context.Context, orderID, paymentID string) error {
	const q = `UPDATE orders SET payment_id = $2, updated_at = now() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, q, orderID, paymentID)
	return err
}

// generatedOrderInput is the cron-internal struct (kept package-private).
type generatedOrderInput struct {
	SubscriptionID string
	UserID         string
	ProductID      string
	HubID          string
	AddressID      string
	Quantity       int
	AmountPaise    int64
}
