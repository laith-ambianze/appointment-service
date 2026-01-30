# Task 06: Handlers and API Routes

**Priority**: High  
**Estimated Time**: 4 hours  
**Dependencies**: TASK_05  
**Status**: Not Started

---

## Objective

Create HTTP handlers for all API endpoints and set up routing.

---

## Prerequisites

- [ ] Task 05 completed
- [ ] Service layer implemented

---

## Steps

### 1. Create Product Handler

**File**: `internal/handlers/product_handler.go`

```go
package handlers

import (
 "net/http"

 "appointment-service/internal/models"
 "appointment-service/internal/service"
 "github.com/gin-gonic/gin"
 "github.com/google/uuid"
)

type ProductHandler struct {
 service *service.ProductService
}

func NewProductHandler(service *service.ProductService) *ProductHandler {
 return &ProductHandler{service: service}
}

// Register godoc
// @Summary Register a new product
// @Description Register a new product and receive API credentials
// @Tags products
// @Accept json
// @Produce json
// @Param request body models.CreateProductRequest true "Product registration request"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/products/register [post]
func (h *ProductHandler) Register(c *gin.Context) {
 var req models.CreateProductRequest
 if err := c.ShouldBindJSON(&req); err != nil {
  c.JSON(http.StatusBadRequest, gin.H{
   "success": false,
   "error":   err.Error(),
  })
  return
 }

 product, apiSecret, err := h.service.Register(c.Request.Context(), &req)
 if err != nil {
  c.JSON(http.StatusInternalServerError, gin.H{
   "success": false,
   "error":   err.Error(),
  })
  return
 }

 c.JSON(http.StatusCreated, gin.H{
  "success": true,
  "data": gin.H{
   "productId": product.ID,
   "apiKey":    product.APIKey,
   "apiSecret": apiSecret,
   "message":   "Store these credentials securely. The secret will not be shown again.",
  },
 })
}

// GetMe godoc
// @Summary Get current product information
// @Description Get information about the authenticated product
// @Tags products
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /v1/products/me [get]
func (h *ProductHandler) GetMe(c *gin.Context) {
 productID, exists := c.Get("product_id")
 if !exists {
  c.JSON(http.StatusUnauthorized, gin.H{
   "success": false,
   "error":   "Unauthorized",
  })
  return
 }

 product, err := h.service.GetByID(c.Request.Context(), productID.(uuid.UUID))
 if err != nil {
  c.JSON(http.StatusInternalServerError, gin.H{
   "success": false,
   "error":   err.Error(),
  })
  return
 }

 c.JSON(http.StatusOK, gin.H{
  "success": true,
  "data": gin.H{
   "id":          product.ID,
   "name":        product.Name,
   "description": product.Description,
   "isActive":    product.IsActive,
   "createdAt":   product.CreatedAt,
  },
 })
}
```

### 2. Create Appointment Handler

**File**: `internal/handlers/appointment_handler.go`

