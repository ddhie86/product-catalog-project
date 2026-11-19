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
