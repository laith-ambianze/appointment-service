package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/laith-ambianze/appointment-service/internal/models"
	"github.com/laith-ambianze/appointment-service/internal/repository"
	"github.com/laith-ambianze/appointment-service/pkg/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockAppointmentRepository is a mock implementation of AppointmentRepository
type MockAppointmentRepository struct {
	mock.Mock
}

func (m *MockAppointmentRepository) Create(ctx context.Context, appointment *models.Appointment, participants []models.AppointmentParticipant) error {
	args := m.Called(ctx, appointment, participants)
	return args.Error(0)
}

func (m *MockAppointmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Appointment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Appointment), args.Error(1)
}

func (m *MockAppointmentRepository) GetByProductAndUser(ctx context.Context, productID uuid.UUID, externalUserID string, filters repository.AppointmentFilters) ([]models.Appointment, error) {
	args := m.Called(ctx, productID, externalUserID, filters)
	return args.Get(0).([]models.Appointment), args.Error(1)
}

func (m *MockAppointmentRepository) GetByDateRange(ctx context.Context, productID uuid.UUID, startTime, endTime time.Time) ([]models.Appointment, error) {
	args := m.Called(ctx, productID, startTime, endTime)
	return args.Get(0).([]models.Appointment), args.Error(1)
}

func (m *MockAppointmentRepository) Update(ctx context.Context, appointment *models.Appointment) error {
	args := m.Called(ctx, appointment)
	return args.Error(0)
}

func (m *MockAppointmentRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.AppointmentStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockAppointmentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAppointmentRepository) AddParticipant(ctx context.Context, participant *models.AppointmentParticipant) error {
	args := m.Called(ctx, participant)
	return args.Error(0)
}

func (m *MockAppointmentRepository) RemoveParticipant(ctx context.Context, appointmentID uuid.UUID, externalUserID string) error {
	args := m.Called(ctx, appointmentID, externalUserID)
	return args.Error(0)
}

func (m *MockAppointmentRepository) UpdateParticipantStatus(ctx context.Context, appointmentID uuid.UUID, externalUserID string, status models.ParticipantStatus) error {
	args := m.Called(ctx, appointmentID, externalUserID, status)
	return args.Error(0)
}

func (m *MockAppointmentRepository) GetParticipants(ctx context.Context, appointmentID uuid.UUID) ([]models.AppointmentParticipant, error) {
	args := m.Called(ctx, appointmentID)
	return args.Get(0).([]models.AppointmentParticipant), args.Error(1)
}

func newTestService() (*AppointmentService, *MockAppointmentRepository) {
	mockRepo := new(MockAppointmentRepository)
	logger, _ := zap.NewDevelopment()
	service := NewAppointmentService(mockRepo, logger)
	return service, mockRepo
}

func TestAppointmentService_Create_Success(t *testing.T) {
	service, mockRepo := newTestService()
	ctx := context.Background()
	productID := uuid.New()
	userID := "user-123"

	req := CreateAppointmentRequest{
		Title:       "Team Meeting",
		Description: "Weekly sync",
		StartTime:   time.Now().Add(24 * time.Hour),
		EndTime:     time.Now().Add(25 * time.Hour),
		Timezone:    "UTC",
		Participants: []ParticipantRequest{
			{
				ExternalUserID: "user-456",
				Role:           models.ParticipantRoleGuest,
			},
		},
	}

	mockRepo.On("Create", ctx, mock.AnythingOfType("*models.Appointment"), mock.AnythingOfType("[]models.AppointmentParticipant")).Return(nil)
	mockRepo.On("GetByID", ctx, mock.AnythingOfType("uuid.UUID")).Return(&models.Appointment{
		BaseModel: models.BaseModel{ID: uuid.New()},
		ProductID: productID,
		Title:     req.Title,
		Status:    models.AppointmentStatusScheduled,
		CreatedBy: userID,
	}, nil)

	appointment, err := service.Create(ctx, productID, userID, req)

	assert.NoError(t, err)
	assert.NotNil(t, appointment)
	assert.Equal(t, req.Title, appointment.Title)
	mockRepo.AssertExpectations(t)
}

func TestAppointmentService_Create_InvalidTimeRange(t *testing.T) {
	service, _ := newTestService()
	ctx := context.Background()
	productID := uuid.New()
	userID := "user-123"

	req := CreateAppointmentRequest{
		Title:     "Team Meeting",
		StartTime: time.Now().Add(25 * time.Hour),
		EndTime:   time.Now().Add(24 * time.Hour), // End before start
		Participants: []ParticipantRequest{
			{ExternalUserID: "user-456", Role: models.ParticipantRoleGuest},
		},
	}

	_, err := service.Create(ctx, productID, userID, req)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidInput)
}

