package notifications

import "time"

type Notification struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	Channel   string     `json:"channel"` // sms | push
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	Status    string     `json:"status"`
	SentAt    *time.Time `json:"sent_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

type SendRequest struct {
	UserID  string `json:"user_id"  validate:"required,uuid"`
	Channel string `json:"channel"  validate:"required,oneof=sms push"`
	Title   string `json:"title"    validate:"required"`
	Body    string `json:"body"     validate:"required"`
}
