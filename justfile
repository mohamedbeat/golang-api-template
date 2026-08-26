# justfile

set dotenv-load := true  # auto-loads .env file

db_url := "postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?"

# ── Dev Server ─────────────────────────────────────────────────────

# Run the app
run:
    go run ./cmd/api

# Run with live reload (requires air)
dev:
    air

# ── Build ──────────────────────────────────────────────────────────

# Build binary
build:
    go build -o bin/api ./cmd/api

# Build for linux (useful for docker/CI)
build-linux:
    GOOS=linux GOARCH=amd64 go build -o bin/api ./cmd/api

# ── Migrations ─────────────────────────────────────────────────────

# Run all pending migrations
migrate-up:
    goose -dir migrations postgres "{{db_url}}" up

# Roll back last migration
migrate-down:
    goose -dir migrations postgres "{{db_url}}" down

# Roll back all migrations
migrate-reset:
    goose -dir migrations postgres "{{db_url}}" reset

# Show migration status
migrate-status:
    goose -dir migrations postgres "{{db_url}}" status

# Create a new migration file — usage: just migrate-create <name>
migrate-create name:
    goose -dir migrations create {{name}} sql

# ── Database ───────────────────────────────────────────────────────

# Connect to the database via psql
db-shell:
    psql "{{db_url}}"

# Drop and recreate the database (destructive!)
db-reset: migrate-reset migrate-up
    @echo "Database reset complete"

# ── Testing ────────────────────────────────────────────────────────

# Run all tests
test:
    go test ./...

# Run tests with verbose output
test-v:
    go test -v ./...

# Run tests with coverage
test-cov:
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out

# ── Code Quality ───────────────────────────────────────────────────

# Format code
fmt:
    go fmt ./...

# Run linter (requires golangci-lint)
lint:
    golangci-lint run

# Run go vet
vet:
    go vet ./...

# Run fmt + vet + lint
check: fmt vet lint

# ── Docker ─────────────────────────────────────────────────────────

# Start postgres container only
docker-db:
    docker compose up db -d

# Start all services
docker-up:
    docker compose up -d

# Stop all services
docker-down:
    docker compose down

 

# ── Misc ───────────────────────────────────────────────────────────

# List all available recipes
default:
    @just --list

# Tidy go modules
tidy:
    go mod tidy
