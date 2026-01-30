# Task 02: Configuration and Logger Setup

**Priority**: High  
**Estimated Time**: 2 hours  
**Dependencies**: TASK_01  
**Status**: Not Started

---

## Objective

Create configuration loader and structured logging system for the application.

---

## Prerequisites

- [ ] Task 01 completed
- [ ] Dependencies installed

---

## Steps

### 1. Create Configuration Package

**File**: `internal/config/config.go`

```go
package config

import (
 "fmt"
 "os"
 "strconv"
 "time"

 "github.com/joho/godotenv"
)

type Config struct {
 Environment string
 Host        string
 Port        string
 
 Database DatabaseConfig
 Security SecurityConfig
 CORS     CORSConfig
 RateLimit RateLimitConfig
 Logging  LoggingConfig
}

type DatabaseConfig struct {
 Host               string
 Port               string
 User               string
 Password           string
 Name               string
 SSLMode            string
 MaxConnections     int
 MaxIdleConnections int
 MaxLifetime        time.Duration
}

type SecurityConfig struct {
 JWTSecret     string
 SaltRounds    int
}

type CORSConfig struct {
 AllowedOrigins []string
 AllowedMethods []string
 AllowedHeaders []string
}

type RateLimitConfig struct {
 RequestsPerMinute int
 Burst             int
}

type LoggingConfig struct {
 Level  string
 Format string
}

func Load() (*Config, error) {
 // Load .env file in development
 if os.Getenv("GO_ENV") != "production" {
  if err := godotenv.Load(); err != nil {
   // Not fatal, might be using system environment variables
   fmt.Println("Warning: .env file not found, using system environment variables")
  }
 }

 config := &Config{
  Environment: getEnv("GO_ENV", "development"),
  Host:        getEnv("API_HOST", "0.0.0.0"),
  Port:        getEnv("API_PORT", "8080"),
  
  Database: DatabaseConfig{
   Host:               getEnv("DB_HOST", "localhost"),
   Port:               getEnv("DB_PORT", "5432"),
   User:               getEnv("DB_USER", "appointments"),
   Password:           getEnv("DB_PASSWORD", ""),
   Name:               getEnv("DB_NAME", "appointments"),
   SSLMode:            getEnv("DB_SSL_MODE", "disable"),
   MaxConnections:     getEnvInt("DB_MAX_CONNECTIONS", 25),
   MaxIdleConnections: getEnvInt("DB_MAX_IDLE_CONNECTIONS", 5),
   MaxLifetime:        getEnvDuration("DB_CONNECTION_MAX_LIFETIME", 5*time.Minute),
  },
  
  Security: SecurityConfig{
   JWTSecret:  getEnv("JWT_SECRET", ""),
   SaltRounds: getEnvInt("API_SECRET_SALT_ROUNDS", 10),
  },
  
  CORS: CORSConfig{
   AllowedOrigins: getEnvSlice("CORS_ALLOWED_ORIGINS", []string{"*"}),
   AllowedMethods: getEnvSlice("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "PATCH", "DELETE"}),
   AllowedHeaders: getEnvSlice("CORS_ALLOWED_HEADERS", []string{"Content-Type", "Authorization"}),
  },
  
  RateLimit: RateLimitConfig{
   RequestsPerMinute: getEnvInt("RATE_LIMIT_REQUESTS_PER_MINUTE", 100),
   Burst:             getEnvInt("RATE_LIMIT_BURST", 20),
  },
  
  Logging: LoggingConfig{
   Level:  getEnv("LOG_LEVEL", "info"),
   Format: getEnv("LOG_FORMAT", "json"),
  },
 }

 if err := config.Validate(); err != nil {
  return nil, err
 }

 return config, nil
}

func (c *Config) Validate() error {
 if c.Database.Password == "" {
  return fmt.Errorf("DB_PASSWORD is required")
 }
 if c.Security.JWTSecret == "" {
  return fmt.Errorf("JWT_SECRET is required")
 }
 if len(c.Security.JWTSecret) < 32 {
  return fmt.Errorf("JWT_SECRET must be at least 32 characters")
 }
 return nil
}

func (c *Config) DatabaseURL() string {
 return fmt.Sprintf(
  "postgres://%s:%s@%s:%s/%s?sslmode=%s",
  c.Database.User,
  c.Database.Password,
  c.Database.Host,
  c.Database.Port,
  c.Database.Name,
  c.Database.SSLMode,
 )
}

// Helper functions
func getEnv(key, defaultValue string) string {
 if value := os.Getenv(key); value != "" {
  return value
 }
 return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
 if value := os.Getenv(key); value != "" {
  if intValue, err := strconv.Atoi(value); err == nil {
   return intValue
  }
 }
 return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
 if value := os.Getenv(key); value != "" {
  if duration, err := time.ParseDuration(value); err == nil {
   return duration
  }
 }
 return defaultValue
}

func getEnvSlice(key string, defaultValue []string) []string {
 if value := os.Getenv(key); value != "" {
  return splitAndTrim(value, ",")
 }
 return defaultValue
}

func splitAndTrim(s, sep string) []string {
 var result []string
 for _, item := range split(s, sep) {
  if trimmed := trim(item); trimmed != "" {
   result = append(result, trimmed)
  }
 }
 return result
}

func split(s, sep string) []string {
 // Simple split implementation
 var result []string
 start := 0
 for i := 0; i < len(s); i++ {
  if s[i] == sep[0] {
   result = append(result, s[start:i])
   start = i + 1
  }
 }
 result = append(result, s[start:])
 return result
}

func trim(s string) string {
 // Simple trim implementation
 start := 0
 end := len(s)
 for start < end && (s[start] == ' ' || s[start] == '\t') {
  start++
 }
 for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
  end--
 }
 return s[start:end]
}
```

