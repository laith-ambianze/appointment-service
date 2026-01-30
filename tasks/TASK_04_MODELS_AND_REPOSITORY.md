# Task 04: Models and Repository Layer

**Priority**: High  
**Estimated Time**: 4 hours  
**Dependencies**: TASK_03  
**Status**: Not Started

---

## Objective

Create domain models and implement the repository pattern for database operations.

---

## Prerequisites

- [ ] Task 03 completed
- [ ] Database running and migrated

---

## Steps

### 1. Create Models

**File**: `internal/models/product.go`

```go
package models

import (
 "time"
 "github.com/google/uuid"
)

type Product struct {
 ID            uuid.UUID  `json:"id"`
 Name          string     `json:"name"`
 Description   string     `json:"description"`
 APIKey        string     `json:"apiKey"`
 APISecretHash string     `json:"-"` // Never expose in JSON
 CallbackURL   string     `json:"callbackUrl,omitempty"`
 IsActive      bool       `json:"isActive"`
 CreatedAt     time.Time  `json:"createdAt"`
 UpdatedAt     time.Time  `json:"updatedAt"`
}

type CreateProductRequest struct {
 Name        string `json:"name" validate:"required,min=3,max=255"`
 Description string `json:"description" validate:"max=1000"`
 CallbackURL string `json:"callbackUrl" validate:"omitempty,url"`
}
```

**File**: `internal/models/appointment.go`

```go
package models

import (
 "encoding/json"
 "time"
 "github.com/google/uuid"
)

type Appointment struct {
 ID                 uuid.UUID       `json:"id"`
 ProductID          uuid.UUID       `json:"productId"`
 CreatedBy          string          `json:"createdBy"`
 Title              string          `json:"title"`
 Description        string          `json:"description"`
 StartTime          time.Time       `json:"startTime"`
 EndTime            time.Time       `json:"endTime"`
 Location           string          `json:"location"`
 MeetingType        string          `json:"meetingType"`
 Status             string          `json:"status"`
 CancelledBy        *string         `json:"cancelledBy,omitempty"`
 CancellationReason *string         `json:"cancellationReason,omitempty"`
 CancelledAt        *time.Time      `json:"cancelledAt,omitempty"`
 Metadata           json.RawMessage `json:"metadata,omitempty"`
 Participants       []Participant   `json:"participants,omitempty"`
 CreatedAt          time.Time       `json:"createdAt"`
 UpdatedAt          time.Time       `json:"updatedAt"`
}

type Participant struct {
 ID             uuid.UUID       `json:"id"`
 AppointmentID  uuid.UUID       `json:"appointmentId"`
 ExternalUserID string          `json:"externalUserId"`
 Role           string          `json:"role"`
 UserMetadata   json.RawMessage `json:"userMetadata"`
 Status         string          `json:"status"`
 CreatedAt      time.Time       `json:"createdAt"`
 UpdatedAt      time.Time       `json:"updatedAt"`
}

type UserMetadata struct {
 UserID       string                 `json:"userId" validate:"required"`
 FirstName    string                 `json:"firstName" validate:"required"`
 LastName     string                 `json:"lastName" validate:"required"`
 Email        string                 `json:"email" validate:"omitempty,email"`
 Phone        string                 `json:"phone"`
 Timezone     string                 `json:"timezone"`
 CustomFields map[string]interface{} `json:"customFields,omitempty"`
}

type ParticipantRequest struct {
 ExternalUserID string                 `json:"externalUserId" validate:"required"`
 Role           string                 `json:"role" validate:"required,oneof=host guest attendee observer"`
 Metadata       map[string]interface{} `json:"metadata" validate:"required"`
}

type CreateAppointmentRequest struct {
 Title        string               `json:"title" validate:"required,min=3,max=500"`
 Description  string               `json:"description" validate:"max=2000"`
 StartTime    time.Time            `json:"startTime" validate:"required"`
 EndTime      time.Time            `json:"endTime" validate:"required"`
 Location     string               `json:"location" validate:"max=500"`
 MeetingType  string               `json:"meetingType" validate:"omitempty,oneof=online in-person phone"`
 Participants []ParticipantRequest `json:"participants" validate:"required,min=2,dive"`
 Metadata     json.RawMessage      `json:"metadata,omitempty"`
}

type UpdateAppointmentRequest struct {
 Title       *string    `json:"title" validate:"omitempty,min=3,max=500"`
 Description *string    `json:"description" validate:"omitempty,max=2000"`
 StartTime   *time.Time `json:"startTime"`
 EndTime     *time.Time `json:"endTime"`
 Location    *string    `json:"location" validate:"omitempty,max=500"`
 MeetingType *string    `json:"meetingType" validate:"omitempty,oneof=online in-person phone"`
}

type CancelAppointmentRequest struct {
 CancelledBy string `json:"cancelledBy" validate:"required"`
 Reason      string `json:"reason" validate:"required,min=5,max=500"`
}

const (
 StatusPending   = "pending"
 StatusConfirmed = "confirmed"
 StatusCancelled = "cancelled"
 StatusCompleted = "completed"
)

const (
 RoleHost     = "host"
 RoleGuest    = "guest"
 RoleAttendee = "attendee"
 RoleObserver = "observer"
)

const (
 ParticipantStatusPending  = "pending"
 ParticipantStatusAccepted = "accepted"
 ParticipantStatusDeclined = "declined"
)

const (
 MeetingTypeOnline    = "online"
 MeetingTypeInPerson  = "in-person"
 MeetingTypePhone     = "phone"
)
```

