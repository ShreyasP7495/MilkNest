package complaints

import (
	"context"
	"database/sql"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func scanOne(row *sql.Row) (*Complaint, error) {
	c := &Complaint{}
	err := row.Scan(&c.ID, &c.UserID, &c.OrderID, &c.Type, &c.Description, &c.Status, &c.ResolutionNote, &c.CreatedAt, &c.ResolvedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

const baseSelect = `
	SELECT id, user_id, order_id, type, description, status, resolution_note, created_at, resolved_at
	  FROM complaints`

func (r *Repository) Create(ctx context.Context, userID string, req CreateRequest) (*Complaint, error) {
	const q = `
		INSERT INTO complaints (user_id, order_id, type, description, status)
		VALUES ($1, NULLIF($2,''), $3, $4, 'open')
		RETURNING id, user_id, order_id, type, description, status, resolution_note, created_at, resolved_at`
	return scanOne(r.db.QueryRowContext(ctx, q, userID, req.OrderID, req.Type, req.Description))
}

func (r *Repository) ListForUser(ctx context.Context, userID string) ([]Complaint, error) {
	rows, err := r.db.QueryContext(ctx, baseSelect+` WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMany(rows)
}

func (r *Repository) ListAll(ctx context.Context) ([]Complaint, error) {
	rows, err := r.db.QueryContext(ctx, baseSelect+` ORDER BY created_at DESC LIMIT 200`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMany(rows)
}

func (r *Repository) Resolve(ctx context.Context, id, note string) (*Complaint, error) {
	const q = `
		UPDATE complaints
		   SET status = 'resolved', resolution_note = $2, resolved_at = now()
		 WHERE id = $1
		RETURNING id, user_id, order_id, type, description, status, resolution_note, created_at, resolved_at`
	return scanOne(r.db.QueryRowContext(ctx, q, id, note))
}

func scanMany(rows *sql.Rows) ([]Complaint, error) {
	out := make([]Complaint, 0)
	for rows.Next() {
		var c Complaint
		if err := rows.Scan(&c.ID, &c.UserID, &c.OrderID, &c.Type, &c.Description, &c.Status, &c.ResolutionNote, &c.CreatedAt, &c.ResolvedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
