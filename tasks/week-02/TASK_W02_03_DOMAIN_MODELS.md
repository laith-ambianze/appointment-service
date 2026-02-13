# Task W02-03: Domain Models & Entities

**Status**: ⏸️ Not Started  
**Estimated Time**: 2-3 hours  
**Prerequisites**: TASK_W02_02_MIGRATIONS.md  
**Next Task**: TASK_W02_04_PRODUCT_REPOSITORY.md

---

## Objective

Create Go domain models that represent the database entities with proper validation, JSON marshaling, and business logic.

---

## Steps

### 1. Create Base Model

Location: `appointment-service/internal/models/base.go`

```go
package models

import (
 "database/sql"
 "time"

 "github.com/google/uuid"
)

// BaseModel contains common fields for all models
type BaseModel struct {
 ID        uuid.UUID      `json:"id"`
 CreatedAt time.Time      `json:"created_at"`
 UpdatedAt time.Time      `json:"updated_at"`
 DeletedAt sql.NullTime   `json:"deleted_at,omitempty"`
}

// IsDeleted returns true if the model is soft-deleted
func (b *BaseModel) IsDeleted() bool {
 return b.DeletedAt.Valid
}
```

### 2. Create Product Model

Location: `appointment-service/internal/models/product.go`

```go
package models

import (
 "encoding/json"
 "fmt"

 "github.com/google/uuid"
)

// ProductStatus represents the status of a product
type ProductStatus string

const (
 ProductStatusActive   ProductStatus = "active"
 ProductStatusInactive ProductStatus = "inactive"
 ProductStatusSuspended ProductStatus = "suspended"
)

// Product represents a registered product in the system
type Product struct {
 BaseModel
 Name           string            `json:"name"`
 Description    string            `json:"description,omitempty"`
 APIKey         string            `json:"api_key"`
 APISecretHash  string            `json:"-"` // Never expose in JSON
 Status         ProductStatus     `json:"status"`
 WebhookURL     string            `json:"webhook_url,omitempty"`
 WebhookSecret  string            `json:"-"` // Never expose in JSON
 Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// Validate validates the product fields
func (p *Product) Validate() error {
 if p.Name == "" {
  return fmt.Errorf("product name is required")
 }
 
 if len(p.Name) > 255 {
  return fmt.Errorf("product name must be less than 255 characters")
 }
 
 if p.APIKey == "" {
  return fmt.Errorf("API key is required")
 }
 
 if p.APISecretHash == "" {
  return fmt.Errorf("API secret hash is required")
 }
 
 if p.Status == "" {
  p.Status = ProductStatusActive
 }
 
 // Validate status
 if p.Status != ProductStatusActive && 
    p.Status != ProductStatusInactive && 
    p.Status != ProductStatusSuspended {
  return fmt.Errorf("invalid product status: %s", p.Status)
 }
 
 return nil
}

// IsActive returns true if the product is active
func (p *Product) IsActive() bool {
 return p.Status == ProductStatusActive && !p.IsDeleted()
}

// MetadataJSON returns metadata as JSON string
func (p *Product) MetadataJSON() (string, error) {
 if p.Metadata == nil {
  return "{}", nil
 }
 
 bytes, err := json.Marshal(p.Metadata)
 if err != nil {
  return "", fmt.Errorf("failed to marshal metadata: %w", err)
 }
 
 return string(bytes), nil
}

// CreateProductInput represents input for creating a product
type CreateProductInput struct {
 Name          string                 `json:"name" binding:"required,max=255"`
 Description   string                 `json:"description" binding:"max=1000"`
 WebhookURL    string                 `json:"webhook_url" binding:"omitempty,url"`
 Metadata      map[string]interface{} `json:"metadata"`
}

// UpdateProductInput represents input for updating a product
type UpdateProductInput struct {
 Name          *string                 `json:"name" binding:"omitempty,max=255"`
 Description   *string                 `json:"description" binding:"omitempty,max=1000"`
 Status        *ProductStatus          `json:"status" binding:"omitempty,oneof=active inactive suspended"`
 WebhookURL    *string                 `json:"webhook_url" binding:"omitempty,url"`
 WebhookSecret *string                 `json:"webhook_secret"`
 Metadata      map[string]interface{}  `json:"metadata"`
}
```

