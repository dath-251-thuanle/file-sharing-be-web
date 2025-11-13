.PHONY: help run build test clean docker-build docker-run swagger migrate-up migrate-down

# Variables
APP_NAME=file-sharing-backend
BINARY_NAME=server
DOCKER_IMAGE=file-sharing-backend:latest
MAIN_PATH=./cmd/server/main.go
DOCS_PATH=./docs/swagger

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

# Default target
help:
	@echo "Available targets:"
	@echo "  make run              - Run the application in development mode"
	@echo "  make build            - Build the application binary"
	@echo "  make test             - Run all tests"
	@echo "  make test-coverage    - Run tests with coverage report"
	@echo "  make lint             - Run linter"
	@echo "  make fmt              - Format Go code"
	@echo "  make vet              - Run go vet"
	@echo "  make clean            - Clean build artifacts"
	@echo "  make swagger          - Generate Swagger documentation"
	@echo "  make migrate-up       - Apply database migrations"
	@echo "  make migrate-down     - Rollback database migrations"
	@echo "  make db-seed          - Seed database with sample data"
	@echo "  make docker-build     - Build Docker image"
	@echo "  make docker-run       - Run Docker container"
	@echo "  make docker-up        - Start all services with docker-compose"
	@echo "  make docker-down      - Stop all services"
	@echo "  make deps             - Download dependencies"
	@echo "  make tidy             - Tidy go.mod"

# Development
run:
	@echo "Starting $(APP_NAME) in development mode..."
	$(GOCMD) run $(MAIN_PATH)

build:
	@echo "Building $(APP_NAME)..."
	$(GOBUILD) -o bin/$(BINARY_NAME) -v $(MAIN_PATH)

build-linux:
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o bin/$(BINARY_NAME)-linux -v $(MAIN_PATH)

build-windows:
	@echo "Building for Windows..."
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o bin/$(BINARY_NAME).exe -v $(MAIN_PATH)

# Testing
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-unit:
	@echo "Running unit tests..."
	$(GOTEST) -v ./tests/unit/...

test-integration:
	@echo "Running integration tests..."
	$(GOTEST) -v ./tests/integration/...

# Code Quality
lint:
	@echo "Running linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...

vet:
	@echo "Running go vet..."
	$(GOCMD) vet ./...

# Dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download

tidy:
	@echo "Tidying go.mod..."
	$(GOMOD) tidy

vendor:
	@echo "Vendoring dependencies..."
	$(GOMOD) vendor

# Documentation
swagger:
	@echo "Generating Swagger documentation..."
	@if command -v swag > /dev/null; then \
		swag init -g $(MAIN_PATH) -o $(DOCS_PATH); \
		echo "Swagger docs generated in $(DOCS_PATH)"; \
	else \
		echo "swag not installed. Install: go install github.com/swaggo/swag/cmd/swag@latest"; \
	fi

# Database
db-create:
	docker exec -it postgres createdb -U $(DB_USER) $(DB_NAME)

db-drop:
	docker exec -it postgres dropdb -U $(DB_USER) $(DB_NAME)

db-shell:
	docker exec -it postgres psql -U $(DB_USER) -d $(DB_NAME)

# Migrations (requires golang-migrate)
migrate-up:
	@echo "Applying database migrations..."
	@if command -v migrate > /dev/null; then \
		migrate -path ./migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)" up; \
	elif command -v pwsh > /dev/null; then \
		pwsh -NoProfile -ExecutionPolicy Bypass -File ./scripts/migrate.ps1 up; \
	elif command -v bash > /dev/null; then \
		bash ./scripts/migrate.sh up; \
	else \
		echo "No migration tool found. Install golang-migrate or use PowerShell/Bash scripts."; \
	fi

migrate-down:
	@echo "Rolling back last migration..."
	@if command -v migrate > /dev/null; then \
		migrate -path ./migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)" down 1; \
	elif command -v pwsh > /dev/null; then \
		pwsh -NoProfile -ExecutionPolicy Bypass -File ./scripts/migrate.ps1 down; \
	elif command -v bash > /dev/null; then \
		bash ./scripts/migrate.sh down; \
	else \
		echo "No migration tool found."; \
	fi

migrate-version:
	@echo "Checking migration version..."
	@if command -v migrate > /dev/null; then \
		migrate -path ./migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)" version; \
	elif command -v pwsh > /dev/null; then \
		pwsh -NoProfile -ExecutionPolicy Bypass -File ./scripts/migrate.ps1 version; \
	elif command -v bash > /dev/null; then \
		bash ./scripts/migrate.sh version; \
	else \
		echo "No migration tool found."; \
	fi

migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "Usage: make migrate-create name=migration_name"; \
	else \
		migrate create -ext sql -dir ./migrations -seq $(name); \
	fi

migrate-force:
	@if [ -z "$(version)" ]; then \
		echo "Usage: make migrate-force version=1"; \
	else \
		migrate -path ./migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)" force $(version); \
	fi

migrate-drop:
	@echo "WARNING: This will drop all tables!"
	@read -p "Are you sure? (yes/no): " confirm && [ "$$confirm" = "yes" ] && \
		migrate -path ./migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)" drop -f || \
		echo "Cancelled."

db-reset: migrate-drop migrate-up
	@echo "Database reset and seeded"

# Docker
docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE) .

docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 --env-file .env $(DOCKER_IMAGE)

docker-up:
	@echo "Starting all services with docker-compose..."
	docker-compose up -d

docker-down:
	@echo "Stopping all services..."
	docker-compose down

docker-logs:
	@echo "Showing logs..."
	docker-compose logs -f

docker-ps:
	@echo "Listing containers..."
	docker-compose ps

docker-clean:
	@echo "Cleaning up Docker resources..."
	docker-compose down -v
	docker system prune -f

# Security
verify-security:
	@echo "Running security verification..."
	@if command -v pwsh > /dev/null; then \
		pwsh -File ./scripts/verify-security.ps1; \
	elif command -v bash > /dev/null; then \
		bash ./scripts/verify-security.sh; \
	else \
		echo "No PowerShell or Bash found. Please run manually."; \
	fi

scan-image:
	@echo "Scanning Docker image with Trivy..."
	@if command -v trivy > /dev/null; then \
		trivy image --severity HIGH,CRITICAL file-sharing-backend:latest; \
	else \
		echo "Trivy not installed. Install: choco install trivy"; \
	fi

# Cleanup
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf bin/
	rm -rf coverage.out coverage.html
	rm -rf $(DOCS_PATH)
	@echo "Cleanup complete"

# Development utilities
dev-setup:
	@echo "Setting up development environment..."
	@echo "Installing tools..."
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Downloading dependencies..."
	$(GOMOD) download
	@echo "Creating .env from .env.example..."
	@if [ ! -f .env ]; then cp .env.example .env; fi
	@echo "Setup complete! Edit .env with your configuration."

watch:
	@echo "Watching for changes..."
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "air not installed. Install: go install github.com/cosmtrek/air@latest"; \
		echo "Running with go run instead..."; \
		$(GOCMD) run $(MAIN_PATH); \
	fi

# Production
build-prod:
	@echo "Building for production..."
	CGO_ENABLED=0 GOOS=linux $(GOBUILD) -a -installsuffix cgo -ldflags="-w -s" -o bin/$(BINARY_NAME) $(MAIN_PATH)

deploy:
	@echo "Deploying to production..."
	@echo "TODO: Add deployment commands"

# All-in-one commands
all: clean deps build test

install: dev-setup migrate-up
	@echo "Installation complete!"

.DEFAULT_GOAL := help
