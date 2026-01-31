# Task W01-05: Initial Code Skeleton

**Status**: Not Started  
**Estimated Time**: 2-3 hours  
**Prerequisites**: TASK_W01_04_CICD_PIPELINE.md  
**Next Task**: Week 02 Tasks

---

## Objective

Create the initial code skeleton including configuration management, logger, and a basic HTTP server with health check endpoint.

---

## Steps

### 1. Create Configuration Package

Location: `appointment-service/internal/config/config.go`

```go
package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	// Application
	Env     string
	APIHost string
	APIPort string

	// Database
	DBHost               string
	DBPort               string
	DBUser               string
	DBPassword           string
	DBName               string
	DBSSLMode            string
	DBMaxConnections     int
	DBMaxIdleConnections int
	DBMaxLifetime        string

	// Redis
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int

	// Security
	JWTSecret         string
	APISecretRounds   int

	// CORS
	CORSAllowedOrigins string
	CORSAllowedMethods string
	CORSAllowedHeaders string

	// Rate Limiting
	RateLimitRequestsPerMinute int
	RateLimitBurst             int

	// Logging
	LogLevel  string
	LogFormat string

	// Monitoring
	PrometheusEnabled bool
	PrometheusPort    string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists (ignore error if not found)
	_ = godotenv.Load()

	cfg := &Config{
		// Application
		Env:     getEnv("GO_ENV", "development"),
		APIHost: getEnv("API_HOST", "0.0.0.0"),
		APIPort: getEnv("API_PORT", "8080"),

		// Database
		DBHost:               getEnv("DB_HOST", "localhost"),
		DBPort:               getEnv("DB_PORT", "5432"),
		DBUser:               getEnv("DB_USER", "appointments"),
		DBPassword:           getEnv("DB_PASSWORD", "password"),
		DBName:               getEnv("DB_NAME", "appointments_dev"),
		DBSSLMode:            getEnv("DB_SSL_MODE", "disable"),
		DBMaxConnections:     getEnvAsInt("DB_MAX_CONNECTIONS", 25),
		DBMaxIdleConnections: getEnvAsInt("DB_MAX_IDLE_CONNECTIONS", 5),
		DBMaxLifetime:        getEnv("DB_CONNECTION_MAX_LIFETIME", "5m"),

		// Redis
		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvAsInt("REDIS_DB", 0),

		// Security
		JWTSecret:       getEnv("JWT_SECRET", "change-me-in-production"),
		APISecretRounds: getEnvAsInt("API_SECRET_SALT_ROUNDS", 10),

		// CORS
		CORSAllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "*"),
		CORSAllowedMethods: getEnv("CORS_ALLOWED_METHODS", "GET,POST,PUT,PATCH,DELETE,OPTIONS"),
		CORSAllowedHeaders: getEnv("CORS_ALLOWED_HEADERS", "Content-Type,Authorization,X-API-Key,X-API-Secret"),

		// Rate Limiting
		RateLimitRequestsPerMinute: getEnvAsInt("RATE_LIMIT_REQUESTS_PER_MINUTE", 100),
		RateLimitBurst:             getEnvAsInt("RATE_LIMIT_BURST", 20),

		// Logging
		LogLevel:  getEnv("LOG_LEVEL", "debug"),
		LogFormat: getEnv("LOG_FORMAT", "json"),

		// Monitoring
		PrometheusEnabled: getEnvAsBool("PROMETHEUS_ENABLED", true),
		PrometheusPort:    getEnv("PROMETHEUS_PORT", "9090"),
	}

	// Validate required fields
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks if required configuration values are set
func (c *Config) Validate() error {
	if c.DBPassword == "password" && c.Env == "production" {
		return fmt.Errorf("DB_PASSWORD must be changed in production")
	}

	if c.JWTSecret == "change-me-in-production" && c.Env == "production" {
		return fmt.Errorf("JWT_SECRET must be changed in production")
	}

	return nil
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Env == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Env == "production"
}

// DatabaseDSN returns the PostgreSQL connection string
func (c *Config) DatabaseDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

// Helper functions

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}

func getEnvAsBool(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return fallback
}
```

### 2. Create Logger Package

