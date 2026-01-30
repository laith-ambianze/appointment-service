# Task 08: Main Application and Server

**Priority**: High  
**Estimated Time**: 2 hours  
**Dependencies**: TASK_07  
**Status**: Not Started

---

## Objective

Create the main application entry point, wire all components together, and implement graceful shutdown.

---

## Prerequisites

- [ ] All previous tasks (01-07) completed
- [ ] All packages implemented

---

## Steps

### 1. Create Main Application

**File**: `cmd/api/main.go`

```go
package main

import (
 "context"
 "fmt"
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
 "appointment-service/pkg/database"
 "appointment-service/pkg/logger"

 "github.com/gin-gonic/gin"
 "go.uber.org/zap"
)

func main() {
 // Load configuration
 cfg, err := config.Load()
 if err != nil {
  fmt.Printf("Failed to load configuration: %v\n", err)
  os.Exit(1)
 }

 // Initialize logger
 zapLogger, err := logger.NewLogger(cfg.Environment, cfg.Logging.Level)
 if err != nil {
  fmt.Printf("Failed to initialize logger: %v\n", err)
  os.Exit(1)
 }
 defer zapLogger.Sync()

 zapLogger.Info("Starting Appointment Service",
  zap.String("environment", cfg.Environment),
  zap.String("version", "1.0.0"),
 )

 // Connect to database
 dbPool, err := database.NewPostgresPool(cfg, zapLogger)
 if err != nil {
  zapLogger.Fatal("Failed to connect to database", zap.Error(err))
 }
 defer database.ClosePool(dbPool, zapLogger)

 // Initialize repositories
 productRepo := repository.NewProductRepository(dbPool)
 appointmentRepo := repository.NewAppointmentRepository(dbPool)

 // Initialize services
 productService := service.NewProductService(productRepo)
 appointmentService := service.NewAppointmentService(appointmentRepo, productRepo)

 // Initialize handlers
 healthHandler := handlers.NewHealthHandler()
 productHandler := handlers.NewProductHandler(productService)
 appointmentHandler := handlers.NewAppointmentHandler(appointmentService)

 // Setup Gin router
 if cfg.Environment == "production" {
  gin.SetMode(gin.ReleaseMode)
 }

 router := gin.New()

 // Setup middleware
 middleware.SetupMiddleware(router, cfg, zapLogger, productService)

 // Get auth middleware
 authMiddleware := middleware.GetAuthMiddleware(productService)

 // Setup routes
 routes.SetupRoutes(router, healthHandler, productHandler, appointmentHandler, authMiddleware)

 // Create HTTP server
 srv := &http.Server{
  Addr:         fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
  Handler:      router,
  ReadTimeout:  15 * time.Second,
  WriteTimeout: 15 * time.Second,
  IdleTimeout:  60 * time.Second,
 }

 // Start server in a goroutine
 go func() {
  zapLogger.Info("Server starting",
   zap.String("address", srv.Addr),
  )
  if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
   zapLogger.Fatal("Server failed to start", zap.Error(err))
  }
 }()

 zapLogger.Info("Server started successfully",
  zap.String("address", srv.Addr),
  zap.String("environment", cfg.Environment),
 )

 // Wait for interrupt signal for graceful shutdown
 quit := make(chan os.Signal, 1)
 signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
 <-quit

 zapLogger.Info("Shutting down server...")

 // Graceful shutdown with 10 second timeout
 ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
 defer cancel()

 if err := srv.Shutdown(ctx); err != nil {
  zapLogger.Error("Server forced to shutdown", zap.Error(err))
 }

 zapLogger.Info("Server exited gracefully")
}
```

### 2. Add Missing Database Package

**File**: `pkg/database/postgres.go` (if not created in Task 03)

```go
package database

import (
 "context"
 "fmt"
 "time"

 "appointment-service/internal/config"
 "github.com/jackc/pgx/v5/pgxpool"
 "go.uber.org/zap"
)

func NewPostgresPool(cfg *config.Config, logger *zap.Logger) (*pgxpool.Pool, error) {
 poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL())
 if err != nil {
  return nil, fmt.Errorf("failed to parse database URL: %w", err)
 }

 // Configure connection pool
 poolConfig.MaxConns = int32(cfg.Database.MaxConnections)
 poolConfig.MinConns = int32(cfg.Database.MaxIdleConnections)
 poolConfig.MaxConnLifetime = cfg.Database.MaxLifetime
 poolConfig.MaxConnIdleTime = 30 * time.Minute
 poolConfig.HealthCheckPeriod = 1 * time.Minute

 ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
 defer cancel()

 pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
 if err != nil {
  return nil, fmt.Errorf("failed to create connection pool: %w", err)
 }

 // Test connection
 if err := pool.Ping(ctx); err != nil {
  return nil, fmt.Errorf("failed to ping database: %w", err)
 }

 logger.Info("Database connection established",
  zap.String("host", cfg.Database.Host),
  zap.String("database", cfg.Database.Name),
  zap.Int("max_connections", cfg.Database.MaxConnections),
 )

 return pool, nil
}

func ClosePool(pool *pgxpool.Pool, logger *zap.Logger) {
 if pool != nil {
  pool.Close()
  logger.Info("Database connection closed")
 }
}
```

