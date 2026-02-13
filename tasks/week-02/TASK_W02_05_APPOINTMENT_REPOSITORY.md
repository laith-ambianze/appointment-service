# Task W02-05: Appointment Repository Implementation

**Status**: ⏸️ Not Started  
**Estimated Time**: 3-4 hours  
**Prerequisites**: TASK_W02_04_PRODUCT_REPOSITORY.md  
**Next Task**: Week 03 Tasks

---

## Objective

Implement the appointment repository with complex queries including participant management, date range filtering, and multi-table joins using pgx.

---

## Steps

### 1. Create Appointment Repository Interface

Location: `appointment-service/internal/repository/appointment_repo.go`

```go
package repository

import (
 "context"
 "time"

 "github.com/google/uuid"
 "github.com/laith-ambianze/appointment-service/internal/models"
)

// AppointmentRepository defines the interface for appointment data operations
type AppointmentRepository interface {
 // Create creates a new appointment with participants
 Create(ctx context.Context, appointment *models.Appointment, participants []models.AppointmentParticipant) error
 
 // GetByID retrieves an appointment by ID with participants
 GetByID(ctx context.Context, id uuid.UUID) (*models.Appointment, error)
 
 // GetByProductAndUser retrieves appointments for a user in a product
 GetByProductAndUser(ctx context.Context, productID uuid.UUID, externalUserID string, filters AppointmentFilters) ([]models.Appointment, error)
 
 // GetByDateRange retrieves appointments within a date range
 GetByDateRange(ctx context.Context, productID uuid.UUID, startTime, endTime time.Time) ([]models.Appointment, error)
 
 // Update updates an existing appointment
 Update(ctx context.Context, appointment *models.Appointment) error
 
 // UpdateStatus updates appointment status
 UpdateStatus(ctx context.Context, id uuid.UUID, status models.AppointmentStatus) error
 
 // Delete soft-deletes an appointment
 Delete(ctx context.Context, id uuid.UUID) error
 
 // AddParticipant adds a participant to an appointment
 AddParticipant(ctx context.Context, participant *models.AppointmentParticipant) error
 
 // RemoveParticipant removes a participant from an appointment
 RemoveParticipant(ctx context.Context, appointmentID uuid.UUID, externalUserID string) error
 
 // UpdateParticipantStatus updates a participant's status
 UpdateParticipantStatus(ctx context.Context, appointmentID uuid.UUID, externalUserID string, status models.ParticipantStatus) error
 
 // GetParticipants retrieves all participants for an appointment
 GetParticipants(ctx context.Context, appointmentID uuid.UUID) ([]models.AppointmentParticipant, error)
}

// AppointmentFilters defines filters for listing appointments
type AppointmentFilters struct {
 Status         *models.AppointmentStatus
 StartTimeFrom  *time.Time
 StartTimeTo    *time.Time
 IncludeDeleted bool
 Limit          int
 Offset         int
}
```

### 2. Implement Appointment Repository (Part 1 - Core CRUD)

Location: `appointment-service/internal/repository/appointment_repo_impl.go`

