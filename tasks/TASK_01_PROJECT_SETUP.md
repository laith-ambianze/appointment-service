# Task 01: Project Setup and Initialization

**Priority**: High  
**Estimated Time**: 2-3 hours  
**Dependencies**: None  
**Status**: Not Started

---

## Objective

Initialize the Go project structure, configure the development environment, and set up version control.

---

## Prerequisites

- [ ] Go 1.21+ installed
- [ ] Git installed
- [ ] Code editor (VS Code recommended)
- [ ] Docker and Docker Compose installed

---

## Steps

### 1. Initialize Git Repository

```bash
cd "c:\Users\SOKKER\Desktop\Appointment Project"
git init
git remote add origin https://github.com/laith-ambianze/appointment-service.git
```

### 2. Create Go Module

```bash
go mod init appointment-service
```

### 3. Create Project Structure

```bash
# Create directories
mkdir -p cmd\api
mkdir -p internal\handlers
mkdir -p internal\middleware
mkdir -p internal\models
mkdir -p internal\repository
mkdir -p internal\service
mkdir -p internal\config
mkdir -p internal\routes
mkdir -p pkg\auth
mkdir -p pkg\logger
mkdir -p pkg\validator
mkdir -p migrations
mkdir -p tests\unit
mkdir -p tests\integration
mkdir -p docs
```

### 4. Install Core Dependencies

```bash
# Web framework
go get github.com/gin-gonic/gin

# Database
go get github.com/jackc/pgx/v5
go get github.com/jackc/pgx/v5/pgxpool

# Configuration
go get github.com/joho/godotenv

# Logging
go get go.uber.org/zap

# JWT
go get github.com/golang-jwt/jwt/v5

# UUID
go get github.com/google/uuid

# Validation
go get github.com/go-playground/validator/v10

# CORS
go get github.com/gin-contrib/cors

# Testing
go get github.com/stretchr/testify
```

### 5. Create Essential Files

#### .gitignore

```gitignore
# Binaries
bin/
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test binary
*.test

# Output of the go coverage tool
*.out
coverage.html

# Dependency directories
vendor/

# Go workspace file
go.work

# Environment files
.env
.env.local

# IDE
.vscode/
.idea/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db

# Logs
*.log

# Database
*.db
postgres_data/
```

#### .env.example

```env
# Application
GO_ENV=development
API_PORT=8080
API_HOST=0.0.0.0

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=appointments
DB_PASSWORD=your-password-here
DB_NAME=appointments
DB_SSL_MODE=disable
DB_MAX_CONNECTIONS=25
DB_MAX_IDLE_CONNECTIONS=5
DB_CONNECTION_MAX_LIFETIME=5m

# Security
JWT_SECRET=your-jwt-secret-key-min-32-characters-long
API_SECRET_SALT_ROUNDS=10

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001
CORS_ALLOWED_METHODS=GET,POST,PUT,PATCH,DELETE
CORS_ALLOWED_HEADERS=Content-Type,Authorization,X-API-Key,X-API-Secret

# Rate Limiting
RATE_LIMIT_REQUESTS_PER_MINUTE=100
RATE_LIMIT_BURST=20

# Logging
LOG_LEVEL=debug
LOG_FORMAT=json
```

#### README.md

```markdown
# Appointment Service

Appointment as a Service - A microservice for managing appointments across multiple products.

## Quick Start

1. Copy `.env.example` to `.env` and update values
2. Run `docker-compose up -d`
3. Access API at `http://localhost:8080`

## Development

- `make build` - Build the application
- `make run` - Run the application
- `make test` - Run tests
- `make docker-up` - Start with Docker

## Documentation

See [FINAL_PROJECT_PLAN.md](FINAL_PROJECT_PLAN.md) for complete project documentation.
```

### 6. Copy .env.example to .env

```bash
cp .env.example .env
# Edit .env with your actual values
```

### 7. Initialize Git

```bash
git add .
git commit -m "Initial project setup"
git branch -M master
git push -u origin master
```

---

## Acceptance Criteria

- [ ] Go module initialized (`go.mod` exists)
- [ ] All directories created
- [ ] Core dependencies installed
- [ ] `.gitignore` created
- [ ] `.env.example` created
- [ ] `.env` created (not committed)
- [ ] README.md created
- [ ] Initial commit pushed to GitHub

---

## Verification

```bash
# Verify Go module
go mod verify

# Check structure
tree /F

# Test dependencies
go mod download
go mod tidy
```

---

## Next Task

[TASK_02_CONFIG_AND_LOGGER.md](TASK_02_CONFIG_AND_LOGGER.md)

---

## Notes

- Keep sensitive data out of version control
- Update `.env` with real values before running
- Review all installed packages periodically for security updates
