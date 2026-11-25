// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
package config

import "testing"

func TestValidateDatabaseConfig(t *testing.T) {
	tests := []struct {
		name    string
		db      DatabaseConfig
		env     string
		wantErr bool
	}{
		{
			name: "accepts local development configuration",
			db: DatabaseConfig{
				Host: "127.0.0.1", Port: 5432, Name: "catalog",
				User: "catalog", Password: "local-secret", SSLMode: "disable",
			},
			env: "development",
		},
		{
			name: "rejects a missing password",
			db: DatabaseConfig{
				Host: "127.0.0.1", Port: 5432, Name: "catalog",
				User: "catalog", SSLMode: "disable",
			},
			env:     "development",
			wantErr: true,
		},
		{
			name: "rejects an invalid port",
			db: DatabaseConfig{
				Host: "127.0.0.1", Port: 65536, Name: "catalog",
				User: "catalog", Password: "local-secret", SSLMode: "disable",
			},
			env:     "development",
			wantErr: true,
		},
		{
			name: "requires TLS in production",
			db: DatabaseConfig{
				Host: "db.example", Port: 5432, Name: "catalog",
				User: "catalog", Password: "production-secret", SSLMode: "disable",
			},
			env:     "production",
			wantErr: true,
		},
		{
			name: "accepts verified TLS in production",
			db: DatabaseConfig{
				Host: "db.example", Port: 5432, Name: "catalog",
				User: "catalog", Password: "production-secret", SSLMode: "verify-full",
			},
			env: "production",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDatabaseConfig(tt.db, tt.env)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateDatabaseConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateGoogleDriveConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     GoogleDriveConfig
		enabled bool
		wantErr bool
	}{
		{
			name: "image uploads are optional",
			cfg:  GoogleDriveConfig{MaxUploadBytes: 5 << 20},
		},
		{
			name: "complete OAuth configuration enables image uploads",
			cfg: GoogleDriveConfig{
				OAuthClientID: "client-id", OAuthClientSecret: "client-secret",
				RefreshToken: "refresh-token", RootFolderID: "folder-id",
				MaxUploadBytes: 5 << 20,
			},
			enabled: true,
		},
		{
			name: "partial OAuth configuration is rejected",
			cfg: GoogleDriveConfig{
				OAuthClientID:  "client-id",
				MaxUploadBytes: 5 << 20,
			},
			wantErr: true,
		},
		{
			name: "non-positive upload limit is rejected",
			cfg: GoogleDriveConfig{
				MaxUploadBytes: 0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.cfg
			err := validateGoogleDriveConfig(&cfg)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateGoogleDriveConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && cfg.Enabled != tt.enabled {
				t.Fatalf("Enabled = %v, want %v", cfg.Enabled, tt.enabled)
			}
		})
	}
}
