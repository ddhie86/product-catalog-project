// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
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
	"github.com/dedysetiawan/product-catalog-project/pkg/apperror"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	email := strings.ToLower(strings.TrimSpace(os.Getenv("SEED_ADMIN_EMAIL")))
	pass := os.Getenv("SEED_ADMIN_PASSWORD")
	if email == "" || strings.TrimSpace(pass) == "" || len(pass) < 12 || len(pass) > 72 {
		log.Fatal("SEED_ADMIN_EMAIL and SEED_ADMIN_PASSWORD (12-72 bytes) are required")
	}

	ctx := context.Background()
	pool, err := database.NewPostgres(ctx, cfg.Database)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	users := repository.NewPostgresUserRepository(pool)
	roles := repository.NewPostgresRoleRepository(pool)
	_, err = users.FindByEmail(ctx, email)
	if err == nil {
		log.Printf("admin already exists: %s", email)
		return
	}
	if apperror.Code(err) != "NOT_FOUND" {
		log.Fatal(err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}
	roleList, err := roles.ListRoles(ctx)
	if err != nil {
		log.Fatal(err)
	}
	var superAdminRoleID uuid.UUID
	for _, role := range roleList {
		if role.Name == "Super Admin" && role.IsSystem {
			superAdminRoleID = role.ID
			break
		}
	}
	if superAdminRoleID == uuid.Nil {
		log.Fatal("Super Admin role is missing; run migrations before seeding")
	}
	u := &domain.User{ID: uuid.New(), Email: email, PasswordHash: string(hash)}
	if err := users.CreateAdmin(ctx, u, []uuid.UUID{superAdminRoleID}); err != nil {
		log.Fatal(err)
	}
	log.Printf("admin created: %s", email)
}