```go
package handlers

import (
 "net/http"
 "strconv"

 "appointment-service/internal/models"
 "appointment-service/internal/service"
 "github.com/gin-gonic/gin"
 "github.com/google/uuid"
)

type AppointmentHandler struct {
 service *service.AppointmentService
}

func NewAppointmentHandler(service *service.AppointmentService) *AppointmentHandler {
 return &AppointmentHandler{service: service}
}

// Create godoc
// @Summary Create a new appointment
// @Description Create a new appointment between two users
// @Tags appointments
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.CreateAppointmentRequest true "Appointment creation request"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/appointments [post]
func (h *AppointmentHandler) Create(c *gin.Context) {
 productID := getProductID(c)
 if productID == uuid.Nil {
  c.JSON(http.StatusUnauthorized, gin.H{
   "success": false,
   "error":   "Unauthorized",
  })
  return
 }

 var req models.CreateAppointmentRequest
 if err := c.ShouldBindJSON(&req); err != nil {
  c.JSON(http.StatusBadRequest, gin.H{
   "success": false,
   "error":   err.Error(),
  })
  return
 }

 appointment, err := h.service.Create(c.Request.Context(), productID, &req)
 if err != nil {
  c.JSON(http.StatusBadRequest, gin.H{
   "success": false,
   "error":   err.Error(),
  })
  return
 }

 c.JSON(http.StatusCreated, gin.H{
  "success": true,
  "data": gin.H{
   "appointmentId": appointment.ID,
   "status":        appointment.Status,
   "createdAt":     appointment.CreatedAt,
  },
 })
}

// GetAll godoc
// @Summary Get all appointments for product
// @Description Get all appointments for the authenticated product
// @Tags appointments
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(20)
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /v1/appointments [get]
func (h *AppointmentHandler) GetAll(c *gin.Context) {
 productID := getProductID(c)
 if productID == uuid.Nil {
  c.JSON(http.StatusUnauthorized, gin.H{
   "success": false,
   "error":   "Unauthorized",
  })
  return
 }

 page := getIntQuery(c, "page", 1)
 pageSize := getIntQuery(c, "pageSize", 20)

 if pageSize > 100 {
  pageSize = 100
 }

 appointments, err := h.service.GetByProduct(c.Request.Context(), productID, page, pageSize)
 if err != nil {
  c.JSON(http.StatusInternalServerError, gin.H{
   "success": false,
   "error":   err.Error(),
  })
  return
 }

 c.JSON(http.StatusOK, gin.H{
  "success": true,
  "data": gin.H{
   "appointments": appointments,
   "pagination": gin.H{
    "page":     page,
    "pageSize": pageSize,
   },
  },
 })
}

// GetByID godoc
// @Summary Get appointment by ID
// @Description Get a specific appointment by ID
// @Tags appointments
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Appointment ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /v1/appointments/{id} [get]
func (h *AppointmentHandler) GetByID(c *gin.Context) {
 productID := getProductID(c)
 if productID == uuid.Nil {
  c.JSON(http.StatusUnauthorized, gin.H{
   "success": false,
   "error":   "Unauthorized",
  })
  return
 }

 appointmentID, err := uuid.Parse(c.Param("id"))
 if err != nil {
  c.JSON(http.StatusBadRequest, gin.H{
   "success": false,
   "error":   "Invalid appointment ID",
  })
  return
 }

 appointment, err := h.service.GetByID(c.Request.Context(), productID, appointmentID)
 if err != nil {
  c.JSON(http.StatusNotFound, gin.H{
   "success": false,
   "error":   err.Error(),
  })
  return
 }

 c.JSON(http.StatusOK, gin.H{
  "success": true,
  "data":    appointment,
 })
}

// GetByUser godoc
// @Summary Get user appointments
// @Description Get all appointments for a specific user
// @Tags appointments
// @Produce json
// @Security ApiKeyAuth
// @Param userId path string true "User ID"
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(20)
// @Success 200 {object} map[string]interface{}
// @Router /v1/appointments/user/{userId} [get]
func (h *AppointmentHandler) GetByUser(c *gin.Context) {
 productID := getProductID(c)
 if productID == uuid.Nil {
  c.JSON(http.StatusUnauthorized, gin.H{
   "success": false,
   "error":   "Unauthorized",
  })
  return
 }

 userID := c.Param("userId")
 page := getIntQuery(c, "page", 1)
 pageSize := getIntQuery(c, "pageSize", 20)

 if pageSize > 100 {
  pageSize = 100
 }

 appointments, err := h.service.GetByUser(c.Request.Context(), productID, userID, page, pageSize)
 if err != nil {
  c.JSON(http.StatusInternalServerError, gin.H{
   "success": false,
   "error":   err.Error(),
  })
  return
 }

 c.JSON(http.StatusOK, gin.H{
  "success": true,
  "data": gin.H{
   "appointments": appointments,
   "pagination": gin.H{
    "page":     page,
    "pageSize": pageSize,
   },
  },
 })
}

// Cancel godoc
// @Summary Cancel an appointment
// @Description Cancel a specific appointment
// @Tags appointments
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Appointment ID"
// @Param request body models.CancelAppointmentRequest true "Cancellation request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /v1/appointments/{id}/cancel [patch]
func (h *AppointmentHandler) Cancel(c *gin.Context) {
 productID := getProductID(c)
 if productID == uuid.Nil {
  c.JSON(http.StatusUnauthorized, gin.H{
   "success": false,
   "error":   "Unauthorized",
  })
  return
 }

 appointmentID, err := uuid.Parse(c.Param("id"))
 if err != nil {
  c.JSON(http.StatusBadRequest, gin.H{
   "success": false,
   "error":   "Invalid appointment ID",
  })
  return
 }

 var req models.CancelAppointmentRequest
 if err := c.ShouldBindJSON(&req); err != nil {
  c.JSON(http.StatusBadRequest, gin.H{
   "success": false,
   "error":   err.Error(),
  })
  return
 }

 if err := h.service.Cancel(c.Request.Context(), productID, appointmentID, &req); err != nil {
  c.JSON(http.StatusBadRequest, gin.H{
   "success": false,
   "error":   err.Error(),
  })
  return
 }

 c.JSON(http.StatusOK, gin.H{
  "success": true,
  "message": "Appointment cancelled successfully",
 })
}

// UpdateStatus godoc
// @Summary Update appointment status
// @Description Update the status of an appointment
// @Tags appointments
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Appointment ID"
// @Param request body map[string]string true "Status update request"
// @Success 200 {object} map[string]interface{}
// @Router /v1/appointments/{id}/status [patch]
func (h *AppointmentHandler) UpdateStatus(c *gin.Context) {
 productID := getProductID(c)
 if productID == uuid.Nil {
  c.JSON(http.StatusUnauthorized, gin.H{
   "success": false,
   "error":   "Unauthorized",
  })
  return
 }

 appointmentID, err := uuid.Parse(c.Param("id"))
 if err != nil {
  c.JSON(http.StatusBadRequest, gin.H{
   "success": false,
   "error":   "Invalid appointment ID",
  })
  return
 }

 var req struct {
  Status string `json:"status" binding:"required"`
 }
 if err := c.ShouldBindJSON(&req); err != nil {
  c.JSON(http.StatusBadRequest, gin.H{
   "success": false,
   "error":   err.Error(),
  })
  return
 }

 if err := h.service.UpdateStatus(c.Request.Context(), productID, appointmentID, req.Status); err != nil {
  c.JSON(http.StatusBadRequest, gin.H{
   "success": false,
   "error":   err.Error(),
  })
  return
 }

 c.JSON(http.StatusOK, gin.H{
  "success": true,
  "message": "Status updated successfully",
 })
}

// Helper functions
func getProductID(c *gin.Context) uuid.UUID {
 productID, exists := c.Get("product_id")
 if !exists {
  return uuid.Nil
 }
 return productID.(uuid.UUID)
}

func getIntQuery(c *gin.Context, key string, defaultValue int) int {
 valueStr := c.DefaultQuery(key, strconv.Itoa(defaultValue))
 value, err := strconv.Atoi(valueStr)
 if err != nil {
  return defaultValue
 }
 return value
}
```

