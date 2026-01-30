# Appointment-as-a-Service - Final Project Plan

## Project Overview

A microservice-based appointment management system designed to be integrated into multiple products, allowing users from different platforms to schedule and manage appointments with each other.

---

## 🚀 Quick Start Guide

### Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- PostgreSQL 15+ (if running locally)
- Make (optional, for convenience)

### Setup Steps

#### 1. Clone and Initialize

```bash
# Create project directory
mkdir appointment-service
cd appointment-service

# Initialize Go module
go mod init appointment-service

# Install dependencies
go get github.com/gin-gonic/gin
go get github.com/jackc/pgx/v5
go get github.com/golang-jwt/jwt/v5
go get go.uber.org/zap
go get github.com/google/uuid
go get github.com/joho/godotenv
```

#### 2. Configure Environment

```bash
# Copy environment template
cp .env.example .env

# Edit .env with your settings
nano .env
```

#### 3. Start with Docker

```bash
# Build and start all services
docker-compose up -d

# Check logs
docker-compose logs -f app

# Run migrations
docker-compose exec app ./main -migrate
```

#### 4. Test the API

```bash
# Health check
curl http://localhost:8080/health

# Register a product
curl -X POST http://localhost:8080/v1/products/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Product",
    "description": "My test application"
  }'
```

### Local Development (Without Docker)

```bash
# Install migration tool
make install-migrate

# Start PostgreSQL locally
# (Assumes PostgreSQL is installed and running)

# Run migrations
make migrate-up

# Build the application
make build

# Run the application
make run

# Or use hot reload (requires air)
make dev
```

---

## 🎯 Core Objectives

- **Multi-tenant Service**: Support multiple products/applications
- **Production-Ready**: Dockerized deployment
- **Secure**: Token-based authentication for product integration
- **Scalable**: Own database, isolated from client applications
- **Easy Integration**: RESTful API with comprehensive documentation

---

## 🏗️ System Architecture

### Architecture Type

**Microservice Architecture** - Standalone service with REST API

### Components

```md
┌─────────────────────────────────────────────────┐
│              Product 1 (Client)                  │
│         (Frontend + Backend)                     │
└──────────────────┬──────────────────────────────┘
                   │ API Calls (Token Auth)
                   ▼
┌─────────────────────────────────────────────────┐
│      Appointment Service (Docker Container)      │
│                                                   │
│  ┌──────────────┐      ┌──────────────┐        │
│  │ API Gateway   │──────│ Auth Layer   │        │
│  │   (Go/Gin)    │      │ (JWT Token)  │        │
│  └──────────────┘      └──────────────┘        │
│          │                                       │
│  ┌──────────────────────────────┐              │
│  │    Business Logic Layer       │              │
│  │  - Appointment Management     │              │
│  │  - Product Management         │              │
│  │  - User Metadata Management   │              │
│  └──────────────┬───────────────┘              │
│                 │                                 │
│  ┌──────────────▼───────────────┐              │
│  │    PostgreSQL Database        │              │
│  │  - Products                   │              │
│  │  - Appointments               │              │
│  │  - User Metadata              │              │
│  └──────────────────────────────┘              │
└─────────────────────────────────────────────────┘
```

---

## 💻 Technology Stack

### Backend

- **Language**: Go (Golang) 1.21+
- **Framework**: Gin / Fiber
- **Router**: Chi (alternative)
- **Validation**: go-playground/validator
- **Documentation**: Swagger/OpenAPI (swag)

### Database

- **Primary DB**: PostgreSQL 15+
- **Database Driver**: pgx (PostgreSQL driver)
- **ORM/Query Builder**: GORM / sqlc
- **Migrations**: golang-migrate / goose

### Security

- **Authentication**: JWT (golang-jwt/jwt)
- **Encryption**: bcrypt from crypto/bcrypt
- **Rate Limiting**: golang.org/x/time/rate
- **CORS**: gin-contrib/cors or rs/cors

### DevOps

- **Containerization**: Docker + Docker Compose
- **Environment**: godotenv / viper
- **Logging**: zap / logrus
- **Testing**: Go testing package + testify

---

## 🗄️ Database Schema

### Tables

#### 1. **products**

Stores registered products/applications that use the service.

```sql
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    api_key VARCHAR(255) UNIQUE NOT NULL,
    api_secret_hash VARCHAR(255) NOT NULL,
    callback_url VARCHAR(500),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### 2. **appointments**

Core table for storing appointment data.

```sql
CREATE TABLE appointments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id),
    
    -- User 1 (Initiator)
    user1_id VARCHAR(255) NOT NULL,
    user1_metadata JSONB NOT NULL,
    
    -- User 2 (Recipient)
    user2_id VARCHAR(255) NOT NULL,
    user2_metadata JSONB NOT NULL,
    
    -- Appointment Details
    title VARCHAR(500) NOT NULL,
    description TEXT,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    location VARCHAR(500),
    meeting_type VARCHAR(50), -- 'online', 'in-person', 'phone'
    
    -- Status Management
    status VARCHAR(50) DEFAULT 'pending', -- 'pending', 'confirmed', 'cancelled', 'completed'
    cancelled_by VARCHAR(255),
    cancellation_reason TEXT,
    cancelled_at TIMESTAMP,
    
    -- Additional Data
    additional_metadata JSONB,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT valid_time_range CHECK (end_time > start_time)
);

CREATE INDEX idx_appointments_product ON appointments(product_id);
CREATE INDEX idx_appointments_user1 ON appointments(user1_id);
CREATE INDEX idx_appointments_user2 ON appointments(user2_id);
CREATE INDEX idx_appointments_status ON appointments(status);
CREATE INDEX idx_appointments_time ON appointments(start_time, end_time);
```

#### 3. **user_metadata_schema**

User metadata structure (stored as JSONB):

```json
{
    "userId": "string (required)",
    "firstName": "string (required)",
    "lastName": "string (required)",
    "email": "string (optional)",
    "phone": "string (optional)",
    "timezone": "string (optional)",
    "customFields": {
        "key": "value"
    }
}
```

#### 4. **appointment_history** (Audit Trail)

Track all changes to appointments.

```sql
CREATE TABLE appointment_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    appointment_id UUID NOT NULL REFERENCES appointments(id),
    action VARCHAR(50) NOT NULL, -- 'created', 'updated', 'cancelled', 'confirmed'
    changed_by VARCHAR(255),
    changes JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

