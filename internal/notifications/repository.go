package notifications

import (
	"context"
	"database/sql"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

// Log persists a notification row and returns the assigned id.
func (r *Repository) Log(ctx context.Context, userID, channel, title, body, status string) (string, error) {
	const q = `
		INSERT INTO notifications (user_id, channel, title, body, status, sent_at)
		VALUES ($1, $2, $3, $4, $5, CASE WHEN $5 = 'sent' THEN now() ELSE NULL END)
		RETURNING id`
	var id string
	err := r.db.QueryRowContext(ctx, q, userID, channel, title, body, status).Scan(&id)
	return id, err
}

// ListForUser returns the most recent notifications for a user.
func (r *Repository) ListForUser(ctx context.Context, userID string, limit int) ([]Notification, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	const q = `
		SELECT id, user_id, channel, title, body, status, sent_at, created_at
		  FROM notifications
		 WHERE user_id = $1
		 ORDER BY created_at DESC
		 LIMIT $2`
	rows, err := r.db.QueryContext(ctx, q, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Notification, 0, limit)
	for rows.Next() {
		var n Notification
		if err := rows.Scan(&n.ID, &n.UserID, &n.Channel, &n.Title, &n.Body, &n.Status, &n.SentAt, &n.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}