### 3. Create Appointment Model

Location: `appointment-service/internal/models/appointment.go`

```go
package models

import (
 "encoding/json"
 "fmt"
 "time"

 "github.com/google/uuid"
)

// AppointmentStatus represents the status of an appointment
type AppointmentStatus string

const (
 AppointmentStatusScheduled AppointmentStatus = "scheduled"
 AppointmentStatusConfirmed AppointmentStatus = "confirmed"
 AppointmentStatusCancelled AppointmentStatus = "cancelled"
 AppointmentStatusCompleted AppointmentStatus = "completed"
 AppointmentStatusNoShow    AppointmentStatus = "no_show"
)

// Appointment represents a scheduled appointment
type Appointment struct {
 BaseModel
 ProductID    uuid.UUID                  `json:"product_id"`
 Title        string                     `json:"title"`
 Description  string                     `json:"description,omitempty"`
 StartTime    time.Time                  `json:"start_time"`
 EndTime      time.Time                  `json:"end_time"`
 Timezone     string                     `json:"timezone"`
 Location     string                     `json:"location,omitempty"`
 Status       AppointmentStatus          `json:"status"`
 CreatedBy    string                     `json:"created_by"`
 Metadata     map[string]interface{}     `json:"metadata,omitempty"`
 Participants []AppointmentParticipant   `json:"participants,omitempty"`
}

// Validate validates the appointment fields
func (a *Appointment) Validate() error {
 if a.ProductID == uuid.Nil {
  return fmt.Errorf("product ID is required")
 }
 
 if a.Title == "" {
  return fmt.Errorf("appointment title is required")
 }
 
 if len(a.Title) > 255 {
  return fmt.Errorf("appointment title must be less than 255 characters")
 }
 
 if a.StartTime.IsZero() {
  return fmt.Errorf("start time is required")
 }
 
 if a.EndTime.IsZero() {
  return fmt.Errorf("end time is required")
 }
 
 if a.EndTime.Before(a.StartTime) || a.EndTime.Equal(a.StartTime) {
  return fmt.Errorf("end time must be after start time")
 }
 
 if a.Timezone == "" {
  a.Timezone = "UTC"
 }
 
 if a.CreatedBy == "" {
  return fmt.Errorf("created_by is required")
 }
 
 if a.Status == "" {
  a.Status = AppointmentStatusScheduled
 }
 
 // Validate status
 validStatuses := []AppointmentStatus{
  AppointmentStatusScheduled,
  AppointmentStatusConfirmed,
  AppointmentStatusCancelled,
  AppointmentStatusCompleted,
  AppointmentStatusNoShow,
 }
 
 isValidStatus := false
 for _, status := range validStatuses {
  if a.Status == status {
   isValidStatus = true
   break
  }
 }
 
 if !isValidStatus {
  return fmt.Errorf("invalid appointment status: %s", a.Status)
 }
 
 return nil
}

// IsActive returns true if the appointment is not cancelled or deleted
func (a *Appointment) IsActive() bool {
 return a.Status != AppointmentStatusCancelled && !a.IsDeleted()
}

// IsPast returns true if the appointment end time is in the past
func (a *Appointment) IsPast() bool {
 return a.EndTime.Before(time.Now())
}

// IsFuture returns true if the appointment start time is in the future
func (a *Appointment) IsFuture() bool {
 return a.StartTime.After(time.Now())
}

// Duration returns the duration of the appointment
func (a *Appointment) Duration() time.Duration {
 return a.EndTime.Sub(a.StartTime)
}

// MetadataJSON returns metadata as JSON string
func (a *Appointment) MetadataJSON() (string, error) {
 if a.Metadata == nil {
  return "{}", nil
 }
 
 bytes, err := json.Marshal(a.Metadata)
 if err != nil {
  return "", fmt.Errorf("failed to marshal metadata: %w", err)
 }
 
 return string(bytes), nil
}

// CreateAppointmentInput represents input for creating an appointment
type CreateAppointmentInput struct {
 ProductID    uuid.UUID                  `json:"product_id" binding:"required"`
 Title        string                     `json:"title" binding:"required,max=255"`
 Description  string                     `json:"description" binding:"max=2000"`
 StartTime    time.Time                  `json:"start_time" binding:"required"`
 EndTime      time.Time                  `json:"end_time" binding:"required,gtfield=StartTime"`
 Timezone     string                     `json:"timezone" binding:"required"`
 Location     string                     `json:"location" binding:"max=500"`
 CreatedBy    string                     `json:"created_by" binding:"required"`
 Metadata     map[string]interface{}     `json:"metadata"`
 Participants []ParticipantInput         `json:"participants" binding:"required,min=1,dive"`
}

// UpdateAppointmentInput represents input for updating an appointment
type UpdateAppointmentInput struct {
 Title       *string                `json:"title" binding:"omitempty,max=255"`
 Description *string                `json:"description" binding:"omitempty,max=2000"`
 StartTime   *time.Time             `json:"start_time"`
 EndTime     *time.Time             `json:"end_time"`
 Timezone    *string                `json:"timezone"`
 Location    *string                `json:"location" binding:"omitempty,max=500"`
 Status      *AppointmentStatus     `json:"status"`
 Metadata    map[string]interface{} `json:"metadata"`
}
```