```go
package repository

import (
 "context"
 "database/sql"
 "encoding/json"
 "errors"
 "fmt"
 "strings"
 "time"

 "github.com/google/uuid"
 "github.com/jackc/pgx/v5"
 "github.com/jackc/pgx/v5/pgxpool"
 "github.com/laith-ambianze/appointment-service/internal/models"
 "go.uber.org/zap"
)

// appointmentRepository implements AppointmentRepository interface
type appointmentRepository struct {
 db     *pgxpool.Pool
 logger *zap.Logger
}

// NewAppointmentRepository creates a new appointment repository
func NewAppointmentRepository(db *pgxpool.Pool, logger *zap.Logger) AppointmentRepository {
 return &appointmentRepository{
  db:     db,
  logger: logger,
 }
}

// Create creates a new appointment with participants in a transaction
func (r *appointmentRepository) Create(ctx context.Context, appointment *models.Appointment, participants []models.AppointmentParticipant) error {
 // Validate appointment
 if err := appointment.Validate(); err != nil {
  return fmt.Errorf("%w: %v", ErrInvalidInput, err)
 }

 // Validate participants
 if len(participants) == 0 {
  return fmt.Errorf("%w: at least one participant is required", ErrInvalidInput)
 }

 for i := range participants {
  if err := participants[i].Validate(); err != nil {
   return fmt.Errorf("%w: participant %d: %v", ErrInvalidInput, i, err)
  }
 }

 // Generate ID if not provided
 if appointment.ID == uuid.Nil {
  appointment.ID = uuid.New()
 }

 // Marshal metadata
 metadataJSON, err := appointment.MetadataJSON()
 if err != nil {
  return fmt.Errorf("%w: %v", ErrInvalidInput, err)
 }

 // Start transaction
 tx, err := r.db.Begin(ctx)
 if err != nil {
  return fmt.Errorf("%w: failed to begin transaction: %v", ErrDatabase, err)
 }
 defer tx.Rollback(ctx)

 // Insert appointment
 appointmentQuery := `
  INSERT INTO appointments (
   id, product_id, title, description, start_time, end_time,
   timezone, location, status, created_by, metadata,
   created_at, updated_at
  ) VALUES (
   $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
  )
 `

 now := time.Now()
 appointment.CreatedAt = now
 appointment.UpdatedAt = now

 _, err = tx.Exec(ctx, appointmentQuery,
  appointment.ID,
  appointment.ProductID,
  appointment.Title,
  appointment.Description,
  appointment.StartTime,
  appointment.EndTime,
  appointment.Timezone,
  nullString(appointment.Location),
  appointment.Status,
  appointment.CreatedBy,
  metadataJSON,
  appointment.CreatedAt,
  appointment.UpdatedAt,
 )

 if err != nil {
  return fmt.Errorf("%w: failed to create appointment: %v", ErrDatabase, err)
 }

 // Insert participants
 participantQuery := `
  INSERT INTO appointment_participants (
   id, appointment_id, external_user_id, role, status,
   user_metadata, created_at, updated_at
  ) VALUES (
   $1, $2, $3, $4, $5, $6, $7, $8
  )
 `

 for i := range participants {
  participant := &participants[i]
  
  // Generate ID if not provided
  if participant.ID == uuid.Nil {
   participant.ID = uuid.New()
  }

  participant.AppointmentID = appointment.ID
  participant.CreatedAt = now
  participant.UpdatedAt = now

  userMetadataJSON, err := participant.UserMetadataJSON()
  if err != nil {
   return fmt.Errorf("%w: failed to marshal participant metadata: %v", ErrInvalidInput, err)
  }

  _, err = tx.Exec(ctx, participantQuery,
   participant.ID,
   participant.AppointmentID,
   participant.ExternalUserID,
   participant.Role,
   participant.Status,
   userMetadataJSON,
   participant.CreatedAt,
   participant.UpdatedAt,
  )

  if err != nil {
   return fmt.Errorf("%w: failed to create participant: %v", ErrDatabase, err)
  }
 }

 // Commit transaction
 if err := tx.Commit(ctx); err != nil {
  return fmt.Errorf("%w: failed to commit transaction: %v", ErrDatabase, err)
 }

 appointment.Participants = participants

 r.logger.Info("Appointment created",
  zap.String("appointment_id", appointment.ID.String()),
  zap.String("product_id", appointment.ProductID.String()),
  zap.Int("participants", len(participants)),
 )

 return nil
}

// GetByID retrieves an appointment by ID with participants
func (r *appointmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Appointment, error) {
 query := `
  SELECT 
   id, product_id, title, description, start_time, end_time,
   timezone, location, status, created_by, metadata,
   created_at, updated_at, deleted_at
  FROM appointments
  WHERE id = $1 AND deleted_at IS NULL
 `

 var appointment models.Appointment
 var description, location sql.NullString
 var metadataJSON string

 err := r.db.QueryRow(ctx, query, id).Scan(
  &appointment.ID,
  &appointment.ProductID,
  &appointment.Title,
  &description,
  &appointment.StartTime,
  &appointment.EndTime,
  &appointment.Timezone,
  &location,
  &appointment.Status,
  &appointment.CreatedBy,
  &metadataJSON,
  &appointment.CreatedAt,
  &appointment.UpdatedAt,
  &appointment.DeletedAt,
 )

 if err != nil {
  if errors.Is(err, pgx.ErrNoRows) {
   return nil, ErrNotFound
  }
  return nil, fmt.Errorf("%w: %v", ErrDatabase, err)
 }

 // Set nullable fields
 appointment.Description = description.String
 appointment.Location = location.String

 // Unmarshal metadata
 if metadataJSON != "" && metadataJSON != "{}" {
  if err := json.Unmarshal([]byte(metadataJSON), &appointment.Metadata); err != nil {
   r.logger.Warn("Failed to unmarshal appointment metadata",
    zap.String("appointment_id", id.String()),
    zap.Error(err),
   )
  }
 }

 // Load participants
 participants, err := r.GetParticipants(ctx, id)
 if err != nil {
  return nil, err
 }
 appointment.Participants = participants

 return &appointment, nil
}

// GetByProductAndUser retrieves appointments for a user in a product
func (r *appointmentRepository) GetByProductAndUser(ctx context.Context, productID uuid.UUID, externalUserID string, filters AppointmentFilters) ([]models.Appointment, error) {
 query := `
  SELECT DISTINCT
   a.id, a.product_id, a.title, a.description, a.start_time, a.end_time,
   a.timezone, a.location, a.status, a.created_by, a.metadata,
   a.created_at, a.updated_at, a.deleted_at
  FROM appointments a
  INNER JOIN appointment_participants ap ON a.id = ap.appointment_id
  WHERE a.product_id = $1
    AND ap.external_user_id = $2
 `

 var conditions []string
 args := []interface{}{productID, externalUserID}
 argIndex := 3

 // Apply filters
 if !filters.IncludeDeleted {
  conditions = append(conditions, "a.deleted_at IS NULL")
 }

 if filters.Status != nil {
  conditions = append(conditions, fmt.Sprintf("a.status = $%d", argIndex))
  args = append(args, *filters.Status)
  argIndex++
 }

 if filters.StartTimeFrom != nil {
  conditions = append(conditions, fmt.Sprintf("a.start_time >= $%d", argIndex))
  args = append(args, *filters.StartTimeFrom)
  argIndex++
 }

 if filters.StartTimeTo != nil {
  conditions = append(conditions, fmt.Sprintf("a.start_time <= $%d", argIndex))
  args = append(args, *filters.StartTimeTo)
  argIndex++
 }

 // Add conditions to query
 if len(conditions) > 0 {
  query += " AND " + strings.Join(conditions, " AND ")
 }

 // Add ordering
 query += " ORDER BY a.start_time DESC"

 // Add pagination
 if filters.Limit > 0 {
  query += fmt.Sprintf(" LIMIT $%d", argIndex)
  args = append(args, filters.Limit)
  argIndex++
 }

 if filters.Offset > 0 {
  query += fmt.Sprintf(" OFFSET $%d", argIndex)
  args = append(args, filters.Offset)
 }

 // Execute query
 rows, err := r.db.Query(ctx, query, args...)
 if err != nil {
  return nil, fmt.Errorf("%w: %v", ErrDatabase, err)
 }
 defer rows.Close()

 // Scan results
 appointments := []models.Appointment{}
 for rows.Next() {
  appointment, err := r.scanAppointment(rows)
  if err != nil {
   return nil, err
  }

  // Load participants
  participants, err := r.GetParticipants(ctx, appointment.ID)
  if err != nil {
   r.logger.Warn("Failed to load participants",
    zap.String("appointment_id", appointment.ID.String()),
    zap.Error(err),
   )
  } else {
   appointment.Participants = participants
  }

  appointments = append(appointments, *appointment)
 }

 if err := rows.Err(); err != nil {
  return nil, fmt.Errorf("%w: %v", ErrDatabase, err)
 }

 return appointments, nil
}

// GetByDateRange retrieves appointments within a date range
func (r *appointmentRepository) GetByDateRange(ctx context.Context, productID uuid.UUID, startTime, endTime time.Time) ([]models.Appointment, error) {
 query := `
  SELECT 
   id, product_id, title, description, start_time, end_time,
   timezone, location, status, created_by, metadata,
   created_at, updated_at, deleted_at
  FROM appointments
  WHERE product_id = $1
    AND start_time >= $2
    AND end_time <= $3
    AND deleted_at IS NULL
  ORDER BY start_time ASC
 `

 rows, err := r.db.Query(ctx, query, productID, startTime, endTime)
 if err != nil {
  return nil, fmt.Errorf("%w: %v", ErrDatabase, err)
 }
 defer rows.Close()

 appointments := []models.Appointment{}
 for rows.Next() {
  appointment, err := r.scanAppointment(rows)
  if err != nil {
   return nil, err
  }

  // Load participants
  participants, err := r.GetParticipants(ctx, appointment.ID)
  if err != nil {
   r.logger.Warn("Failed to load participants",
    zap.String("appointment_id", appointment.ID.String()),
    zap.Error(err),
   )
  } else {
   appointment.Participants = participants
  }

  appointments = append(appointments, *appointment)
 }

 if err := rows.Err(); err != nil {
  return nil, fmt.Errorf("%w: %v", ErrDatabase, err)
 }

 return appointments, nil
}

// Update updates an existing appointment
func (r *appointmentRepository) Update(ctx context.Context, appointment *models.Appointment) error {
 // Validate appointment
 if err := appointment.Validate(); err != nil {
  return fmt.Errorf("%w: %v", ErrInvalidInput, err)
 }

 // Marshal metadata
 metadataJSON, err := appointment.MetadataJSON()
 if err != nil {
  return fmt.Errorf("%w: %v", ErrInvalidInput, err)
 }

 query := `
  UPDATE appointments SET
   title = $1,
   description = $2,
   start_time = $3,
   end_time = $4,
   timezone = $5,
   location = $6,
   status = $7,
   metadata = $8,
   updated_at = $9
  WHERE id = $10 AND deleted_at IS NULL
 `

 appointment.UpdatedAt = time.Now()

 result, err := r.db.Exec(ctx, query,
  appointment.Title,
  appointment.Description,
  appointment.StartTime,
  appointment.EndTime,
  appointment.Timezone,
  nullString(appointment.Location),
  appointment.Status,
  metadataJSON,
  appointment.UpdatedAt,
  appointment.ID,
 )

 if err != nil {
  return fmt.Errorf("%w: %v", ErrDatabase, err)
 }

 if result.RowsAffected() == 0 {
  return ErrNotFound
 }

 r.logger.Info("Appointment updated",
  zap.String("appointment_id", appointment.ID.String()),
 )

 return nil
}

// UpdateStatus updates appointment status
func (r *appointmentRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.AppointmentStatus) error {
 query := `
  UPDATE appointments 
  SET status = $1, updated_at = $2
  WHERE id = $3 AND deleted_at IS NULL
 `

 result, err := r.db.Exec(ctx, query, status, time.Now(), id)
 if err != nil {
  return fmt.Errorf("%w: %v", ErrDatabase, err)
 }

 if result.RowsAffected() == 0 {
  return ErrNotFound
 }

 r.logger.Info("Appointment status updated",
  zap.String("appointment_id", id.String()),
  zap.String("status", string(status)),
 )

 return nil
}

// Delete soft-deletes an appointment
func (r *appointmentRepository) Delete(ctx context.Context, id uuid.UUID) error {
 query := `
  UPDATE appointments 
  SET deleted_at = $1, updated_at = $2
  WHERE id = $3 AND deleted_at IS NULL
 `

 now := time.Now()
 result, err := r.db.Exec(ctx, query, now, now, id)
 if err != nil {
  return fmt.Errorf("%w: %v", ErrDatabase, err)
 }

 if result.RowsAffected() == 0 {
  return ErrNotFound
 }

 r.logger.Info("Appointment soft-deleted",
  zap.String("appointment_id", id.String()),
 )

 return nil
}

// Participant operations

// AddParticipant adds a participant to an appointment
func (r *appointmentRepository) AddParticipant(ctx context.Context, participant *models.AppointmentParticipant) error {
 if err := participant.Validate(); err != nil {
  return fmt.Errorf("%w: %v", ErrInvalidInput, err)
 }

 if participant.ID == uuid.Nil {
  participant.ID = uuid.New()
 }

 userMetadataJSON, err := participant.UserMetadataJSON()
 if err != nil {
  return fmt.Errorf("%w: %v", ErrInvalidInput, err)
 }

 query := `
  INSERT INTO appointment_participants (
   id, appointment_id, external_user_id, role, status,
   user_metadata, created_at, updated_at
  ) VALUES (
   $1, $2, $3, $4, $5, $6, $7, $8
  )
 `

 now := time.Now()
 participant.CreatedAt = now
 participant.UpdatedAt = now

 _, err = r.db.Exec(ctx, query,
  participant.ID,
  participant.AppointmentID,
  participant.ExternalUserID,
  participant.Role,
  participant.Status,
  userMetadataJSON,
  participant.CreatedAt,
  participant.UpdatedAt,
 )

 if err != nil {
  return fmt.Errorf("%w: %v", ErrDatabase, err)
 }

 return nil
}

// RemoveParticipant removes a participant from an appointment
func (r *appointmentRepository) RemoveParticipant(ctx context.Context, appointmentID uuid.UUID, externalUserID string) error {
 query := `
  DELETE FROM appointment_participants
  WHERE appointment_id = $1 AND external_user_id = $2
 `

 result, err := r.db.Exec(ctx, query, appointmentID, externalUserID)
 if err != nil {
  return fmt.Errorf("%w: %v", ErrDatabase, err)
 }

 if result.RowsAffected() == 0 {
  return ErrNotFound
 }

 return nil
}

// UpdateParticipantStatus updates a participant's status
func (r *appointmentRepository) UpdateParticipantStatus(ctx context.Context, appointmentID uuid.UUID, externalUserID string, status models.ParticipantStatus) error {
 query := `
  UPDATE appointment_participants
  SET status = $1, updated_at = $2
  WHERE appointment_id = $3 AND external_user_id = $4
 `

 result, err := r.db.Exec(ctx, query, status, time.Now(), appointmentID, externalUserID)
 if err != nil {
  return fmt.Errorf("%w: %v", ErrDatabase, err)
 }

 if result.RowsAffected() == 0 {
  return ErrNotFound
 }

 return nil
}

// GetParticipants retrieves all participants for an appointment
func (r *appointmentRepository) GetParticipants(ctx context.Context, appointmentID uuid.UUID) ([]models.AppointmentParticipant, error) {
 query := `
  SELECT 
   id, appointment_id, external_user_id, role, status,
   user_metadata, created_at, updated_at
  FROM appointment_participants
  WHERE appointment_id = $1
  ORDER BY CASE role 
   WHEN 'host' THEN 1
   WHEN 'guest' THEN 2
   WHEN 'attendee' THEN 3
   ELSE 4
  END, created_at ASC
 `

 rows, err := r.db.Query(ctx, query, appointmentID)
 if err != nil {
  return nil, fmt.Errorf("%w: %v", ErrDatabase, err)
 }
 defer rows.Close()

 participants := []models.AppointmentParticipant{}
 for rows.Next() {
  var participant models.AppointmentParticipant
  var userMetadataJSON string

  err := rows.Scan(
   &participant.ID,
   &participant.AppointmentID,
   &participant.ExternalUserID,
   &participant.Role,
   &participant.Status,
   &userMetadataJSON,
   &participant.CreatedAt,
   &participant.UpdatedAt,
  )

  if err != nil {
   return nil, fmt.Errorf("%w: %v", ErrDatabase, err)
  }

  // Unmarshal user metadata
  if userMetadataJSON != "" && userMetadataJSON != "{}" {
   if err := json.Unmarshal([]byte(userMetadataJSON), &participant.UserMetadata); err != nil {
    r.logger.Warn("Failed to unmarshal participant user metadata",
     zap.String("participant_id", participant.ID.String()),
     zap.Error(err),
    )
   }
  }

  participants = append(participants, participant)
 }

 if err := rows.Err(); err != nil {
  return nil, fmt.Errorf("%w: %v", ErrDatabase, err)
 }

 return participants, nil
}

// Helper function to scan appointment from rows
func (r *appointmentRepository) scanAppointment(rows pgx.Rows) (*models.Appointment, error) {
 var appointment models.Appointment
 var description, location sql.NullString
 var metadataJSON string

 err := rows.Scan(
  &appointment.ID,
  &appointment.ProductID,
  &appointment.Title,
  &description,
  &appointment.StartTime,
  &appointment.EndTime,
  &appointment.Timezone,
  &location,
  &appointment.Status,
  &appointment.CreatedBy,
  &metadataJSON,
  &appointment.CreatedAt,
  &appointment.UpdatedAt,
  &appointment.DeletedAt,
 )

 if err != nil {
  return nil, fmt.Errorf("%w: %v", ErrDatabase, err)
 }

 // Set nullable fields
 appointment.Description = description.String
 appointment.Location = location.String

 // Unmarshal metadata
 if metadataJSON != "" && metadataJSON != "{}" {
  if err := json.Unmarshal([]byte(metadataJSON), &appointment.Metadata); err != nil {
   r.logger.Warn("Failed to unmarshal appointment metadata",
    zap.String("appointment_id", appointment.ID.String()),
    zap.Error(err),
   )
  }
 }

 return &appointment, nil
}
```