---

## 🔌 API Endpoints (MVP)

### Base URL

```md
https://api.appointment-service.com/v1
```

### Authentication

All endpoints (except product registration) require authentication:

```md
Headers:
  X-API-Key: {product_api_key}
  X-API-Secret: {product_api_secret}
```

---

### 1. Product Management

#### **POST /products/register**

Register a new product and receive API credentials.

**Request Body:**

```json
{
    "name": "Product Name",
    "description": "Product description",
    "callbackUrl": "https://product.com/webhook"
}
```

**Response:**

```json
{
    "success": true,
    "data": {
        "productId": "uuid",
        "apiKey": "prod_xxxxxxxxxxxxx",
        "apiSecret": "secret_xxxxxxxxxxxxx",
        "message": "Store these credentials securely. The secret will not be shown again."
    }
}
```

#### **GET /products/me**

Get current product information.

**Response:**

```json
{
    "success": true,
    "data": {
        "id": "uuid",
        "name": "Product Name",
        "isActive": true,
        "createdAt": "2026-01-30T10:00:00Z"
    }
}
```

---

### 2. Appointment Management

#### **POST /appointments**

Create a new appointment.

**Request Body:**

```json
{
    "user1": {
        "userId": "user_123",
        "firstName": "John",
        "lastName": "Doe",
        "email": "john@example.com",
        "phone": "+1234567890",
        "customFields": {}
    },
    "user2": {
        "userId": "user_456",
        "firstName": "Jane",
        "lastName": "Smith",
        "email": "jane@example.com",
        "phone": "+0987654321",
        "customFields": {}
    },
    "title": "Business Meeting",
    "description": "Discuss Q1 goals",
    "startTime": "2026-02-15T10:00:00Z",
    "endTime": "2026-02-15T11:00:00Z",
    "location": "Conference Room A",
    "meetingType": "in-person",
    "additionalMetadata": {}
}
```

**Response:**

```json
{
    "success": true,
    "data": {
        "appointmentId": "uuid",
        "status": "pending",
        "createdAt": "2026-01-30T10:00:00Z"
    }
}
```

#### **GET /appointments**

Get all appointments for the authenticated product.

**Query Parameters:**

- `page` (default: 1)
- `limit` (default: 20, max: 100)
- `status` (optional: pending, confirmed, cancelled, completed)
- `startDate` (optional: ISO 8601)
- `endDate` (optional: ISO 8601)

**Response:**

```json
{
    "success": true,
    "data": {
        "appointments": [...],
        "pagination": {
            "page": 1,
            "limit": 20,
            "total": 100,
            "totalPages": 5
        }
    }
}
```

#### **GET /appointments/:appointmentId**

Get a specific appointment by ID.

**Response:**

```json
{
    "success": true,
    "data": {
        "id": "uuid",
        "title": "Business Meeting",
        "user1": {...},
        "user2": {...},
        "startTime": "2026-02-15T10:00:00Z",
        "endTime": "2026-02-15T11:00:00Z",
        "status": "pending",
        "createdAt": "2026-01-30T10:00:00Z",
        "updatedAt": "2026-01-30T10:00:00Z"
    }
}
```

#### **GET /appointments/user/:userId**

Get all appointments for a specific user.

**Query Parameters:**

- Same as GET /appointments

**Response:** Same structure as GET /appointments

#### **PATCH /appointments/:appointmentId**

Update appointment details (except status).

**Request Body:**

```json
{
    "title": "Updated Meeting Title",
    "description": "Updated description",
    "startTime": "2026-02-15T11:00:00Z",
    "endTime": "2026-02-15T12:00:00Z",
    "location": "Updated location"
}
```

#### **PATCH /appointments/:appointmentId/status**

Update appointment status.

**Request Body:**

```json
{
    "status": "confirmed"
}
```

**Available Status Transitions:**

- pending → confirmed
- pending → cancelled
- confirmed → cancelled
- confirmed → completed

#### **PATCH /appointments/:appointmentId/cancel**

Cancel an appointment.

**Request Body:**

```json
{
    "cancelledBy": "user_123",
    "reason": "Schedule conflict"
}
```

**Response:**

```json
{
    "success": true,
    "data": {
        "appointmentId": "uuid",
        "status": "cancelled",
        "cancelledBy": "user_123",
        "cancelledAt": "2026-01-30T10:00:00Z"
    }
}
```

#### **DELETE /appointments/:appointmentId**

Permanently delete an appointment (admin only).

---

## 🔐 Security Implementation

### 1. API Authentication Flow

```md
1. Product registers → Receives API Key + Secret
2. Product stores credentials securely
3. For each request:
   - Include X-API-Key and X-API-Secret headers
   - Server validates credentials
   - Server generates short-lived JWT for session
   - Request processed if valid
```

### 2. Security Measures

- API secrets hashed with bcrypt (never stored in plain text)
- Rate limiting: 100 requests/minute per product
- CORS whitelist configuration
- Input validation on all endpoints
- SQL injection prevention (ORM parameterized queries)
- XSS protection
- Request size limits
- HTTPS only in production

### 3. Environment Variables

```env
# Application
GO_ENV=production
API_PORT=8080
API_HOST=0.0.0.0

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=appointments
DB_PASSWORD=your-secure-password
DB_NAME=appointments
DB_SSL_MODE=require
DB_MAX_CONNECTIONS=25
DB_MAX_IDLE_CONNECTIONS=5
DB_CONNECTION_MAX_LIFETIME=5m

# Security
JWT_SECRET=your-jwt-secret-key-min-32-chars
API_SECRET_SALT_ROUNDS=10

# CORS
CORS_ALLOWED_ORIGINS=https://product1.com,https://product2.com
CORS_ALLOWED_METHODS=GET,POST,PUT,PATCH,DELETE
CORS_ALLOWED_HEADERS=Content-Type,Authorization,X-API-Key,X-API-Secret

# Rate Limiting
RATE_LIMIT_REQUESTS_PER_MINUTE=100
RATE_LIMIT_BURST=20

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
```