### 2. Create Logger Package

**File**: `pkg/logger/logger.go`

```go
package logger

import (
 "go.uber.org/zap"
 "go.uber.org/zap/zapcore"
)

func NewLogger(env string, level string) (*zap.Logger, error) {
 var config zap.Config
 
 if env == "production" {
  config = zap.NewProductionConfig()
 } else {
  config = zap.NewDevelopmentConfig()
  config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
 }
 
 // Set log level
 logLevel, err := zapcore.ParseLevel(level)
 if err != nil {
  logLevel = zapcore.InfoLevel
 }
 config.Level = zap.NewAtomicLevelAt(logLevel)
 
 logger, err := config.Build(
  zap.AddCaller(),
  zap.AddStacktrace(zapcore.ErrorLevel),
 )
 if err != nil {
  return nil, err
 }
 
 return logger, nil
}
```

### 3. Create Config Test

**File**: `internal/config/config_test.go`

```go
package config

import (
 "os"
 "testing"
 "github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
 // Set required environment variables
 os.Setenv("DB_PASSWORD", "testpassword")
 os.Setenv("JWT_SECRET", "this-is-a-test-secret-key-with-32-chars-minimum")
 
 cfg, err := Load()
 
 assert.NoError(t, err)
 assert.NotNil(t, cfg)
 assert.Equal(t, "development", cfg.Environment)
 assert.Equal(t, "testpassword", cfg.Database.Password)
}

func TestValidate(t *testing.T) {
 tests := []struct {
  name    string
  config  *Config
  wantErr bool
 }{
  {
   name: "valid config",
   config: &Config{
    Database: DatabaseConfig{Password: "password"},
    Security: SecurityConfig{JWTSecret: "this-is-a-very-long-secret-key-for-testing-purposes"},
   },
   wantErr: false,
  },
  {
   name: "missing DB password",
   config: &Config{
    Database: DatabaseConfig{Password: ""},
    Security: SecurityConfig{JWTSecret: "this-is-a-very-long-secret-key-for-testing-purposes"},
   },
   wantErr: true,
  },
  {
   name: "short JWT secret",
   config: &Config{
    Database: DatabaseConfig{Password: "password"},
    Security: SecurityConfig{JWTSecret: "short"},
   },
   wantErr: true,
  },
 }
 
 for _, tt := range tests {
  t.Run(tt.name, func(t *testing.T) {
   err := tt.config.Validate()
   if tt.wantErr {
    assert.Error(t, err)
   } else {
    assert.NoError(t, err)
   }
  })
 }
}
```

---

## Acceptance Criteria

- [ ] Config package loads environment variables
- [ ] Config validation works correctly
- [ ] Logger package creates zap logger
- [ ] Different log levels for dev/prod
- [ ] Tests pass for config package
- [ ] Database URL generation works

---

## Verification

```bash
# Run tests
go test ./internal/config -v

# Build to check for errors
go build ./...
```

---

## Next Task

[TASK_03_DATABASE_SETUP.md](TASK_03_DATABASE_SETUP.md)

---

## Notes

- Never commit `.env` file with real credentials
- Use strong JWT secrets in production (32+ characters)
- Consider using secrets management in production (AWS Secrets Manager, HashiCorp Vault)
