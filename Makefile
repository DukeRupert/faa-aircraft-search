# Environment variables for development
export POSTGRES_USER ?= postgres
export POSTGRES_PASSWORD ?= postgres
export POSTGRES_DB ?= faa_aircraft
export DB_HOST ?= localhost
export DB_PORT ?= 5432
export DB_SSLMODE ?= disable

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Binary names
WEB_BINARY=faa-web
MIGRATE_BINARY=faa-migrate

# Build directory
BUILD_DIR=bin

.PHONY: all build clean test deps web migrate import-data clear-data count-data dev-setup sqlc-generate templ-generate

# Default target
all: build

# Generate templ templates
templ-generate:
	templ generate

# Generate SQLC code
sqlc-generate:
	sqlc generate

# Generate all code
generate: templ-generate sqlc-generate

# Build all binaries
build: generate web-build migrate-build

# Build web server
web-build:
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(WEB_BINARY) ./cmd/web

# Build migration tool
migrate-build:
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(MIGRATE_BINARY) ./cmd/migrate

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)

# Run tests
test:
	$(GOTEST) -v ./...

# Download dependencies
deps:
	$(GOMOD) tidy
	$(GOMOD) download

# Development setup
dev-setup: deps
	@echo "Setting up development environment..."
	@echo "Make sure Docker is running and then run: make db-up"
	@echo "Then run migrations: make migrate-up"
	@echo "Generate SQLC code: make sqlc-generate"
	@echo "Import data: make import-data"
	@echo "Start web server: make web"

# Run web server
web: generate
	$(GOCMD) run cmd/web/main.go

# Run migration tool
migrate: generate
	$(GOCMD) run cmd/migrate/main.go $(ARGS)

# Database operations
db-up:
	docker-compose up -d

db-down:
	docker-compose down

db-logs:
	docker-compose logs postgres

# Migration operations (requires goose to be installed and env vars set)
migrate-up:
	goose -dir migrations postgres up

migrate-down:
	goose -dir migrations postgres down

migrate-reset:
	goose -dir migrations postgres reset

# Data operations
import-data: generate
	$(GOCMD) run cmd/migrate/main.go -action=import -file=aircraft_data.xlsx

clear-data: generate
	$(GOCMD) run cmd/migrate/main.go -action=clear

count-data: generate
	$(GOCMD) run cmd/migrate/main.go -action=count

# Development workflow shortcuts
dev: db-up generate import-data web

# API testing (requires curl)
test-api:
	@echo "Testing health endpoint..."
	curl -s http://localhost:8080/health | jq .
	@echo "\nTesting search endpoint..."
	curl -s "http://localhost:8080/api/v1/aircraft/search?q=boeing&limit=5" | jq .

# Install development tools
install-tools:
	go install github.com/pressly/goose/v3/cmd/goose@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install github.com/a-h/templ/cmd/templ@latest

# Help
help:
	@echo "Available commands:"
	@echo "  build          - Build all binaries"
	@echo "  generate       - Generate all code (templ + sqlc)"
	@echo "  templ-generate - Generate templ templates"
	@echo "  sqlc-generate  - Generate SQLC code from SQL queries"
	@echo "  web-build      - Build web server binary"
	@echo "  migrate-build  - Build migration tool binary"
	@echo "  clean          - Clean build artifacts"
	@echo "  test           - Run tests"
	@echo "  deps           - Download dependencies"
	@echo "  dev-setup      - Setup development environment"
	@echo "  web            - Run web server"
	@echo "  migrate        - Run migration tool (use ARGS='...' for arguments)"
	@echo "  db-up          - Start database with Docker"
	@echo "  db-down        - Stop database"
	@echo "  db-logs        - Show database logs"
	@echo "  migrate-up     - Run database migrations"
	@echo "  migrate-down   - Rollback last migration"
	@echo "  migrate-reset  - Reset all migrations"
	@echo "  import-data    - Import aircraft data from Excel"
	@echo "  clear-data     - Clear all aircraft data"
	@echo "  count-data     - Count aircraft records"
	@echo "  dev            - Full development setup (db + migrate + sqlc + import + web)"
	@echo "  test-api       - Test API endpoints"
	@echo "  install-tools  - Install development tools"