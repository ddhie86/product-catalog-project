package database

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dedysetiawan/product-catalog-project/internal/config"
)

func NewPostgres(ctx context.Context, db config.DatabaseConfig) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(connectionURL(db))
	if err != nil {
		return nil, fmt.Errorf("parse database configuration: %w", err)
	}
	cfg.MaxConns = 20
	cfg.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return pool, nil
}

func connectionURL(db config.DatabaseConfig) string {
	query := url.Values{}
	query.Set("sslmode", strings.ToLower(db.SSLMode))
	if db.SSLRootCert != "" {
		query.Set("sslrootcert", db.SSLRootCert)
	}
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(db.User, db.Password),
		Host:     net.JoinHostPort(db.Host, strconv.Itoa(db.Port)),
		Path:     "/" + db.Name,
		RawQuery: query.Encode(),
	}
	return u.String()
}
