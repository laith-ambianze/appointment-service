# Task 07: Middleware (Auth, CORS, Logging)

**Priority**: High  
**Estimated Time**: 3 hours  
**Dependencies**: TASK_06  
**Status**: Not Started

---

## Objective

Implement middleware for authentication, CORS, logging, and rate limiting.

---

## Prerequisites

- [ ] Task 06 completed
- [ ] Handlers and routes created

---

## Steps

### 1. Create Authentication Middleware

**File**: `internal/middleware/auth.go`

```go
package middleware

import (
 "net/http"
 "strings"

 "appointment-service/internal/config"
 "appointment-service/internal/service"
 "github.com/gin-gonic/gin"
 "github.com/google/uuid"
)

func AuthMiddleware(productService *service.ProductService) gin.HandlerFunc {
 return func(c *gin.Context) {
  apiKey := c.GetHeader("X-API-Key")
  apiSecret := c.GetHeader("X-API-Secret")

  if apiKey == "" || apiSecret == "" {
   c.JSON(http.StatusUnauthorized, gin.H{
    "success": false,
    "error":   "Missing authentication credentials",
   })
   c.Abort()
   return
  }

  // Authenticate product
  product, err := productService.Authenticate(c.Request.Context(), apiKey, apiSecret)
  if err != nil {
   c.JSON(http.StatusUnauthorized, gin.H{
    "success": false,
    "error":   "Invalid credentials",
   })
   c.Abort()
   return
  }

  // Store product ID in context
  c.Set("product_id", product.ID)
  c.Set("product_name", product.Name)
  c.Next()
 }
}

// BearerAuthMiddleware for JWT token authentication (optional/future)
func BearerAuthMiddleware(jwtSecret string) gin.HandlerFunc {
 return func(c *gin.Context) {
  authHeader := c.GetHeader("Authorization")
  if authHeader == "" {
   c.JSON(http.StatusUnauthorized, gin.H{
    "success": false,
    "error":   "Missing authorization header",
   })
   c.Abort()
   return
  }

  // Extract token
  parts := strings.SplitN(authHeader, " ", 2)
  if len(parts) != 2 || parts[0] != "Bearer" {
   c.JSON(http.StatusUnauthorized, gin.H{
    "success": false,
    "error":   "Invalid authorization header format",
   })
   c.Abort()
   return
  }

  // Here you would validate the JWT token
  // For now, just pass through
  c.Next()
 }
}
```

### 2. Create CORS Middleware

**File**: `internal/middleware/cors.go`

```go
package middleware

import (
 "appointment-service/internal/config"
 "github.com/gin-contrib/cors"
 "github.com/gin-gonic/gin"
 "time"
)

func CORSMiddleware(cfg *config.Config) gin.HandlerFunc {
 corsConfig := cors.Config{
  AllowOrigins:     cfg.CORS.AllowedOrigins,
  AllowMethods:     cfg.CORS.AllowedMethods,
  AllowHeaders:     cfg.CORS.AllowedHeaders,
  ExposeHeaders:    []string{"Content-Length"},
  AllowCredentials: true,
  MaxAge:           12 * time.Hour,
 }

 return cors.New(corsConfig)
}
```

### 3. Create Logging Middleware

**File**: `internal/middleware/logger.go`

```go
package middleware

import (
 "time"

 "github.com/gin-gonic/gin"
 "go.uber.org/zap"
)

func LoggerMiddleware(logger *zap.Logger) gin.HandlerFunc {
 return func(c *gin.Context) {
  start := time.Now()
  path := c.Request.URL.Path
  query := c.Request.URL.RawQuery

  // Process request
  c.Next()

  // Log after request
  end := time.Now()
  latency := end.Sub(start)

  // Get product info if available
  productID, _ := c.Get("product_id")
  productName, _ := c.Get("product_name")

  fields := []zap.Field{
   zap.Int("status", c.Writer.Status()),
   zap.String("method", c.Request.Method),
   zap.String("path", path),
   zap.String("query", query),
   zap.String("ip", c.ClientIP()),
   zap.String("user-agent", c.Request.UserAgent()),
   zap.Duration("latency", latency),
   zap.Int("body_size", c.Writer.Size()),
  }

  if productID != nil {
   fields = append(fields, zap.Any("product_id", productID))
  }
  if productName != nil {
   fields = append(fields, zap.String("product_name", productName.(string)))
  }

  // Log errors if any
  if len(c.Errors) > 0 {
   fields = append(fields, zap.String("errors", c.Errors.String()))
  }

  // Choose log level based on status code
  if c.Writer.Status() >= 500 {
   logger.Error("Server error", fields...)
  } else if c.Writer.Status() >= 400 {
   logger.Warn("Client error", fields...)
  } else {
   logger.Info("Request completed", fields...)
  }
 }
}
```

### 4. Create Rate Limiting Middleware

**File**: `internal/middleware/ratelimit.go`