---

## 🐳 Docker Setup

### Project Structure

```md
appointment-service/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── handlers/
│   │   ├── appointment.go
│   │   └── product.go
│   ├── middleware/
│   │   ├── auth.go
│   │   └── ratelimit.go
│   ├── models/
│   │   ├── appointment.go
│   │   └── product.go
│   ├── repository/
│   │   ├── appointment_repo.go
│   │   └── product_repo.go
│   ├── service/
│   │   ├── appointment_service.go
│   │   └── product_service.go
│   ├── config/
│   │   └── config.go
│   └── routes/
│       └── routes.go
├── pkg/
│   ├── auth/
│   ├── logger/
│   └── validator/
├── migrations/
│   ├── 001_create_products.up.sql
│   ├── 001_create_products.down.sql
│   ├── 002_create_appointments.up.sql
│   └── 002_create_appointments.down.sql
├── tests/
│   ├── integration/
│   └── unit/
├── docs/
│   ├── swagger.yaml
│   └── API_INTEGRATION_GUIDE.md
├── .env.example
├── .dockerignore
├── .gitignore
├── docker-compose.yml
├── Dockerfile
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

### Dockerfile

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

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

# Copy .env if needed (or use environment variables)
COPY .env.example .env

EXPOSE 8080

CMD ["./main"]
```

### docker-compose.yml

```yaml
version: '3.8'

services:
  app:
    build: .
    container_name: appointment-service
    ports:
      - "8080:8080"
    environment:
      GO_ENV: production
      DB_HOST: db
      DB_PORT: 5432
      DB_USER: appointments
      DB_PASSWORD: password
      DB_NAME: appointments
      DB_SSL_MODE: disable
      JWT_SECRET: ${JWT_SECRET}
      API_PORT: 8080
    depends_on:
      db:
        condition: service_healthy
    restart: unless-stopped
    networks:
      - appointment-network
    command: ["/bin/sh", "-c", "sleep 5 && ./main"]

  db:
    image: postgres:15-alpine
    container_name: appointment-db
    environment:
      POSTGRES_USER: appointments
      POSTGRES_PASSWORD: password
      POSTGRES_DB: appointments
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U appointments"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - appointment-network

volumes:
  postgres_data:

networks:
  appointment-network:
    driver: bridge
```

### Deployment Commands

```bash
# Build and start services
docker-compose up -d

# View logs
docker-compose logs -f app

# Stop services
docker-compose down

# Rebuild after changes
docker-compose up -d --build

# Run migrations manually
docker-compose exec app ./main -migrate

# Build locally (without Docker)
make build

# Run locally
make run

# Run tests
make test
```

### Makefile

```makefile
.PHONY: build run test clean migrate-up migrate-down docker-build docker-up docker-down

# Application name
APP_NAME=appointment-service
BUILD_DIR=bin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=$(APP_NAME)

# Database
DB_URL=postgresql://appointments:password@localhost:5432/appointments?sslmode=disable

all: test build

## Build the application
build:
 @echo "Building..."
 @mkdir -p $(BUILD_DIR)
 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) -v ./cmd/api

## Run the application
run: build
 @echo "Running..."
 @./$(BUILD_DIR)/$(BINARY_NAME)

## Run tests
test:
 @echo "Running tests..."
 $(GOTEST) -v -cover ./...

## Run tests with coverage
test-coverage:
 @echo "Running tests with coverage..."
 $(GOTEST) -v -coverprofile=coverage.out ./...
 $(GOCMD) tool cover -html=coverage.out -o coverage.html
 @echo "Coverage report generated: coverage.html"

## Run linter
lint:
 @echo "Running linter..."
 golangci-lint run

## Format code
fmt:
 @echo "Formatting code..."
 $(GOCMD) fmt ./...

## Tidy dependencies
tidy:
 @echo "Tidying dependencies..."
 $(GOMOD) tidy

## Download dependencies
deps:
 @echo "Downloading dependencies..."
 $(GOMOD) download

## Clean build artifacts
clean:
 @echo "Cleaning..."
 @rm -rf $(BUILD_DIR)
 @rm -f coverage.out coverage.html

## Run database migrations up
migrate-up:
 @echo "Running migrations up..."
 migrate -path migrations -database "$(DB_URL)" up

## Run database migrations down
migrate-down:
 @echo "Running migrations down..."
 migrate -path migrations -database "$(DB_URL)" down

## Create a new migration
migrate-create:
 @read -p "Enter migration name: " name; \
 migrate create -ext sql -dir migrations -seq $$name

## Install migration tool
install-migrate:
 @echo "Installing golang-migrate..."
 go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

## Docker build
docker-build:
 @echo "Building Docker image..."
 docker build -t $(APP_NAME):latest .

## Docker compose up
docker-up:
 @echo "Starting Docker containers..."
 docker-compose up -d

## Docker compose down
docker-down:
 @echo "Stopping Docker containers..."
 docker-compose down

## Docker compose logs
docker-logs:
 docker-compose logs -f app

## Development mode with hot reload (requires air)
dev:
 @echo "Running in development mode..."
 air

## Install development tools
install-tools:
 @echo "Installing development tools..."
 go install github.com/cosmtrek/air@latest
 go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
 go install github.com/swaggo/swag/cmd/swag@latest

## Generate Swagger documentation
swagger:
 @echo "Generating Swagger docs..."
 swag init -g cmd/api/main.go -o docs

## Help
help:
 @echo "Available targets:"
 @echo "  make build           - Build the application"
 @echo "  make run             - Run the application"
 @echo "  make test            - Run tests"
 @echo "  make test-coverage   - Run tests with coverage"
 @echo "  make lint            - Run linter"
 @echo "  make fmt             - Format code"
 @echo "  make tidy            - Tidy dependencies"
 @echo "  make clean           - Clean build artifacts"
 @echo "  make migrate-up      - Run database migrations up"
 @echo "  make migrate-down    - Run database migrations down"
 @echo "  make migrate-create  - Create a new migration"
 @echo "  make docker-build    - Build Docker image"
 @echo "  make docker-up       - Start Docker containers"
 @echo "  make docker-down     - Stop Docker containers"
 @echo "  make dev             - Run in development mode with hot reload"
 @echo "  make install-tools   - Install development tools"
 @echo "  make swagger         - Generate Swagger documentation"
```

