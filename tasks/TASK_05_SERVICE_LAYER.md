# Task 05: Service Layer (Business Logic)

**Priority**: High  
**Estimated Time**: 3 hours  
**Dependencies**: TASK_04  
**Status**: Not Started

---

## Objective

Implement business logic layer with validation, authentication helpers, and service methods.

---

## Prerequisites

- [ ] Task 04 completed
- [ ] Models and repositories implemented

---

## Steps

### 1. Create Auth Helper Package

**File**: `pkg/auth/jwt.go`

```go
package auth

import (
 "fmt"
 "time"

 "github.com/golang-jwt/jwt/v5"
 "github.com/google/uuid"
)

type Claims struct {
 ProductID uuid.UUID `json:"product_id"`
 jwt.RegisteredClaims
}

func GenerateToken(productID uuid.UUID, secret string, expirationHours int) (string, error) {
 claims := &Claims{
  ProductID: productID,
  RegisteredClaims: jwt.RegisteredClaims{
   ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(expirationHours))),
   IssuedAt:  jwt.NewNumericDate(time.Now()),
   NotBefore: jwt.NewNumericDate(time.Now()),
  },
 }

 token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
 return token.SignedString([]byte(secret))
}

func ValidateToken(tokenString string, secret string) (*Claims, error) {
 token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
  if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
   return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
  }
  return []byte(secret), nil
 })

 if err != nil {
  return nil, err
 }

 if claims, ok := token.Claims.(*Claims); ok && token.Valid {
  return claims, nil
 }

 return nil, fmt.Errorf("invalid token")
}
```

**File**: `pkg/auth/hash.go`

```go
package auth

import (
 "crypto/rand"
 "encoding/base64"
 "fmt"

 "golang.org/x/crypto/bcrypt"
)

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
 bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
 return string(bytes), err
}

// CheckPasswordHash compares a password with a hash
func CheckPasswordHash(password, hash string) bool {
 err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
 return err == nil
}

// GenerateAPIKey generates a random API key
func GenerateAPIKey(prefix string) (string, error) {
 bytes := make([]byte, 32)
 if _, err := rand.Read(bytes); err != nil {
  return "", err
 }
 return fmt.Sprintf("%s_%s", prefix, base64.URLEncoding.EncodeToString(bytes)), nil
}

// GenerateAPISecret generates a random API secret
func GenerateAPISecret() (string, error) {
 bytes := make([]byte, 48)
 if _, err := rand.Read(bytes); err != nil {
  return "", err
 }
 return base64.URLEncoding.EncodeToString(bytes), nil
}
```

### 2. Create Validator Package

**File**: `pkg/validator/validator.go`

```go
package validator

import (
 "fmt"
 "strings"

 "github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
 validate = validator.New()
}

func ValidateStruct(s interface{}) error {
 err := validate.Struct(s)
 if err != nil {
  return formatValidationErrors(err)
 }
 return nil
}

func formatValidationErrors(err error) error {
 if validationErrors, ok := err.(validator.ValidationErrors); ok {
  var messages []string
  for _, e := range validationErrors {
   messages = append(messages, fmt.Sprintf("%s: %s", e.Field(), getErrorMessage(e)))
  }
  return fmt.Errorf(strings.Join(messages, "; "))
 }
 return err
}

func getErrorMessage(e validator.FieldError) string {
 switch e.Tag() {
 case "required":
  return "is required"
 case "email":
  return "must be a valid email"
 case "url":
  return "must be a valid URL"
 case "min":
  return fmt.Sprintf("must be at least %s characters", e.Param())
 case "max":
  return fmt.Sprintf("must be at most %s characters", e.Param())
 case "oneof":
  return fmt.Sprintf("must be one of: %s", e.Param())
 default:
  return "is invalid"
 }
}
```

### 3. Create Product Service

**File**: `internal/service/product_service.go`

