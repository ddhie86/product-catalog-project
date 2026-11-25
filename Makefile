-include .env
export

APP     := product-catalog
DB_URL  = postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

.PHONY: run build test cover fmt vet tidy check-db-config migrate-up migrate-down seed bootstrap up down google-drive-auth

run:            ## Jalankan API
	go run ./cmd/api

build:
	go build -o bin/$(APP) ./cmd/api

test:
	go test ./... -race -count=1

cover:
	go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out | tail -1

fmt:
	gofmt -w .

vet:
	go vet ./...

tidy:
	go mod tidy

# Requires the golang-migrate CLI: brew install golang-migrate
check-db-config:
	@test -n "$(DB_HOST)" -a -n "$(DB_PORT)" -a -n "$(DB_NAME)" -a \
		-n "$(DB_USER)" -a -n "$(DB_PASSWORD)" -a -n "$(DB_SSLMODE)" || \
		(echo "Set DB_HOST, DB_PORT, DB_NAME, DB_USER, DB_PASSWORD, and DB_SSLMODE in .env"; exit 1)

migrate-up: check-db-config
	migrate -path migrations -database "$(DB_URL)" up

migrate-down: check-db-config
	migrate -path migrations -database "$(DB_URL)" down 1

seed:           ## Create the initial admin from SEED_ADMIN_*
	go run ./cmd/seed

bootstrap:      ## Apply migrations and create the initial admin
	$(MAKE) migrate-up
	$(MAKE) seed

up:             ## Start PostgreSQL and Mailpit for local development
	docker compose up -d

down:
	docker compose down

google-drive-auth:
	go run ./cmd/google-drive-auth
