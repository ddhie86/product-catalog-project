-include .env
export

APP     := product-catalog
DB_URL  ?= $(DATABASE_URL)

.PHONY: run build test cover fmt vet tidy migrate-up migrate-down seed up down

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

# Butuh CLI golang-migrate: brew install golang-migrate
migrate-up:
	migrate -path migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path migrations -database "$(DB_URL)" down 1

seed:           ## Buat admin awal dari SEED_ADMIN_*
	go run ./cmd/seed

up:             ## Postgres + Mailpit (SMTP dev)
	docker compose up -d

down:
	docker compose down
