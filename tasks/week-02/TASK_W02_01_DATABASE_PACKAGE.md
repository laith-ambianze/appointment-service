# Task W02-01: Database Connection Package

**Status**: ⏸️ Not Started  
**Estimated Time**: 3-4 hours  
**Prerequisites**: Week 01 Complete  
**Next Task**: TASK_W02_02_MIGRATIONS.md

---

## Objective

Create a robust database connection package using pgx v5 with connection pooling, health checks, and transaction support.

---

## Steps

### 1. Install Database Dependencies

```bash
# Install pgx v5 (PostgreSQL driver)
go get github.com/jackc/pgx/v5
go get github.com/jackc/pgx/v5/pgxpool

# Install golang-migrate for migrations
go get -tags 'postgres' github.com/golang-migrate/migrate/v4
go get github.com/golang-migrate/migrate/v4/database/postgres
go get github.com/golang-migrate/migrate/v4/source/file

# Update go.mod
go mod tidy
```

### 2. Update Configuration for Database

Location: `appointment-service/internal/config/config.go`

Add validation method (if not already present):

```go
// Validate checks if required configuration values are set
func (c *Config) Validate() error {
 if c.DBPassword == "password" && c.Env == "production" {
  return fmt.Errorf("default database password cannot be used in production")
 }
 
 if c.JWTSecret == "your-jwt-secret-key-minimum-32-characters-change-in-production" && c.Env == "production" {
  return fmt.Errorf("default JWT secret cannot be used in production")
 }
 
 return nil
}
```

### 3. Create Database Package

Location: `appointment-service/pkg/database/postgres.go`

```go
package database

import (
 "context"
 "fmt"
 "time"

 "github.com/jackc/pgx/v5/pgxpool"
 "go.uber.org/zap"
)

// Config holds database configuration
type Config struct {
 Host            string
 Port            string
 User            string
 Password        string
 Database        string
 SSLMode         string
 MaxConnections  int
 MaxIdleConns    int
 ConnMaxLifetime time.Duration
}

// PostgresDB wraps the pgx connection pool
type PostgresDB struct {
 Pool   *pgxpool.Pool
 logger *zap.Logger
}

// NewPostgresDB creates a new PostgreSQL connection pool
func NewPostgresDB(cfg Config, logger *zap.Logger) (*PostgresDB, error) {
 // Build connection string
 dsn := fmt.Sprintf(
  "host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
  cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode,
 )

 // Configure connection pool
 poolConfig, err := pgxpool.ParseConfig(dsn)
 if err != nil {
  return nil, fmt.Errorf("failed to parse database config: %w", err)
 }

 // Set pool configuration
 poolConfig.MaxConns = int32(cfg.MaxConnections)
 poolConfig.MinConns = int32(cfg.MaxIdleConns)
 poolConfig.MaxConnLifetime = cfg.ConnMaxLifetime
 poolConfig.MaxConnIdleTime = 30 * time.Minute
 poolConfig.HealthCheckPeriod = 1 * time.Minute

 // Create connection pool
 ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
 defer cancel()

 pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
 if err != nil {
  return nil, fmt.Errorf("failed to create connection pool: %w", err)
 }

 // Test connection
 if err := pool.Ping(ctx); err != nil {
  pool.Close()
  return nil, fmt.Errorf("failed to ping database: %w", err)
 }

 logger.Info("Database connection established",
  zap.String("host", cfg.Host),
  zap.String("database", cfg.Database),
  zap.Int("max_connections", cfg.MaxConnections),
 )

 return &PostgresDB{
  Pool:   pool,
  logger: logger,
 }, nil
}

// Close closes the database connection pool
func (db *PostgresDB) Close() {
 if db.Pool != nil {
  db.logger.Info("Closing database connection pool")
  db.Pool.Close()
 }
}

// Ping checks if the database is accessible
func (db *PostgresDB) Ping(ctx context.Context) error {
 return db.Pool.Ping(ctx)
}

// Stats returns connection pool statistics
func (db *PostgresDB) Stats() *pgxpool.Stat {
 return db.Pool.Stat()
}
```

### 4. Create Transaction Support

Location: `appointment-service/pkg/database/transaction.go`