### 3. Create Integration Tests

Location: `appointment-service/tests/integration/appointment_repo_test.go`

```go
package integration

import (
 "context"
 "testing"
 "time"

 "github.com/google/uuid"
 "github.com/laith-ambianze/appointment-service/internal/models"
 "github.com/laith-ambianze/appointment-service/internal/repository"
 "github.com/laith-ambianze/appointment-service/pkg/database"
 "github.com/stretchr/testify/assert"
 "github.com/stretchr/testify/require"
 "go.uber.org/zap"
 "golang.org/x/crypto/bcrypt"
)

func TestAppointmentRepository_CreateAndGet(t *testing.T) {
 if testing.Short() {
  t.Skip("Skipping integration test")
 }

 // Setup
 db := setupTestDB(t)
 defer db.Close()

 productRepo := repository.NewProductRepository(db.Pool, zap.NewNop())
 appointmentRepo := repository.NewAppointmentRepository(db.Pool, zap.NewNop())
 ctx := context.Background()

 // Create test product
 product := createTestProduct(t, productRepo, ctx)
 defer productRepo.HardDelete(ctx, product.ID)

 // Create test appointment
 appointment := &models.Appointment{
  ProductID:   product.ID,
  Title:       "Test Appointment",
  Description: "A test appointment",
  StartTime:   time.Now().Add(24 * time.Hour),
  EndTime:     time.Now().Add(25 * time.Hour),
  Timezone:    "UTC",
  Location:    "Test Location",
  Status:      models.AppointmentStatusScheduled,
  CreatedBy:   "user_test_123",
  Metadata: map[string]interface{}{
   "test": true,
  },
 }

 participants := []models.AppointmentParticipant{
  {
   ExternalUserID: "user_test_123",
   Role:           models.ParticipantRoleHost,
   Status:         models.ParticipantStatusAccepted,
   UserMetadata: map[string]interface{}{
    "name":  "Test Host",
    "email": "host@test.com",
   },
  },
  {
   ExternalUserID: "user_test_456",
   Role:           models.ParticipantRoleGuest,
   Status:         models.ParticipantStatusPending,
   UserMetadata: map[string]interface{}{
    "name":  "Test Guest",
    "email": "guest@test.com",
   },
  },
 }

 // Test create
 err := appointmentRepo.Create(ctx, appointment, participants)
 require.NoError(t, err)
 assert.NotEqual(t, uuid.Nil, appointment.ID)

 // Test get by ID
 retrieved, err := appointmentRepo.GetByID(ctx, appointment.ID)
 require.NoError(t, err)
 require.NotNil(t, retrieved)
 assert.Equal(t, appointment.Title, retrieved.Title)
 assert.Len(t, retrieved.Participants, 2)
}

func TestAppointmentRepository_GetByProductAndUser(t *testing.T) {
 if testing.Short() {
  t.Skip("Skipping integration test")
 }

 // Setup
 db := setupTestDB(t)
 defer db.Close()

 productRepo := repository.NewProductRepository(db.Pool, zap.NewNop())
 appointmentRepo := repository.NewAppointmentRepository(db.Pool, zap.NewNop())
 ctx := context.Background()

 // Create test product
 product := createTestProduct(t, productRepo, ctx)
 defer productRepo.HardDelete(ctx, product.ID)

 // Create test appointments
 appointment1 := createTestAppointment(t, appointmentRepo, ctx, product.ID, "user_123")
 appointment2 := createTestAppointment(t, appointmentRepo, ctx, product.ID, "user_123")

 // Test
 filters := repository.AppointmentFilters{}
 appointments, err := appointmentRepo.GetByProductAndUser(ctx, product.ID, "user_123", filters)
 
 require.NoError(t, err)
 assert.GreaterOrEqual(t, len(appointments), 2)

 // Verify appointments are for the correct user
 for _, appt := range appointments {
  hasUser := false
  for _, p := range appt.Participants {
   if p.ExternalUserID == "user_123" {
    hasUser = true
    break
   }
  }
  assert.True(t, hasUser, "Appointment should have user_123 as participant")
 }

 // Cleanup
 _ = appointmentRepo.Delete(ctx, appointment1.ID)
 _ = appointmentRepo.Delete(ctx, appointment2.ID)
}

// Helper functions
func setupTestDB(t *testing.T) *database.PostgresDB {
 cfg := database.Config{
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

 db, err := database.NewPostgresDB(cfg, zap.NewNop())
 require.NoError(t, err)

 return db
}

func createTestProduct(t *testing.T, repo repository.ProductRepository, ctx context.Context) *models.Product {
 apiSecretHash, _ := bcrypt.GenerateFromPassword([]byte("secret"), 10)
 product := &models.Product{
  Name:          "Test Product " + uuid.NewString()[:8],
  APIKey:        "test_key_" + uuid.NewString()[:8],
  APISecretHash: string(apiSecretHash),
  Status:        models.ProductStatusActive,
 }

 err := repo.Create(ctx, product)
 require.NoError(t, err)

 return product
}

func createTestAppointment(t *testing.T, repo repository.AppointmentRepository, ctx context.Context, productID uuid.UUID, userID string) *models.Appointment {
 appointment := &models.Appointment{
  ProductID:   productID,
  Title:       "Test " + uuid.NewString()[:8],
  StartTime:   time.Now().Add(24 * time.Hour),
  EndTime:     time.Now().Add(25 * time.Hour),
  Timezone:    "UTC",
  Status:      models.AppointmentStatusScheduled,
  CreatedBy:   userID,
 }

 participants := []models.AppointmentParticipant{
  {
   ExternalUserID: userID,
   Role:           models.ParticipantRoleHost,
   Status:         models.ParticipantStatusAccepted,
  },
 }

 err := repo.Create(ctx, appointment, participants)
 require.NoError(t, err)

 return appointment
}
```