---

## � Sample Go Implementation

### Main Application Entry (cmd/api/main.go)

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

    "appointment-service/internal/config"
    "appointment-service/internal/handlers"
    "appointment-service/internal/middleware"
    "appointment-service/internal/repository"
    "appointment-service/internal/routes"
    "appointment-service/internal/service"
    "appointment-service/pkg/logger"

    "github.com/gin-gonic/gin"
    "github.com/jackc/pgx/v5/pgxpool"
    "go.uber.org/zap"
)

func main() {
    // Load configuration
    cfg, err := config.Load()
    if err != nil {
        log.Fatal(\"Failed to load config:\", err)
    }

    // Initialize logger
    zapLogger, err := logger.NewLogger(cfg.Environment)
    if err != nil {
        log.Fatal(\"Failed to initialize logger:\", err)
    }
    defer zapLogger.Sync()

    // Connect to database
    dbPool, err := connectDB(cfg.DatabaseURL)
    if err != nil {
        zapLogger.Fatal(\"Failed to connect to database\", zap.Error(err))
    }
    defer dbPool.Close()

    zapLogger.Info(\"Database connection established\")

    // Initialize repositories
    productRepo := repository.NewProductRepository(dbPool)
    appointmentRepo := repository.NewAppointmentRepository(dbPool)

    // Initialize services
    productSvc := service.NewProductService(productRepo)
    appointmentSvc := service.NewAppointmentService(appointmentRepo, productRepo)

    // Initialize handlers
    productHandler := handlers.NewProductHandler(productSvc)
    appointmentHandler := handlers.NewAppointmentHandler(appointmentSvc)

    // Setup router
    router := setupRouter(cfg, zapLogger, productHandler, appointmentHandler)

    // Start server
    srv := &http.Server{
        Addr:         fmt.Sprintf(\"%s:%s\", cfg.Host, cfg.Port),
        Handler:      router,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    // Graceful shutdown
    go func() {
        zapLogger.Info(\"Starting server\", zap.String(\"address\", srv.Addr))
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            zapLogger.Fatal(\"Server failed to start\", zap.Error(err))
        }
    }()

    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    zapLogger.Info(\"Shutting down server...\")

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        zapLogger.Fatal(\"Server forced to shutdown\", zap.Error(err))
    }

    zapLogger.Info(\"Server exited\")
}

func connectDB(url string) (*pgxpool.Pool, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    pool, err := pgxpool.New(ctx, url)
    if err != nil {
        return nil, err
    }

    if err := pool.Ping(ctx); err != nil {
        return nil, err
    }

    return pool, nil
}

func setupRouter(cfg *config.Config, logger *zap.Logger, productHandler *handlers.ProductHandler, appointmentHandler *handlers.AppointmentHandler) *gin.Engine {
    if cfg.Environment == \"production\" {
        gin.SetMode(gin.ReleaseMode)
    }

    router := gin.New()
    router.Use(gin.Recovery())
    router.Use(middleware.Logger(logger))
    router.Use(middleware.CORS(cfg.CORSOrigins))

    // Health check
    router.GET(\"/health\", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{\"status\": \"healthy\"})
    })

    // API v1 routes
    v1 := router.Group(\"/v1\")
    {
        routes.RegisterProductRoutes(v1, productHandler)
        routes.RegisterAppointmentRoutes(v1, appointmentHandler, middleware.AuthMiddleware(cfg))
    }

    return router
}
```

### Authentication Middleware (internal/middleware/auth.go)

```go
package middleware

import (
    \"crypto/subtle\"
    \"net/http\"

    \"appointment-service/internal/config\"
    \"github.com/gin-gonic/gin\"
)

func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
    return func(c *gin.Context) {
        apiKey := c.GetHeader(\"X-API-Key\")
        apiSecret := c.GetHeader(\"X-API-Secret\")

        if apiKey == \"\" || apiSecret == \"\" {
            c.JSON(http.StatusUnauthorized, gin.H{
                \"error\": \"Missing authentication credentials\",
            })
            c.Abort()
            return
        }

        // Validate credentials (simplified - should check database)
        // In production, look up product by API key and verify secret hash
        if !isValidCredentials(apiKey, apiSecret) {
            c.JSON(http.StatusUnauthorized, gin.H{
                \"error\": \"Invalid credentials\",
            })
            c.Abort()
            return
        }

        // Store product ID in context for use in handlers
        c.Set(\"product_id\", extractProductID(apiKey))
        c.Next()
    }
}

func isValidCredentials(apiKey, apiSecret string) bool {
    // This should query the database and verify bcrypt hash
    // Simplified for example
    return subtle.ConstantTimeCompare([]byte(apiKey), []byte(\"test-key\")) == 1
}

func extractProductID(apiKey string) string {
    // Extract product ID from API key
    return \"product-id\"
}
```

### Appointment Handler (internal/handlers/appointment.go)

```go
package handlers

import (
    \"net/http\"
    \"time\"

    \"appointment-service/internal/models\"
    \"appointment-service/internal/service\"
    \"github.com/gin-gonic/gin\"
    \"github.com/google/uuid\"
)

type AppointmentHandler struct {
    service *service.AppointmentService
}

func NewAppointmentHandler(service *service.AppointmentService) *AppointmentHandler {
    return &AppointmentHandler{service: service}
}

type CreateAppointmentRequest struct {
    User1           UserMetadata `json:\"user1\" binding:\"required\"`
    User2           UserMetadata `json:\"user2\" binding:\"required\"`
    Title           string       `json:\"title\" binding:\"required,min=3,max=500\"`
    Description     string       `json:\"description\"`
    StartTime       time.Time    `json:\"startTime\" binding:\"required\"`
    EndTime         time.Time    `json:\"endTime\" binding:\"required\"`
    Location        string       `json:\"location\"`
    MeetingType     string       `json:\"meetingType\" binding:\"omitempty,oneof=online in-person phone\"`
    AdditionalData  interface{}  `json:\"additionalMetadata\"`
}

type UserMetadata struct {
    UserID       string            `json:\"userId\" binding:\"required\"`
    FirstName    string            `json:\"firstName\" binding:\"required\"`
    LastName     string            `json:\"lastName\" binding:\"required\"`
    Email        string            `json:\"email\" binding:\"omitempty,email\"`
    Phone        string            `json:\"phone\"`
    CustomFields map[string]string `json:\"customFields\"`
}

func (h *AppointmentHandler) Create(c *gin.Context) {
    var req CreateAppointmentRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            \"success\": false,
            \"error\":   err.Error(),
        })
        return
    }

    // Validate time range
    if !req.EndTime.After(req.StartTime) {
        c.JSON(http.StatusBadRequest, gin.H{
            \"success\": false,
            \"error\":   \"End time must be after start time\",
        })
        return
    }

    productID := c.GetString(\"product_id\")

    appointment := &models.Appointment{
        ID:              uuid.New(),
        ProductID:       uuid.MustParse(productID),
        User1ID:         req.User1.UserID,
        User1Metadata:   req.User1,
        User2ID:         req.User2.UserID,
        User2Metadata:   req.User2,
        Title:           req.Title,
        Description:     req.Description,
        StartTime:       req.StartTime,
        EndTime:         req.EndTime,
        Location:        req.Location,
        MeetingType:     req.MeetingType,
        Status:          \"pending\",
        AdditionalData:  req.AdditionalData,
    }

    if err := h.service.CreateAppointment(c.Request.Context(), appointment); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            \"success\": false,
            \"error\":   \"Failed to create appointment\",
        })
        return
    }

    c.JSON(http.StatusCreated, gin.H{
        \"success\": true,
        \"data\": gin.H{
            \"appointmentId\": appointment.ID,
            \"status\":        appointment.Status,
            \"createdAt\":     appointment.CreatedAt,
        },
    })
}

func (h *AppointmentHandler) GetAll(c *gin.Context) {
    productID := c.GetString(\"product_id\")
    
    appointments, err := h.service.GetAppointmentsByProduct(c.Request.Context(), productID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            \"success\": false,
            \"error\":   \"Failed to fetch appointments\",
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        \"success\": true,
        \"data\": gin.H{
            \"appointments\": appointments,
        },
    })
}
```

---

## �📚 Integration Documentation

### For Backend Developers (Client Products)

#### Step 1: Register Your Product

```bash
curl -X POST https://api.appointment-service.com/v1/products/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Product",
    "description": "Product description",
    "callbackUrl": "https://myproduct.com/webhook"
  }'
```

**Save the `apiKey` and `apiSecret` securely!**

#### Step 2: Store Credentials

```bash
# Backend environment variables (.env)
APPOINTMENT_SERVICE_URL=https://api.appointment-service.com/v1
APPOINTMENT_API_KEY=prod_xxxxxxxxxxxxx
APPOINTMENT_API_SECRET=secret_xxxxxxxxxxxxx
```

#### Step 3: Create Appointment Helper

**Go Example:**

```go
package appointment

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
    "time"
)

type AppointmentClient struct {
    BaseURL   string
    APIKey    string
    APISecret string
    HTTPClient *http.Client
}

func NewClient() *AppointmentClient {
    return &AppointmentClient{
        BaseURL:   os.Getenv("APPOINTMENT_SERVICE_URL"),
        APIKey:    os.Getenv("APPOINTMENT_API_KEY"),
        APISecret: os.Getenv("APPOINTMENT_API_SECRET"),
        HTTPClient: &http.Client{
            Timeout: 10 * time.Second,
        },
    }
}

type CreateAppointmentRequest struct {
    User1 UserMetadata `json:"user1"`
    User2 UserMetadata `json:"user2"`
    Title string `json:"title"`
    Description string `json:"description"`
    StartTime time.Time `json:"startTime"`
    EndTime time.Time `json:"endTime"`
    Location string `json:"location"`
    MeetingType string `json:"meetingType"`
}

type UserMetadata struct {
    UserID string `json:"userId"`
    FirstName string `json:"firstName"`
    LastName string `json:"lastName"`
    Email string `json:"email"`
    Phone string `json:"phone"`
}

type AppointmentResponse struct {
    Success bool `json:"success"`
    Data map[string]interface{} `json:"data"`
}

func (c *AppointmentClient) CreateAppointment(req CreateAppointmentRequest) (*AppointmentResponse, error) {
    jsonData, err := json.Marshal(req)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal request: %w", err)
    }

    httpReq, err := http.NewRequest("POST", c.BaseURL+"/appointments", bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }

    httpReq.Header.Set("Content-Type", "application/json")
    httpReq.Header.Set("X-API-Key", c.APIKey)
    httpReq.Header.Set("X-API-Secret", c.APISecret)

    resp, err := c.HTTPClient.Do(httpReq)
    if err != nil {
        return nil, fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response: %w", err)
    }

    if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
        return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
    }

    var result AppointmentResponse
    if err := json.Unmarshal(body, &result); err != nil {
        return nil, fmt.Errorf("failed to unmarshal response: %w", err)
    }

    return &result, nil
}