func TestAppointmentService_Create_PastStartTime(t *testing.T) {
	service, _ := newTestService()
	ctx := context.Background()
	productID := uuid.New()
	userID := "user-123"

	req := CreateAppointmentRequest{
		Title:     "Team Meeting",
		StartTime: time.Now().Add(-1 * time.Hour), // In the past
		EndTime:   time.Now().Add(1 * time.Hour),
		Participants: []ParticipantRequest{
			{ExternalUserID: "user-456", Role: models.ParticipantRoleGuest},
		},
	}

	_, err := service.Create(ctx, productID, userID, req)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidInput)
}

func TestAppointmentService_GetByID_Success(t *testing.T) {
	service, mockRepo := newTestService()
	ctx := context.Background()
	productID := uuid.New()
	appointmentID := uuid.New()

	expectedAppointment := &models.Appointment{
		BaseModel: models.BaseModel{ID: appointmentID},
		ProductID: productID,
		Title:     "Test Appointment",
	}

	mockRepo.On("GetByID", ctx, appointmentID).Return(expectedAppointment, nil)

	appointment, err := service.GetByID(ctx, productID, appointmentID)

	assert.NoError(t, err)
	assert.Equal(t, expectedAppointment, appointment)
	mockRepo.AssertExpectations(t)
}

func TestAppointmentService_GetByID_WrongProduct(t *testing.T) {
	service, mockRepo := newTestService()
	ctx := context.Background()
	productID := uuid.New()
	wrongProductID := uuid.New()
	appointmentID := uuid.New()

	expectedAppointment := &models.Appointment{
		BaseModel: models.BaseModel{ID: appointmentID},
		ProductID: wrongProductID, // Different product
		Title:     "Test Appointment",
	}

	mockRepo.On("GetByID", ctx, appointmentID).Return(expectedAppointment, nil)

	_, err := service.GetByID(ctx, productID, appointmentID)

	assert.Error(t, err)
	assert.Equal(t, ErrAppointmentNotFound, err)
	mockRepo.AssertExpectations(t)
}

func TestAppointmentService_GetByID_NotFound(t *testing.T) {
	service, mockRepo := newTestService()
	ctx := context.Background()
	productID := uuid.New()
	appointmentID := uuid.New()

	mockRepo.On("GetByID", ctx, appointmentID).Return(nil, repository.ErrNotFound)

	_, err := service.GetByID(ctx, productID, appointmentID)

	assert.Error(t, err)
	assert.Equal(t, ErrAppointmentNotFound, err)
	mockRepo.AssertExpectations(t)
}

func TestAppointmentService_Respond_Success(t *testing.T) {
	service, mockRepo := newTestService()
	ctx := context.Background()
	productID := uuid.New()
	appointmentID := uuid.New()
	userID := "admin-user"

	existingAppointment := &models.Appointment{
		BaseModel: models.BaseModel{ID: appointmentID},
		ProductID: productID,
		Title:     "Test Appointment",
		Status:    models.AppointmentStatusScheduled,
		CreatedBy: "other-user",
	}

	mockRepo.On("GetByID", ctx, appointmentID).Return(existingAppointment, nil)
	mockRepo.On("UpdateStatus", ctx, appointmentID, models.AppointmentStatusConfirmed).Return(nil)

	req := RespondToAppointmentRequest{
		Status: models.AppointmentStatusConfirmed,
	}

	appointment, err := service.Respond(ctx, productID, userID, auth.RoleAdmin, appointmentID, req)

	assert.NoError(t, err)
	assert.Equal(t, models.AppointmentStatusConfirmed, appointment.Status)
	mockRepo.AssertExpectations(t)
}

func TestAppointmentService_Respond_ForbiddenForUser(t *testing.T) {
	service, _ := newTestService()
	ctx := context.Background()
	productID := uuid.New()
	appointmentID := uuid.New()
	userID := "regular-user"

	req := RespondToAppointmentRequest{
		Status: models.AppointmentStatusConfirmed,
	}

	// User role cannot respond to appointments
	_, err := service.Respond(ctx, productID, userID, auth.RoleUser, appointmentID, req)

	assert.Error(t, err)
	assert.Equal(t, ErrForbidden, err)
}

func TestAppointmentService_Respond_InvalidTransition(t *testing.T) {
	service, mockRepo := newTestService()
	ctx := context.Background()
	productID := uuid.New()
	appointmentID := uuid.New()
	userID := "admin-user"

	// Cancelled appointment
	existingAppointment := &models.Appointment{
		BaseModel: models.BaseModel{ID: appointmentID},
		ProductID: productID,
		Title:     "Test Appointment",
		Status:    models.AppointmentStatusCancelled, // Already cancelled
		CreatedBy: "other-user",
	}

	mockRepo.On("GetByID", ctx, appointmentID).Return(existingAppointment, nil)

	req := RespondToAppointmentRequest{
		Status: models.AppointmentStatusConfirmed,
	}

	_, err := service.Respond(ctx, productID, userID, auth.RoleAdmin, appointmentID, req)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidStatusTransition)
	mockRepo.AssertExpectations(t)
}