```go
package service

import (
 "context"
 "fmt"

 "appointment-service/internal/models"
 "appointment-service/internal/repository"
 "appointment-service/pkg/auth"
 "appointment-service/pkg/validator"
 "github.com/google/uuid"
)

type ProductService struct {
 repo repository.ProductRepository
}

func NewProductService(repo repository.ProductRepository) *ProductService {
 return &ProductService{repo: repo}
}

func (s *ProductService) Register(ctx context.Context, req *models.CreateProductRequest) (*models.Product, string, error) {
 // Validate request
 if err := validator.ValidateStruct(req); err != nil {
  return nil, "", fmt.Errorf("validation error: %w", err)
 }

 // Generate API credentials
 apiKey, err := auth.GenerateAPIKey("prod")
 if err != nil {
  return nil, "", fmt.Errorf("failed to generate API key: %w", err)
 }

 apiSecret, err := auth.GenerateAPISecret()
 if err != nil {
  return nil, "", fmt.Errorf("failed to generate API secret: %w", err)
 }

 // Hash the API secret
 hashedSecret, err := auth.HashPassword(apiSecret)
 if err != nil {
  return nil, "", fmt.Errorf("failed to hash secret: %w", err)
 }

 // Create product
 product := &models.Product{
  ID:            uuid.New(),
  Name:          req.Name,
  Description:   req.Description,
  APIKey:        apiKey,
  APISecretHash: hashedSecret,
  CallbackURL:   req.CallbackURL,
  IsActive:      true,
 }

 if err := s.repo.Create(ctx, product); err != nil {
  return nil, "", fmt.Errorf("failed to create product: %w", err)
 }

 // Return product and plaintext secret (only time it's available)
 return product, apiSecret, nil
}

func (s *ProductService) GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
 return s.repo.GetByID(ctx, id)
}

func (s *ProductService) Authenticate(ctx context.Context, apiKey, apiSecret string) (*models.Product, error) {
 product, err := s.repo.GetByAPIKey(ctx, apiKey)
 if err != nil {
  return nil, fmt.Errorf("invalid credentials")
 }

 if !product.IsActive {
  return nil, fmt.Errorf("product is inactive")
 }

 if !auth.CheckPasswordHash(apiSecret, product.APISecretHash) {
  return nil, fmt.Errorf("invalid credentials")
 }

 return product, nil
}

func (s *ProductService) List(ctx context.Context, page, pageSize int) ([]*models.Product, error) {
 offset := (page - 1) * pageSize
 return s.repo.List(ctx, pageSize, offset)
}
```

### 4. Create Appointment Service

**File**: `internal/service/appointment_service.go`