### 4. Create Participant Model

Location: `appointment-service/internal/models/participant.go`

```go
package models

import (
 "encoding/json"
 "fmt"

 "github.com/google/uuid"
)

// ParticipantRole represents the role of a participant
type ParticipantRole string

const (
 ParticipantRoleHost     ParticipantRole = "host"
 ParticipantRoleGuest    ParticipantRole = "guest"
 ParticipantRoleAttendee ParticipantRole = "attendee"
 ParticipantRoleObserver ParticipantRole = "observer"
)

// ParticipantStatus represents the status of a participant
type ParticipantStatus string

const (
 ParticipantStatusPending   ParticipantStatus = "pending"
 ParticipantStatusAccepted  ParticipantStatus = "accepted"
 ParticipantStatusDeclined  ParticipantStatus = "declined"
 ParticipantStatusTentative ParticipantStatus = "tentative"
)

// AppointmentParticipant represents a participant in an appointment
type AppointmentParticipant struct {
 BaseModel
 AppointmentID  uuid.UUID              `json:"appointment_id"`
 ExternalUserID string                 `json:"external_user_id"`
 Role           ParticipantRole        `json:"role"`
 Status         ParticipantStatus      `json:"status"`
 UserMetadata   map[string]interface{} `json:"user_metadata,omitempty"`
}

// Validate validates the participant fields
func (p *AppointmentParticipant) Validate() error {
 if p.AppointmentID == uuid.Nil {
  return fmt.Errorf("appointment ID is required")
 }
 
 if p.ExternalUserID == "" {
  return fmt.Errorf("external user ID is required")
 }
 
 if len(p.ExternalUserID) > 255 {
  return fmt.Errorf("external user ID must be less than 255 characters")
 }
 
 if p.Role == "" {
  p.Role = ParticipantRoleGuest
 }
 
 // Validate role
 validRoles := []ParticipantRole{
  ParticipantRoleHost,
  ParticipantRoleGuest,
  ParticipantRoleAttendee,
  ParticipantRoleObserver,
 }
 
 isValidRole := false
 for _, role := range validRoles {
  if p.Role == role {
   isValidRole = true
   break
  }
 }
 
 if !isValidRole {
  return fmt.Errorf("invalid participant role: %s", p.Role)
 }
 
 if p.Status == "" {
  p.Status = ParticipantStatusPending
 }
 
 // Validate status
 validStatuses := []ParticipantStatus{
  ParticipantStatusPending,
  ParticipantStatusAccepted,
  ParticipantStatusDeclined,
  ParticipantStatusTentative,
 }
 
 isValidStatus := false
 for _, status := range validStatuses {
  if p.Status == status {
   isValidStatus = true
   break
  }
 }
 
 if !isValidStatus {
  return fmt.Errorf("invalid participant status: %s", p.Status)
 }
 
 return nil
}

// IsHost returns true if the participant is the host
func (p *AppointmentParticipant) IsHost() bool {
 return p.Role == ParticipantRoleHost
}

// HasAccepted returns true if the participant has accepted
func (p *AppointmentParticipant) HasAccepted() bool {
 return p.Status == ParticipantStatusAccepted
}

// UserMetadataJSON returns user metadata as JSON string
func (p *AppointmentParticipant) UserMetadataJSON() (string, error) {
 if p.UserMetadata == nil {
  return "{}", nil
 }
 
 bytes, err := json.Marshal(p.UserMetadata)
 if err != nil {
  return "", fmt.Errorf("failed to marshal user metadata: %w", err)
 }
 
 return string(bytes), nil
}

// ParticipantInput represents input for adding a participant
type ParticipantInput struct {
 ExternalUserID string                 `json:"external_user_id" binding:"required,max=255"`
 Role           ParticipantRole        `json:"role" binding:"required,oneof=host guest attendee observer"`
 Status         ParticipantStatus      `json:"status" binding:"omitempty,oneof=pending accepted declined tentative"`
 UserMetadata   map[string]interface{} `json:"user_metadata"`
}

// UpdateParticipantInput represents input for updating a participant
type UpdateParticipantInput struct {
 Role         *ParticipantRole       `json:"role" binding:"omitempty,oneof=host guest attendee observer"`
 Status       *ParticipantStatus     `json:"status" binding:"omitempty,oneof=pending accepted declined tentative"`
 UserMetadata map[string]interface{} `json:"user_metadata"`
}
```

