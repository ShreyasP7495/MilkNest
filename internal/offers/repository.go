package offers

import (
	"context"
	"database/sql"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

// ListActive returns active offers sorted by min_quantity ASC.
func (r *Repository) ListActive(ctx context.Context) ([]Offer, error) {
	const q = `
		SELECT id, name, min_quantity, free_quantity, active, created_at
		  FROM offers WHERE active = TRUE ORDER BY min_quantity ASC`
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Offer, 0)
	for rows.Next() {
		var o Offer
		if err := rows.Scan(&o.ID, &o.Name, &o.MinQuantity, &o.FreeQuantity, &o.Active, &o.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

// HighestApplicable returns the offer with the largest min_quantity that is
// still <= packets. Returns (nil, sql.ErrNoRows) when none applies.
func (r *Repository) HighestApplicable(ctx context.Context, packets int) (*Offer, error) {
	const q = `
		SELECT id, name, min_quantity, free_quantity, active, created_at
		  FROM offers
		 WHERE active = TRUE AND min_quantity <= $1
		 ORDER BY min_quantity DESC
		 LIMIT 1`
	o := &Offer{}
	err := r.db.QueryRowContext(ctx, q, packets).Scan(
		&o.ID, &o.Name, &o.MinQuantity, &o.FreeQuantity, &o.Active, &o.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return o, nil
}
