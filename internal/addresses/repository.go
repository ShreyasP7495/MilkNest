package addresses

import (
	"context"
	"database/sql"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

// Create inserts an address row. The is_default invariant (at most one
// default per user) is enforced inside a transaction.
func (r *Repository) Create(ctx context.Context, userID string, req UpsertRequest) (*Address, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	if req.IsDefault {
		if _, err := tx.ExecContext(ctx, `UPDATE addresses SET is_default = FALSE WHERE user_id = $1`, userID); err != nil {
			return nil, err
		}
	}

	const q = `
		INSERT INTO addresses (user_id, house_number, street, city, state, pincode, latitude, longitude, is_default)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, user_id, house_number, street, city, state, pincode, latitude, longitude, is_default, created_at`
	a := &Address{}
	if err := tx.QueryRowContext(ctx, q,
		userID, req.HouseNumber, req.Street, req.City, req.State, req.Pincode,
		req.Latitude, req.Longitude, req.IsDefault,
	).Scan(
		&a.ID, &a.UserID, &a.HouseNumber, &a.Street, &a.City, &a.State, &a.Pincode,
		&a.Latitude, &a.Longitude, &a.IsDefault, &a.CreatedAt,
	); err != nil {
		return nil, err
	}
	return a, tx.Commit()
}

func (r *Repository) ListForUser(ctx context.Context, userID string) ([]Address, error) {
	const q = `
		SELECT id, user_id, house_number, street, city, state, pincode, latitude, longitude, is_default, created_at
		  FROM addresses WHERE user_id = $1 ORDER BY is_default DESC, created_at DESC`
	rows, err := r.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Address, 0)
	for rows.Next() {
		var a Address
		if err := rows.Scan(&a.ID, &a.UserID, &a.HouseNumber, &a.Street, &a.City, &a.State, &a.Pincode, &a.Latitude, &a.Longitude, &a.IsDefault, &a.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *Repository) GetForUser(ctx context.Context, userID, id string) (*Address, error) {
	const q = `
		SELECT id, user_id, house_number, street, city, state, pincode, latitude, longitude, is_default, created_at
		  FROM addresses WHERE id = $1 AND user_id = $2`
	a := &Address{}
	err := r.db.QueryRowContext(ctx, q, id, userID).Scan(
		&a.ID, &a.UserID, &a.HouseNumber, &a.Street, &a.City, &a.State, &a.Pincode,
		&a.Latitude, &a.Longitude, &a.IsDefault, &a.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (r *Repository) Update(ctx context.Context, userID, id string, req UpsertRequest) (*Address, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	if req.IsDefault {
		if _, err := tx.ExecContext(ctx, `UPDATE addresses SET is_default = FALSE WHERE user_id = $1 AND id <> $2`, userID, id); err != nil {
			return nil, err
		}
	}

	const q = `
		UPDATE addresses
		   SET house_number = $3, street = $4, city = $5, state = $6, pincode = $7,
		       latitude = $8, longitude = $9, is_default = $10
		 WHERE id = $1 AND user_id = $2
		 RETURNING id, user_id, house_number, street, city, state, pincode, latitude, longitude, is_default, created_at`
	a := &Address{}
	if err := tx.QueryRowContext(ctx, q,
		id, userID, req.HouseNumber, req.Street, req.City, req.State, req.Pincode,
		req.Latitude, req.Longitude, req.IsDefault,
	).Scan(
		&a.ID, &a.UserID, &a.HouseNumber, &a.Street, &a.City, &a.State, &a.Pincode,
		&a.Latitude, &a.Longitude, &a.IsDefault, &a.CreatedAt,
	); err != nil {
		return nil, err
	}
	return a, tx.Commit()
}

func (r *Repository) Delete(ctx context.Context, userID, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM addresses WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}