**File**: `internal/models/history.go`

```go
package models

import (
 "encoding/json"
 "time"
 "github.com/google/uuid"
)

type AppointmentHistory struct {
 ID            uuid.UUID       `json:"id"`
 AppointmentID uuid.UUID       `json:"appointmentId"`
 Action        string          `json:"action"`
 ChangedBy     *string         `json:"changedBy,omitempty"`
 Changes       json.RawMessage `json:"changes,omitempty"`
 CreatedAt     time.Time       `json:"createdAt"`
}

const (
 ActionCreated   = "created"
 ActionUpdated   = "updated"
 ActionCancelled = "cancelled"
 ActionConfirmed = "confirmed"
 ActionCompleted = "completed"
)
```

### 2. Create Product Repository

**File**: `internal/repository/product_repository.go`

```go
package repository

import (
 "context"
 "fmt"

 "appointment-service/internal/models"
 "github.com/google/uuid"
 "github.com/jackc/pgx/v5"
 "github.com/jackc/pgx/v5/pgxpool"
)

type ProductRepository interface {
 Create(ctx context.Context, product *models.Product) error
 GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error)
 GetByAPIKey(ctx context.Context, apiKey string) (*models.Product, error)
 Update(ctx context.Context, product *models.Product) error
 Delete(ctx context.Context, id uuid.UUID) error
 List(ctx context.Context, limit, offset int) ([]*models.Product, error)
}

type productRepository struct {
 pool *pgxpool.Pool
}

func NewProductRepository(pool *pgxpool.Pool) ProductRepository {
 return &productRepository{pool: pool}
}

func (r *productRepository) Create(ctx context.Context, product *models.Product) error {
 query := `
  INSERT INTO products (id, name, description, api_key, api_secret_hash, callback_url, is_active)
  VALUES ($1, $2, $3, $4, $5, $6, $7)
  RETURNING created_at, updated_at
 `
 return r.pool.QueryRow(ctx, query,
  product.ID,
  product.Name,
  product.Description,
  product.APIKey,
  product.APISecretHash,
  product.CallbackURL,
  product.IsActive,
 ).Scan(&product.CreatedAt, &product.UpdatedAt)
}

func (r *productRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
 query := `
  SELECT id, name, description, api_key, api_secret_hash, callback_url, is_active, created_at, updated_at
  FROM products
  WHERE id = $1
 `
 
 product := &models.Product{}
 err := r.pool.QueryRow(ctx, query, id).Scan(
  &product.ID,
  &product.Name,
  &product.Description,
  &product.APIKey,
  &product.APISecretHash,
  &product.CallbackURL,
  &product.IsActive,
  &product.CreatedAt,
  &product.UpdatedAt,
 )
 
 if err == pgx.ErrNoRows {
  return nil, fmt.Errorf("product not found")
 }
 if err != nil {
  return nil, fmt.Errorf("failed to get product: %w", err)
 }
 
 return product, nil
}

func (r *productRepository) GetByAPIKey(ctx context.Context, apiKey string) (*models.Product, error) {
 query := `
  SELECT id, name, description, api_key, api_secret_hash, callback_url, is_active, created_at, updated_at
  FROM products
  WHERE api_key = $1
 `
 
 product := &models.Product{}
 err := r.pool.QueryRow(ctx, query, apiKey).Scan(
  &product.ID,
  &product.Name,
  &product.Description,
  &product.APIKey,
  &product.APISecretHash,
  &product.CallbackURL,
  &product.IsActive,
  &product.CreatedAt,
  &product.UpdatedAt,
 )
 
 if err == pgx.ErrNoRows {
  return nil, fmt.Errorf("product not found")
 }
 if err != nil {
  return nil, fmt.Errorf("failed to get product: %w", err)
 }
 
 return product, nil
}

func (r *productRepository) Update(ctx context.Context, product *models.Product) error {
 query := `
  UPDATE products
  SET name = $2, description = $3, callback_url = $4, is_active = $5
  WHERE id = $1
  RETURNING updated_at
 `
 return r.pool.QueryRow(ctx, query,
  product.ID,
  product.Name,
  product.Description,
  product.CallbackURL,
  product.IsActive,
 ).Scan(&product.UpdatedAt)
}

func (r *productRepository) Delete(ctx context.Context, id uuid.UUID) error {
 query := `DELETE FROM products WHERE id = $1`
 result, err := r.pool.Exec(ctx, query, id)
 if err != nil {
  return fmt.Errorf("failed to delete product: %w", err)
 }
 
 if result.RowsAffected() == 0 {
  return fmt.Errorf("product not found")
 }
 
 return nil
}

func (r *productRepository) List(ctx context.Context, limit, offset int) ([]*models.Product, error) {
 query := `
  SELECT id, name, description, api_key, api_secret_hash, callback_url, is_active, created_at, updated_at
  FROM products
  ORDER BY created_at DESC
  LIMIT $1 OFFSET $2
 `
 
 rows, err := r.pool.Query(ctx, query, limit, offset)
 if err != nil {
  return nil, fmt.Errorf("failed to list products: %w", err)
 }
 defer rows.Close()
 
 var products []*models.Product
 for rows.Next() {
  product := &models.Product{}
  err := rows.Scan(
   &product.ID,
   &product.Name,
   &product.Description,
   &product.APIKey,
   &product.APISecretHash,
   &product.CallbackURL,
   &product.IsActive,
   &product.CreatedAt,
   &product.UpdatedAt,
  )
  if err != nil {
   return nil, fmt.Errorf("failed to scan product: %w", err)
  }
  products = append(products, product)
 }
 
 return products, rows.Err()
}
```

