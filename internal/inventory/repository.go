package inventory

import (
	"context"
	"database/sql"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

// DBTX is the subset of *sql.DB / *sql.Tx that all our queries need. Lets the
// service pass a transaction in for atomic reservation.
type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// Reserve atomically reserves `qty` units of `productID` at `hubID`. It
// returns sql.ErrNoRows if no row had enough free stock (caller maps to a
// business error).
func (r *Repository) Reserve(ctx context.Context, tx DBTX, productID, hubID string, qty int) error {
	const q = `
		UPDATE inventory
		   SET reserved_quantity = reserved_quantity + $3,
		       updated_at = now()
		 WHERE product_id = $1
		   AND hub_id     = $2
		   AND quantity - reserved_quantity >= $3`
	res, err := tx.ExecContext(ctx, q, productID, hubID, qty)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// Release decrements reserved_quantity by qty (used on subscription cancel).
func (r *Repository) Release(ctx context.Context, tx DBTX, productID, hubID string, qty int) error {
	const q = `
		UPDATE inventory
		   SET reserved_quantity = GREATEST(reserved_quantity - $3, 0),
		       updated_at = now()
		 WHERE product_id = $1 AND hub_id = $2`
	_, err := tx.ExecContext(ctx, q, productID, hubID, qty)
	return err
}

// Consume removes qty from both quantity and reserved_quantity. Used by the
// nightly order-generation cron when reservation becomes a real deduction.
func (r *Repository) Consume(ctx context.Context, tx DBTX, productID, hubID string, qty int) error {
	const q = `
		UPDATE inventory
		   SET quantity          = quantity - $3,
		       reserved_quantity = GREATEST(reserved_quantity - $3, 0),
		       updated_at        = now()
		 WHERE product_id = $1
		   AND hub_id     = $2
		   AND quantity >= $3`
	res, err := tx.ExecContext(ctx, q, productID, hubID, qty)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// AdminUpsert sets quantity to an absolute value, creating the row if needed.
func (r *Repository) AdminUpsert(ctx context.Context, req UpdateRequest) (*Inventory, error) {
	const q = `
		INSERT INTO inventory (product_id, hub_id, quantity, reserved_quantity)
		VALUES ($1, $2, $3, 0)
		ON CONFLICT (product_id, hub_id)
		  DO UPDATE SET quantity = EXCLUDED.quantity, updated_at = now()
		RETURNING id, product_id, hub_id, quantity, reserved_quantity, updated_at`
	inv := &Inventory{}
	err := r.db.QueryRowContext(ctx, q, req.ProductID, req.HubID, req.Quantity).Scan(
		&inv.ID, &inv.ProductID, &inv.HubID, &inv.Quantity, &inv.ReservedQuantity, &inv.UpdatedAt,
	)
	return inv, err
}

// List returns all inventory rows (admin view).
func (r *Repository) List(ctx context.Context) ([]Inventory, error) {
	const q = `
		SELECT id, product_id, hub_id, quantity, reserved_quantity, updated_at
		  FROM inventory ORDER BY updated_at DESC`
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Inventory, 0)
	for rows.Next() {
		var i Inventory
		if err := rows.Scan(&i.ID, &i.ProductID, &i.HubID, &i.Quantity, &i.ReservedQuantity, &i.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, i)
	}
	return out, rows.Err()
}