```go
package database

import (
 "context"
 "fmt"

 "github.com/jackc/pgx/v5"
)

// TxFunc represents a function that executes within a transaction
type TxFunc func(tx pgx.Tx) error

// WithTransaction executes a function within a database transaction
// If the function returns an error, the transaction is rolled back
// Otherwise, the transaction is committed
func (db *PostgresDB) WithTransaction(ctx context.Context, fn TxFunc) error {
 // Begin transaction
 tx, err := db.Pool.Begin(ctx)
 if err != nil {
  return fmt.Errorf("failed to begin transaction: %w", err)
 }

 // Ensure transaction is either committed or rolled back
 defer func() {
  if p := recover(); p != nil {
   // Panic occurred, rollback
   _ = tx.Rollback(ctx)
   panic(p) // Re-throw panic after rollback
  } else if err != nil {
   // Error occurred, rollback
   _ = tx.Rollback(ctx)
  } else {
   // Success, commit
   err = tx.Commit(ctx)
   if err != nil {
    db.logger.Error("Failed to commit transaction", zap.Error(err))
   }
  }
 }()

 // Execute the function within the transaction
 err = fn(tx)
 return err
}
```

### 5. Create Database Health Check

Location: `appointment-service/pkg/database/health.go`

```go
package database

import (
 "context"
 "time"
)

// HealthCheck performs a database health check
type HealthCheck struct {
 Status      string            `json:"status"`
 Message     string            `json:"message,omitempty"`
 Connections ConnectionStats   `json:"connections"`
 ResponseTime string           `json:"response_time"`
}

// ConnectionStats holds connection pool statistics
type ConnectionStats struct {
 TotalConnections  int32 `json:"total_connections"`
 IdleConnections   int32 `json:"idle_connections"`
 AcquireCount      int64 `json:"acquire_count"`
 AcquireDuration   int64 `json:"acquire_duration_ms"`
 MaxConnections    int32 `json:"max_connections"`
}

// Health performs a health check on the database
func (db *PostgresDB) Health(ctx context.Context) HealthCheck {
 start := time.Now()
 
 // Try to ping the database
 err := db.Ping(ctx)
 responseTime := time.Since(start)
 
 if err != nil {
  return HealthCheck{
   Status:      "unhealthy",
   Message:     err.Error(),
   ResponseTime: responseTime.String(),
  }
 }
 
 // Get connection pool stats
 stats := db.Stats()
 
 return HealthCheck{
  Status: "healthy",
  Connections: ConnectionStats{
   TotalConnections: stats.TotalConns(),
   IdleConnections:  stats.IdleConns(),
   AcquireCount:     stats.AcquireCount(),
   AcquireDuration:  stats.AcquireDuration().Milliseconds(),
   MaxConnections:   stats.MaxConns(),
  },
  ResponseTime: responseTime.String(),
 }
}
```

### 6. Update docker-compose.yml

Location: `appointment-service/docker-compose.yml`

Add PostgreSQL service:

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    container_name: appointment-postgres
    environment:
      POSTGRES_USER: appointments
      POSTGRES_PASSWORD: password
      POSTGRES_DB: appointments_dev
    ports:
      - "1998:1998"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U appointments"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - appointment-network

  redis:
    image: redis:7-alpine
    container_name: appointment-redis
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - appointment-network

  app:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: appointment-service
    ports:
      - "8081:8081"
    environment:
      GO_ENV: development
      API_PORT: 8081
      DB_HOST: postgres
      DB_PORT: 1998
      DB_USER: appointments
      DB_PASSWORD: password
      DB_NAME: appointments_dev
      REDIS_HOST: redis
      REDIS_PORT: 6379
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    networks:
      - appointment-network
    volumes:
      - .:/app
    command: ["./bin/appointment-service"]

volumes:
  postgres_data:

networks:
  appointment-network:
    driver: bridge