func (c *AppointmentClient) GetUserAppointments(userID string) (*AppointmentResponse, error) {
    url := fmt.Sprintf("%s/appointments/user/%s", c.BaseURL, userID)
    httpReq, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }

    httpReq.Header.Set("X-API-Key", c.APIKey)
    httpReq.Header.Set("X-API-Secret", c.APISecret)

    resp, err := c.HTTPClient.Do(httpReq)
    if err != nil {
        return nil, fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response: %w", err)
    }

    var result AppointmentResponse
    if err := json.Unmarshal(body, &result); err != nil {
        return nil, fmt.Errorf("failed to unmarshal response: %w", err)
    }

    return &result, nil
}

func (c *AppointmentClient) CancelAppointment(appointmentID, userID, reason string) (*AppointmentResponse, error) {
    url := fmt.Sprintf("%s/appointments/%s/cancel", c.BaseURL, appointmentID)
    cancelReq := map[string]string{
        "cancelledBy": userID,
        "reason": reason,
    }
    jsonData, _ := json.Marshal(cancelReq)

    httpReq, err := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }

    httpReq.Header.Set("Content-Type", "application/json")
    httpReq.Header.Set("X-API-Key", c.APIKey)
    httpReq.Header.Set("X-API-Secret", c.APISecret)

    resp, err := c.HTTPClient.Do(httpReq)
    if err != nil {
        return nil, fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response: %w", err)
    }

    var result AppointmentResponse
    if err := json.Unmarshal(body, &result); err != nil {
        return nil, fmt.Errorf("failed to unmarshal response: %w", err)
    }

    return &result, nil
}
```

```javascript
const axios = require('axios');