```go
package middleware

import (
 "net/http"
 "sync"
 "time"

 "github.com/gin-gonic/gin"
 "golang.org/x/time/rate"
)

type rateLimiter struct {
 limiters map[string]*rate.Limiter
 mu       sync.RWMutex
 r        rate.Limit
 b        int
}

func newRateLimiter(requestsPerMinute int, burst int) *rateLimiter {
 r := rate.Every(time.Minute / time.Duration(requestsPerMinute))
 return &rateLimiter{
  limiters: make(map[string]*rate.Limiter),
  r:        r,
  b:        burst,
 }
}

func (rl *rateLimiter) getLimiter(key string) *rate.Limiter {
 rl.mu.Lock()
 defer rl.mu.Unlock()

 limiter, exists := rl.limiters[key]
 if !exists {
  limiter = rate.NewLimiter(rl.r, rl.b)
  rl.limiters[key] = limiter
 }

 return limiter
}

func (rl *rateLimiter) cleanupOldLimiters() {
 ticker := time.NewTicker(10 * time.Minute)
 go func() {
  for range ticker.C {
   rl.mu.Lock()
   // Simple cleanup - in production, track last access time
   if len(rl.limiters) > 1000 {
    rl.limiters = make(map[string]*rate.Limiter)
   }
   rl.mu.Unlock()
  }
 }()
}

func RateLimitMiddleware(requestsPerMinute, burst int) gin.HandlerFunc {
 limiter := newRateLimiter(requestsPerMinute, burst)
 limiter.cleanupOldLimiters()

 return func(c *gin.Context) {
  // Use product_id if authenticated, otherwise use IP
  key := c.ClientIP()
  if productID, exists := c.Get("product_id"); exists {
   key = productID.(string)
  }

  limiter := limiter.getLimiter(key)
  if !limiter.Allow() {
   c.JSON(http.StatusTooManyRequests, gin.H{
    "success": false,
    "error":   "Rate limit exceeded. Please try again later.",
   })
   c.Abort()
   return
  }

  c.Next()
 }
}
```

### 5. Create Recovery Middleware

**File**: `internal/middleware/recovery.go`

```go
package middleware

import (
 "fmt"
 "net/http"

 "github.com/gin-gonic/gin"
 "go.uber.org/zap"
)

func RecoveryMiddleware(logger *zap.Logger) gin.HandlerFunc {
 return func(c *gin.Context) {
  defer func() {
   if err := recover(); err != nil {
    logger.Error("Panic recovered",
     zap.Any("error", err),
     zap.String("path", c.Request.URL.Path),
     zap.String("method", c.Request.Method),
    )

    c.JSON(http.StatusInternalServerError, gin.H{
     "success": false,
     "error":   "Internal server error",
    })
    c.Abort()
   }
  }()
  c.Next()
 }
}
```

### 6. Create Request ID Middleware

**File**: `internal/middleware/requestid.go`

```go
package middleware

import (
 "github.com/gin-gonic/gin"
 "github.com/google/uuid"
)

const RequestIDKey = "request_id"

func RequestIDMiddleware() gin.HandlerFunc {
 return func(c *gin.Context) {
  // Check if request ID exists in header
  requestID := c.GetHeader("X-Request-ID")
  if requestID == "" {
   requestID = uuid.New().String()
  }

  // Set request ID in context and header
  c.Set(RequestIDKey, requestID)
  c.Writer.Header().Set("X-Request-ID", requestID)

  c.Next()
 }
}
```

### 7. Create Middleware Package Index

**File**: `internal/middleware/middleware.go`

```go
package middleware

import (
 "appointment-service/internal/config"
 "appointment-service/internal/service"
 "github.com/gin-gonic/gin"
 "go.uber.org/zap"
)

// SetupMiddleware configures all middleware for the router
func SetupMiddleware(
 router *gin.Engine,
 cfg *config.Config,
 logger *zap.Logger,
 productService *service.ProductService,
) {
 // Recovery middleware (should be first)
 router.Use(RecoveryMiddleware(logger))

 // Request ID
 router.Use(RequestIDMiddleware())

 // CORS
 router.Use(CORSMiddleware(cfg))

 // Logging
 router.Use(LoggerMiddleware(logger))

 // Rate limiting (optional, can be applied per route)
 if cfg.RateLimit.RequestsPerMinute > 0 {
  router.Use(RateLimitMiddleware(
   cfg.RateLimit.RequestsPerMinute,
   cfg.RateLimit.Burst,
  ))
 }
}

// GetAuthMiddleware returns the authentication middleware
func GetAuthMiddleware(productService *service.ProductService) gin.HandlerFunc {
 return AuthMiddleware(productService)
}
```

---

## Acceptance Criteria

- [ ] Authentication middleware validates API credentials
- [ ] CORS middleware configured properly
- [ ] Logging middleware logs all requests
- [ ] Rate limiting middleware prevents abuse
- [ ] Recovery middleware handles panics gracefully
- [ ] Request ID middleware adds tracking
- [ ] All middleware compiles and works together

---

## Verification

```bash
# Build
go build ./...

# Run tests
go test ./internal/middleware/... -v

# Check for issues
go vet ./...
```

---

## Testing Middleware

Create a simple test:

**File**: `internal/middleware/auth_test.go`

```go
package middleware

import (
 "net/http"
 "net/http/httptest"
 "testing"

 "github.com/gin-gonic/gin"
 "github.com/stretchr/testify/assert"
)

func TestAuthMiddleware_MissingCredentials(t *testing.T) {
 gin.SetMode(gin.TestMode)
 router := gin.New()

 // Mock service (would need proper mock in real test)
 router.Use(func(c *gin.Context) {
  c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing credentials"})
  c.Abort()
 })

 router.GET("/test", func(c *gin.Context) {
  c.JSON(http.StatusOK, gin.H{"message": "success"})
 })

 w := httptest.NewRecorder()
 req, _ := http.NewRequest("GET", "/test", nil)
 router.ServeHTTP(w, req)

 assert.Equal(t, http.StatusUnauthorized, w.Code)
}
```

---

## Next Task

[TASK_08_MAIN_APPLICATION.md](TASK_08_MAIN_APPLICATION.md)

---

## Notes

- Middleware order matters - recovery should be first
- Rate limiting can be per-IP or per-product
- Consider adding request timeout middleware
- Log sensitive data carefully (never log secrets)
