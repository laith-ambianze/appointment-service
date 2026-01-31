# Week 01: Immediate Actions - Project Foundation

**Priority**: Critical  
**Estimated Time**: 1 Week (40 hours)  
**Team Size**: 3-4 people  
**Status**: Not Started

---

## Overview

This week focuses on establishing the foundation for the appointment microservice project. The goal is to align the team, finalize technical decisions, and set up the development infrastructure.

---

## Team Roles

- **Tech Lead**: Architecture decisions, code review setup
- **Backend Developer 1**: Environment setup, repository structure
- **Backend Developer 2**: Database setup, migration planning
- **DevOps Engineer**: CI/CD pipeline, Docker configuration

---

## Day 1: Monday - Document Review & Architecture Alignment

### Day 1 Morning (9:00 AM - 12:00 PM)

#### Task 1.1: Team Document Review Meeting

**Duration**: 2 hours

**Attendees**: All team members + Product Owner

**Agenda**:

1. Review MIGRATION_STRATEGY_AND_ARCHITECTURE.md (60 min)
2. Review DESIGN_DECISION_PARTICIPANTS.md (30 min)
3. Q&A and clarifications (30 min)

**Action Items**:

- [ ] Schedule meeting room/video call
- [ ] Send documents to all attendees 24h in advance
- [ ] Prepare presentation slides for key concepts
- [ ] Create shared Google Doc for questions/notes

**Deliverables**:

- Meeting notes document
- List of clarification questions
- Team sign-off on approach

#### Task 1.2: Technical Architecture Deep Dive

**Duration**: 1 hour

**Participants**: Tech Lead + Senior Developers

**Topics**:

- Multi-tenant isolation strategy
- pgx vs GORM decision confirmation
- API authentication approach
- Webhook vs polling for events
- Caching strategy (Redis)

**Deliverables**:

- Architecture Decision Records (ADRs) document
- List of technical risks
- Mitigation strategies

### Day 1 Afternoon (1:00 PM - 5:00 PM)

#### Task 1.3: Create Technical Specifications

**Duration**: 3 hours

**Assignee**: Tech Lead + Backend Dev 1

**Activities**:

- [ ] Define API contract (OpenAPI/Swagger spec)
- [ ] Document database schema with relationships
- [ ] Define data migration steps
- [ ] Create sequence diagrams for key flows
- [ ] Document error handling strategy

**Deliverables**:

- `docs/API_SPECIFICATION.yaml` (OpenAPI 3.0)
- `docs/DATABASE_SCHEMA.md` with ER diagrams
- `docs/MIGRATION_PLAN.md` detailed steps
- `docs/SEQUENCE_DIAGRAMS.md`

**Template**:

```yaml
# API_SPECIFICATION.yaml
openapi: 3.0.0
info:
  title: Appointment Service API
  version: 1.0.0
  description: Multi-tenant appointment management microservice

servers:
  - url: https://api.appointments.company.com/v1
    description: Production
  - url: http://localhost:8080/v1
    description: Development

paths:
  /products/register:
    post:
      summary: Register a new product
      # ... details
```

#### Task 1.4: Risk Assessment Workshop

**Duration**: 1 hour

**Participants**: All team members

**Activities**:

- [ ] Identify technical risks
- [ ] Identify business risks
- [ ] Assess impact and probability
- [ ] Define mitigation strategies
- [ ] Create risk register

**Deliverables**:

- `docs/RISK_REGISTER.md`

**Risk Categories**:

- Data migration failures
- Performance issues
- Security vulnerabilities
- Integration complexity
- Timeline overruns

---

## Day 2: Tuesday - Finalize Decisions & Planning

### Day 2 Morning (9:00 AM - 12:00 PM)

#### Task 2.1: Technology Stack Finalization

**Duration**: 2 hours

**Participants**: Tech Lead + DevOps

**Decisions to Make**:

1. **Database Access Layer**
   - ✅ **Decision**: pgx (raw SQL) with repository pattern
   - **Rationale**: Performance, JSONB support, complex queries
   - **Document**: Create ADR-001-database-access-layer.md