```

### 7. Update main.go to Initialize Database

Location: `appointment-service/cmd/api/main.go`

Update the main function:

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

 "github.com/gin-gonic/gin"
 "github.com/laith-ambianze/appointment-service/internal/config"
 "github.com/laith-ambianze/appointment-service/internal/handlers"
 "github.com/laith-ambianze/appointment-service/pkg/database"
 "github.com/laith-ambianze/appointment-service/pkg/logger"
 "go.uber.org/zap"
)

func main() {
 // Load configuration
 cfg, err := config.Load()
 if err != nil {
  fmt.Printf("Failed to load configuration: %v\n", err)
  os.Exit(1)
 }

 // Validate configuration
 if err := cfg.Validate(); err != nil {
  fmt.Printf("Invalid configuration: %v\n", err)
  os.Exit(1)
 }

 // Initialize logger
 log, err := logger.New(cfg.LogLevel, cfg.LogFormat)
 if err != nil {
  fmt.Printf("Failed to initialize logger: %v\n", err)
  os.Exit(1)
 }
 defer log.Sync()

 log.Info("Starting Appointment Service",
  zap.String("env", cfg.Env),
  zap.String("port", cfg.APIPort),
  zap.String("log_level", cfg.LogLevel),
 )

 // Initialize database
 dbConfig := database.Config{
  Host:            cfg.DBHost,
  Port:            cfg.DBPort,
  User:            cfg.DBUser,
  Password:        cfg.DBPassword,
  Database:        cfg.DBName,
  SSLMode:         cfg.DBSSLMode,
  MaxConnections:  cfg.DBMaxConnections,
  MaxIdleConns:    cfg.DBMaxIdleConnections,
  ConnMaxLifetime: parseDuration(cfg.DBMaxLifetime, 5*time.Minute),
 }

 db, err := database.NewPostgresDB(dbConfig, log)
 if err != nil {
  log.Fatal("Failed to connect to database", zap.Error(err))
 }
 defer db.Close()

 log.Info("Database connection established")

 // Set Gin mode
 if cfg.IsProduction() {
  gin.SetMode(gin.ReleaseMode)
 }

 // Create Gin router
 router := gin.New()

 // Add middleware
 router.Use(gin.Recovery())
 router.Use(loggerMiddleware(log))

 // Setup routes
 setupRoutes(router, db)

 // Create HTTP server
 addr := fmt.Sprintf("%s:%s", cfg.APIHost, cfg.APIPort)
 srv := &http.Server{
  Addr:              addr,
  Handler:           router,
  ReadTimeout:       10 * time.Second,
  WriteTimeout:      10 * time.Second,
  IdleTimeout:       60 * time.Second,
  ReadHeaderTimeout: 5 * time.Second,
 }

 // Start server in goroutine
 go func() {
  log.Info("Server starting", zap.String("address", addr))
  if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
   log.Fatal("Failed to start server", zap.Error(err))
  }
 }()

 // Wait for interrupt signal
 quit := make(chan os.Signal, 1)
 signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
 <-quit

 log.Info("Shutting down server...")

 // Graceful shutdown with 5 second timeout
 ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
 defer cancel()

 if err := srv.Shutdown(ctx); err != nil {
  log.Fatal("Server forced to shutdown", zap.Error(err))
 }

 log.Info("Server exited")
}

// setupRoutes configures all application routes
func setupRoutes(router *gin.Engine, db *database.PostgresDB) {
 // Health check handler
 healthHandler := handlers.NewHealthHandler()

 // Health check endpoints (no version prefix)
 router.GET("/health", healthHandler.Health)
 router.GET("/ready", func(c *gin.Context) {
  // Check database health
  dbHealth := db.Health(c.Request.Context())
  
  if dbHealth.Status != "healthy" {
   c.JSON(http.StatusServiceUnavailable, gin.H{
    "status": "not ready",
    "database": dbHealth,
   })
   return
  }
  
  c.JSON(http.StatusOK, gin.H{
   "status": "ready",
   "database": dbHealth,
  })
 })

 // API v1 routes (will be implemented in later tasks)
 v1 := router.Group("/v1")
 {
  v1.GET("/ping", func(c *gin.Context) {
   c.JSON(http.StatusOK, gin.H{"message": "pong"})
  })
 }
}

// loggerMiddleware creates a Gin middleware for logging requests
func loggerMiddleware(log *logger.Logger) gin.HandlerFunc {
 return func(c *gin.Context) {
  start := time.Now()
  path := c.Request.URL.Path
  query := c.Request.URL.RawQuery

  c.Next()

  duration := time.Since(start)
  statusCode := c.Writer.Status()

  log.Info("HTTP Request",
   zap.String("method", c.Request.Method),
   zap.String("path", path),
   zap.String("query", query),
   zap.Int("status", statusCode),
   zap.Duration("duration", duration),
   zap.String("client_ip", c.ClientIP()),
  )
 }
}

// parseDuration parses a duration string, returns fallback if parsing fails
func parseDuration(s string, fallback time.Duration) time.Duration {
 d, err := time.ParseDuration(s)
 if err != nil {
  return fallback
 }
 return d
}
```

### 8. Update Makefile with Database Commands

Location: `appointment-service/Makefile`

Add database-related commands:

```makefile
# Database commands
.PHONY: db-start db-stop db-reset db-console

db-start: ## Start database container
 @echo "$(CYAN)Starting database...$(NC)"
 docker-compose up -d postgres
 @echo "$(GREEN)Database started$(NC)"

db-stop: ## Stop database container
 @echo "$(CYAN)Stopping database...$(NC)"
 docker-compose stop postgres
 @echo "$(GREEN)Database stopped$(NC)"

db-reset: ## Reset database (drops and recreates)
 @echo "$(CYAN)Resetting database...$(NC)"
 docker-compose down -v postgres
 docker-compose up -d postgres
 sleep 5
 $(MAKE) migrate-up
 @echo "$(GREEN)Database reset complete$(NC)"

db-console: ## Open PostgreSQL console
 @echo "$(CYAN)Opening database console...$(NC)"
 docker-compose exec postgres psql -U appointments -d appointments_dev
```

### 9. Create Test File for Database Package

Location: `appointment-service/pkg/database/postgres_test.go`

```go
package database

import (
 "context"
 "testing"
 "time"

 "github.com/stretchr/testify/assert"
 "github.com/stretchr/testify/require"
 "go.uber.org/zap"
)

func TestNewPostgresDB(t *testing.T) {
 // Skip if not in integration test mode
 if testing.Short() {
  t.Skip("Skipping integration test")
 }

 logger := zap.NewNop()
 
 cfg := Config{
  Host:            "localhost",
  Port:            "1998",
  User:            "appointments",
  Password:        "password",
  Database:        "appointments_dev",
  SSLMode:         "disable",
  MaxConnections:  10,
  MaxIdleConns:    5,
  ConnMaxLifetime: 5 * time.Minute,
 }

 db, err := NewPostgresDB(cfg, logger)
 require.NoError(t, err)
 require.NotNil(t, db)
 defer db.Close()

 // Test connection
 ctx := context.Background()
 err = db.Ping(ctx)
 assert.NoError(t, err)

 // Test stats
 stats := db.Stats()
 assert.NotNil(t, stats)
}

func TestPostgresDB_Health(t *testing.T) {
 if testing.Short() {
  t.Skip("Skipping integration test")
 }

 logger := zap.NewNop()
 
 cfg := Config{
  Host:            "localhost",
  Port:            "1998",
  User:            "appointments",
  Password:        "password",
  Database:        "appointments_dev",
  SSLMode:         "disable",
  MaxConnections:  10,
  MaxIdleConns:    5,
  ConnMaxLifetime: 5 * time.Minute,
 }

 db, err := NewPostgresDB(cfg, logger)
 require.NoError(t, err)
 defer db.Close()

 ctx := context.Background()
 health := db.Health(ctx)

 assert.Equal(t, "healthy", health.Status)
 assert.NotZero(t, health.Connections.TotalConnections)
}
```

---

## Verification Checklist

- [ ] Database dependencies installed
- [ ] Database package created (postgres.go, transaction.go, health.go)
- [ ] docker-compose.yml updated with PostgreSQL service
- [ ] main.go updated to initialize database
- [ ] Makefile updated with database commands
- [ ] Test file created
- [ ] Code compiles without errors: `make build`
- [ ] Database starts: `make db-start`
- [ ] Application connects to database: `make run`
- [ ] Health endpoint includes database status: `curl http://localhost:8081/ready`

---

## Testing

```bash
# Start database
make db-start

# Wait for database to be ready
sleep 5

# Run the application
make run

# In another terminal, test endpoints
curl http://localhost:8081/health
curl http://localhost:8081/ready

# Should see database connection in logs
# Expected output includes: "Database connection established"
```

---

## Expected Output

When running `make run`, you should see:

```md
{"level":"info","ts":...,"msg":"Starting Appointment Service","env":"development","port":"8081","log_level":"debug"}
{"level":"info","ts":...,"msg":"Database connection established","host":"localhost","database":"appointments_dev","max_connections":25}
{"level":"info","ts":...,"msg":"Server starting","address":"0.0.0.0:8081"}
```

When calling `/ready`:

```json
{
  "status": "ready",
  "database": {
    "status": "healthy",
    "connections": {
      "total_connections": 1,
      "idle_connections": 1,
      "acquire_count": 2,
      "acquire_duration_ms": 0,
      "max_connections": 25
    },
    "response_time": "145.2µs"
  }
}
```

---

## Troubleshooting

**Issue**: Cannot connect to database  
**Solution**:

- Check PostgreSQL is running: `docker-compose ps`
- Check connection string in .env file
- Check PostgreSQL logs: `docker-compose logs postgres`

**Issue**: "too many clients" error  
**Solution**: Reduce `DB_MAX_CONNECTIONS` in .env file

**Issue**: Tests fail with "database does not exist"  
**Solution**: Create database first: `make db-start`

---

## Next Task

After completing this task, proceed to [TASK_W02_02_MIGRATIONS.md](TASK_W02_02_MIGRATIONS.md) to set up the migration system and create the database schema.
