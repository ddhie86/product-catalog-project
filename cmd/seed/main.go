// Command seed: membuat akun admin awal dari SEED_ADMIN_EMAIL / SEED_ADMIN_PASSWORD.
package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/dedysetiawan/product-catalog-project/internal/config"
	"github.com/dedysetiawan/product-catalog-project/internal/modules/auth/domain"
	"github.com/dedysetiawan/product-catalog-project/internal/modules/auth/repository"
	"github.com/dedysetiawan/product-catalog-project/internal/platform/database"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	email := strings.ToLower(os.Getenv("SEED_ADMIN_EMAIL"))
	pass := os.Getenv("SEED_ADMIN_PASSWORD")
	if email == "" || len(pass) < 8 {
		log.Fatal("SEED_ADMIN_EMAIL and SEED_ADMIN_PASSWORD (min 8 chars) are required")
	}

	ctx := context.Background()
	pool, err := database.NewPostgres(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}
	u := &domain.User{ID: uuid.New(), Email: email, PasswordHash: string(hash)}
	if err := repository.NewPostgresUserRepository(pool).Create(ctx, u); err != nil {
		log.Fatal(err)
	}
	log.Printf("admin created: %s", email)
}