class AppointmentService {
    constructor() {
        this.baseURL = process.env.APPOINTMENT_SERVICE_URL;
        this.apiKey = process.env.APPOINTMENT_API_KEY;
        this.apiSecret = process.env.APPOINTMENT_API_SECRET;
    }

    async createAppointment(data) {
        try {
            const response = await axios.post(
                `${this.baseURL}/appointments`,
                data,
                {
                    headers: {
                        'Content-Type': 'application/json',
                        'X-API-Key': this.apiKey,
                        'X-API-Secret': this.apiSecret
                    }
                }
            );
            return response.data;
        } catch (error) {
            console.error('Error creating appointment:', error.response?.data);
            throw error;
        }
    }

    async getUserAppointments(userId) {
        try {
            const response = await axios.get(
                `${this.baseURL}/appointments/user/${userId}`,
                {
                    headers: {
                        'X-API-Key': this.apiKey,
                        'X-API-Secret': this.apiSecret
                    }
                }
            );
            return response.data;
        } catch (error) {
            console.error('Error fetching appointments:', error.response?.data);
            throw error;
        }
    }

    async cancelAppointment(appointmentId, userId, reason) {
        try {
            const response = await axios.patch(
                `${this.baseURL}/appointments/${appointmentId}/cancel`,
                {
                    cancelledBy: userId,
                    reason: reason
                },
                {
                    headers: {
                        'Content-Type': 'application/json',
                        'X-API-Key': this.apiKey,
                        'X-API-Secret': this.apiSecret
                    }
                }
            );
            return response.data;
        } catch (error) {
            console.error('Error cancelling appointment:', error.response?.data);
            throw error;
        }
    }
}

module.exports = new AppointmentService();
```

#### Step 4: Usage in Your Backend

**Go Example:**

```go
package handlers

import (
    "encoding/json"
    "net/http"
    "time"
    "yourapp/appointment"
    "yourapp/models"
)

type AppointmentHandler struct {
    appointmentClient *appointment.AppointmentClient
    userRepo *models.UserRepository
}

func NewAppointmentHandler(client *appointment.AppointmentClient, userRepo *models.UserRepository) *AppointmentHandler {
    return &AppointmentHandler{
        appointmentClient: client,
        userRepo: userRepo,
    }
}

type ScheduleMeetingRequest struct {
    User1ID   string    `json:"user1Id"`
    User2ID   string    `json:"user2Id"`
    Title     string    `json:"title"`
    StartTime time.Time `json:"startTime"`
    EndTime   time.Time `json:"endTime"`
}

func (h *AppointmentHandler) ScheduleMeeting(w http.ResponseWriter, r *http.Request) {
    var req ScheduleMeetingRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Fetch user details from your database
    user1, err := h.userRepo.GetByID(req.User1ID)
    if err != nil {
        http.Error(w, "User 1 not found", http.StatusNotFound)
        return
    }

    user2, err := h.userRepo.GetByID(req.User2ID)
    if err != nil {
        http.Error(w, "User 2 not found", http.StatusNotFound)
        return
    }

    // Create appointment
    appointmentReq := appointment.CreateAppointmentRequest{
        User1: appointment.UserMetadata{
            UserID:    user1.ID,
            FirstName: user1.FirstName,
            LastName:  user1.LastName,
            Email:     user1.Email,
            Phone:     user1.Phone,
        },
        User2: appointment.UserMetadata{
            UserID:    user2.ID,
            FirstName: user2.FirstName,
            LastName:  user2.LastName,
            Email:     user2.Email,
            Phone:     user2.Phone,
        },
        Title:       req.Title,
        StartTime:   req.StartTime,
        EndTime:     req.EndTime,
        MeetingType: "online",
    }

    result, err := h.appointmentClient.CreateAppointment(appointmentReq)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "appointment": result,
    })
}
```

**Node.js Example:**

```javascript
// In your route handler
const appointmentService = require('./services/appointmentService');