Location: `appointment-service/pkg/logger/logger.go`

```go
package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap logger
type Logger struct {
	*zap.Logger
}

// New creates a new logger instance
func New(level, format string) (*Logger, error) {
	var config zap.Config

	// Configure based on format
	if format == "json" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// Set log level
	logLevel, err := parseLevel(level)
	if err != nil {
		logLevel = zapcore.InfoLevel
	}
	config.Level = zap.NewAtomicLevelAt(logLevel)

	// Build logger
	zapLogger, err := config.Build(
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return nil, err
	}

	return &Logger{zapLogger}, nil
}

// NewDefault creates a development logger
func NewDefault() *Logger {
	logger, _ := New("debug", "console")
	return logger
}

// parseLevel converts string level to zapcore.Level
func parseLevel(level string) (zapcore.Level, error) {
	switch level {
	case "debug":
		return zapcore.DebugLevel, nil
	case "info":
		return zapcore.InfoLevel, nil
	case "warn", "warning":
		return zapcore.WarnLevel, nil
	case "error":
		return zapcore.ErrorLevel, nil
	case "fatal":
		return zapcore.FatalLevel, nil
	default:
		return zapcore.InfoLevel, nil
	}
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}

// WithFields adds fields to the logger
func (l *Logger) WithFields(fields ...zap.Field) *Logger {
	return &Logger{l.Logger.With(fields...)}
}

// InfoF logs a formatted info message
func (l *Logger) InfoF(template string, args ...interface{}) {
	l.Sugar().Infof(template, args...)
}

// ErrorF logs a formatted error message
func (l *Logger) ErrorF(template string, args ...interface{}) {
	l.Sugar().Errorf(template, args...)
}

// DebugF logs a formatted debug message
func (l *Logger) DebugF(template string, args ...interface{}) {
	l.Sugar().Debugf(template, args...)
}

// WarnF logs a formatted warning message
func (l *Logger) WarnF(template string, args ...interface{}) {
	l.Sugar().Warnf(template, args...)
}

// FatalF logs a formatted fatal message and exits
func (l *Logger) FatalF(template string, args ...interface{}) {
	l.Sugar().Fatalf(template, args...)
	os.Exit(1)
}
```

### 3. Create Health Handler

Location: `appointment-service/internal/handlers/health.go`

```go
package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	startTime time.Time
}

// NewHealthHandler creates a new health handler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{
		startTime: time.Now(),
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Service   string    `json:"service"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
	Uptime    string    `json:"uptime"`
}

// Health returns the service health status
// @Summary Health check endpoint
// @Description Returns the health status of the service
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	uptime := time.Since(h.startTime)

	c.JSON(http.StatusOK, HealthResponse{
		Status:    "ok",
		Service:   "appointment-service",
		Version:   "1.0.0",
		Timestamp: time.Now().UTC(),
		Uptime:    uptime.String(),
	})
}

// Ready returns readiness status (for Kubernetes)
// @Summary Readiness check endpoint
// @Description Returns whether the service is ready to accept traffic
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /ready [get]
func (h *HealthHandler) Ready(c *gin.Context) {
	// TODO: Add actual readiness checks (database connection, etc.)
	c.JSON(http.StatusOK, HealthResponse{
		Status:    "ready",
		Service:   "appointment-service",
		Version:   "1.0.0",
		Timestamp: time.Now().UTC(),
		Uptime:    time.Since(h.startTime).String(),
	})
}

// Live returns liveness status (for Kubernetes)
// @Summary Liveness check endpoint
// @Description Returns whether the service is alive
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /live [get]
func (h *HealthHandler) Live(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status:    "alive",
		Service:   "appointment-service",
		Version:   "1.0.0",
		Timestamp: time.Now().UTC(),
		Uptime:    time.Since(h.startTime).String(),
	})
}
```

### 4. Create Main Application

Location: `appointment-service/cmd/api/main.go`

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
	setupRoutes(router)

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
func setupRoutes(router *gin.Engine) {
	// Health check handler
	healthHandler := handlers.NewHealthHandler()

	// Health check endpoints (no version prefix)
	router.GET("/health", healthHandler.Health)
	router.GET("/ready", healthHandler.Ready)
	router.GET("/live", healthHandler.Live)

	// API v1 routes
	v1 := router.Group("/v1")
	{
		// TODO: Add API endpoints here
		v1.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "pong",
			})
		})
	}
}

// loggerMiddleware creates a Gin middleware for logging
func loggerMiddleware(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log request
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		if raw != "" {
			path = path + "?" + raw
		}

		log.Info("HTTP Request",
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("client_ip", clientIP),
			zap.String("error", errorMessage),
		)
	}
}
```