func TestAppointmentService_Cancel_Success(t *testing.T) {
	service, mockRepo := newTestService()
	ctx := context.Background()
	productID := uuid.New()
	appointmentID := uuid.New()
	userID := "user-123"

	existingAppointment := &models.Appointment{
		BaseModel: models.BaseModel{ID: appointmentID},
		ProductID: productID,
		Title:     "Test Appointment",
		Status:    models.AppointmentStatusScheduled,
		CreatedBy: userID, // Same user
	}

	mockRepo.On("GetByID", ctx, appointmentID).Return(existingAppointment, nil)
	mockRepo.On("UpdateStatus", ctx, appointmentID, models.AppointmentStatusCancelled).Return(nil)

	appointment, err := service.Cancel(ctx, productID, userID, auth.RoleUser, appointmentID)

	assert.NoError(t, err)
	assert.Equal(t, models.AppointmentStatusCancelled, appointment.Status)
	mockRepo.AssertExpectations(t)
}

func TestAppointmentService_Cancel_ByAdmin(t *testing.T) {
	service, mockRepo := newTestService()
	ctx := context.Background()
	productID := uuid.New()
	appointmentID := uuid.New()
	adminUserID := "admin-user"

	existingAppointment := &models.Appointment{
		BaseModel: models.BaseModel{ID: appointmentID},
		ProductID: productID,
		Title:     "Test Appointment",
		Status:    models.AppointmentStatusScheduled,
		CreatedBy: "other-user", // Different user
	}

	mockRepo.On("GetByID", ctx, appointmentID).Return(existingAppointment, nil)
	mockRepo.On("UpdateStatus", ctx, appointmentID, models.AppointmentStatusCancelled).Return(nil)

	// Admin can cancel anyone's appointment
	appointment, err := service.Cancel(ctx, productID, adminUserID, auth.RoleAdmin, appointmentID)

	assert.NoError(t, err)
	assert.Equal(t, models.AppointmentStatusCancelled, appointment.Status)
	mockRepo.AssertExpectations(t)
}

func TestAppointmentService_Delete_OnlyAdmin(t *testing.T) {
	service, mockRepo := newTestService()
	ctx := context.Background()
	productID := uuid.New()
	appointmentID := uuid.New()

	existingAppointment := &models.Appointment{
		BaseModel: models.BaseModel{ID: appointmentID},
		ProductID: productID,
	}

	// Admin can delete
	mockRepo.On("GetByID", ctx, appointmentID).Return(existingAppointment, nil)
	mockRepo.On("Delete", ctx, appointmentID).Return(nil)

	err := service.Delete(ctx, productID, "admin-user", auth.RoleAdmin, appointmentID)
	assert.NoError(t, err)

	// User cannot delete
	err = service.Delete(ctx, productID, "user", auth.RoleUser, uuid.New())
	assert.Equal(t, ErrForbidden, err)

	// Provider cannot delete
	err = service.Delete(ctx, productID, "provider", auth.RoleProvider, uuid.New())
	assert.Equal(t, ErrForbidden, err)
}

func TestAppointmentService_isValidStatusTransition(t *testing.T) {
	service, _ := newTestService()

	tests := []struct {
		name     string
		from     models.AppointmentStatus
		to       models.AppointmentStatus
		expected bool
	}{
		{"scheduled to confirmed", models.AppointmentStatusScheduled, models.AppointmentStatusConfirmed, true},
		{"scheduled to cancelled", models.AppointmentStatusScheduled, models.AppointmentStatusCancelled, true},
		{"scheduled to completed", models.AppointmentStatusScheduled, models.AppointmentStatusCompleted, false},
		{"confirmed to completed", models.AppointmentStatusConfirmed, models.AppointmentStatusCompleted, true},
		{"confirmed to cancelled", models.AppointmentStatusConfirmed, models.AppointmentStatusCancelled, true},
		{"confirmed to no_show", models.AppointmentStatusConfirmed, models.AppointmentStatusNoShow, true},
		{"cancelled to confirmed", models.AppointmentStatusCancelled, models.AppointmentStatusConfirmed, false},
		{"completed to cancelled", models.AppointmentStatusCompleted, models.AppointmentStatusCancelled, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.isValidStatusTransition(tt.from, tt.to)
			assert.Equal(t, tt.expected, result)
		})
	}
}
