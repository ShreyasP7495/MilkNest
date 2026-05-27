package payments

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	rzp "github.com/razorpay/razorpay-go"

	"github.com/milknest/backend/pkg/config"
)

// Gateway is the narrow interface the service depends on so it can be
// swapped/mocked in tests.
type Gateway interface {
	CreateOrder(amountPaise int64, receipt string, notes map[string]interface{}) (string, error)
	VerifyWebhookSignature(payload []byte, signature string) error
}

// RealRazorpay wraps the official SDK and the webhook secret.
type RealRazorpay struct {
	client        *rzp.Client
	webhookSecret string
}

// NewRazorpay returns a gateway implementation. If the key id/secret are
// blank we return a stub that succeeds so the rest of the app stays runnable.
func NewRazorpay(cfg config.RazorpayConfig) Gateway {
	if cfg.KeyID == "" || cfg.KeySecret == "" {
		return &StubRazorpay{}
	}
	return &RealRazorpay{
		client:        rzp.NewClient(cfg.KeyID, cfg.KeySecret),
		webhookSecret: cfg.WebhookSecret,
	}
}

func (r *RealRazorpay) CreateOrder(amountPaise int64, receipt string, notes map[string]interface{}) (string, error) {
	body := map[string]interface{}{
		"amount":   amountPaise,
		"currency": "INR",
		"receipt":  receipt,
	}
	if notes != nil {
		body["notes"] = notes
	}
	out, err := r.client.Order.Create(body, nil)
	if err != nil {
		return "", fmt.Errorf("razorpay create order: %w", err)
	}
	id, _ := out["id"].(string)
	if id == "" {
		return "", errors.New("razorpay create order: empty id in response")
	}
	return id, nil
}

func (r *RealRazorpay) VerifyWebhookSignature(payload []byte, signature string) error {
	if r.webhookSecret == "" {
		return errors.New("RAZORPAY_WEBHOOK_SECRET is not configured")
	}
	mac := hmac.New(sha256.New, []byte(r.webhookSecret))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(signature)) {
		return errors.New("invalid signature")
	}
	return nil
}

// StubRazorpay is used in development when credentials are not yet provided.
// It produces deterministic-ish order ids and accepts every signature so the
// flow remains testable without real Razorpay calls.
type StubRazorpay struct{}

func (StubRazorpay) CreateOrder(amountPaise int64, receipt string, _ map[string]interface{}) (string, error) {
	return "stub_order_" + receipt, nil
}

func (StubRazorpay) VerifyWebhookSignature(_ []byte, _ string) error { return nil }
