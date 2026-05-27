package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config holds every tunable used by the application. Values come from
// environment variables, defaulting to the values in .env.example. Production
// deployments must override the secrets.
type Config struct {
	App      AppConfig
	DB       DBConfig
	Redis    RedisConfig
	JWT      JWTConfig
	OTP      OTPConfig
	Razorpay RazorpayConfig
	SMS      SMSConfig
	FCM      FCMConfig
	Business BusinessConfig
	Cron     CronConfig
}

type AppConfig struct {
	Env  string
	Port string
	Name string
}

type DBConfig struct {
	Host         string
	Port         string
	User         string
	Password     string
	Name         string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
}

func (d DBConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode)
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", r.Host, r.Port)
}

type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
}

type OTPConfig struct {
	TTL    time.Duration
	Length int
	Mock   bool
}

type RazorpayConfig struct {
	KeyID         string
	KeySecret     string
	WebhookSecret string
}

type SMSConfig struct {
	Provider string
	APIKey   string
	SenderID string
}

type FCMConfig struct {
	ServerKey string
}

type BusinessConfig struct {
	MaxPacketsPerDay     int
	BNPLSurchargePercent int
}

type CronConfig struct {
	OrderGenSpec string
	Timezone     string
}

// Load reads .env (if present) and binds env vars into Config. It panics on
// fatal misconfiguration (e.g. blank JWT secret in production).
func Load() (*Config, error) {
	_ = godotenv.Load() // best-effort: production envs may not have a file

	v := viper.New()
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Defaults mirror .env.example
	v.SetDefault("APP_ENV", "development")
	v.SetDefault("APP_PORT", "8080")
	v.SetDefault("APP_NAME", "milknest")

	v.SetDefault("DB_HOST", "localhost")
	v.SetDefault("DB_PORT", "5432")
	v.SetDefault("DB_USER", "milk")
	v.SetDefault("DB_PASSWORD", "milk123")
	v.SetDefault("DB_NAME", "milknest")
	v.SetDefault("DB_SSLMODE", "disable")
	v.SetDefault("DB_MAX_OPEN_CONNS", 25)
	v.SetDefault("DB_MAX_IDLE_CONNS", 5)

	v.SetDefault("REDIS_HOST", "localhost")
	v.SetDefault("REDIS_PORT", "6379")
	v.SetDefault("REDIS_PASSWORD", "")
	v.SetDefault("REDIS_DB", 0)

	v.SetDefault("JWT_ACCESS_TTL_MINUTES", 1440)
	v.SetDefault("JWT_REFRESH_TTL_DAYS", 30)

	v.SetDefault("OTP_TTL_SECONDS", 300)
	v.SetDefault("OTP_LENGTH", 6)
	v.SetDefault("OTP_MOCK", true)

	v.SetDefault("SMS_PROVIDER", "stub")
	v.SetDefault("MAX_PACKETS_PER_DAY", 10)
	v.SetDefault("BNPL_SURCHARGE_PERCENT", 10)

	v.SetDefault("ORDER_GEN_CRON", "0 0 * * *")
	v.SetDefault("TIMEZONE", "Asia/Kolkata")

	cfg := &Config{
		App: AppConfig{
			Env:  v.GetString("APP_ENV"),
			Port: v.GetString("APP_PORT"),
			Name: v.GetString("APP_NAME"),
		},
		DB: DBConfig{
			Host:         v.GetString("DB_HOST"),
			Port:         v.GetString("DB_PORT"),
			User:         v.GetString("DB_USER"),
			Password:     v.GetString("DB_PASSWORD"),
			Name:         v.GetString("DB_NAME"),
			SSLMode:      v.GetString("DB_SSLMODE"),
			MaxOpenConns: v.GetInt("DB_MAX_OPEN_CONNS"),
			MaxIdleConns: v.GetInt("DB_MAX_IDLE_CONNS"),
		},
		Redis: RedisConfig{
			Host:     v.GetString("REDIS_HOST"),
			Port:     v.GetString("REDIS_PORT"),
			Password: v.GetString("REDIS_PASSWORD"),
			DB:       v.GetInt("REDIS_DB"),
		},
		JWT: JWTConfig{
			AccessSecret:  v.GetString("JWT_ACCESS_SECRET"),
			RefreshSecret: v.GetString("JWT_REFRESH_SECRET"),
			AccessTTL:     time.Duration(v.GetInt("JWT_ACCESS_TTL_MINUTES")) * time.Minute,
			RefreshTTL:    time.Duration(v.GetInt("JWT_REFRESH_TTL_DAYS")) * 24 * time.Hour,
		},
		OTP: OTPConfig{
			TTL:    time.Duration(v.GetInt("OTP_TTL_SECONDS")) * time.Second,
			Length: v.GetInt("OTP_LENGTH"),
			Mock:   v.GetBool("OTP_MOCK"),
		},
		Razorpay: RazorpayConfig{
			KeyID:         v.GetString("RAZORPAY_KEY_ID"),
			KeySecret:     v.GetString("RAZORPAY_KEY_SECRET"),
			WebhookSecret: v.GetString("RAZORPAY_WEBHOOK_SECRET"),
		},
		SMS: SMSConfig{
			Provider: v.GetString("SMS_PROVIDER"),
			APIKey:   v.GetString("SMS_API_KEY"),
			SenderID: v.GetString("SMS_SENDER_ID"),
		},
		FCM: FCMConfig{
			ServerKey: v.GetString("FCM_SERVER_KEY"),
		},
		Business: BusinessConfig{
			MaxPacketsPerDay:     v.GetInt("MAX_PACKETS_PER_DAY"),
			BNPLSurchargePercent: v.GetInt("BNPL_SURCHARGE_PERCENT"),
		},
		Cron: CronConfig{
			OrderGenSpec: v.GetString("ORDER_GEN_CRON"),
			Timezone:     v.GetString("TIMEZONE"),
		},
	}

	if cfg.JWT.AccessSecret == "" || cfg.JWT.RefreshSecret == "" {
		if cfg.App.Env == "production" {
			return nil, fmt.Errorf("JWT_ACCESS_SECRET and JWT_REFRESH_SECRET are required in production")
		}
		// Dev fallback - obvious, must NOT be used in prod.
		cfg.JWT.AccessSecret = "dev-access-secret-change-me"
		cfg.JWT.RefreshSecret = "dev-refresh-secret-change-me"
	}

	return cfg, nil
}
