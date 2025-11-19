# =============================================================================
# Makefile for File Sharing API - Docker & Database Management
# =============================================================================
# Author: DATH-251 Team
# Description: Simplified commands for managing Docker services and database migrations
# Usage: make [command]
# =============================================================================

# Load .env file if it exists
-include .env
export

# =============================================================================
# Configuration
# =============================================================================

# Docker Compose
COMPOSE := docker-compose
COMPOSE_FILE := docker-compose.yml

# Database Configuration (with defaults)
DB_USER ?= postgres
DB_PASSWORD ?= postgres
DB_NAME ?= file_sharing_db
DB_HOST ?= postgres
DB_PORT ?= 5432
DB_SSLMODE ?= disable

# Database URL for migrations
DATABASE_URL := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

# Colors for output
GREEN := \033[0;32m
YELLOW := \033[1;33m
RED := \033[0;31m
NC := \033[0m # No Color

# =============================================================================
# Help Command (Default)
# =============================================================================

.PHONY: help
help: ## Show this help message
	@echo "$(GREEN)═══════════════════════════════════════════════════════════════$(NC)"
	@echo "$(GREEN)  File Sharing API - Docker & Database Management$(NC)"
	@echo "$(GREEN)═══════════════════════════════════════════════════════════════$(NC)"
	@echo ""
	@echo "$(YELLOW)Available commands:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(YELLOW)Examples:$(NC)"
	@echo "  make setup              # First time setup (start DB + run migrations)"
	@echo "  make dev                # Start development environment"
	@echo "  make migrate-up         # Run pending migrations"
	@echo "  make logs               # View all logs"
	@echo ""

# =============================================================================
# Docker Commands
# =============================================================================

.PHONY: docker-up
docker-up: ## Start all Docker services (DB only)
	@echo "$(GREEN)Starting Docker services...$(NC)"
	$(COMPOSE) up -d postgres
	@echo "$(GREEN)✓ Services started!$(NC)"

.PHONY: docker-down
docker-down: ## Stop all Docker services
	@echo "$(YELLOW)Stopping Docker services...$(NC)"
	$(COMPOSE) down
	@echo "$(GREEN)✓ Services stopped!$(NC)"

.PHONY: docker-restart
docker-restart: docker-down docker-up ## Restart all Docker services

.PHONY: docker-clean
docker-clean: ## Stop and remove all containers, volumes, and images
	@echo "$(RED)WARNING: This will remove all data!$(NC)"
	@read -p "Are you sure? (yes/no): " confirm; \
	if [ "$$confirm" = "yes" ]; then \
		$(COMPOSE) down -v --remove-orphans; \
		echo "$(GREEN)✓ Cleaned up!$(NC)"; \
	else \
		echo "$(YELLOW)Cancelled.$(NC)"; \
	fi

.PHONY: dev
dev: ## Start development environment (with hot reload)
	@echo "$(GREEN)Starting development environment...$(NC)"
	$(COMPOSE) --profile dev up -d
	@echo "$(GREEN)✓ Dev environment started!$(NC)"
	@echo "$(YELLOW)API running at: http://localhost:8080$(NC)"
	@echo "$(YELLOW)View logs: make logs-app$(NC)"

.PHONY: prod
prod: ## Start production environment
	@echo "$(GREEN)Starting production environment...$(NC)"
	$(COMPOSE) --profile prod up -d
	@echo "$(GREEN)✓ Production environment started!$(NC)"

.PHONY: tools
tools: ## Start additional tools (Adminer)
	@echo "$(GREEN)Starting tools...$(NC)"
	$(COMPOSE) --profile tools up -d
	@echo "$(GREEN)✓ Tools started!$(NC)"
	@echo "$(YELLOW)Adminer: http://localhost:8081$(NC)"

# =============================================================================
# Database Migration Commands
# =============================================================================

.PHONY: migrate-up
migrate-up: ## Run all pending migrations
	@echo "$(GREEN)Running migrations UP...$(NC)"
	$(COMPOSE) run --rm migrate -path=/migrations -database "$(DATABASE_URL)" up
	@echo "$(GREEN)✓ Migrations completed!$(NC)"

.PHONY: migrate-down
migrate-down: ## Rollback last migration
	@echo "$(YELLOW)Rolling back last migration...$(NC)"
	$(COMPOSE) run --rm migrate -path=/migrations -database "$(DATABASE_URL)" down 1
	@echo "$(GREEN)✓ Rollback completed!$(NC)"

.PHONY: migrate-down-all
migrate-down-all: ## Rollback all migrations
	@echo "$(RED)WARNING: This will rollback ALL migrations!$(NC)"
	@read -p "Are you sure? (yes/no): " confirm; \
	if [ "$$confirm" = "yes" ]; then \
		$(COMPOSE) run --rm migrate -path=/migrations -database "$(DATABASE_URL)" down -all; \
		echo "$(GREEN)✓ All migrations rolled back!$(NC)"; \
	else \
		echo "$(YELLOW)Cancelled.$(NC)"; \
	fi

.PHONY: migrate-drop
migrate-drop: ## Drop all tables (DANGEROUS!)
	@echo "$(RED)WARNING: This will DROP ALL TABLES!$(NC)"
	@read -p "Type 'DROP' to confirm: " confirm; \
	if [ "$$confirm" = "DROP" ]; then \
		$(COMPOSE) run --rm migrate -path=/migrations -database "$(DATABASE_URL)" drop -f; \
		echo "$(GREEN)✓ All tables dropped!$(NC)"; \
	else \
		echo "$(YELLOW)Cancelled.$(NC)"; \
	fi

.PHONY: migrate-version
migrate-version: ## Show current migration version
	@echo "$(GREEN)Current migration version:$(NC)"
	@$(COMPOSE) run --rm migrate -path=/migrations -database "$(DATABASE_URL)" version

.PHONY: migrate-create
migrate-create: ## Create new migration (usage: make migrate-create NAME=my_migration)
ifndef NAME
	@echo "$(RED)Error: NAME is required$(NC)"
	@echo "$(YELLOW)Usage: make migrate-create NAME=add_new_table$(NC)"
	@exit 1
endif
	@echo "$(GREEN)Creating migration: $(NAME)$(NC)"
	$(COMPOSE) run --rm migrate create -ext sql -dir /migrations -seq $(NAME)
	@echo "$(GREEN)✓ Migration files created!$(NC)"

.PHONY: migrate-force
migrate-force: ## Force set migration version (usage: make migrate-force VERSION=1)
ifndef VERSION
	@echo "$(RED)Error: VERSION is required$(NC)"
	@echo "$(YELLOW)Usage: make migrate-force VERSION=1$(NC)"
	@exit 1
endif
	@echo "$(YELLOW)Forcing migration version to $(VERSION)...$(NC)"
	$(COMPOSE) run --rm migrate -path=/migrations -database "$(DATABASE_URL)" force $(VERSION)
	@echo "$(GREEN)✓ Version forced to $(VERSION)!$(NC)"

# =============================================================================
# Database Commands
# =============================================================================

.PHONY: db-shell
db-shell: ## Connect to PostgreSQL shell
	@echo "$(GREEN)Connecting to database...$(NC)"
	$(COMPOSE) exec postgres psql -U $(DB_USER) -d $(DB_NAME)

.PHONY: db-status
db-status: ## Check database connection status
	@echo "$(GREEN)Checking database status...$(NC)"
	@$(COMPOSE) exec postgres pg_isready -U $(DB_USER) || echo "$(RED)Database is not ready$(NC)"

.PHONY: db-backup
db-backup: ## Backup database to ./backups/
	@mkdir -p backups
	@echo "$(GREEN)Creating database backup...$(NC)"
	$(COMPOSE) exec -T postgres pg_dump -U $(DB_USER) $(DB_NAME) > backups/backup_$(shell date +%Y%m%d_%H%M%S).sql
	@echo "$(GREEN)✓ Backup created in ./backups/$(NC)"

.PHONY: db-restore
db-restore: ## Restore database from backup (usage: make db-restore FILE=backup.sql)
ifndef FILE
	@echo "$(RED)Error: FILE is required$(NC)"
	@echo "$(YELLOW)Usage: make db-restore FILE=backups/backup_20240101_120000.sql$(NC)"
	@exit 1
endif
	@echo "$(YELLOW)Restoring database from $(FILE)...$(NC)"
	cat $(FILE) | $(COMPOSE) exec -T postgres psql -U $(DB_USER) -d $(DB_NAME)
	@echo "$(GREEN)✓ Database restored!$(NC)"

# =============================================================================
# Logs Commands
# =============================================================================

.PHONY: logs
logs: ## View logs from all services
	$(COMPOSE) logs -f

.PHONY: logs-app
logs-app: ## View logs from application
	$(COMPOSE) logs -f app-dev app

.PHONY: logs-db
logs-db: ## View logs from database
	$(COMPOSE) logs -f postgres

# =============================================================================
# Testing & Verification Commands
# =============================================================================

.PHONY: test-db
test-db: ## Test database connection and schema
	@echo "$(GREEN)Testing database connection...$(NC)"
	@$(COMPOSE) exec postgres psql -U $(DB_USER) -d $(DB_NAME) -c "SELECT version();" > /dev/null 2>&1 && echo "$(GREEN)✓ Connection OK$(NC)" || echo "$(RED)✗ Connection Failed$(NC)"
	@echo "$(GREEN)Checking tables...$(NC)"
	@$(COMPOSE) exec postgres psql -U $(DB_USER) -d $(DB_NAME) -c "\dt" 2>/dev/null || echo "$(YELLOW)No tables found$(NC)"

.PHONY: verify
verify: ## Verify complete setup (DB + migrations)
	@echo "$(GREEN)Verifying setup...$(NC)"
	@echo ""
	@echo "$(YELLOW)1. Database Status:$(NC)"
	@$(COMPOSE) exec postgres pg_isready -U $(DB_USER) && echo "$(GREEN)   ✓ Database is ready$(NC)" || echo "$(RED)   ✗ Database is not ready$(NC)"
	@echo ""
	@echo "$(YELLOW)2. Migration Version:$(NC)"
	@$(COMPOSE) run --rm migrate -path=/migrations -database "$(DATABASE_URL)" version 2>/dev/null || echo "$(YELLOW)   No migrations run yet$(NC)"
	@echo ""
	@echo "$(YELLOW)3. Tables:$(NC)"
	@$(COMPOSE) exec postgres psql -U $(DB_USER) -d $(DB_NAME) -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';" 2>/dev/null | xargs echo "   Tables found:" || echo "$(YELLOW)   Could not count tables$(NC)"
	@echo ""

# =============================================================================
# Setup & Installation Commands
# =============================================================================

.PHONY: setup
setup: ## Complete setup: start services + run migrations (one command!)
	@echo "$(GREEN)═══════════════════════════════════════════════════════════════$(NC)"
	@echo "$(GREEN)  Starting Complete Setup$(NC)"
	@echo "$(GREEN)═══════════════════════════════════════════════════════════════$(NC)"
	@echo ""
	@echo "$(YELLOW)Step 1/4: Checking .env file...$(NC)"
	@if [ ! -f .env ]; then \
		if [ -f .env.example ]; then \
			echo "$(YELLOW)Creating .env from .env.example...$(NC)"; \
			cp .env.example .env; \
			echo "$(GREEN)✓ .env file created!$(NC)"; \
			echo "$(YELLOW)⚠ Using default values. Update .env if needed.$(NC)"; \
		fi \
	else \
		echo "$(GREEN)✓ .env file exists$(NC)"; \
	fi
	@echo ""
	@echo "$(YELLOW)Step 2/4: Starting Docker services...$(NC)"
	@$(MAKE) docker-up
	@echo ""
	@echo "$(YELLOW)Step 3/4: Waiting for database to be ready...$(NC)"
	@sleep 5
	@echo ""
	@echo "$(YELLOW)Step 4/4: Running migrations...$(NC)"
	@$(MAKE) migrate-up
	@echo ""
	@echo "$(GREEN)═══════════════════════════════════════════════════════════════$(NC)"
	@echo "$(GREEN)  Setup Complete! 🎉$(NC)"
	@echo "$(GREEN)═══════════════════════════════════════════════════════════════$(NC)"
	@echo ""
	@echo "$(YELLOW)Next steps:$(NC)"
	@echo "  • Start development: $(GREEN)make dev$(NC)"
	@echo "  • View database: $(GREEN)make db-shell$(NC)"
	@echo "  • Check status: $(GREEN)make verify$(NC)"
	@echo "  • View all commands: $(GREEN)make help$(NC)"
	@echo ""

.PHONY: init
init: ## First-time initialization (create .env from example)
	@if [ ! -f .env ]; then \
		if [ -f .env.example ]; then \
			echo "$(GREEN)Creating .env from .env.example...$(NC)"; \
			cp .env.example .env; \
			echo "$(GREEN)✓ .env file created!$(NC)"; \
			echo "$(YELLOW)⚠ Please review and update .env file (especially DB_PASSWORD)$(NC)"; \
		else \
			echo "$(RED)Error: .env.example not found$(NC)"; \
			exit 1; \
		fi \
	else \
		echo "$(YELLOW).env file already exists$(NC)"; \
	fi

.PHONY: reset
reset: docker-clean setup ## Reset everything (clean + setup)

.PHONY: quick-start
quick-start: setup dev ## Quick start for first-time users (setup + dev)

# =============================================================================
# Utility Commands
# =============================================================================

.PHONY: check
check: ## Quick health check (DB + migrations)
	@echo "$(GREEN)Running health checks...$(NC)"
	@echo ""
	@echo "$(YELLOW)Database:$(NC)"
	@$(COMPOSE) exec postgres pg_isready -U $(DB_USER) 2>/dev/null && echo "  $(GREEN)✓ Ready$(NC)" || echo "  $(RED)✗ Not ready$(NC)"
	@echo ""
	@echo "$(YELLOW)Migration Version:$(NC)"
	@$(COMPOSE) run --rm migrate -path=/migrations -database "$(DATABASE_URL)" version 2>/dev/null | head -n 1 || echo "  $(YELLOW)No migrations$(NC)"
	@echo ""

.PHONY: ps
ps: ## Show running containers
	$(COMPOSE) ps

.PHONY: stats
stats: ## Show container resource usage
	docker stats $(shell $(COMPOSE) ps -q)

.PHONY: shell
shell: ## Open shell in app container
	$(COMPOSE) exec app-dev sh || echo "$(YELLOW)Dev container not running. Try: make dev$(NC)"

.PHONY: clean-logs
clean-logs: ## Clean up Docker logs
	@echo "$(YELLOW)Cleaning Docker logs...$(NC)"
	$(COMPOSE) logs --no-color > /dev/null 2>&1 || true
	@echo "$(GREEN)✓ Logs cleaned!$(NC)"

# =============================================================================
# Special Targets
# =============================================================================

.DEFAULT_GOAL := help

# Mark all targets as phony (not files)
.PHONY: all

