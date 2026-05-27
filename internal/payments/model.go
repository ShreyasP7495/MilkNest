package payments

import "time"

type Payment struct {
	ID                string    `json:"id"`
	OrderID           *string   `json:"order_id,omitempty"`
	UserID            string    `json:"user_id"`
	AmountPaise       int64     `json:"amount_paise"`
	SurchargePaise    int64     `json:"surcharge_paise"`
	Method            string    `json:"method"`
	RazorpayOrderID   string    `json:"razorpay_order_id,omitempty"`
	RazorpayPaymentID string    `json:"razorpay_payment_id,omitempty"`
	Status            string    `json:"status"`
	IdempotencyKey    string    `json:"-"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// CreateRequest matches API Spec §8.1.
type CreateRequest struct {
	OrderID string `json:"order_id" validate:"required,uuid"`
	Method  string `json:"method"   validate:"required,oneof=upi card netbanking bnpl wallet"`
}

type CreateResponse struct {
	PaymentID       string `json:"payment_id"`
	RazorpayOrderID string `json:"razorpay_order_id,omitempty"`
	AmountPaise     int64  `json:"amount_paise"`
	SurchargePaise  int64  `json:"surcharge_paise"`
	Currency        string `json:"currency"`
	KeyID           string `json:"key_id,omitempty"`
	Status          string `json:"status"`
}

// WebhookEvent is the slice of the Razorpay webhook payload we use.
type WebhookEvent struct {
	Event   string                 `json:"event"`
	Payload map[string]interface{} `json:"payload"`
}
