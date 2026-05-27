package payments

import (
	"context"
	"database/sql"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

// Create inserts a payments row in `created` state. idempotencyKey may be
// nil/empty.
func (r *Repository) Create(ctx context.Context, userID string, orderID *string, amount, surcharge int64, method, rzpOrderID, idemKey string) (*Payment, error) {
	const q = `
		INSERT INTO payments (user_id, order_id, amount_paise, surcharge_paise, method, razorpay_order_id, idempotency_key, status)
		VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7,''), 'created')
		RETURNING id, order_id, user_id, amount_paise, surcharge_paise, method, COALESCE(razorpay_order_id,''),
		          COALESCE(razorpay_payment_id,''), status, created_at, updated_at`
	p := &Payment{}
	err := r.db.QueryRowContext(ctx, q, userID, orderID, amount, surcharge, method, rzpOrderID, idemKey).Scan(
		&p.ID, &p.OrderID, &p.UserID, &p.AmountPaise, &p.SurchargePaise, &p.Method,
		&p.RazorpayOrderID, &p.RazorpayPaymentID, &p.Status, &p.CreatedAt, &p.UpdatedAt,
	)
	return p, err
}

func (r *Repository) Get(ctx context.Context, id string) (*Payment, error) {
	const q = `
		SELECT id, order_id, user_id, amount_paise, surcharge_paise, method,
		       COALESCE(razorpay_order_id,''), COALESCE(razorpay_payment_id,''), status, created_at, updated_at
		  FROM payments WHERE id = $1`
	p := &Payment{}
	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&p.ID, &p.OrderID, &p.UserID, &p.AmountPaise, &p.SurchargePaise, &p.Method,
		&p.RazorpayOrderID, &p.RazorpayPaymentID, &p.Status, &p.CreatedAt, &p.UpdatedAt,
	)
	return p, err
}

func (r *Repository) FindByRazorpayOrder(ctx context.Context, rzpOrderID string) (*Payment, error) {
	const q = `
		SELECT id, order_id, user_id, amount_paise, surcharge_paise, method,
		       COALESCE(razorpay_order_id,''), COALESCE(razorpay_payment_id,''), status, created_at, updated_at
		  FROM payments WHERE razorpay_order_id = $1`
	p := &Payment{}
	err := r.db.QueryRowContext(ctx, q, rzpOrderID).Scan(
		&p.ID, &p.OrderID, &p.UserID, &p.AmountPaise, &p.SurchargePaise, &p.Method,
		&p.RazorpayOrderID, &p.RazorpayPaymentID, &p.Status, &p.CreatedAt, &p.UpdatedAt,
	)
	return p, err
}

// MarkPaid sets razorpay_payment_id and status='paid'. Idempotent on
// razorpay_payment_id UNIQUE constraint.
func (r *Repository) MarkPaid(ctx context.Context, paymentID, rzpPaymentID string) error {
	const q = `
		UPDATE payments
		   SET razorpay_payment_id = $2, status = 'paid', updated_at = now()
		 WHERE id = $1 AND status <> 'paid'`
	_, err := r.db.ExecContext(ctx, q, paymentID, rzpPaymentID)
	return err
}