### 5. Create Local .env File

```bash
# Copy example to actual .env
cp .env.example .env

# The .env file will be used for local development
```

### 6. Install Dependencies

```bash
# Install all Go dependencies
make install

# Or manually
go get github.com/gin-gonic/gin
go get github.com/joho/godotenv
go get go.uber.org/zap
go mod tidy
```

### 7. Build and Run

```bash
# Build the application
make build

# Or run directly
make run
```

### 8. Test the Application

```bash
# In another terminal, test endpoints

# Health check
curl http://localhost:8080/health

# Expected response:
# {
#   "status": "ok",
#   "service": "appointment-service",
#   "version": "1.0.0",
#   "timestamp": "2026-01-31T12:00:00Z",
#   "uptime": "5.123456s"
# }

# Ready check
curl http://localhost:8080/ready

# Live check
curl http://localhost:8080/live

# Ping endpoint
curl http://localhost:8080/v1/ping

# Expected: {"message":"pong"}
```

### 9. Test with Docker Compose

```bash
# Start all services
make docker-up

# Check logs
make docker-logs

# Test health endpoint
curl http://localhost:8080/health

# Stop services
make docker-down
```

### 10. Commit Initial Code

```bash
# Add all files
git add .

# Commit
git commit -m "feat: add initial application skeleton

- Add configuration management with validation
- Add Zap logger with structured logging
- Add health check endpoints (health, ready, live)
- Add main application with graceful shutdown
- Add logging middleware
- Add basic route setup
- Support for development and production modes"

# Push
git push origin master
```

---

## Verification Checklist

- [ ] `internal/config/config.go` created and tested
- [ ] `pkg/logger/logger.go` created and working
- [ ] `internal/handlers/health.go` created with 3 endpoints
- [ ] `cmd/api/main.go` created with full server setup
- [ ] `.env` file created from `.env.example`
- [ ] Dependencies installed successfully
- [ ] Application builds without errors (`make build`)
- [ ] Application runs successfully (`make run`)
- [ ] Health endpoint returns 200 OK
- [ ] Ready endpoint returns 200 OK
- [ ] Live endpoint returns 200 OK
- [ ] Ping endpoint returns pong
- [ ] Logs are output correctly
- [ ] Docker Compose starts successfully
- [ ] Code committed and pushed

---

## Expected Console Output

When running `make run`, you should see:

```
2026-01-31T12:00:00.000Z	INFO	appointment-service/main.go:35	Starting Appointment Service	{"env": "development", "port": "8080", "log_level": "debug"}
2026-01-31T12:00:00.000Z	INFO	appointment-service/main.go:59	Server starting	{"address": "0.0.0.0:8080"}
```

When making a request:

```
2026-01-31T12:00:05.000Z	INFO	appointment-service/main.go:95	HTTP Request	{"method": "GET", "path": "/health", "status": 200, "latency": "123.456µs", "client_ip": "127.0.0.1", "error": ""}
```

---

## Testing CI Pipeline

After pushing, the GitHub Actions CI pipeline should:
1. ✅ Lint the code successfully
2. ✅ Run tests (none yet, but no failures)
3. ✅ Build the binary
4. ✅ Build Docker image (if on master)

---

## Next Steps

**Week 01 Complete!** 🎉

You now have:
- ✅ Project structure
- ✅ Configuration files
- ✅ ADR documentation
- ✅ CI/CD pipeline
- ✅ Working HTTP server with health checks

**Proceed to Week 02 tasks:**
- Database setup and migrations
- Repository layer implementation
- Service layer business logic
- API endpoints for appointments

---

**Status**: ⏸️ Ready to Start
