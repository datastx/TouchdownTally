# TouchdownTally - Family Football Pool Application
# Makefile for development and deployment tasks

# Variables
GO_MODULE := touchdown-tally
BACKEND_DIR := backend
FRONTEND_DIR := frontend
DOCKER_COMPOSE := docker-compose

# Colors for output
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
NC := \033[0m # No Color

.PHONY: help setup clean build test run dev deploy

# Default target
help: ## Show this help message
	@echo "TouchdownTally - Family Football Pool Application"
	@echo "================================================"
	@echo "Available commands:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

setup: ## Initial project setup (run after container creation)
	@echo "$(YELLOW)Setting up TouchdownTally development environment...$(NC)"
	@if [ -f .devcontainer/startup.sh ]; then \
		.devcontainer/startup.sh; \
	fi
	@$(MAKE) setup-backend
	@$(MAKE) setup-frontend
	@echo "$(GREEN)Setup complete! Ready for development.$(NC)"

setup-backend: ## Setup Go backend
	@echo "$(YELLOW)Setting up Go backend...$(NC)"
	@mkdir -p $(BACKEND_DIR)
	@cd $(BACKEND_DIR) && \
		if [ ! -f go.mod ]; then \
			go mod init $(GO_MODULE); \
		fi && \
		go mod tidy
	@echo "$(GREEN)Backend setup complete.$(NC)"

setup-frontend: ## Setup Vue.js frontend
	@echo "$(YELLOW)Setting up Vue.js frontend...$(NC)"
	@if [ ! -d "$(FRONTEND_DIR)" ]; then \
		vue create $(FRONTEND_DIR) --preset default --packageManager npm; \
		cd $(FRONTEND_DIR) && npm install vuetify @mdi/font axios @types/node; \
	fi
	@echo "$(GREEN)Frontend setup complete.$(NC)"

install: ## Install dependencies for both backend and frontend
	@echo "$(YELLOW)Installing dependencies...$(NC)"
	@cd $(BACKEND_DIR) && go mod download
	@cd $(FRONTEND_DIR) && npm install
	@echo "$(GREEN)Dependencies installed.$(NC)"

# Development commands
dev: ## Start development servers (backend and frontend)
	@echo "$(YELLOW)Starting development servers...$(NC)"
	@$(MAKE) -j2 dev-backend dev-frontend

dev-backend: ## Start Go backend in development mode
	@echo "$(YELLOW)Starting Go backend on :8080...$(NC)"
	@cd $(BACKEND_DIR) && \
		go run -ldflags="-X main.version=dev" ./cmd/server

dev-frontend: ## Start Vue.js frontend in development mode
	@echo "$(YELLOW)Starting Vue.js frontend on :3000...$(NC)"
	@cd $(FRONTEND_DIR) && npm run serve

# Database commands
db-up: ## Start database services
	@echo "$(YELLOW)Starting database services...$(NC)"
	@$(DOCKER_COMPOSE) up -d postgres redis
	@echo "$(GREEN)Database services started.$(NC)"

db-down: ## Stop database services
	@echo "$(YELLOW)Stopping database services...$(NC)"
	@$(DOCKER_COMPOSE) down
	@echo "$(GREEN)Database services stopped.$(NC)"

db-reset: ## Reset database (WARNING: destroys all data)
	@echo "$(RED)WARNING: This will destroy all database data!$(NC)"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		echo ""; \
		$(DOCKER_COMPOSE) down -v; \
		$(DOCKER_COMPOSE) up -d postgres redis; \
		echo "$(GREEN)Database reset complete.$(NC)"; \
	else \
		echo ""; \
		echo "$(YELLOW)Database reset cancelled.$(NC)"; \
	fi

db-migrate: ## Run database migrations
	@echo "$(YELLOW)Running database migrations...$(NC)"
	@cd $(BACKEND_DIR) && go run ./cmd/migrate

db-seed: ## Seed database with sample data
	@echo "$(YELLOW)Seeding database with sample data...$(NC)"
	@cd $(BACKEND_DIR) && go run ./cmd/seed

db-console: ## Connect to PostgreSQL console
	@echo "$(YELLOW)Connecting to PostgreSQL console...$(NC)"
	@docker exec -it $$($(DOCKER_COMPOSE) ps -q postgres) psql -U touchdown_user -d touchdown_tally

# Testing commands
test: ## Run all tests
	@echo "$(YELLOW)Running all tests...$(NC)"
	@$(MAKE) test-backend
	@$(MAKE) test-frontend
	@echo "$(GREEN)All tests completed.$(NC)"

test-backend: ## Run Go backend tests
	@echo "$(YELLOW)Running Go backend tests...$(NC)"
	@cd $(BACKEND_DIR) && go test -v ./...

test-frontend: ## Run Vue.js frontend tests
	@echo "$(YELLOW)Running Vue.js frontend tests...$(NC)"
	@cd $(FRONTEND_DIR) && npm run test:unit

test-coverage: ## Run tests with coverage report
	@echo "$(YELLOW)Running tests with coverage...$(NC)"
	@cd $(BACKEND_DIR) && go test -coverprofile=coverage.out ./...
	@cd $(BACKEND_DIR) && go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated: $(BACKEND_DIR)/coverage.html$(NC)"

# Build commands
build: ## Build both backend and frontend for production
	@echo "$(YELLOW)Building for production...$(NC)"
	@$(MAKE) build-backend
	@$(MAKE) build-frontend
	@echo "$(GREEN)Build complete.$(NC)"

build-backend: ## Build Go backend binary
	@echo "$(YELLOW)Building Go backend...$(NC)"
	@cd $(BACKEND_DIR) && \
		CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o bin/server ./cmd/server
	@echo "$(GREEN)Backend build complete: $(BACKEND_DIR)/bin/server$(NC)"

build-frontend: ## Build Vue.js frontend for production
	@echo "$(YELLOW)Building Vue.js frontend...$(NC)"
	@cd $(FRONTEND_DIR) && npm run build
	@echo "$(GREEN)Frontend build complete: $(FRONTEND_DIR)/dist/$(NC)"

# Code quality commands
lint: ## Run linters for all code
	@echo "$(YELLOW)Running linters...$(NC)"
	@$(MAKE) lint-backend
	@$(MAKE) lint-frontend
	@echo "$(GREEN)Linting complete.$(NC)"

lint-backend: ## Run Go linters
	@echo "$(YELLOW)Running Go linters...$(NC)"
	@cd $(BACKEND_DIR) && golangci-lint run

lint-frontend: ## Run frontend linters
	@echo "$(YELLOW)Running frontend linters...$(NC)"
	@cd $(FRONTEND_DIR) && npm run lint

format: ## Format all code
	@echo "$(YELLOW)Formatting code...$(NC)"
	@$(MAKE) format-backend
	@$(MAKE) format-frontend
	@echo "$(GREEN)Code formatting complete.$(NC)"

format-backend: ## Format Go code
	@cd $(BACKEND_DIR) && gofmt -s -w .
	@cd $(BACKEND_DIR) && goimports -w .

format-frontend: ## Format frontend code
	@cd $(FRONTEND_DIR) && npm run lint -- --fix

# Utility commands
clean: ## Clean build artifacts and dependencies
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	@rm -rf $(BACKEND_DIR)/bin
	@rm -rf $(FRONTEND_DIR)/dist
	@rm -rf $(FRONTEND_DIR)/node_modules
	@cd $(BACKEND_DIR) && go clean -modcache
	@echo "$(GREEN)Clean complete.$(NC)"

logs: ## Show logs from all services
	@$(DOCKER_COMPOSE) logs -f

logs-db: ## Show database logs
	@$(DOCKER_COMPOSE) logs -f postgres

generate: ## Generate Go code (models, mocks, etc.)
	@echo "$(YELLOW)Generating Go code...$(NC)"
	@cd $(BACKEND_DIR) && go generate ./...

update: ## Update all dependencies
	@echo "$(YELLOW)Updating dependencies...$(NC)"
	@cd $(BACKEND_DIR) && go get -u ./... && go mod tidy
	@cd $(FRONTEND_DIR) && npm update
	@echo "$(GREEN)Dependencies updated.$(NC)"

# NFL Data commands
fetch-nfl-data: ## Fetch latest NFL data from MySportsFeeds
	@echo "$(YELLOW)Fetching latest NFL data...$(NC)"
	@cd $(BACKEND_DIR) && go run ./cmd/fetch-nfl-data

update-scores: ## Update game scores
	@echo "$(YELLOW)Updating game scores...$(NC)"
	@cd $(BACKEND_DIR) && go run ./cmd/update-scores

# Development utilities
mock: ## Generate mocks for testing
	@echo "$(YELLOW)Generating mocks...$(NC)"
	@cd $(BACKEND_DIR) && go generate ./...

api-docs: ## Generate API documentation
	@echo "$(YELLOW)Generating API documentation...$(NC)"
	@cd $(BACKEND_DIR) && swag init -g ./cmd/server/main.go

# Docker commands
docker-build: ## Build Docker image
	@echo "$(YELLOW)Building Docker image...$(NC)"
	@docker build -t touchdown-tally:latest .

docker-run: ## Run application in Docker
	@echo "$(YELLOW)Running application in Docker...$(NC)"
	@docker-compose up -d

docker-stop: ## Stop Docker containers
	@$(DOCKER_COMPOSE) down

# Production/Deployment commands
deploy-staging: ## Deploy to staging environment
	@echo "$(YELLOW)Deploying to staging...$(NC)"
	@$(MAKE) build
	@echo "$(GREEN)Staging deployment complete.$(NC)"

deploy-prod: ## Deploy to production (Railway)
	@echo "$(YELLOW)Deploying to production...$(NC)"
	@railway deploy
	@echo "$(GREEN)Production deployment complete.$(NC)"

# Security commands
security-scan: ## Run security scans
	@echo "$(YELLOW)Running security scans...$(NC)"
	@cd $(BACKEND_DIR) && gosec ./...
	@cd $(FRONTEND_DIR) && npm audit

# Health checks
health: ## Check health of all services
	@echo "$(YELLOW)Checking service health...$(NC)"
	@curl -f http://localhost:8080/health || echo "$(RED)Backend unhealthy$(NC)"
	@curl -f http://localhost:3000 || echo "$(RED)Frontend unhealthy$(NC)"
	@$(DOCKER_COMPOSE) exec postgres pg_isready -U touchdown_user || echo "$(RED)Database unhealthy$(NC)"

# Environment info
info: ## Show environment information
	@echo "TouchdownTally Development Environment"
	@echo "===================================="
	@echo "Go version: $$(go version)"
	@if command -v node > /dev/null; then echo "Node version: $$(node --version)"; fi
	@if command -v npm > /dev/null; then echo "npm version: $$(npm --version)"; fi
	@if command -v vue > /dev/null; then echo "Vue CLI version: $$(vue --version)"; fi
	@echo "Docker version: $$(docker --version)"
	@echo "Docker Compose version: $$(docker-compose --version)"
	@echo ""
	@echo "Docker services status:"
	@$(DOCKER_COMPOSE) ps

docker-status: ## Show status of Docker services
	@echo "$(YELLOW)Docker services status:$(NC)"
	@$(DOCKER_COMPOSE) ps
	@echo ""
	@echo "$(YELLOW)Docker system info:$(NC)"
	@docker system df

# Quick start for new developers
quickstart: ## Quick start for new developers
	@echo "$(GREEN)Welcome to TouchdownTally!$(NC)"
	@echo "$(YELLOW)Running initial setup...$(NC)"
	@$(MAKE) setup
	@$(MAKE) db-up
	@sleep 5
	@$(MAKE) db-migrate
	@$(MAKE) db-seed
	@echo "$(GREEN)Setup complete! Run 'make dev' to start development servers.$(NC)"