2. **API Framework**
   - ✅ **Decision**: Gin (recommended)
   - **Alternative**: Fiber (if performance is critical)
   - **Document**: Create ADR-002-api-framework.md

3. **Logging**
   - ✅ **Decision**: Zap (structured logging)
   - **Document**: Create ADR-003-logging.md

4. **Configuration Management**
   - ✅ **Decision**: godotenv + viper
   - **Document**: Create ADR-004-configuration.md

5. **Testing Framework**
   - ✅ **Decision**: Go testing + testify
   - **Document**: Create ADR-005-testing.md

6. **API Documentation**
   - ✅ **Decision**: Swagger/OpenAPI + swag
   - **Document**: Create ADR-006-api-documentation.md

**Deliverables**:

- 6 ADR documents in `docs/adr/` folder
- Updated `docs/TECHNOLOGY_STACK.md`

**ADR Template**:

```markdown
# ADR-001: Database Access Layer

**Status**: Accepted  
**Date**: 2026-01-31  
**Deciders**: Tech Lead, Senior Developers

## Context
We need to choose between GORM (ORM) and pgx (raw SQL) for database access.

## Decision
We will use pgx with raw SQL.

## Rationale
1. Performance: 20-30% faster for our use case
2. JSONB support: Better control over JSONB queries
3. Complex joins: Participants pattern requires optimized queries
4. Transparency: Full control over SQL queries

## Consequences
- More SQL code to write
- Need to maintain query strings
- Easier to optimize performance
- Steeper learning curve for junior developers

## Alternatives Considered
- GORM: Easier to use but slower, less control
- sqlx: Middle ground but less features than pgx
```

#### Task 2.2: Sprint Planning

**Duration**: 2 hours

**Participants**: All team members + Product Owner

**Activities**:

- [ ] Break down 16-week roadmap into 2-week sprints
- [ ] Define Sprint 1 goals (Weeks 1-2)
- [ ] Create user stories with acceptance criteria
- [ ] Estimate story points
- [ ] Assign tasks to team members

**Deliverables**:

- Sprint 1 backlog in Jira/GitHub Projects
- Velocity target: 40 story points
- Definition of Done checklist

**Sprint 1 Goals**:

1. Project setup complete
2. Database schema created
3. Authentication middleware working
4. Product registration API functional
5. Basic appointment CRUD APIs

### Day 2 Afternoon (1:00 PM - 5:00 PM)

#### Task 2.3: Development Environment Standards

**Duration**: 2 hours

**Assignee**: DevOps + Backend Dev 1

**Activities**:

- [ ] Define development machine requirements
- [ ] Create development setup guide
- [ ] Document coding standards
- [ ] Set up pre-commit hooks
- [ ] Configure IDE settings (VS Code)

**Deliverables**:

- `docs/DEVELOPMENT_SETUP.md`
- `docs/CODING_STANDARDS.md`
- `.editorconfig` file
- `.vscode/settings.json`
- `.golangci.yml` (linter config)

Example: **Coding Standards**

```go
// Package naming: lowercase, single word
package appointment

// Function naming: PascalCase for exported, camelCase for private
func CreateAppointment() {} // Exported
func validateRequest() {}   // Private

// Error handling: always check errors
appointment, err := service.Create(req)
if err != nil {
    return fmt.Errorf("failed to create appointment: %w", err)
}

// Constants: PascalCase with descriptive names
const (
    StatusPending   = "pending"
    StatusConfirmed = "confirmed"
)

// Comments: Required for all exported functions
// CreateAppointment creates a new appointment with participants.
// It validates the request, checks availability, and stores the appointment.
func CreateAppointment(ctx context.Context, req CreateRequest) (*Appointment, error) {
    // Implementation
}
```

#### Task 2.4: Communication & Collaboration Setup

**Duration**: 1 hour

**Assignee**: Tech Lead

**Activities**:

- [ ] Create Slack/Teams channels
  - `#appointments-dev` - Development discussions
  - `#appointments-deployments` - Deployment notifications
  - `#appointments-alerts` - Production alerts
- [ ] Set up daily standup schedule (9:30 AM)
- [ ] Configure GitHub notifications
- [ ] Create wiki pages for documentation
- [ ] Set up code review guidelines