---

## Verification Checklist

- [ ] Appointment repository interface defined
- [ ] All methods implemented
- [ ] Transaction support for creating appointments with participants
- [ ] Complex queries implemented (joins, date ranges, filters)
- [ ] Participant operations working
- [ ] Integration tests written and passing
- [ ] Code compiles: `go build ./...`
- [ ] Tests pass: `make test`

---

## Testing

```bash
# Start database
make db-start
sleep 5

# Run migrations
make migrate-up

# Run all tests
make test

# Run integration tests only
go test ./tests/integration/... -v

# Run with coverage
go test ./... -cover

# Test specific scenarios
go test ./internal/repository/... -run TestAppointment -v
```

---

## Expected Output

Running integration tests:

```md
=== RUN   TestAppointmentRepository_CreateAndGet
--- PASS: TestAppointmentRepository_CreateAndGet (0.05s)
=== RUN   TestAppointmentRepository_GetByProductAndUser
--- PASS: TestAppointmentRepository_GetByProductAndUser (0.08s)
PASS
ok      github.com/laith-ambianze/appointment-service/tests/integration  0.150s
```

---

## Week 02 Summary

Congratulations! You've completed Week 02. You now have:

✅ **Database Infrastructure**

- PostgreSQL connection pool with pgx
- Migration system with 3 tables
- Database health checks
- Transaction support

✅ **Domain Models**

- Product, Appointment, Participant models
- Validation logic
- JSON marshaling/unmarshaling
- Business logic methods

✅ **Repository Layer**

- Product repository with CRUD operations
- Appointment repository with complex queries
- Participant management
- Proper error handling

✅ **Testing**

- Unit tests for models
- Integration tests for repositories
- Test fixtures and helpers

---

## Next Steps: Week 03

Week 03 will focus on:

1. **Product Management API** - REST endpoints for product operations
2. **API Key Authentication** - Middleware for product authentication
3. **Request Validation** - Input validation and error handling
4. **API Documentation** - Swagger/OpenAPI specs

---

## Congratulations! 🎉

Week 02 is complete. Your appointment service now has a solid data layer with proper models and repositories.