### 5. Create Model Tests

Location: `appointment-service/internal/models/appointment_test.go`

```go
package models

import (
 "testing"
 "time"

 "github.com/google/uuid"
 "github.com/stretchr/testify/assert"
 "github.com/stretchr/testify/require"
)

func TestAppointment_Validate(t *testing.T) {
 tests := []struct {
  name        string
  appointment *Appointment
  wantErr     bool
  errContains string
 }{
  {
   name: "valid appointment",
   appointment: &Appointment{
    ProductID:  uuid.New(),
    Title:      "Test Appointment",
    StartTime:  time.Now().Add(1 * time.Hour),
    EndTime:    time.Now().Add(2 * time.Hour),
    Timezone:   "UTC",
    CreatedBy:  "user_123",
    Status:     AppointmentStatusScheduled,
   },
   wantErr: false,
  },
  {
   name: "missing product ID",
   appointment: &Appointment{
    Title:      "Test",
    StartTime:  time.Now(),
    EndTime:    time.Now().Add(1 * time.Hour),
    CreatedBy:  "user_123",
   },
   wantErr:     true,
   errContains: "product ID is required",
  },
  {
   name: "missing title",
   appointment: &Appointment{
    ProductID:  uuid.New(),
    StartTime:  time.Now(),
    EndTime:    time.Now().Add(1 * time.Hour),
    CreatedBy:  "user_123",
   },
   wantErr:     true,
   errContains: "title is required",
  },
  {
   name: "end time before start time",
   appointment: &Appointment{
    ProductID:  uuid.New(),
    Title:      "Test",
    StartTime:  time.Now().Add(2 * time.Hour),
    EndTime:    time.Now().Add(1 * time.Hour),
    CreatedBy:  "user_123",
   },
   wantErr:     true,
   errContains: "end time must be after start time",
  },
 }

 for _, tt := range tests {
  t.Run(tt.name, func(t *testing.T) {
   err := tt.appointment.Validate()
   
   if tt.wantErr {
    require.Error(t, err)
    if tt.errContains != "" {
     assert.Contains(t, err.Error(), tt.errContains)
    }
   } else {
    require.NoError(t, err)
   }
  })
 }
}

func TestAppointment_IsActive(t *testing.T) {
 appointment := &Appointment{
  Status: AppointmentStatusScheduled,
 }
 
 assert.True(t, appointment.IsActive())
 
 appointment.Status = AppointmentStatusCancelled
 assert.False(t, appointment.IsActive())
}

func TestAppointment_Duration(t *testing.T) {
 start := time.Now()
 end := start.Add(1 * time.Hour)
 
 appointment := &Appointment{
  StartTime: start,
  EndTime:   end,
 }
 
 assert.Equal(t, 1*time.Hour, appointment.Duration())
}
```

