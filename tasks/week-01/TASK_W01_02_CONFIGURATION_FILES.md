# Task W01-02: Configuration Files

**Status**: ✅ COMPLETED  
**Estimated Time**: 2-3 hours  
**Prerequisites**: TASK_W01_01_PROJECT_SETUP.md  
**Next Task**: TASK_W01_03_ADR_DOCUMENTS.md

---

## Objective

Create essential configuration files including Makefile, Dockerfile, docker-compose.yml, and environment configuration templates.

---

## Steps

### 1. Create .env.example

Location: `appointment-service/.env.example`

```bash
# Application
GO_ENV=development
API_PORT=8081
API_HOST=0.0.0.0

# Database
DB_HOST=localhost
DB_PORT=1998
DB_USER=appointments
DB_PASSWORD=change-me-in-production
DB_NAME=appointments_dev
DB_SSL_MODE=disable
DB_MAX_CONNECTIONS=25
DB_MAX_IDLE_CONNECTIONS=5
DB_CONNECTION_MAX_LIFETIME=5m

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Security
JWT_SECRET=your-jwt-secret-key-minimum-32-characters-change-in-production
API_SECRET_SALT_ROUNDS=10

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8081
CORS_ALLOWED_METHODS=GET,POST,PUT,PATCH,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=Content-Type,Authorization,X-API-Key,X-API-Secret

# Rate Limiting
RATE_LIMIT_REQUESTS_PER_MINUTE=100
RATE_LIMIT_BURST=20

# Logging
LOG_LEVEL=debug
LOG_FORMAT=json

# Monitoring
PROMETHEUS_ENABLED=true
PROMETHEUS_PORT=9090
```

### 2. Create Makefile

Location: `appointment-service/Makefile`

```makefile
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
DB_DSN=postgresql://appointments:password@localhost:1998/appointments_dev?sslmode=disable

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
```

### 3. Create Dockerfile

Location: `appointment-service/Dockerfile`

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o main ./cmd/api

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/main .

# Copy migrations
COPY --from=builder /app/migrations ./migrations

# Copy environment template
COPY --from=builder /app/.env.example .env.example

# Expose port
EXPOSE 8081

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8081/health || exit 1

# Run the application
CMD ["./main"]
```

### 4. Create .dockerignore

Location: `appointment-service/.dockerignore`

```md
# Git
.git
.gitignore
.github

# Documentation
*.md
docs/

# IDE
.vscode/
.idea/
*.swp
*.swo

# Environment
.env
.env.*
!.env.example

# Build artifacts
bin/
dist/
*.exe

# Tests
*_test.go
tests/
coverage.*

# Temporary files
tmp/
temp/
*.log

# Dependencies (will be downloaded in container)
vendor/
```

### 5. Create docker-compose.yml

Location: `appointment-service/docker-compose.yml`

```yaml
version: '3.8'

services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: appointment-service
    ports:
      - "8081:8081"
    environment:
      - GO_ENV=development
      - DB_HOST=postgres
      - DB_PORT=1998
      - DB_USER=appointments
      - DB_PASSWORD=secure_password
      - DB_NAME=appointments_dev
      - DB_SSL_MODE=disable
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - LOG_LEVEL=debug
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    networks:
      - appointment-network
    restart: unless-stopped
    volumes:
      - .:/app  # Mount for development

  postgres:
    image: postgres:15-alpine
    container_name: appointment-postgres
    environment:
      - POSTGRES_USER=appointments
      - POSTGRES_PASSWORD=secure_password
      - POSTGRES_DB=appointments_dev
    ports:
      - "1998:1998"
    volumes:
      - postgres-data:/var/lib/postgresql/data
    networks:
      - appointment-network
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U appointments"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    container_name: appointment-redis
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data
    networks:
      - appointment-network
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 3s
      retries: 5
    command: redis-server --appendonly yes

  prometheus:
    image: prom/prometheus:latest
    container_name: appointment-prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./deployments/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus-data:/prometheus
    networks:
      - appointment-network
    restart: unless-stopped
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'

  grafana:
    image: grafana/grafana:latest
    container_name: appointment-grafana
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_USER=admin
      - GF_SECURITY_ADMIN_PASSWORD=admin
      - GF_USERS_ALLOW_SIGN_UP=false
    volumes:
      - grafana-data:/var/lib/grafana
    networks:
      - appointment-network
    restart: unless-stopped
    depends_on:
      - prometheus

networks:
  appointment-network:
    driver: bridge

volumes:
  postgres-data:
  redis-data:
  prometheus-data:
  grafana-data:
```

### 6. Create .air.toml (Hot Reload Config)

Location: `appointment-service/.air.toml`

```toml
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = []
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main ./cmd/api"
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata", "tests", "docs", "deployments"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html"]
  include_file = []
  kill_delay = "0s"
  log = "build-errors.log"
  poll = false
  poll_interval = 0
  rerun = false
  rerun_delay = 500
  send_interrupt = false
  stop_on_error = false

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  main_only = false
  time = false

[misc]
  clean_on_exit = false

[screen]
  clear_on_rebuild = false
  keep_scroll = true
```

### 7. Create Prometheus Config

Location: `appointment-service/deployments/prometheus/prometheus.yml`

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'appointment-service'
    static_configs:
      - targets: ['app:8081']
    metrics_path: '/metrics'
```

### 8. Create .editorconfig

Location: `appointment-service/.editorconfig`

```ini
root = true

[*]
charset = utf-8
end_of_line = lf
insert_final_newline = true
trim_trailing_whitespace = true

[*.go]
indent_style = tab
indent_size = 4

[*.{yaml,yml}]
indent_style = space
indent_size = 2

[*.{json,md}]
indent_style = space
indent_size = 2

[Makefile]
indent_style = tab
```

### 9. Create .golangci.yml (Linter Config)

Location: `appointment-service/.golangci.yml`

```yaml
run:
  timeout: 5m
  tests: true
  skip-dirs:
    - vendor
    - tmp
    - docs
    - deployments

linters:
  enable:
    - errcheck
    - gosimple
    - govet
    - ineffassign
    - staticcheck
    - unused
    - gofmt
    - goimports
    - misspell
    - gocritic
    - revive

linters-settings:
  errcheck:
    check-blank: true
  govet:
    check-shadowing: true
  gofmt:
    simplify: true
  revive:
    rules:
      - name: exported
        disabled: false

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - errcheck
```

### 10. Commit Changes

```bash
# Add all configuration files
git add .

# Commit
git commit -m "chore: add configuration files

- Add Makefile with development commands
- Add Dockerfile and docker-compose.yml
- Add .env.example with all config options
- Add .air.toml for hot reload
- Add .editorconfig and .golangci.yml
- Add Prometheus configuration"

# Push
git push origin master
```

---

## Verification Checklist

- [x] .env.example created with all configuration options
- [x] Makefile created with all development commands
- [x] Dockerfile created for production builds
- [x] .dockerignore configured
- [x] docker-compose.yml created with all services
- [x] .air.toml created for hot reload
- [x] .editorconfig created
- [x] .golangci.yml created for linting
- [x] Prometheus config created
- [x] All files committed and pushed

---

## Testing

```bash
# Test Makefile
make help  # Should show all available commands

# Test Docker Compose (don't start yet, just validate)
docker-compose config  # Should show no errors

# Verify files exist
ls -la | grep -E "Makefile|Dockerfile|docker-compose"
```

---

## Next Steps

Proceed to **TASK_W01_03_ADR_DOCUMENTS.md** to create Architecture Decision Records.

---

**Status**: ✅ COMPLETED (Commit: 5701af2)
