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
