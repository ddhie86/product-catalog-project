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

	Database DatabaseConfig

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

type DatabaseConfig struct {
	Host        string
	Port        int
	Name        string
	User        string
	Password    string
	SSLMode     string
	SSLRootCert string
}

func Load() (*Config, error) {
	ttl, err := time.ParseDuration(get("JWT_TTL", "24h"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_TTL: %w", err)
	}
	dbPort, err := strconv.Atoi(get("DB_PORT", "5432"))
	if err != nil {
		return nil, fmt.Errorf("DB_PORT must be an integer between 1 and 65535")
	}
	cfg := &Config{
		Env:            get("APP_ENV", "development"),
		Port:           get("APP_PORT", "8080"),
		AllowedOrigins: strings.Split(get("CORS_ALLOWED_ORIGINS", "http://localhost:5173"), ","),
		Database: DatabaseConfig{
			Host:        os.Getenv("DB_HOST"),
			Port:        dbPort,
			Name:        os.Getenv("DB_NAME"),
			User:        os.Getenv("DB_USER"),
			Password:    os.Getenv("DB_PASSWORD"),
			SSLMode:     strings.ToLower(get("DB_SSLMODE", "disable")),
			SSLRootCert: os.Getenv("DB_SSLROOTCERT"),
		},
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
	if err := validateDatabaseConfig(cfg.Database, cfg.Env); err != nil {
		return nil, err
	}
	if len(cfg.JWTSecret) < 16 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 16 characters")
	}
	return cfg, nil
}

func validateDatabaseConfig(db DatabaseConfig, env string) error {
	if strings.TrimSpace(db.Host) == "" {
		return fmt.Errorf("DB_HOST is required")
	}
	if db.Port < 1 || db.Port > 65535 {
		return fmt.Errorf("DB_PORT must be an integer between 1 and 65535")
	}
	if strings.TrimSpace(db.Name) == "" {
		return fmt.Errorf("DB_NAME is required")
	}
	if strings.TrimSpace(db.User) == "" {
		return fmt.Errorf("DB_USER is required")
	}
	if db.Password == "" {
		return fmt.Errorf("DB_PASSWORD is required")
	}

	switch strings.ToLower(db.SSLMode) {
	case "disable", "allow", "prefer", "require", "verify-ca", "verify-full":
	default:
		return fmt.Errorf("DB_SSLMODE must be a valid PostgreSQL SSL mode")
	}
	if strings.EqualFold(env, "production") {
		switch strings.ToLower(db.SSLMode) {
		case "require", "verify-ca", "verify-full":
		default:
			return fmt.Errorf("DB_SSLMODE must be require or stronger in production")
		}
	}
	return nil
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