### 3. Create Appointment Repository (Simplified for now)

**File**: `internal/repository/appointment_repository.go`

```go
package repository

import (
 "context"
 "fmt"

 "appointment-service/internal/models"
 "github.com/google/uuid"
 "github.com/jackc/pgx/v5"
 "github.com/jackc/pgx/v5/pgxpool"
)

type AppointmentRepository interface {
 Create(ctx context.Context, appointment *models.Appointment, participants []models.Participant) error
 GetByID(ctx context.Context, id uuid.UUID) (*models.Appointment, error)
 GetByProductID(ctx context.Context, productID uuid.UUID, limit, offset int) ([]*models.Appointment, error)
 GetByUserID(ctx context.Context, productID uuid.UUID, userID string, limit, offset int) ([]*models.Appointment, error)
 GetParticipants(ctx context.Context, appointmentID uuid.UUID) ([]models.Participant, error)
 Update(ctx context.Context, appointment *models.Appointment) error
 Delete(ctx context.Context, id uuid.UUID) error
}

type appointmentRepository struct {
 pool *pgxpool.Pool
}

func NewAppointmentRepository(pool *pgxpool.Pool) AppointmentRepository {
 return &appointmentRepository{pool: pool}
}

func (r *appointmentRepository) Create(ctx context.Context, appointment *models.Appointment, participants []models.Participant) error {
 // Start transaction
 tx, err := r.pool.Begin(ctx)
 if err != nil {
  return fmt.Errorf("failed to begin transaction: %w", err)
 }
 defer tx.Rollback(ctx)

 // Insert appointment
 appointmentQuery := `
  INSERT INTO appointments (
   id, product_id, created_by, title, description, 
   start_time, end_time, location, meeting_type, status, metadata
  )
  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
  RETURNING created_at, updated_at
 `
 err = tx.QueryRow(ctx, appointmentQuery,
  appointment.ID,
  appointment.ProductID,
  appointment.CreatedBy,
  appointment.Title,
  appointment.Description,
  appointment.StartTime,
  appointment.EndTime,
  appointment.Location,
  appointment.MeetingType,
  appointment.Status,
  appointment.Metadata,
 ).Scan(&appointment.CreatedAt, &appointment.UpdatedAt)
 if err != nil {
  return fmt.Errorf("failed to create appointment: %w", err)
 }

 // Insert participants
 participantQuery := `
  INSERT INTO appointment_participants (
   id, appointment_id, external_user_id, role, user_metadata, status
  )
  VALUES ($1, $2, $3, $4, $5, $6)
  RETURNING created_at, updated_at
 `
 for i := range participants {
  err = tx.QueryRow(ctx, participantQuery,
   participants[i].ID,
   appointment.ID,
   participants[i].ExternalUserID,
   participants[i].Role,
   participants[i].UserMetadata,
   participants[i].Status,
  ).Scan(&participants[i].CreatedAt, &participants[i].UpdatedAt)
  if err != nil {
   return fmt.Errorf("failed to create participant: %w", err)
  }
 }

 // Commit transaction
 if err = tx.Commit(ctx); err != nil {
  return fmt.Errorf("failed to commit transaction: %w", err)
 }

 appointment.Participants = participants
 return nil
}

func (r *appointmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Appointment, error) {
 query := `
  SELECT id, product_id, created_by, title, description, 
         start_time, end_time, location, meeting_type,
         status, cancelled_by, cancellation_reason, cancelled_at,
         metadata, created_at, updated_at
  FROM appointments
  WHERE id = $1
 `
 
 appointment := &models.Appointment{}
 err := r.pool.QueryRow(ctx, query, id).Scan(
  &appointment.ID,
  &appointment.ProductID,
  &appointment.CreatedBy,
  &appointment.Title,
  &appointment.Description,
  &appointment.StartTime,
  &appointment.EndTime,
  &appointment.Location,
  &appointment.MeetingType,
  &appointment.Status,
  &appointment.CancelledBy,
  &appointment.CancellationReason,
  &appointment.CancelledAt,
  &appointment.Metadata,
  &appointment.CreatedAt,
  &appointment.UpdatedAt,
 )
 
 if err == pgx.ErrNoRows {
  return nil, fmt.Errorf("appointment not found")
 }
 if err != nil {
  return nil, fmt.Errorf("failed to get appointment: %w", err)
 }

 // Load participants
 participants, err := r.GetParticipants(ctx, id)
 if err != nil {
  return nil, fmt.Errorf("failed to get participants: %w", err)
 }
 appointment.Participants = participants
 
 return appointment, nil
}

func (r *appointmentRepository) GetByProductID(ctx context.Context, productID uuid.UUID, limit, offset int) ([]*models.Appointment, error) {
 query := `
  SELECT id, product_id, created_by, title, description, 
         start_time, end_time, location, meeting_type,
         status, cancelled_by, cancellation_reason, cancelled_at,
         metadata, created_at, updated_at
  FROM appointments
  WHERE product_id = $1
  ORDER BY start_time DESC
  LIMIT $2 OFFSET $3
 `
 
 rows, err := r.pool.Query(ctx, query, productID, limit, offset)
 if err != nil {
  return nil, fmt.Errorf("failed to query appointments: %w", err)
 }
 defer rows.Close()
 
 return scanAppointments(rows)
}

func (r *appointmentRepository) GetByUserID(ctx context.Context, productID uuid.UUID, userID string, limit, offset int) ([]*models.Appointment, error) {
 query := `
  SELECT DISTINCT a.id, a.product_id, a.created_by, a.title, a.description,
         a.start_time, a.end_time, a.location, a.meeting_type,
         a.status, a.cancelled_by, a.cancellation_reason, a.cancelled_at,
         a.metadata, a.created_at, a.updated_at
  FROM appointments a
  INNER JOIN appointment_participants ap ON a.id = ap.appointment_id
  WHERE a.product_id = $1 AND ap.external_user_id = $2
  ORDER BY a.start_time DESC
  LIMIT $3 OFFSET $4
 `
 
 rows, err := r.pool.Query(ctx, query, productID, userID, limit, offset)
 if err != nil {
  return nil, fmt.Errorf("failed to query appointments: %w", err)
 }
 defer rows.Close()
 
 return scanAppointments(rows)
}

func (r *appointmentRepository) GetParticipants(ctx context.Context, appointmentID uuid.UUID) ([]models.Participant, error) {
 query := `
  SELECT id, appointment_id, external_user_id, role, user_metadata, status, created_at, updated_at
  FROM appointment_participants
  WHERE appointment_id = $1
  ORDER BY CASE role 
   WHEN 'host' THEN 1 
   WHEN 'guest' THEN 2 
   WHEN 'attendee' THEN 3 
   ELSE 4 
  END
 `
 
 rows, err := r.pool.Query(ctx, query, appointmentID)
 if err != nil {
  return nil, fmt.Errorf("failed to query participants: %w", err)
 }
 defer rows.Close()
 
 var participants []models.Participant
 for rows.Next() {
  var p models.Participant
  err := rows.Scan(
   &p.ID,
   &p.AppointmentID,
   &p.ExternalUserID,
   &p.Role,
   &p.UserMetadata,
   &p.Status,
   &p.CreatedAt,
   &p.UpdatedAt,
  )
  if err != nil {
   return nil, fmt.Errorf("failed to scan participant: %w", err)
  }
  participants = append(participants, p)
 }
 
 return participants, rows.Err()
}

func (r *appointmentRepository) Update(ctx context.Context, appointment *models.Appointment) error {
 query := `
  UPDATE appointments
  SET title = $2, description = $3, start_time = $4, end_time = $5,
      location = $6, meeting_type = $7, status = $8,
      cancelled_by = $9, cancellation_reason = $10, cancelled_at = $11
  WHERE id = $1
  RETURNING updated_at
 `
 return r.pool.QueryRow(ctx, query,
  appointment.ID,
  appointment.Title,
  appointment.Description,
  appointment.StartTime,
  appointment.EndTime,
  appointment.Location,
  appointment.MeetingType,
  appointment.Status,
  appointment.CancelledBy,
  appointment.CancellationReason,
  appointment.CancelledAt,
 ).Scan(&appointment.UpdatedAt)
}

func (r *appointmentRepository) Delete(ctx context.Context, id uuid.UUID) error {
 query := `DELETE FROM appointments WHERE id = $1`
 result, err := r.pool.Exec(ctx, query, id)
 if err != nil {
  return fmt.Errorf("failed to delete appointment: %w", err)
 }
 
 if result.RowsAffected() == 0 {
  return fmt.Errorf("appointment not found")
 }
 
 return nil
}

func scanAppointments(rows pgx.Rows) ([]*models.Appointment, error) {
 var appointments []*models.Appointment
 for rows.Next() {
  appointment := &models.Appointment{}
  err := rows.Scan(
   &appointment.ID,
   &appointment.ProductID,
   &appointment.CreatedBy,
   &appointment.Title,
   &appointment.Description,
   &appointment.StartTime,
   &appointment.EndTime,
   &appointment.Location,
   &appointment.MeetingType,
   &appointment.Status,
   &appointment.CancelledBy,
   &appointment.CancellationReason,
   &appointment.CancelledAt,
   &appointment.Metadata,
   &appointment.CreatedAt,
   &appointment.UpdatedAt,
  )
  if err != nil {
   return nil, fmt.Errorf("failed to scan appointment: %w", err)
  }
  appointments = append(appointments, appointment)
 }
 
 return appointments, rows.Err()
}
```

---

## Acceptance Criteria

- [ ] All model structs created
- [ ] Product repository implemented
- [ ] Appointment repository implemented
- [ ] All CRUD operations defined
- [ ] Code compiles without errors
- [ ] Repository interfaces defined

---

## Verification

```bash
# Build to check for errors
go build ./...

# Run go vet
go vet ./...

# Check for issues
golangci-lint run ./...
```

---

## Next Task

[TASK_05_SERVICE_LAYER.md](TASK_05_SERVICE_LAYER.md)

---

## Notes

- Repository pattern provides abstraction over data access
- Use interfaces for easier testing and mocking
- Consider connection pooling settings for production
- Add logging to repository methods later