```go
package service

import (
 "context"
 "encoding/json"
 "fmt"
 "time"

 "appointment-service/internal/models"
 "appointment-service/internal/repository"
 "appointment-service/pkg/validator"
 "github.com/google/uuid"
)

type AppointmentService struct {
 repo        repository.AppointmentRepository
 productRepo repository.ProductRepository
}

func NewAppointmentService(repo repository.AppointmentRepository, productRepo repository.ProductRepository) *AppointmentService {
 return &AppointmentService{
  repo:        repo,
  productRepo: productRepo,
 }
}

func (s *AppointmentService) Create(ctx context.Context, productID uuid.UUID, req *models.CreateAppointmentRequest) (*models.Appointment, error) {
 // Validate request
 if err := validator.ValidateStruct(req); err != nil {
  return nil, fmt.Errorf("validation error: %w", err)
 }

 // Validate time range
 if !req.EndTime.After(req.StartTime) {
  return nil, fmt.Errorf("end time must be after start time")
 }

 // Validate not in the past
 if req.StartTime.Before(time.Now()) {
  return nil, fmt.Errorf("start time cannot be in the past")
 }

 // Validate at least 2 participants
 if len(req.Participants) < 2 {
  return nil, fmt.Errorf("appointment must have at least 2 participants")
 }

 // Verify product exists and is active
 product, err := s.productRepo.GetByID(ctx, productID)
 if err != nil {
  return nil, fmt.Errorf("product not found")
 }
 if !product.IsActive {
  return nil, fmt.Errorf("product is not active")
 }

 // Find host participant (creator)
 var creatorID string
 for _, p := range req.Participants {
  if p.Role == models.RoleHost {
   creatorID = p.ExternalUserID
   break
  }
 }
 if creatorID == "" {
  // If no explicit host, use first participant as creator
  creatorID = req.Participants[0].ExternalUserID
 }

 // Create appointment
 appointment := &models.Appointment{
  ID:          uuid.New(),
  ProductID:   productID,
  CreatedBy:   creatorID,
  Title:       req.Title,
  Description: req.Description,
  StartTime:   req.StartTime,
  EndTime:     req.EndTime,
  Location:    req.Location,
  MeetingType: req.MeetingType,
  Status:      models.StatusPending,
  Metadata:    req.Metadata,
 }

 // Create participants
 participants := make([]models.Participant, len(req.Participants))
 for i, pReq := range req.Participants {
  // Convert metadata to JSON
  metadataJSON, err := json.Marshal(pReq.Metadata)
  if err != nil {
   return nil, fmt.Errorf("failed to marshal participant metadata: %w", err)
  }

  participants[i] = models.Participant{
   ID:             uuid.New(),
   AppointmentID:  appointment.ID,
   ExternalUserID: pReq.ExternalUserID,
   Role:           pReq.Role,
   UserMetadata:   metadataJSON,
   Status:         models.ParticipantStatusPending,
  }
 }

 if err := s.repo.Create(ctx, appointment, participants); err != nil {
  return nil, fmt.Errorf("failed to create appointment: %w", err)
 }

 appointment.Participants = participants
 return appointment, nil
}

func (s *AppointmentService) GetByID(ctx context.Context, productID, appointmentID uuid.UUID) (*models.Appointment, error) {
 appointment, err := s.repo.GetByID(ctx, appointmentID)
 if err != nil {
  return nil, err
 }

 // Verify appointment belongs to product
 if appointment.ProductID != productID {
  return nil, fmt.Errorf("appointment not found")
 }

 return appointment, nil
}

func (s *AppointmentService) GetByProduct(ctx context.Context, productID uuid.UUID, page, pageSize int) ([]*models.Appointment, error) {
 offset := (page - 1) * pageSize
 return s.repo.GetByProductID(ctx, productID, pageSize, offset)
}

func (s *AppointmentService) GetByUser(ctx context.Context, productID uuid.UUID, userID string, page, pageSize int) ([]*models.Appointment, error) {
 offset := (page - 1) * pageSize
 return s.repo.GetByUserID(ctx, productID, userID, pageSize, offset)
}

func (s *AppointmentService) Cancel(ctx context.Context, productID, appointmentID uuid.UUID, req *models.CancelAppointmentRequest) error {
 // Validate request
 if err := validator.ValidateStruct(req); err != nil {
  return fmt.Errorf("validation error: %w", err)
 }

 // Get appointment
 appointment, err := s.GetByID(ctx, productID, appointmentID)
 if err != nil {
  return err
 }

 // Check if already cancelled
 if appointment.Status == models.StatusCancelled {
  return fmt.Errorf("appointment is already cancelled")
 }

 // Check if completed
 if appointment.Status == models.StatusCompleted {
  return fmt.Errorf("cannot cancel completed appointment")
 }

 // Update appointment
 now := time.Now()
 appointment.Status = models.StatusCancelled
 appointment.CancelledBy = &req.CancelledBy
 appointment.CancellationReason = &req.Reason
 appointment.CancelledAt = &now

 if err := s.repo.Update(ctx, appointment); err != nil {
  return fmt.Errorf("failed to cancel appointment: %w", err)
 }

 return nil
}

func (s *AppointmentService) UpdateStatus(ctx context.Context, productID, appointmentID uuid.UUID, status string) error {
 // Validate status
 validStatuses := map[string]bool{
  models.StatusPending:   true,
  models.StatusConfirmed: true,
  models.StatusCancelled: true,
  models.StatusCompleted: true,
 }
 if !validStatuses[status] {
  return fmt.Errorf("invalid status")
 }

 // Get appointment
 appointment, err := s.GetByID(ctx, productID, appointmentID)
 if err != nil {
  return err
 }

 // Update status
 appointment.Status = status

 if err := s.repo.Update(ctx, appointment); err != nil {
  return fmt.Errorf("failed to update appointment status: %w", err)
 }

 return nil
}
```

---

## Acceptance Criteria

- [ ] Auth package created with JWT and hashing functions
- [ ] Validator package created
- [ ] Product service implemented
- [ ] Appointment service implemented
- [ ] Business validation logic added
- [ ] Code compiles without errors

---

## Verification

```bash
# Build to check for errors
go build ./...

# Run tests (if any)
go test ./...

# Check for issues
go vet ./...
```

---

## Next Task

[TASK_06_HANDLERS_AND_ROUTES.md](TASK_06_HANDLERS_AND_ROUTES.md)

---

## Notes

- Service layer contains all business logic
- Keep handlers thin - delegate to services
- Services should not know about HTTP
- Add comprehensive error handling