### 3. Update Makefile

Add to existing Makefile:

```makefile
# Build and Run
.PHONY: build run dev

build:
 @echo "Building application..."
 @mkdir -p bin
 go build -o bin/appointment-service ./cmd/api

run: build
 @echo "Running application..."
 @./bin/appointment-service

dev:
 @echo "Running in development mode..."
 air

# Full application lifecycle
.PHONY: start stop restart logs

start: db-start
 @echo "Starting application..."
 @sleep 3
 make db-migrate-up
 make build
 ./bin/appointment-service

stop:
 @echo "Stopping services..."
 @pkill -f appointment-service || true
 make db-stop

restart: stop start

logs:
 @tail -f logs/app.log
```

### 4. Create Air Configuration (Hot Reload)

**File**: `.air.toml`

```toml
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
args_bin = []
bin = "./tmp/main"
cmd = "go build -o ./tmp/main ./cmd/api"
delay = 1000
exclude_dir = ["assets", "tmp", "vendor", "testdata", "migrations"]
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

### 5. Create Startup Script (Optional)

**File**: `scripts/start.sh` (Linux/Mac)

```bash
#!/bin/bash

echo "🚀 Starting Appointment Service..."

# Check if .env exists
if [ ! -f .env ]; then
    echo "⚠️  .env file not found. Copying from .env.example..."
    cp .env.example .env
    echo "✅ Please edit .env with your configuration"
    exit 1
fi

# Start database
echo "📦 Starting database..."
make db-start

# Wait for database
echo "⏳ Waiting for database to be ready..."
sleep 5

# Run migrations
echo "🗄️  Running database migrations..."
make db-migrate-up

# Build application
echo "🔨 Building application..."
make build

# Run application
echo "🎯 Starting server..."
./bin/appointment-service
```

**File**: `scripts/start.ps1` (Windows)

```powershell
Write-Host "🚀 Starting Appointment Service..." -ForegroundColor Green

# Check if .env exists
if (-not (Test-Path .env)) {
    Write-Host "⚠️  .env file not found. Copying from .env.example..." -ForegroundColor Yellow
    Copy-Item .env.example .env
    Write-Host "✅ Please edit .env with your configuration" -ForegroundColor Green
    exit 1
}

# Start database
Write-Host "📦 Starting database..." -ForegroundColor Cyan
make db-start

# Wait for database
Write-Host "⏳ Waiting for database to be ready..." -ForegroundColor Yellow
Start-Sleep -Seconds 5

# Run migrations
Write-Host "🗄️  Running database migrations..." -ForegroundColor Cyan
make db-migrate-up

# Build application
Write-Host "🔨 Building application..." -ForegroundColor Cyan
make build

# Run application
Write-Host "🎯 Starting server..." -ForegroundColor Green
.\bin\appointment-service.exe
```

### 6. Update go.mod

Ensure all dependencies are listed:

```bash
go mod tidy
go mod verify
```

---

## Acceptance Criteria

- [ ] Main application compiles successfully
- [ ] All components properly wired together
- [ ] Server starts and listens on configured port
- [ ] Health check endpoint responds
- [ ] Graceful shutdown works
- [ ] Logging outputs properly
- [ ] Database connection established
- [ ] Configuration loads correctly

---

## Verification

```bash
# Build application
make build

# Start database
make db-start

# Run migrations
make db-migrate-up

# Run application
make run

# In another terminal, test health endpoint
curl http://localhost:8080/health

# Test product registration
curl -X POST http://localhost:8080/v1/products/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Product",
    "description": "My test application"
  }'

# Stop with Ctrl+C and verify graceful shutdown
```

---

## Next Task

[TASK_09_UNIT_TESTS.md](TASK_09_UNIT_TESTS.md)

---

## Notes

- Always test graceful shutdown
- Monitor logs for any errors
- Verify database connections are properly closed
- Check memory usage and goroutine leaks
- Test with different environment configurations
