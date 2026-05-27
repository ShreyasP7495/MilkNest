package auth

import "time"

// User mirrors the users table - shared shape used by other modules.
type User struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	PhoneNumber string    `json:"phone_number"`
	Email       *string   `json:"email,omitempty"`
	Role        string    `json:"role"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SendOTPRequest matches API Spec §2.1.
type SendOTPRequest struct {
	PhoneNumber string `json:"phone_number" validate:"required,e164|len=10"`
}

type SendOTPResponse struct {
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
	// MockOTP is only populated when OTP_MOCK=true to ease development. Never
	// returned in production.
	MockOTP string `json:"mock_otp,omitempty"`
}

// VerifyOTPRequest matches API Spec §2.2.
type VerifyOTPRequest struct {
	PhoneNumber string `json:"phone_number" validate:"required"`
	OTP         string `json:"otp"          validate:"required"`
}

type VerifyOTPResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	User         AuthUser `json:"user"`
}

type AuthUser struct {
	ID   string `json:"id"`
	Role string `json:"role"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