app.post('/api/schedule-meeting', async (req, res) => {
    try {
        const { user1Id, user2Id, title, startTime, endTime } = req.body;
        
        // Fetch user details from your database
        const user1 = await getUserById(user1Id);
        const user2 = await getUserById(user2Id);
        
        // Create appointment
        const appointment = await appointmentService.createAppointment({
            user1: {
                userId: user1.id,
                firstName: user1.firstName,
                lastName: user1.lastName,
                email: user1.email,
                phone: user1.phone
            },
            user2: {
                userId: user2.id,
                firstName: user2.firstName,
                lastName: user2.lastName,
                email: user2.email,
                phone: user2.phone
            },
            title: title,
            startTime: startTime,
            endTime: endTime,
            meetingType: 'online'
        });
        
        res.json({ success: true, appointment });
    } catch (error) {
        res.status(500).json({ error: error.message });
    }
});
```

---

## 📋 Development Roadmap

### Phase 1: MVP (Weeks 1-4)

- [ ] Project setup & configuration
- [ ] Database schema & migrations
- [ ] Product registration API
- [ ] Appointment CRUD APIs
- [ ] Authentication middleware
- [ ] Basic validation & error handling
- [ ] Docker setup
- [ ] API documentation
- [ ] Integration guide

### Phase 2: Testing & Refinement (Weeks 5-6)

- [ ] Unit tests (>80% coverage)
- [ ] Integration tests
- [ ] Load testing
- [ ] Security audit
- [ ] Bug fixes
- [ ] Performance optimization

### Phase 3: Post-MVP Features (Weeks 7-12)

- [ ] **User Availability Settings**
  - Set working hours per day
  - Block specific time slots
  - Recurring availability patterns
  - Timezone handling
- [ ] **Conflict Detection**
  - Check for overlapping appointments
  - Suggest alternative times
- [ ] **Notifications** (optional)
  - Webhook callbacks to products
  - Appointment reminders
- [ ] **Recurring Appointments**
  - Daily, weekly, monthly patterns
- [ ] **Advanced Search & Filters**
  - By date range, status, users
  - Full-text search
- [ ] **Analytics Dashboard**
  - Appointment statistics per product
  - Usage metrics

---

## 🧪 Testing Strategy

### Unit Tests

**Go Testing Example:**

```go
package service

