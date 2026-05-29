.PHONY: help up down run seed migrate test test-unit test-integration lint build swagger docker-build

help:
	@echo "Available targets:"
	@echo "  make up              - Start postgres + backend containers (docker-compose)"
	@echo "  make down            - Stop containers"
	@echo "  make run             - Run backend locally (requires go + postgres)"
	@echo "  make seed            - Load initial data into postgres"
	@echo "  make migrate         - Run DB migrations"
	@echo "  make test            - Run all tests (unit + integration)"
	@echo "  make test-unit       - Run unit tests only"
	@echo "  make test-integration - Run integration tests (requires postgres)"
	@echo "  make lint            - Lint code (golangci-lint)"
	@echo "  make build           - Build binary"
	@echo "  make swagger         - Generate Swagger docs"
	@echo "  make docker-build    - Build Docker image"

up:
	docker-compose up -d

down:
	docker-compose down

run:
	go run ./cmd/api

seed:
	@echo "Seed runs automatically in development mode on startup."
	@echo "To seed manually, ensure DATABASE_URL is set and run:"
	@echo "  go run ./cmd/api"

migrate:
	@echo "Migrations run automatically on startup via golang-migrate embedded runner."
	@echo "To run migrations manually against a local DB:"
	@echo "  DATABASE_URL='postgres://localhost:5432/issuetracker?sslmode=disable' go run ./cmd/api"

test:
	go test -v ./...

test-unit:
	go test -v -short ./internal/...

test-integration:
	go test -v ./tests -run Integration

lint:
	@if ! command -v golangci-lint &> /dev/null; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	golangci-lint run ./...

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/api ./cmd/api

swagger:
	@echo "Swagger generation not yet implemented"

docker-build:
	docker build -t issue-tracker-backend:latest .
