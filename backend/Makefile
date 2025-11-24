# =============================================================================
# Makefile for File Sharing API - Docker & Database Management
# =============================================================================
# Author: DATH-251 Team
# Usage: make [command]
# =============================================================================

-include .env
export

# =============================================================================
# Configuration
# =============================================================================

COMPOSE := docker compose
GREEN := \033[0;32m
YELLOW := \033[1;33m
RED := \033[0;31m
NC := \033[0m

# =============================================================================
# Help Command (Default)
# =============================================================================

.PHONY: help
help: ## Show available commands
	@echo "$(GREEN)═══════════════════════════════════════════════════════$(NC)"
	@echo "$(GREEN)  File Sharing API - Make Commands$(NC)"
	@echo "$(GREEN)═══════════════════════════════════════════════════════$(NC)"
	@echo ""
	@echo "$(YELLOW)Main Commands:$(NC)"
	@echo "  $(GREEN)make dev$(NC)          - Start development (hot reload)"
	@echo "  $(GREEN)make app$(NC)          - Start production app"
	@echo "  $(GREEN)make build$(NC)        - Build Docker images"
	@echo "  $(GREEN)make down$(NC)         - Stop all services"
	@echo "  $(GREEN)make logs$(NC)         - View logs"
	@echo ""
	@echo "$(YELLOW)Database:$(NC)"
	@echo "  $(GREEN)make db-reset$(NC)     - Reset database (clean + setup)"
	@echo "  $(GREEN)make db-shell$(NC)     - Open PostgreSQL shell"
	@echo ""
	@echo "$(YELLOW)Other:$(NC)"
	@echo "  $(GREEN)make clean$(NC)        - Remove all containers & volumes"
	@echo ""

# =============================================================================
# Main Commands
# =============================================================================

.PHONY: build
build: ## Build Docker images
	@echo "$(GREEN)Building Docker images...$(NC)"
	$(COMPOSE) build
	@echo "$(GREEN)✓ Build complete!$(NC)"

.PHONY: dev
dev: ## Start development (hot reload, port 8082)
	@echo "$(GREEN)Starting development environment...$(NC)"
	$(COMPOSE) up -d postgres app-dev
	@echo "$(GREEN)✓ Dev started!$(NC)"
	@echo "$(YELLOW)API: http://localhost:8082$(NC)"
	@echo "$(YELLOW)Logs: make logs$(NC)"

.PHONY: app
app: ## Start production app (port 8080)
	@echo "$(GREEN)Starting production app...$(NC)"
	$(COMPOSE) up -d postgres app
	@echo "$(GREEN)✓ App started!$(NC)"
	@echo "$(YELLOW)API: http://localhost:8080$(NC)"

.PHONY: down
down: ## Stop all services
	@echo "$(YELLOW)Stopping services...$(NC)"
	$(COMPOSE) down
	@echo "$(GREEN)✓ Stopped!$(NC)"

.PHONY: restart
restart: down dev ## Restart dev environment

# =============================================================================
# Database Commands
# =============================================================================

.PHONY: db-reset
db-reset: ## Reset database (clean + restart)
	@echo "$(YELLOW)Resetting database...$(NC)"
	$(COMPOSE) down -v
	$(COMPOSE) up -d postgres
	@echo "$(GREEN)✓ Database reset complete!$(NC)"

.PHONY: db-shell
db-shell: ## Open PostgreSQL shell
	@echo "$(GREEN)Opening database shell...$(NC)"
	$(COMPOSE) exec postgres psql -U postgres -d file_sharing_db


# =============================================================================
# Logs & Monitoring
# =============================================================================

.PHONY: logs
logs: ## View logs (all services)
	$(COMPOSE) logs -f

.PHONY: logs-dev
logs-dev: ## View dev logs only
	$(COMPOSE) logs -f app-dev

.PHONY: logs-app
logs-app: ## View production app logs only
	$(COMPOSE) logs -f app

.PHONY: ps
ps: ## Show running containers
	$(COMPOSE) ps

# =============================================================================
# Cleanup
# =============================================================================

.PHONY: clean
clean: ## Remove all containers, volumes, and data
	@echo "$(RED)WARNING: This will remove all data!$(NC)"
	@read -p "Type 'yes' to confirm: " confirm; \
	if [ "$$confirm" = "yes" ]; then \
		$(COMPOSE) down -v --remove-orphans; \
		echo "$(GREEN)✓ Cleaned!$(NC)"; \
	else \
		echo "$(YELLOW)Cancelled.$(NC)"; \
	fi

# =============================================================================
# Default
# =============================================================================

.DEFAULT_GOAL := help