**Deliverables**:

- Communication channels created
- Meeting invites sent
- `docs/CODE_REVIEW_GUIDELINES.md`
- `docs/COMMUNICATION_PROTOCOL.md`

---

## Day 3: Wednesday - Development Environment Setup

### Day 3 Morning (9:00 AM - 12:00 PM)

#### Task 3.1: Create Git Repository

**Duration**: 1 hour

**Assignee**: Tech Lead

**Activities**:

- [ ] Create GitHub repository: `appointment-service`
- [ ] Set up branch protection rules
  - `master` branch protected
  - Require PR reviews (minimum 1)
  - Require status checks to pass
  - Require linear history
- [ ] Create branch naming convention
  - `feature/*` for new features
  - `bugfix/*` for bug fixes
  - `hotfix/*` for production fixes
- [ ] Set up GitHub labels
- [ ] Configure repository settings

**Deliverables**:

- Repository URL: `https://github.com/laith-ambianze/appointment-service`
- Branch protection configured
- `README.md` with project overview
- `CONTRIBUTING.md` with contribution guidelines
- `CODE_OF_CONDUCT.md`

**README.md Template**:

```markdown
# Appointment Service

Multi-tenant appointment management microservice.

## Features

- 🏢 Multi-tenant with product isolation
- 👥 Flexible participant model (1-on-1 and group appointments)
- 🔐 Secure API key authentication
- 🌍 Timezone support
- 📅 Availability management
- 🔔 Webhook notifications

## Tech Stack

- **Language**: Go 1.21+
- **Framework**: Gin
- **Database**: PostgreSQL 15+
- **Cache**: Redis 7+
- **Deployment**: Docker + Kubernetes

## Quick Start

See [DEVELOPMENT_SETUP.md](docs/DEVELOPMENT_SETUP.md)

## Documentation

- [API Specification](docs/API_SPECIFICATION.yaml)
- [Architecture](MIGRATION_STRATEGY_AND_ARCHITECTURE.md)
- [Database Schema](docs/DATABASE_SCHEMA.md)

## License

MIT
```

#### Task 3.2: Initialize Go Project Structure

**Duration**: 2 hours

**Assignee**: Backend Dev 1

**Activities**:

- [ ] Initialize Go module
- [ ] Create directory structure
- [ ] Add `.gitignore`
- [ ] Create Makefile
- [ ] Set up go.mod dependencies

**Commands**:

```bash
# Create project directory
mkdir appointment-service
cd appointment-service

# Initialize Go module
go mod init github.com/laith-ambianze/appointment-service

# Create directory structure
mkdir -p cmd/api
mkdir -p internal/{handlers,middleware,models,repository,service,config}
mkdir -p pkg/{auth,logger,validator,database}
mkdir -p migrations
mkdir -p tests/{unit,integration}
mkdir -p docs
mkdir -p scripts
mkdir -p deployments/{docker,kubernetes}

# Create initial files
touch cmd/api/main.go
touch internal/config/config.go
touch pkg/logger/logger.go
touch .env.example
touch .gitignore
touch Makefile
touch docker-compose.yml
touch Dockerfile
```

**Project Structure**:

```md
appointment-service/
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── internal/                    # Private application code
│   ├── handlers/                # HTTP handlers
│   │   ├── appointment.go
│   │   ├── product.go
│   │   └── health.go
│   ├── middleware/              # HTTP middleware
│   │   ├── auth.go
│   │   ├── cors.go
│   │   ├── logger.go
│   │   └── ratelimit.go
│   ├── models/                  # Domain models
│   │   ├── appointment.go
│   │   ├── product.go
│   │   └── participant.go
│   ├── repository/              # Data access layer
│   │   ├── appointment_repo.go
│   │   └── product_repo.go
│   ├── service/                 # Business logic
│   │   ├── appointment_service.go
│   │   └── product_service.go
│   ├── config/                  # Configuration
│   │   └── config.go
│   └── routes/                  # Route definitions
│       └── routes.go
├── pkg/                         # Public packages
│   ├── auth/                    # Authentication utilities
│   │   ├── jwt.go
│   │   └── hash.go
│   ├── logger/                  # Logging utilities
│   │   └── logger.go
│   ├── validator/               # Validation utilities
│   │   └── validator.go
│   └── database/                # Database utilities
│       └── postgres.go
├── migrations/                  # Database migrations
│   ├── 000001_create_products.up.sql
│   ├── 000001_create_products.down.sql
│   ├── 000002_create_appointments.up.sql
│   └── 000002_create_appointments.down.sql
├── tests/                       # Tests
│   ├── unit/                    # Unit tests
│   └── integration/             # Integration tests
├── docs/                        # Documentation
│   ├── API_SPECIFICATION.yaml
│   ├── DATABASE_SCHEMA.md
│   └── adr/                     # Architecture Decision Records
├── scripts/                     # Utility scripts
│   ├── migrate.sh
│   └── seed.sh
├── deployments/                 # Deployment configs
│   ├── docker/
│   │   └── Dockerfile
│   └── kubernetes/
│       ├── deployment.yaml
│       └── service.yaml
├── .env.example                 # Environment variables template
├── .gitignore                   # Git ignore rules
├── .golangci.yml               # Linter configuration
├── docker-compose.yml          # Docker Compose config
├── Dockerfile                  # Docker image definition
├── Makefile                    # Build automation
├── README.md                   # Project README
├── go.mod                      # Go module definition
└── go.sum                      # Go dependencies checksum
```

**Deliverables**:

- Complete directory structure
- Initialized Go module
- `.gitignore` configured
- `Makefile` with common commands

### Day 3 Afternoon (1:00 PM - 5:00 PM)

#### Task 3.3: Create Essential Configuration Files

**Duration**: 2 hours

**Assignee**: Backend Dev 1 + DevOps

**Files to Create**:

1. **`.gitignore`**

```gitignore
# Binaries
*.exe
*.exe~
*.dll
*.so
*.dylib
bin/
dist/

# Test binary
*.test

# Output of the go coverage tool
*.out

# Dependency directories
vendor/

# Go workspace file
go.work

# Environment variables
.env
.env.local
.env.*.local

# IDE
.vscode/
.idea/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db

# Logs
*.log
logs/

# Database
*.db
*.sqlite

# Temporary files
tmp/
temp/
```

1. **`.env.example`**

```bash
# Application
GO_ENV=development
API_PORT=8080
API_HOST=0.0.0.0

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=appointments
DB_PASSWORD=your-secure-password
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
JWT_SECRET=your-jwt-secret-key-min-32-chars-change-in-production
API_SECRET_SALT_ROUNDS=10

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:3000
CORS_ALLOWED_METHODS=GET,POST,PUT,PATCH,DELETE
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

1. **`Makefile`**

```makefile
.PHONY: help build run test clean migrate-up migrate-down docker-build docker-up

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
BINARY_NAME=$(APP_NAME)

