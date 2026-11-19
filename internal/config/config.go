// Package config memuat konfigurasi aplikasi dari environment variable.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Env            string
	Port           string
	AllowedOrigins []string

	DatabaseURL string

	JWTSecret string
	JWTTTL    time.Duration

	XenditAPIKey        string
	XenditCallbackToken string
	PaymentRedirectURL  string

	SMTPHost string
	SMTPPort int
	SMTPUser string
	SMTPPass string
	SMTPFrom string

	LowStockThreshold int
	EmailQueueSize    int
}

func Load() (*Config, error) {
	ttl, err := time.ParseDuration(get("JWT_TTL", "24h"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_TTL: %w", err)
	}
	cfg := &Config{
		Env:                 get("APP_ENV", "development"),
		Port:                get("APP_PORT", "8080"),
		AllowedOrigins:      strings.Split(get("CORS_ALLOWED_ORIGINS", "http://localhost:5173"), ","),
		DatabaseURL:         os.Getenv("DATABASE_URL"),
		JWTSecret:           os.Getenv("JWT_SECRET"),
		JWTTTL:              ttl,
		XenditAPIKey:        os.Getenv("XENDIT_API_KEY"),
		XenditCallbackToken: os.Getenv("XENDIT_CALLBACK_TOKEN"),
		PaymentRedirectURL:  os.Getenv("PAYMENT_SUCCESS_REDIRECT_URL"),
		SMTPHost:            get("SMTP_HOST", "localhost"),
		SMTPPort:            getInt("SMTP_PORT", 1025),
		SMTPUser:            os.Getenv("SMTP_USER"),
		SMTPPass:            os.Getenv("SMTP_PASS"),
		SMTPFrom:            get("SMTP_FROM", "no-reply@catalog.local"),
		LowStockThreshold:   getInt("LOW_STOCK_THRESHOLD", 5),
		EmailQueueSize:      getInt("EMAIL_QUEUE_SIZE", 256),
	}
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if len(cfg.JWTSecret) < 16 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 16 characters")
	}
	return cfg, nil
}

func get(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
