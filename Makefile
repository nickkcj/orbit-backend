.PHONY: migrate-up migrate-down migrate-status migrate-create sqlc build run

# Load .env file
include .env
export

# Database migrations
migrate-up:
	goose -dir sql/schema postgres "$(DATABASE_URL)" up

migrate-down:
	goose -dir sql/schema postgres "$(DATABASE_URL)" down

migrate-status:
	goose -dir sql/schema postgres "$(DATABASE_URL)" status

migrate-create:
	@read -p "Migration name: " name; \
	goose -dir sql/schema create $$name sql

# SQLC
sqlc:
	sqlc generate

# Build & Run
build:
	go build -o bin/orbit ./cmd/orbit

run:
	go run ./cmd/orbit