Location: `appointment-service/internal/models/participant_test.go`

```go
package models

import (
 "testing"

 "github.com/google/uuid"
 "github.com/stretchr/testify/assert"
 "github.com/stretchr/testify/require"
)

func TestAppointmentParticipant_Validate(t *testing.T) {
 tests := []struct {
  name        string
  participant *AppointmentParticipant
  wantErr     bool
  errContains string
 }{
  {
   name: "valid participant",
   participant: &AppointmentParticipant{
    AppointmentID:  uuid.New(),
    ExternalUserID: "user_123",
    Role:           ParticipantRoleHost,
    Status:         ParticipantStatusAccepted,
   },
   wantErr: false,
  },
  {
   name: "missing appointment ID",
   participant: &AppointmentParticipant{
    ExternalUserID: "user_123",
    Role:           ParticipantRoleHost,
   },
   wantErr:     true,
   errContains: "appointment ID is required",
  },
  {
   name: "missing external user ID",
   participant: &AppointmentParticipant{
    AppointmentID: uuid.New(),
    Role:          ParticipantRoleHost,
   },
   wantErr:     true,
   errContains: "external user ID is required",
  },
  {
   name: "invalid role",
   participant: &AppointmentParticipant{
    AppointmentID:  uuid.New(),
    ExternalUserID: "user_123",
    Role:           "invalid_role",
   },
   wantErr:     true,
   errContains: "invalid participant role",
  },
 }

 for _, tt := range tests {
  t.Run(tt.name, func(t *testing.T) {
   err := tt.participant.Validate()
   
   if tt.wantErr {
    require.Error(t, err)
    if tt.errContains != "" {
     assert.Contains(t, err.Error(), tt.errContains)
    }
   } else {
    require.NoError(t, err)
   }
  })
 }
}

func TestAppointmentParticipant_IsHost(t *testing.T) {
 participant := &AppointmentParticipant{
  Role: ParticipantRoleHost,
 }
 
 assert.True(t, participant.IsHost())
 
 participant.Role = ParticipantRoleGuest
 assert.False(t, participant.IsHost())
}
```

---

## Verification Checklist

- [ ] base.go created with BaseModel
- [ ] product.go created with Product model and validation
- [ ] appointment.go created with Appointment model and validation
- [ ] participant.go created with Participant model and validation
- [ ] All models compile without errors: `go build ./...`
- [ ] Tests written for models
- [ ] Tests pass: `make test`
- [ ] Models use proper types (uuid.UUID, time.Time)
- [ ] JSON tags correctly defined
- [ ] Sensitive fields excluded from JSON (marked with `-`)

---

## Testing

```bash
# Build to check for compilation errors
go build ./internal/models/...

# Run tests
go test ./internal/models/... -v

# Run tests with coverage
go test ./internal/models/... -cover

# Full test suite
make test
```

---

## Expected Output

Running tests:

```md
=== RUN   TestAppointment_Validate
=== RUN   TestAppointment_Validate/valid_appointment
=== RUN   TestAppointment_Validate/missing_product_ID
=== RUN   TestAppointment_Validate/missing_title
=== RUN   TestAppointment_Validate/end_time_before_start_time
--- PASS: TestAppointment_Validate (0.00s)
    --- PASS: TestAppointment_Validate/valid_appointment (0.00s)
    --- PASS: TestAppointment_Validate/missing_product_ID (0.00s)
    --- PASS: TestAppointment_Validate/missing_title (0.00s)
    --- PASS: TestAppointment_Validate/end_time_before_start_time (0.00s)
=== RUN   TestAppointment_IsActive
--- PASS: TestAppointment_IsActive (0.00s)
=== RUN   TestAppointment_Duration
--- PASS: TestAppointment_Duration (0.00s)
PASS
ok      github.com/laith-ambianze/appointment-service/internal/models   0.123s
```

---

## Next Task

After completing this task, proceed to [TASK_W02_04_PRODUCT_REPOSITORY.md](TASK_W02_04_PRODUCT_REPOSITORY.md) to implement the repository layer for products.
