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

	productRepo := repository.NewProductRepository(db.GetPool(), zap.NewNop())
	appointmentRepo := repository.NewAppointmentRepository(db.GetPool(), zap.NewNop())
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

	productRepo := repository.NewProductRepository(db.GetPool(), zap.NewNop())
	appointmentRepo := repository.NewAppointmentRepository(db.GetPool(), zap.NewNop())
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

func TestAppointmentRepository_GetByDateRange(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup
	db := setupTestDB(t)
	defer db.Close()

	productRepo := repository.NewProductRepository(db.GetPool(), zap.NewNop())
	appointmentRepo := repository.NewAppointmentRepository(db.GetPool(), zap.NewNop())
	ctx := context.Background()

	// Create test product
	product := createTestProduct(t, productRepo, ctx)
	defer productRepo.HardDelete(ctx, product.ID)

	// Create test appointment
	appointment := createTestAppointment(t, appointmentRepo, ctx, product.ID, "user_123")
	defer appointmentRepo.Delete(ctx, appointment.ID)

	// Test
	startTime := time.Now()
	endTime := time.Now().Add(48 * time.Hour)
	appointments, err := appointmentRepo.GetByDateRange(ctx, product.ID, startTime, endTime)

	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(appointments), 1)
}

func TestAppointmentRepository_Update(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup
	db := setupTestDB(t)
	defer db.Close()

	productRepo := repository.NewProductRepository(db.GetPool(), zap.NewNop())
	appointmentRepo := repository.NewAppointmentRepository(db.GetPool(), zap.NewNop())
	ctx := context.Background()

	// Create test product
	product := createTestProduct(t, productRepo, ctx)
	defer productRepo.HardDelete(ctx, product.ID)

	// Create test appointment
	appointment := createTestAppointment(t, appointmentRepo, ctx, product.ID, "user_123")
	defer appointmentRepo.Delete(ctx, appointment.ID)

	// Update appointment
	appointment.Title = "Updated Title"
	appointment.Status = models.AppointmentStatusConfirmed
	err := appointmentRepo.Update(ctx, appointment)
	require.NoError(t, err)

	// Verify
	retrieved, err := appointmentRepo.GetByID(ctx, appointment.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", retrieved.Title)
	assert.Equal(t, models.AppointmentStatusConfirmed, retrieved.Status)
}

func TestAppointmentRepository_UpdateStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup
	db := setupTestDB(t)
	defer db.Close()

	productRepo := repository.NewProductRepository(db.GetPool(), zap.NewNop())
	appointmentRepo := repository.NewAppointmentRepository(db.GetPool(), zap.NewNop())
	ctx := context.Background()

	// Create test product
	product := createTestProduct(t, productRepo, ctx)
	defer productRepo.HardDelete(ctx, product.ID)

	// Create test appointment
	appointment := createTestAppointment(t, appointmentRepo, ctx, product.ID, "user_123")
	defer appointmentRepo.Delete(ctx, appointment.ID)

	// Update status
	err := appointmentRepo.UpdateStatus(ctx, appointment.ID, models.AppointmentStatusCancelled)
	require.NoError(t, err)

	// Verify
	retrieved, err := appointmentRepo.GetByID(ctx, appointment.ID)
	require.NoError(t, err)
	assert.Equal(t, models.AppointmentStatusCancelled, retrieved.Status)
}

func TestAppointmentRepository_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup
	db := setupTestDB(t)
	defer db.Close()

	productRepo := repository.NewProductRepository(db.GetPool(), zap.NewNop())
	appointmentRepo := repository.NewAppointmentRepository(db.GetPool(), zap.NewNop())
	ctx := context.Background()

	// Create test product
	product := createTestProduct(t, productRepo, ctx)
	defer productRepo.HardDelete(ctx, product.ID)

	// Create test appointment
	appointment := createTestAppointment(t, appointmentRepo, ctx, product.ID, "user_123")

	// Delete
	err := appointmentRepo.Delete(ctx, appointment.ID)
	require.NoError(t, err)

	// Verify not found
	_, err = appointmentRepo.GetByID(ctx, appointment.ID)
	assert.ErrorIs(t, err, repository.ErrNotFound)
}

func TestAppointmentRepository_ParticipantOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup
	db := setupTestDB(t)
	defer db.Close()

	productRepo := repository.NewProductRepository(db.GetPool(), zap.NewNop())
	appointmentRepo := repository.NewAppointmentRepository(db.GetPool(), zap.NewNop())
	ctx := context.Background()

	// Create test product
	product := createTestProduct(t, productRepo, ctx)
	defer productRepo.HardDelete(ctx, product.ID)

	// Create test appointment
	appointment := createTestAppointment(t, appointmentRepo, ctx, product.ID, "user_123")
	defer appointmentRepo.Delete(ctx, appointment.ID)

	// Test AddParticipant
	newParticipant := &models.AppointmentParticipant{
		AppointmentID:  appointment.ID,
		ExternalUserID: "user_789",
		Role:           models.ParticipantRoleAttendee,
		Status:         models.ParticipantStatusPending,
		UserMetadata: map[string]interface{}{
			"name": "New Attendee",
		},
	}
	err := appointmentRepo.AddParticipant(ctx, newParticipant)
	require.NoError(t, err)

	// Verify participant added
	participants, err := appointmentRepo.GetParticipants(ctx, appointment.ID)
	require.NoError(t, err)
	assert.Len(t, participants, 2)

	// Test UpdateParticipantStatus
	err = appointmentRepo.UpdateParticipantStatus(ctx, appointment.ID, "user_789", models.ParticipantStatusAccepted)
	require.NoError(t, err)

	// Verify status updated
	participants, err = appointmentRepo.GetParticipants(ctx, appointment.ID)
	require.NoError(t, err)
	for _, p := range participants {
		if p.ExternalUserID == "user_789" {
			assert.Equal(t, models.ParticipantStatusAccepted, p.Status)
		}
	}

	// Test RemoveParticipant
	err = appointmentRepo.RemoveParticipant(ctx, appointment.ID, "user_789")
	require.NoError(t, err)

	// Verify participant removed
	participants, err = appointmentRepo.GetParticipants(ctx, appointment.ID)
	require.NoError(t, err)
	assert.Len(t, participants, 1)
}

// Helper functions

func setupTestDB(t *testing.T) *database.PostgresDB {
	cfg := database.Config{
		Host:            "127.0.0.1",
		Port:            "5433",
		User:            "appointments",
		Password:        "password123",
		Database:        "appointments_dev",
		SSLMode:         "disable",
		MaxConnections:  10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
	}

	db, err := database.NewPostgresDB(cfg, zap.NewNop())
	require.NoError(t, err)

	// Wait for database to be ready
	time.Sleep(1 * time.Second)

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
		ProductID: productID,
		Title:     "Test " + uuid.NewString()[:8],
		StartTime: time.Now().Add(24 * time.Hour),
		EndTime:   time.Now().Add(25 * time.Hour),
		Timezone:  "UTC",
		Status:    models.AppointmentStatusScheduled,
		CreatedBy: userID,
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
