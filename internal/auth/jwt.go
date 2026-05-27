package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/milknest/backend/pkg/config"
	"github.com/milknest/backend/pkg/middleware"
)

// TokenService issues and validates JWT access + refresh tokens.
type TokenService struct {
	cfg config.JWTConfig
}

func NewTokenService(cfg config.JWTConfig) *TokenService {
	return &TokenService{cfg: cfg}
}

// IssuePair returns a signed access token and refresh token for the user.
func (s *TokenService) IssuePair(userID, role string) (access, refresh string, err error) {
	access, err = s.sign(userID, role, "access", s.cfg.AccessSecret, s.cfg.AccessTTL)
	if err != nil {
		return "", "", err
	}
	refresh, err = s.sign(userID, role, "refresh", s.cfg.RefreshSecret, s.cfg.RefreshTTL)
	if err != nil {
		return "", "", err
	}
	return access, refresh, nil
}

// ParseRefresh validates the refresh token signature/expiry and returns claims.
func (s *TokenService) ParseRefresh(token string) (*middleware.Claims, error) {
	claims := &middleware.Claims{}
	tok, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(s.cfg.RefreshSecret), nil
	})
	if err != nil || !tok.Valid {
		return nil, errors.New("invalid refresh token")
	}
	if claims.Type != "refresh" {
		return nil, errors.New("wrong token type")
	}
	return claims, nil
}

func (s *TokenService) sign(userID, role, typ, secret string, ttl time.Duration) (string, error) {
	claims := middleware.Claims{
		UserID: userID,
		Role:   role,
		Type:   typ,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID,
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}
