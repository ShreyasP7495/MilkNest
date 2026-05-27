package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"math/big"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/milknest/backend/internal/notifications"
	"github.com/milknest/backend/pkg/config"
	"github.com/milknest/backend/pkg/logger"
	"github.com/milknest/backend/pkg/utils"
)

// Service wires OTP issuance, JWT issuance, and user upsert. It depends only
// on interfaces from other modules so it stays unit-testable.
type Service struct {
	repo    *Repository
	tokens  *TokenService
	rdb     *redis.Client
	cfg     *config.Config
	notifer notifications.Sender
}

// NewService constructs the auth service.
func NewService(repo *Repository, tokens *TokenService, rdb *redis.Client, cfg *config.Config, notifer notifications.Sender) *Service {
	return &Service{repo: repo, tokens: tokens, rdb: rdb, cfg: cfg, notifer: notifer}
}

// SendOTP generates a numeric OTP, stores it in Redis under otp:{phone} with
// the configured TTL, and asks the notifications layer to deliver it. In mock
// mode the OTP is also returned in the response to ease development.
func (s *Service) SendOTP(ctx context.Context, phone string) (*SendOTPResponse, error) {
	otp, err := generateOTP(s.cfg.OTP.Length)
	if err != nil {
		return nil, utils.NewInternal(err)
	}

	key := otpKey(phone)
	if err := s.rdb.Set(ctx, key, otp, s.cfg.OTP.TTL).Err(); err != nil {
		return nil, utils.NewInternal(fmt.Errorf("store otp: %w", err))
	}

	requestID := uuid.NewString()
	body := fmt.Sprintf("Your MilkNest OTP is %s. Valid for %d minutes.", otp, int(s.cfg.OTP.TTL.Minutes()))
	if err := s.notifer.SendSMS(ctx, phone, body); err != nil {
		logger.L().Warn().Err(err).Str("phone", phone).Msg("otp sms send failed (continuing)")
	}

	resp := &SendOTPResponse{Message: "OTP sent successfully", RequestID: requestID}
	if s.cfg.OTP.Mock {
		resp.MockOTP = otp
	}
	return resp, nil
}

// VerifyOTP checks the stored OTP, upserts the user, and returns a JWT pair.
func (s *Service) VerifyOTP(ctx context.Context, phone, submitted string) (*VerifyOTPResponse, error) {
	key := otpKey(phone)
	stored, err := s.rdb.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return nil, utils.NewBadRequest("otp_expired", "OTP expired or not requested")
	}
	if err != nil {
		return nil, utils.NewInternal(err)
	}
	if stored != submitted {
		return nil, utils.NewBadRequest("otp_invalid", "OTP does not match")
	}
	_ = s.rdb.Del(ctx, key).Err() // best-effort burn

	user, err := s.repo.FindByPhone(ctx, phone)
	if errors.Is(err, sql.ErrNoRows) {
		user, err = s.repo.CreateCustomer(ctx, phone)
	}
	if err != nil {
		return nil, utils.NewInternal(err)
	}

	if err := s.repo.EnsureWallet(ctx, user.ID); err != nil {
		logger.L().Warn().Err(err).Msg("ensure wallet failed (non-fatal)")
	}

	access, refresh, err := s.tokens.IssuePair(user.ID, user.Role)
	if err != nil {
		return nil, utils.NewInternal(err)
	}
	return &VerifyOTPResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		User:         AuthUser{ID: user.ID, Role: user.Role},
	}, nil
}

// Refresh validates the refresh token and rotates both tokens.
func (s *Service) Refresh(ctx context.Context, refresh string) (*RefreshResponse, error) {
	claims, err := s.tokens.ParseRefresh(refresh)
	if err != nil {
		return nil, utils.NewUnauthorized("invalid refresh token")
	}
	user, err := s.repo.FindByID(ctx, claims.UserID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, utils.NewUnauthorized("user no longer exists")
	}
	if err != nil {
		return nil, utils.NewInternal(err)
	}
	access, newRefresh, err := s.tokens.IssuePair(user.ID, user.Role)
	if err != nil {
		return nil, utils.NewInternal(err)
	}
	return &RefreshResponse{AccessToken: access, RefreshToken: newRefresh}, nil
}

func otpKey(phone string) string { return "otp:" + phone }

// generateOTP returns a cryptographically-random numeric OTP of length n.
func generateOTP(n int) (string, error) {
	if n <= 0 {
		n = 6
	}
	const digits = "0123456789"
	out := make([]byte, n)
	for i := 0; i < n; i++ {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		out[i] = digits[idx.Int64()]
	}
	return string(out), nil
}