# Colors for output
CYAN=\033[0;36m
NC=\033[0m # No Color

help: ## Show this help
 @echo "$(CYAN)Available targets:$(NC)"
 @grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(CYAN)%-20s$(NC) %s\n", $$1, $$2}'

install: ## Install dependencies
 @echo "$(CYAN)Installing dependencies...$(NC)"
 $(GOGET) github.com/gin-gonic/gin
 $(GOGET) github.com/jackc/pgx/v5
 $(GOGET) github.com/golang-jwt/jwt/v5
 $(GOGET) go.uber.org/zap
 $(GOGET) github.com/google/uuid
 $(GOGET) github.com/joho/godotenv
 $(GOGET) github.com/go-playground/validator/v10
 $(GOGET) golang.org/x/crypto/bcrypt
 $(GOMOD) tidy

build: ## Build the application
 @echo "$(CYAN)Building $(APP_NAME)...$(NC)"
 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

run: ## Run the application
 @echo "$(CYAN)Running $(APP_NAME)...$(NC)"
 $(GOCMD) run $(MAIN_PATH)

test: ## Run tests
 @echo "$(CYAN)Running tests...$(NC)"
 $(GOTEST) -v -race -coverprofile=coverage.out ./...

test-coverage: test ## Run tests with coverage report
 @echo "$(CYAN)Generating coverage report...$(NC)"
 $(GOCMD) tool cover -html=coverage.out -o coverage.html
 @echo "Coverage report: coverage.html"

lint: ## Run linter
 @echo "$(CYAN)Running linter...$(NC)"
 golangci-lint run ./...

fmt: ## Format code
 @echo "$(CYAN)Formatting code...$(NC)"
 $(GOCMD) fmt ./...
 gofmt -s -w .

clean: ## Clean build artifacts
 @echo "$(CYAN)Cleaning...$(NC)"
 rm -rf $(BUILD_DIR)
 rm -f coverage.out coverage.html

migrate-up: ## Run database migrations up
 @echo "$(CYAN)Running migrations up...$(NC)"
 migrate -path migrations -database "postgresql://appointments:password@localhost:5432/appointments_dev?sslmode=disable" up

migrate-down: ## Run database migrations down
 @echo "$(CYAN)Running migrations down...$(NC)"
 migrate -path migrations -database "postgresql://appointments:password@localhost:5432/appointments_dev?sslmode=disable" down

migrate-create: ## Create a new migration file (usage: make migrate-create name=create_table)
 @echo "$(CYAN)Creating migration: $(name)$(NC)"
 migrate create -ext sql -dir migrations -seq $(name)

docker-build: ## Build Docker image
 @echo "$(CYAN)Building Docker image...$(NC)"
 docker build -t $(APP_NAME):latest .

docker-up: ## Start services with Docker Compose
 @echo "$(CYAN)Starting services...$(NC)"
 docker-compose up -d

docker-down: ## Stop services with Docker Compose
 @echo "$(CYAN)Stopping services...$(NC)"
 docker-compose down

docker-logs: ## View Docker Compose logs
 docker-compose logs -f

dev: ## Run in development mode with hot reload (requires air)
 @echo "$(CYAN)Starting development server...$(NC)"
 air

install-tools: ## Install development tools
 @echo "$(CYAN)Installing development tools...$(NC)"
 go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
 go install github.com/cosmtrek/air@latest
 go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
 go install github.com/swaggo/swag/cmd/swag@latest

swagger: ## Generate Swagger documentation
 @echo "$(CYAN)Generating Swagger docs...$(NC)"
 swag init -g cmd/api/main.go -o docs/swagger

.DEFAULT_GOAL := help
```

1. **`docker-compose.yml`**

```yaml
version: '3.8'

services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      - GO_ENV=development
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=appointments
      - DB_PASSWORD=secure_password
      - DB_NAME=appointments_dev
      - REDIS_HOST=redis
      - REDIS_PORT=6379
    depends_on:
      - postgres
      - redis
    volumes:
      - .:/app
    networks:
      - appointment-network
    restart: unless-stopped

  postgres:
    image: postgres:15-alpine
    environment:
      - POSTGRES_USER=appointments
      - POSTGRES_PASSWORD=secure_password
      - POSTGRES_DB=appointments_dev
    ports:
      - "5432:5432"
    volumes:
      - postgres-data:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d
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

  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./deployments/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus-data:/prometheus
    networks:
      - appointment-network
    restart: unless-stopped

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
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

**Deliverables**:

- All configuration files created and committed
- Documentation updated with setup instructions

#### Task 3.4: Install Development Tools

**Duration**: 1 hour

**Assignee**: All Developers

**Activities**:

- [ ] Install Go 1.21+
- [ ] Install PostgreSQL 15+
- [ ] Install Redis 7+
- [ ] Install Docker Desktop
- [ ] Install make
- [ ] Install golang-migrate
- [ ] Install golangci-lint
- [ ] Install air (hot reload)
- [ ] Configure VS Code extensions

**Required VS Code Extensions**:

- Go (golang.go)
- Docker (ms-azuretools.vscode-docker)
- PostgreSQL (ckolkman.vscode-postgres)
- REST Client (humao.rest-client)
- GitLens (eamodio.gitlens)
- Error Lens (usernamehw.errorlens)

**Verification Commands**:

```bash
go version          # Should show 1.21+
postgres --version  # Should show 15+
redis-server --version  # Should show 7+
docker --version
make --version
migrate -version
golangci-lint --version
air -v
```

**Deliverables**:

- All tools installed and verified
- Screenshot of successful installations

---

## Day 4: Thursday - CI/CD Pipeline Setup

### Day 4 Morning (9:00 AM - 12:00 PM)

#### Task 4.1: GitHub Actions Workflow

**Duration**: 2 hours

**Assignee**: DevOps

**Activities**:

- [ ] Create `.github/workflows/ci.yml`
- [ ] Configure automated testing
- [ ] Set up linting checks
- [ ] Configure build verification
- [ ] Add test coverage reporting

**`.github/workflows/ci.yml`**:

```yaml
name: CI

on:
  push:
    branches: [ master, develop ]
  pull_request:
    branches: [ master, develop ]

jobs:
  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: golangci-lint
        uses: golangci/golangci-lint-action@v3
        with:
          version: latest

  test:
    name: Test
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_USER: appointments
          POSTGRES_PASSWORD: test_password
          POSTGRES_DB: appointments_test
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
      
      redis:
        image: redis:7
        ports:
          - 6379:6379
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Install dependencies
        run: make install
      
      - name: Run migrations
        run: make migrate-up
        env:
          DB_HOST: localhost
          DB_PORT: 5432
          DB_USER: appointments
          DB_PASSWORD: test_password
          DB_NAME: appointments_test
      
      - name: Run tests
        run: make test
        env:
          DB_HOST: localhost
          DB_PORT: 5432
          REDIS_HOST: localhost
          REDIS_PORT: 6379
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out

  build:
    name: Build
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Build
        run: make build
      
      - name: Upload artifact
        uses: actions/upload-artifact@v3
        with:
          name: appointment-service
          path: bin/appointment-service
```

**Deliverables**:

- GitHub Actions workflow configured
- Automated tests running on PR
- Build artifacts generated

#### Task 4.2: Docker Image Build Pipeline

**Duration**: 1 hour

**Assignee**: DevOps

**Activities**:

- [ ] Create production Dockerfile
- [ ] Create `.dockerignore`
- [ ] Set up Docker Hub repository
- [ ] Create image build workflow

**`Dockerfile`**:

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/main .
COPY --from=builder /app/migrations ./migrations

# Copy .env.example as template
COPY .env.example .env

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["./main"]
```

**`.dockerignore`**:

```md
# Git
.git
.gitignore

# Documentation
*.md
docs/

# IDE
.vscode/
.idea/

# Environment
.env
.env.*
!.env.example

# Build artifacts
bin/
dist/

# Tests
*_test.go
tests/
coverage.*

# Temporary files
tmp/
temp/
*.log
```

**Deliverables**:

- Production-ready Dockerfile
- Docker image build working locally
- `.dockerignore` configured

### Day 4 Afternoon (1:00 PM - 5:00 PM)

#### Task 4.3: Development Workflow Documentation

**Duration**: 2 hours

**Assignee**: Backend Dev 2

**Activities**:

- [ ] Create `docs/DEVELOPMENT_SETUP.md`
- [ ] Create `docs/TESTING_GUIDE.md`
- [ ] Create `docs/DEPLOYMENT_GUIDE.md`
- [ ] Create `docs/TROUBLESHOOTING.md`

Example: **DEVELOPMENT_SETUP.md**

```markdown
# Development Setup Guide

## Prerequisites

- Go 1.21 or higher
- PostgreSQL 15 or higher
- Redis 7 or higher
- Docker and Docker Compose
- Make

## Initial Setup

### 1. Clone Repository

\`\`\`bash
git clone https://github.com/laith-ambianze/appointment-service.git
cd appointment-service
\`\`\`

### 2. Install Dependencies

\`\`\`bash
make install
make install-tools
\`\`\`

### 3. Configure Environment

\`\`\`bash
cp .env.example .env
# Edit .env with your local settings
\`\`\`

### 4. Start Database

\`\`\`bash
docker-compose up -d postgres redis
\`\`\`

### 5. Run Migrations

\`\`\`bash
make migrate-up
\`\`\`

### 6. Start Application

\`\`\`bash
# Development mode with hot reload
make dev

# Or standard mode
make run
\`\`\`

## Verification

Visit http://localhost:8080/health - should return:
\`\`\`json
{"status": "ok"}
\`\`\`

## Common Commands

\`\`\`bash
make help          # Show all available commands
make test          # Run tests
make lint          # Run linter
make fmt           # Format code
make build         # Build binary
make clean         # Clean artifacts
\`\`\`

## Troubleshooting

See [TROUBLESHOOTING.md](TROUBLESHOOTING.md)
```

**Deliverables**:

- Complete development documentation
- Step-by-step setup guide
- Common issues and solutions

#### Task 4.4: Team Knowledge Transfer

**Duration**: 1 hour

**Participants**: All team members

**Activities**:

- [ ] Walk through project structure
- [ ] Demonstrate development workflow
- [ ] Show how to run tests
- [ ] Explain CI/CD pipeline
- [ ] Q&A session

**Deliverables**:

- All team members can run project locally
- All team members understand workflow
- Knowledge transfer recorded (video/notes)

---

## Day 5: Friday - Final Preparations & Week Review

### Day 5 Morning (9:00 AM - 12:00 PM)

#### Task 5.1: Complete Initial Module Setup

**Duration**: 2 hours

**Assignee**: Backend Dev 1

**Activities**:

- [ ] Create `cmd/api/main.go` skeleton
- [ ] Create `internal/config/config.go`
- [ ] Create `pkg/logger/logger.go`
- [ ] Add health check endpoint
- [ ] Verify everything compiles

**`cmd/api/main.go`**:

```go
package main

import (
 "context"
 "fmt"
 "log"
 "net/http"
 "os"
 "os/signal"
 "syscall"
 "time"

 "github.com/gin-gonic/gin"
 "github.com/laith-ambianze/appointment-service/internal/config"
 "github.com/laith-ambianze/appointment-service/pkg/logger"
)

func main() {
 // Load configuration
 cfg, err := config.Load()
 if err != nil {
  log.Fatalf("Failed to load config: %v", err)
 }

 // Initialize logger
 zapLogger, err := logger.New(cfg.LogLevel, cfg.LogFormat)
 if err != nil {
  log.Fatalf("Failed to initialize logger: %v", err)
 }
 defer zapLogger.Sync()

 // Create Gin router
 if cfg.Env == "production" {
  gin.SetMode(gin.ReleaseMode)
 }
 router := gin.Default()

 // Health check endpoint
 router.GET("/health", func(c *gin.Context) {
  c.JSON(http.StatusOK, gin.H{
   "status": "ok",
   "service": "appointment-service",
   "version": "1.0.0",
  })
 })

 // Create HTTP server
 srv := &http.Server{
  Addr:    fmt.Sprintf("%s:%s", cfg.APIHost, cfg.APIPort),
  Handler: router,
 }

 // Start server in goroutine
 go func() {
  zapLogger.Info(fmt.Sprintf("Starting server on %s", srv.Addr))
  if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
   zapLogger.Fatal(fmt.Sprintf("Failed to start server: %v", err))
  }
 }()

 // Graceful shutdown
 quit := make(chan os.Signal, 1)
 signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
 <-quit

 zapLogger.Info("Shutting down server...")

 ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
 defer cancel()

 if err := srv.Shutdown(ctx); err != nil {
  zapLogger.Fatal(fmt.Sprintf("Server forced to shutdown: %v", err))
 }

 zapLogger.Info("Server exited")
}
```

**Deliverables**:

- Application starts successfully
- Health endpoint accessible
- Logs output correctly
- Configuration loaded from .env

#### Task 5.2: Week 1 Retrospective

**Duration**: 1 hour

**Participants**: All team members

**Agenda**:

1. Review completed tasks (30 min)
2. Discuss challenges faced (15 min)
3. Identify improvements (15 min)

**Questions to Answer**:

- What went well?
- What could be improved?
- What blockers did we face?
- What have we learned?
- Are we ready for Week 2?

**Deliverables**:

- Retrospective notes
- Action items for next week
- Updated risk register

### Day 5 Afternoon (1:00 PM - 5:00 PM)

#### Task 5.3: Prepare for Week 2

**Duration**: 2 hours

**Assignee**: Tech Lead + All Developers

**Activities**:

- [ ] Review Week 2 tasks (database setup)
- [ ] Ensure all prerequisites are met
- [ ] Assign tasks for Monday
- [ ] Create detailed task breakdown
- [ ] Update project board

**Week 2 Preview**:

- Task 01: Project Setup ✅ DONE
- Task 02: Config and Logger (Start Monday)
- Task 03: Database Setup (Start Tuesday)
- Task 04: Models and Repository (Start Wednesday)

**Deliverables**:

- Week 2 tasks assigned
- Sprint board updated
- Team ready to start coding

#### Task 5.4: Documentation Review & Cleanup

**Duration**: 1 hour

**Assignee**: Tech Lead

**Activities**:

- [ ] Review all created documentation
- [ ] Fix any inconsistencies
- [ ] Ensure all links work
- [ ] Update README with current status
- [ ] Commit and push all changes

**Deliverables**:

- All documentation up to date
- Repository clean and organized
- README reflects current state

---

## Week 1 Deliverables Summary

### ✅ Completed Artifacts

1. **Documentation**
   - [ ] MIGRATION_STRATEGY_AND_ARCHITECTURE.md reviewed
   - [ ] API_SPECIFICATION.yaml created
   - [ ] DATABASE_SCHEMA.md created
   - [ ] 6 ADR documents created
   - [ ] DEVELOPMENT_SETUP.md created
   - [ ] CODE_REVIEW_GUIDELINES.md created
   - [ ] RISK_REGISTER.md created

2. **Repository Setup**
   - [ ] GitHub repository created
   - [ ] Branch protection configured
   - [ ] Project structure created
   - [ ] Go module initialized
   - [ ] .gitignore configured
   - [ ] README.md created

3. **Configuration**
   - [ ] .env.example created
   - [ ] Makefile created
   - [ ] docker-compose.yml created
   - [ ] Dockerfile created
   - [ ] .golangci.yml created

4. **CI/CD**
   - [ ] GitHub Actions workflow created
   - [ ] Automated testing configured
   - [ ] Docker build pipeline setup

5. **Development Environment**
   - [ ] All tools installed
   - [ ] Database running locally
   - [ ] Application skeleton created
   - [ ] Health endpoint working

### 📊 Success Metrics

- [ ] All team members can run project locally
- [ ] CI/CD pipeline passing
- [ ] Health endpoint returns 200 OK
- [ ] Code compiles without errors
- [ ] All documentation accessible

### 🎯 Week 1 Goals Status

- [x] Document review with team
- [x] Architecture decisions finalized
- [x] Development environment set up
- [x] Project repository created
- [x] Go module structure initialized

---

## Next Week Preview

### Week 2: Core Implementation

**Focus**: Database setup and basic APIs

**Key Tasks**:

- TASK_02: Config and Logger
- TASK_03: Database Setup and Migrations
- TASK_04: Models and Repository Layer
- TASK_05: Service Layer (start)

**Expected Deliverables**:

- Database schema deployed
- Configuration management working
- Repository pattern implemented
- Basic CRUD operations functional

---

## Appendix: Useful Resources

### Documentation

- [Go Documentation](https://go.dev/doc/)
- [Gin Framework](https://gin-gonic.com/docs/)
- [pgx Documentation](https://pkg.go.dev/github.com/jackc/pgx/v5)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)

### Tools

- [golang-migrate](https://github.com/golang-migrate/migrate)
- [golangci-lint](https://golangci-lint.run/)
- [Air (hot reload)](https://github.com/cosmtrek/air)
- [Swagger](https://swagger.io/tools/swagger-ui/)

### Learning Resources

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [PostgreSQL Tutorial](https://www.postgresqltutorial.com/)

---

**Document Status**: Ready for Implementation  
**Next Review**: End of Week 1  
**Maintained By**: Tech Lead
