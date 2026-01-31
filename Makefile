.PHONY: help install build run test lint fmt clean migrate-up migrate-down migrate-create docker-build docker-up docker-down dev install-tools swagger

# Application
APP_NAME=appointment-service
BUILD_DIR=bin
MAIN_PATH=cmd/api/main.go

# Go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GORUN=$(GOCMD) run
GOFMT=$(GOCMD) fmt
BINARY_NAME=$(APP_NAME)

# Database
DB_DSN=postgresql://appointments:password@localhost:5432/appointments_dev?sslmode=disable

# Colors
CYAN=\033[0;36m
GREEN=\033[0;32m
RED=\033[0;31m
NC=\033[0m

help: ## Show this help message
	@echo "$(CYAN)Available targets:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(CYAN)%-20s$(NC) %s\n", $$1, $$2}'

install: ## Install Go dependencies
	@echo "$(CYAN)Installing dependencies...$(NC)"
	$(GOGET) github.com/gin-gonic/gin@latest
	$(GOGET) github.com/jackc/pgx/v5@latest
	$(GOGET) github.com/jackc/pgx/v5/pgxpool@latest
	$(GOGET) github.com/golang-jwt/jwt/v5@latest
	$(GOGET) go.uber.org/zap@latest
	$(GOGET) github.com/google/uuid@latest
	$(GOGET) github.com/joho/godotenv@latest
	$(GOGET) github.com/go-playground/validator/v10@latest
	$(GOGET) golang.org/x/crypto/bcrypt@latest
	$(GOGET) github.com/redis/go-redis/v9@latest
	$(GOGET) github.com/stretchr/testify@latest
	$(GOMOD) tidy
	@echo "$(GREEN)Dependencies installed successfully$(NC)"

build: ## Build the application
	@echo "$(CYAN)Building $(APP_NAME)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "$(GREEN)Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

run: ## Run the application
	@echo "$(CYAN)Running $(APP_NAME)...$(NC)"
	$(GORUN) $(MAIN_PATH)

test: ## Run tests
	@echo "$(CYAN)Running tests...$(NC)"
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	@echo "$(GREEN)Tests complete$(NC)"

test-coverage: test ## Generate test coverage report
	@echo "$(CYAN)Generating coverage report...$(NC)"
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report: coverage.html$(NC)"

lint: ## Run linter
	@echo "$(CYAN)Running linter...$(NC)"
	golangci-lint run ./...
	@echo "$(GREEN)Linting complete$(NC)"

fmt: ## Format code
	@echo "$(CYAN)Formatting code...$(NC)"
	$(GOFMT) ./...
	gofmt -s -w .
	@echo "$(GREEN)Formatting complete$(NC)"

clean: ## Clean build artifacts
	@echo "$(CYAN)Cleaning...$(NC)"
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	@echo "$(GREEN)Clean complete$(NC)"

migrate-up: ## Run database migrations up
	@echo "$(CYAN)Running migrations up...$(NC)"
	migrate -path migrations -database "$(DB_DSN)" up
	@echo "$(GREEN)Migrations applied$(NC)"

migrate-down: ## Run database migrations down
	@echo "$(CYAN)Running migrations down...$(NC)"
	migrate -path migrations -database "$(DB_DSN)" down
	@echo "$(GREEN)Migrations reverted$(NC)"

migrate-create: ## Create new migration (usage: make migrate-create name=create_users)
	@echo "$(CYAN)Creating migration: $(name)$(NC)"
	@if [ -z "$(name)" ]; then \
		echo "$(RED)Error: name is required. Usage: make migrate-create name=create_users$(NC)"; \
		exit 1; \
	fi
	migrate create -ext sql -dir migrations -seq $(name)
	@echo "$(GREEN)Migration created$(NC)"

docker-build: ## Build Docker image
	@echo "$(CYAN)Building Docker image...$(NC)"
	docker build -t $(APP_NAME):latest .
	@echo "$(GREEN)Docker image built$(NC)"

docker-up: ## Start services with Docker Compose
	@echo "$(CYAN)Starting services...$(NC)"
	docker-compose up -d
	@echo "$(GREEN)Services started$(NC)"

docker-down: ## Stop services with Docker Compose
	@echo "$(CYAN)Stopping services...$(NC)"
	docker-compose down
	@echo "$(GREEN)Services stopped$(NC)"

docker-logs: ## View Docker Compose logs
	docker-compose logs -f

dev: ## Run in development mode with hot reload (requires air)
	@echo "$(CYAN)Starting development server with hot reload...$(NC)"
	@if ! command -v air > /dev/null; then \
		echo "$(RED)Error: air is not installed. Run: make install-tools$(NC)"; \
		exit 1; \
	fi
	air

install-tools: ## Install development tools
	@echo "$(CYAN)Installing development tools...$(NC)"
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/cosmtrek/air@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	@echo "$(GREEN)Development tools installed$(NC)"

swagger: ## Generate Swagger documentation
	@echo "$(CYAN)Generating Swagger documentation...$(NC)"
	swag init -g cmd/api/main.go -o docs/swagger
	@echo "$(GREEN)Swagger documentation generated$(NC)"

.DEFAULT_GOAL := help
