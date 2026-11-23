// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
package database

import (
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dedysetiawan/product-catalog-project/internal/config"
)

func TestConnectionURLEscapesCredentials(t *testing.T) {
	db := config.DatabaseConfig{
		Host: "localhost", Port: 5432, Name: "catalog",
		User: "catalog user", Password: "p@ss:/?#", SSLMode: "disable",
	}

	cfg, err := pgxpool.ParseConfig(connectionURL(db))
	if err != nil {
		t.Fatalf("parse generated connection URL: %v", err)
	}
	if cfg.ConnConfig.User != db.User {
		t.Errorf("user = %q, want %q", cfg.ConnConfig.User, db.User)
	}
	if cfg.ConnConfig.Password != db.Password {
		t.Errorf("password was not preserved after URL encoding")
	}
}