### 3. Create Health Check Handler

**File**: `internal/handlers/health_handler.go`

```go
package handlers

import (
 "net/http"
 "github.com/gin-gonic/gin"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
 return &HealthHandler{}
}

// Health godoc
// @Summary Health check
// @Description Check if the service is running
// @Tags health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
func (h *HealthHandler) Check(c *gin.Context) {
 c.JSON(http.StatusOK, gin.H{
  "status":  "healthy",
  "service": "appointment-service",
 })
}
```

### 4. Create Routes

**File**: `internal/routes/routes.go`

```go
package routes

import (
 "appointment-service/internal/handlers"
 "appointment-service/internal/middleware"
 "github.com/gin-gonic/gin"
)

func SetupRoutes(
 router *gin.Engine,
 healthHandler *handlers.HealthHandler,
 productHandler *handlers.ProductHandler,
 appointmentHandler *handlers.AppointmentHandler,
 authMiddleware gin.HandlerFunc,
) {
 // Health check (no auth)
 router.GET("/health", healthHandler.Check)

 // API v1
 v1 := router.Group("/v1")
 {
  // Product routes
  products := v1.Group("/products")
  {
   products.POST("/register", productHandler.Register)
   products.GET("/me", authMiddleware, productHandler.GetMe)
  }

  // Appointment routes (all require auth)
  appointments := v1.Group("/appointments")
  appointments.Use(authMiddleware)
  {
   appointments.POST("", appointmentHandler.Create)
   appointments.GET("", appointmentHandler.GetAll)
   appointments.GET("/:id", appointmentHandler.GetByID)
   appointments.GET("/user/:userId", appointmentHandler.GetByUser)
   appointments.PATCH("/:id/cancel", appointmentHandler.Cancel)
   appointments.PATCH("/:id/status", appointmentHandler.UpdateStatus)
  }
 }
}
```

---

## Acceptance Criteria

- [ ] Product handler created with Register and GetMe endpoints
- [ ] Appointment handler created with all CRUD operations
- [ ] Health check handler created
- [ ] Routes properly configured
- [ ] All handlers return consistent JSON responses
- [ ] Error handling implemented
- [ ] Code compiles without errors

---

## Verification

```bash
# Build
go build ./...

# Check for errors
go vet ./...

# Format code
go fmt ./...
```

---

## Next Task

[TASK_07_MIDDLEWARE.md](TASK_07_MIDDLEWARE.md)

---

## Notes

- Keep handlers thin - delegate to services
- Use consistent response format
- Handle all error cases
- Add swagger comments for documentation
