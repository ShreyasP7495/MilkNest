package payments

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/milknest/backend/internal/orders"
	"github.com/milknest/backend/pkg/config"
	"github.com/milknest/backend/pkg/logger"
	"github.com/milknest/backend/pkg/utils"
)

// WalletCrediter is implemented by the wallet service. We accept it as an
// interface to avoid importing the wallet package and creating a cycle (wallet
// also calls into payments to create a Razorpay order for top-ups).
type WalletCrediter interface {
	CreditFromTopup(ctx context.Context, paymentID string, userID string, amountPaise int64) error
}

type Service struct {
	repo   *Repository
	orders *orders.Repository
	gw     Gateway
	cfg    *config.Config
	wallet WalletCrediter
}

func NewService(repo *Repository, ordersRepo *orders.Repository, gw Gateway, cfg *config.Config) *Service {
	return &Service{repo: repo, orders: ordersRepo, gw: gw, cfg: cfg}
}

// SetWallet allows the wallet module to inject itself once both are
// constructed. Avoids the cycle at package init time.
func (s *Service) SetWallet(w WalletCrediter) { s.wallet = w }

// CreateForOrder turns an order into a Razorpay order + payments row.
// Applies the 10% BNPL surcharge defined in HLD §9 / LLD §11.
func (s *Service) CreateForOrder(ctx context.Context, userID string, req CreateRequest) (*CreateResponse, error) {
	order, err := s.orders.Get(ctx, req.OrderID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, utils.NewNotFound("order")
	}
	if err != nil {
		return nil, utils.NewInternal(err)
	}
	if order.UserID != userID {
		return nil, utils.NewForbidden("order does not belong to user")
	}

	amount := order.AmountPaise
	var surcharge int64
	if req.Method == "bnpl" {
		surcharge = (amount * int64(s.cfg.Business.BNPLSurchargePercent)) / 100
		amount += surcharge
	}

	receipt := fmt.Sprintf("ord_%s", order.ID)
	rzpOrderID, err := s.gw.CreateOrder(amount, receipt, map[string]interface{}{
		"order_id": order.ID,
		"user_id":  userID,
	})
	if err != nil {
		return nil, utils.NewInternal(err)
	}

	orderIDPtr := order.ID
	p, err := s.repo.Create(ctx, userID, &orderIDPtr, amount, surcharge, req.Method, rzpOrderID, receipt)
	if err != nil {
		return nil, utils.NewInternal(err)
	}

	return &CreateResponse{
		PaymentID:       p.ID,
		RazorpayOrderID: rzpOrderID,
		AmountPaise:     amount,
		SurchargePaise:  surcharge,
		Currency:        "INR",
		KeyID:           s.cfg.Razorpay.KeyID,
		Status:          p.Status,
	}, nil
}

// CreateForWalletTopup is called by the wallet service. Returns the payments
// row id so the wallet can correlate it with the topup ledger entry on
// webhook receipt.
func (s *Service) CreateForWalletTopup(ctx context.Context, userID string, amountPaise int64) (*CreateResponse, error) {
	receipt := fmt.Sprintf("wal_%s_%d", userID, amountPaise)
	rzpOrderID, err := s.gw.CreateOrder(amountPaise, receipt, map[string]interface{}{
		"user_id": userID,
		"purpose": "wallet_topup",
	})
	if err != nil {
		return nil, utils.NewInternal(err)
	}
	p, err := s.repo.Create(ctx, userID, nil, amountPaise, 0, "upi", rzpOrderID, receipt)
	if err != nil {
		return nil, utils.NewInternal(err)
	}
	return &CreateResponse{
		PaymentID:       p.ID,
		RazorpayOrderID: rzpOrderID,
		AmountPaise:     amountPaise,
		Currency:        "INR",
		KeyID:           s.cfg.Razorpay.KeyID,
		Status:          p.Status,
	}, nil
}

func (s *Service) Get(ctx context.Context, id string) (*Payment, error) {
	p, err := s.repo.Get(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, utils.NewNotFound("payment")
	}
	if err != nil {
		return nil, utils.NewInternal(err)
	}
	return p, nil
}

// HandleWebhook verifies the signature, then marks the matching payment as
// paid and advances the linked order (or credits the wallet for top-ups).
func (s *Service) HandleWebhook(ctx context.Context, body []byte, signature string) error {
	if err := s.gw.VerifyWebhookSignature(body, signature); err != nil {
		return utils.NewBadRequest("invalid_signature", err.Error())
	}

	var evt WebhookEvent
	if err := json.Unmarshal(body, &evt); err != nil {
		return utils.NewBadRequest("invalid_body", "could not parse webhook payload")
	}

	// We only react to the 'captured' event for MVP. Failures advance the
	// payment row only in production once we wire payment.failed.
	if evt.Event != "payment.captured" {
		logger.L().Debug().Str("event", evt.Event).Msg("webhook event ignored")
		return nil
	}

	rzpOrderID, rzpPaymentID := extractIDs(evt)
	if rzpOrderID == "" || rzpPaymentID == "" {
		return utils.NewBadRequest("invalid_payload", "missing razorpay ids")
	}

	pay, err := s.repo.FindByRazorpayOrder(ctx, rzpOrderID)
	if errors.Is(err, sql.ErrNoRows) {
		logger.L().Warn().Str("rzp_order_id", rzpOrderID).Msg("payment not found for webhook")
		return nil // idempotent: ignore unknown orders
	}
	if err != nil {
		return utils.NewInternal(err)
	}

	if err := s.repo.MarkPaid(ctx, pay.ID, rzpPaymentID); err != nil {
		return utils.NewInternal(err)
	}

	switch {
	case pay.OrderID != nil:
		if err := s.orders.AttachPayment(ctx, *pay.OrderID, pay.ID); err != nil {
			logger.L().Warn().Err(err).Msg("attach payment to order failed")
		}
	case s.wallet != nil:
		if err := s.wallet.CreditFromTopup(ctx, pay.ID, pay.UserID, pay.AmountPaise); err != nil {
			logger.L().Error().Err(err).Msg("wallet credit failed on topup")
			return utils.NewInternal(err)
		}
	}
	return nil
}

// extractIDs digs the razorpay order/payment ids out of the webhook payload.
// Shape: payload.payment.entity.{id, order_id}
func extractIDs(evt WebhookEvent) (orderID, paymentID string) {
	payment, ok := evt.Payload["payment"].(map[string]interface{})
	if !ok {
		return "", ""
	}
	entity, ok := payment["entity"].(map[string]interface{})
	if !ok {
		return "", ""
	}
	if v, ok := entity["order_id"].(string); ok {
		orderID = v
	}
	if v, ok := entity["id"].(string); ok {
		paymentID = v
	}
	return orderID, paymentID
}