import (
    "testing"
    "time"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

func TestCreateAppointment(t *testing.T) {
    // Arrange
    mockRepo := new(MockAppointmentRepository)
    service := NewAppointmentService(mockRepo)
    
    appointment := &Appointment{
        ProductID: "prod-123",
        User1ID: "user-1",
        User2ID: "user-2",
        Title: "Test Meeting",
        StartTime: time.Now(),
        EndTime: time.Now().Add(time.Hour),
    }
    
    mockRepo.On("Create", mock.Anything).Return(nil)
    
    // Act
    err := service.CreateAppointment(appointment)
    
    // Assert
    assert.NoError(t, err)
    mockRepo.AssertExpectations(t)
}
```

**Run tests:**

```bash
go test ./... -v
go test ./... -cover
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

- Service layer logic
- Validation functions
- Utility functions
- Target: >80% coverage

### Integration Tests

**Go Integration Test Example:**

```go
package integration

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestCreateAppointmentEndpoint(t *testing.T) {
    // Setup test database and router
    router := setupTestRouter()
    
    payload := map[string]interface{}{
        "user1": map[string]string{
            "userId": "user-1",
            "firstName": "John",
            "lastName": "Doe",
        },
        "title": "Test Meeting",
    }
    
    jsonData, _ := json.Marshal(payload)
    req := httptest.NewRequest("POST", "/v1/appointments", bytes.NewBuffer(jsonData))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-API-Key", "test-key")
    req.Header.Set("X-API-Secret", "test-secret")
    
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    assert.Equal(t, http.StatusCreated, w.Code)
}
```

- API endpoint testing
- Database operations
- Authentication flow
- Error handling

### Load Tests

**Using vegeta (Go load testing tool):**

```bash
# Install vegeta
go install github.com/tsenart/vegeta@latest

# Run load test
echo "GET http://localhost:8080/v1/appointments" | vegeta attack -duration=30s -rate=100 | vegeta report
```

- Concurrent requests handling
- Database connection pooling
- Rate limiting effectiveness

---

## 📊 Monitoring & Logging

### Logging Levels

**Go Zap Logger Configuration:**

```go
package logger

import (
    "go.uber.org/zap"
    "go.uber.org/zap"
)

func NewLogger(env string) (*zap.Logger, error) {
    if env == "production" {
        return zap.NewProduction()
    }
    return zap.NewDevelopment()
}

// Usage:
// logger.Info("API request", zap.String("method", "POST"), zap.String("path", "/appointments"))
// logger.Error("Database error", zap.Error(err))
// logger.Warn("Rate limit approaching", zap.Int("requests", 95))
// logger.Debug("Request details", zap.Any("payload", data))
```

**Log Levels:**

- `error`: Errors that need immediate attention
- `warn`: Warning messages
- `info`: General information (API calls, etc.)
- `debug`: Detailed debugging information

### Metrics to Track

**Prometheus Metrics Example:**

```go
var (
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )
    
    httpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "HTTP request duration in seconds",
        },
        []string{"method", "endpoint"},
    )
)
```

**Key Metrics:**

- Request count per endpoint
- Response times (p50, p95, p99)
- Error rates
- Active products
- Total appointments
- Database query performance
- Goroutine count
- Memory usage

### Recommended Tools

- **Logging**: Zap / Logrus + Elasticsearch + Kibana
- **Monitoring**: Prometheus + Grafana
- **Error Tracking**: Sentry
- **APM**: New Relic / Datadog
- **Tracing**: Jaeger / OpenTelemetry

---

## 🚀 Deployment Checklist

### Pre-Production

- [ ] Environment variables configured
- [ ] Database backups enabled
- [ ] SSL certificates installed
- [ ] CORS whitelist configured
- [ ] Rate limits tested
- [ ] API documentation published
- [ ] Load testing completed
- [ ] Security audit passed

### Production

- [ ] Docker containers running
- [ ] Database migrations applied
- [ ] Monitoring tools active
- [ ] Logging configured
- [ ] Backup strategy implemented
- [ ] Rollback plan documented
- [ ] Integration guide published

---

## 📖 Documentation Deliverables

1. **API_INTEGRATION_GUIDE.md**
   - Quick start guide
   - Authentication setup
   - Code examples in multiple languages
   - Error handling
   - Best practices

2. **API_REFERENCE.md**
   - Complete endpoint documentation
   - Request/response schemas
   - Error codes
   - Rate limits

3. **DEPLOYMENT_GUIDE.md**
   - Docker setup
   - Environment configuration
   - Scaling strategies
   - Backup procedures

4. **DEVELOPER_GUIDE.md**
   - Project structure
   - Development workflow
   - Testing guidelines
   - Contribution guidelines

---

## 🔮 Future Enhancements (Post-MVP)

### Availability Management System

**Go Struct Example:**

```go
type UserAvailability struct {
    UserID        string              `json:"userId"`
    ProductID     string              `json:"productId"`
    Timezone      string              `json:"timezone"`
    WorkingHours  map[string][]TimeSlot `json:"workingHours"`
    BlockedSlots  []BlockedSlot       `json:"blockedSlots"`
    MinimumNotice int                 `json:"minimumNotice"` // hours
    BufferTime    int                 `json:"bufferTime"`    // minutes
}

type TimeSlot struct {
    Start string `json:"start"` // "09:00"
    End   string `json:"end"`   // "17:00"
}

type BlockedSlot struct {
    Start  time.Time `json:"start"`
    End    time.Time `json:"end"`
    Reason string    `json:"reason"`
}

// Example usage
availability := UserAvailability{
    UserID:    "user_123",
    ProductID: "prod_456",
    Timezone:  "America/New_York",
    WorkingHours: map[string][]TimeSlot{
        "monday":    {{Start: "09:00", End: "17:00"}},
        "tuesday":   {{Start: "09:00", End: "17:00"}},
        "wednesday": {{Start: "09:00", End: "17:00"}},
        "thursday":  {{Start: "09:00", End: "17:00"}},
        "friday":    {{Start: "09:00", End: "17:00"}},
        "saturday":  {},
        "sunday":    {},
    },
    BlockedSlots: []BlockedSlot{
        {
            Start:  time.Date(2026, 2, 15, 12, 0, 0, 0, time.UTC),
            End:    time.Date(2026, 2, 15, 13, 0, 0, 0, time.UTC),
            Reason: "Lunch break",
        },
    },
    MinimumNotice: 24,
    BufferTime:    15,
}
```

### Additional Features

- Calendar integration (Google, Outlook)
- Video conferencing links (Zoom, Meet)
- SMS/Email notifications
- Multi-language support
- Custom branding per product
- Advanced scheduling algorithms
- Resource booking (rooms, equipment)

---

## 💰 Cost Estimation (Monthly)

### Development Phase

- Developer time: 4-6 weeks for MVP
- Testing & QA: 1-2 weeks

### Infrastructure (Production)

- **Hosting**: $20-50/month
  - Docker container (2GB RAM, 1 CPU)
  - PostgreSQL instance
- **Domain & SSL**: $10-15/month
- **Monitoring Tools**: $0-50/month (free tiers available)
- **Backup Storage**: $5-10/month

**Total Estimated**: $35-125/month

---

## ✅ Success Metrics

### Technical Metrics

- API uptime: >99.9%
- Response time: <200ms (p95)
- Error rate: <0.1%
- Test coverage: >80%

### Business Metrics

- Number of integrated products
- Total appointments created
- API usage growth
- Product retention rate

---

## 🤝 Support & Maintenance

### Support Channels

- Documentation website
- Email support: <support@appointment-service.com>
- GitHub issues (for bug reports)
- Integration assistance for new products

### Maintenance Schedule

- Security patches: As needed
- Feature updates: Monthly
- Dependency updates: Quarterly
- Database optimization: Quarterly

---

## 📝 Notes & Considerations

1. **Data Privacy**: Ensure compliance with GDPR/CCPA if applicable
2. **Data Retention**: Define policy for old appointments (e.g., archive after 2 years)
3. **Scalability**: Database partitioning strategy if exceeding 1M appointments
4. **Multi-region**: Consider CDN and multiple database regions for global users
5. **Versioning**: Use API versioning (v1, v2) for backward compatibility
6. **Migration**: Provide data export functionality for products
7. **Go Best Practices**:
   - Use context for request cancellation and timeouts
   - Implement proper error handling with wrapped errors
   - Use goroutines wisely with proper synchronization
   - Keep dependencies minimal and well-maintained

---

## 📚 Go Resources & Libraries

### Essential Go Packages

```bash
# Web Framework
go get github.com/gin-gonic/gin
# Alternative: go get github.com/gofiber/fiber/v2

# Database
go get github.com/jackc/pgx/v5
go get github.com/jackc/pgx/v5/pgxpool

# Configuration
go get github.com/joho/godotenv
go get github.com/spf13/viper

# Logging
go get go.uber.org/zap
# Alternative: go get github.com/sirupsen/logrus

# Authentication
go get github.com/golang-jwt/jwt/v5

# Validation
go get github.com/go-playground/validator/v10

# Testing
go get github.com/stretchr/testify

# Migrations
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# API Documentation
go get github.com/swaggo/swag/cmd/swag
go get github.com/swaggo/gin-swagger
go get github.com/swaggo/files

# UUID
go get github.com/google/uuid

# CORS
go get github.com/gin-contrib/cors

# Rate Limiting
go get golang.org/x/time/rate
```

### Recommended Learning Resources

- **Official Go Documentation**: <https://go.dev/doc/>
- **Effective Go**: <https://go.dev/doc/effective_go>
- **Go by Example**: <https://gobyexample.com/>
- **Gin Framework**: <https://gin-gonic.com/docs/>
- **PostgreSQL Driver (pgx)**: <https://github.com/jackc/pgx>
- **Go Database Patterns**: <https://go.dev/doc/database/>

---

## 📞 Contact & Resources

- **Project Repository**: <https://github.com/laith-ambianze/appointment-service>
- **Documentation**: <https://docs.appointment-service.com>
- **API Status**: <https://status.appointment-service.com>
- **Go Community**: <https://go.dev/help>
- **Gin Framework Docs**: <https://gin-gonic.com/docs/>

---

**Document Version**: 2.0  
**Last Updated**: January 30, 2026  
**Author**: Project Planning Team  
**Technology Stack**: Go (Golang) 1.21+  
**Status**: Final - Ready for Implementation
