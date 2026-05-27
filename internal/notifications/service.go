package notifications

import (
	"context"
	"errors"

	"github.com/milknest/backend/pkg/config"
	"github.com/milknest/backend/pkg/logger"
)

// Sender is the narrow interface other modules depend on (e.g. auth for OTP
// delivery). Keeping it small means the real provider (Twilio/MSG91/FCM) can
// be swapped without touching callers.
type Sender interface {
	SendSMS(ctx context.Context, phone, body string) error
	SendPush(ctx context.Context, userID, title, body string) error
}

// Service implements Sender. In MVP we log payloads instead of calling a real
// provider; the real implementation goes into provider_*.go files.
type Service struct {
	cfg  *config.Config
	repo *Repository
}

func NewService(cfg *config.Config, repo *Repository) *Service {
	return &Service{cfg: cfg, repo: repo}
}

// SendSMS implements Sender.SendSMS. It logs the message and, in production,
// would call Twilio/MSG91 here. We do not persist SMS rows tied to a user
// because OTP SMS is sent before the user has a row.
func (s *Service) SendSMS(ctx context.Context, phone, body string) error {
	if phone == "" || body == "" {
		return errors.New("phone and body required")
	}
	logger.L().Info().
		Str("provider", s.cfg.SMS.Provider).
		Str("to", phone).
		Str("body", body).
		Msg("[stub] sms send")
	return nil
}

// SendPush logs the push payload and stores a notifications row for history.
func (s *Service) SendPush(ctx context.Context, userID, title, body string) error {
	logger.L().Info().
		Str("user_id", userID).
		Str("title", title).
		Str("body", body).
		Msg("[stub] fcm push")
	if _, err := s.repo.Log(ctx, userID, "push", title, body, "sent"); err != nil {
		return err
	}
	return nil
}

// SendInternal is invoked from POST /notifications/send by other services.
// Channel must be 'sms' or 'push'.
func (s *Service) SendInternal(ctx context.Context, req SendRequest) (string, error) {
	switch req.Channel {
	case "push":
		return "", s.SendPush(ctx, req.UserID, req.Title, req.Body)
	case "sms":
		return "", s.SendSMS(ctx, req.UserID, req.Body)
	default:
		return "", errors.New("unknown channel")
	}
}

// List returns the recent notifications for the calling user.
func (s *Service) List(ctx context.Context, userID string) ([]Notification, error) {
	return s.repo.ListForUser(ctx, userID, 50)
}
