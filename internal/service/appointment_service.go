package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/laith-ambianze/appointment-service/internal/models"
	"github.com/laith-ambianze/appointment-service/internal/repository"
	"github.com/laith-ambianze/appointment-service/pkg/auth"
	"go.uber.org/zap"
)

// Service errors
var (
	ErrAppointmentNotFound     = errors.New("appointment not found")
	ErrUnauthorized            = errors.New("unauthorized access")
	ErrForbidden               = errors.New("forbidden: insufficient permissions")
	ErrInvalidInput            = errors.New("invalid input")
	ErrAppointmentInPast       = errors.New("cannot modify past appointment")
	ErrInvalidStatusTransition = errors.New("invalid status transition")
	ErrParticipantNotFound     = errors.New("participant not found")
	ErrDuplicateParticipant    = errors.New("participant already exists")
)

// AppointmentService handles appointment business logic
type AppointmentService struct {
	repo   repository.AppointmentRepository
	logger *zap.Logger
}

// NewAppointmentService creates a new appointment service
func NewAppointmentService(repo repository.AppointmentRepository, logger *zap.Logger) *AppointmentService {
	return &AppointmentService{
		repo:   repo,
		logger: logger,
	}
}

// CreateAppointmentRequest contains data for creating an appointment
type CreateAppointmentRequest struct {
	Title        string                 `json:"title" binding:"required,min=1,max=255"`
	Description  string                 `json:"description" binding:"max=2000"`
	StartTime    time.Time              `json:"start_time" binding:"required"`
	EndTime      time.Time              `json:"end_time" binding:"required"`
	Timezone     string                 `json:"timezone"`
	Location     string                 `json:"location" binding:"max=500"`
	Metadata     map[string]interface{} `json:"metadata"`
	Participants []ParticipantRequest   `json:"participants" binding:"required,min=1,dive"`
}

// ParticipantRequest contains data for adding a participant
type ParticipantRequest struct {
	ExternalUserID string                 `json:"external_user_id" binding:"required,max=255"`
	Role           models.ParticipantRole `json:"role" binding:"required"`
	UserMetadata   map[string]interface{} `json:"user_metadata"`
}

// UpdateAppointmentRequest contains data for updating an appointment
type UpdateAppointmentRequest struct {
	Title       *string                `json:"title" binding:"omitempty,min=1,max=255"`
	Description *string                `json:"description" binding:"omitempty,max=2000"`
	StartTime   *time.Time             `json:"start_time"`
	EndTime     *time.Time             `json:"end_time"`
	Timezone    *string                `json:"timezone"`
	Location    *string                `json:"location" binding:"omitempty,max=500"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// RespondToAppointmentRequest contains data for responding to an appointment
type RespondToAppointmentRequest struct {
	Status models.AppointmentStatus `json:"status" binding:"required"`
}

// ListAppointmentsRequest contains filters for listing appointments
type ListAppointmentsRequest struct {
	Status        *models.AppointmentStatus `json:"status"`
	StartTimeFrom *time.Time                `json:"start_time_from"`
	StartTimeTo   *time.Time                `json:"start_time_to"`
	Limit         int                       `json:"limit"`
	Offset        int                       `json:"offset"`
}

// Create creates a new appointment
// The appointment is created with status "scheduled" and the creator is automatically added as host
func (s *AppointmentService) Create(ctx context.Context, productID uuid.UUID, userID string, req CreateAppointmentRequest) (*models.Appointment, error) {
	s.logger.Debug("Creating appointment",
		zap.String("product_id", productID.String()),
		zap.String("user_id", userID),
		zap.String("title", req.Title),
	)

	// Validate time range
	if req.EndTime.Before(req.StartTime) || req.EndTime.Equal(req.StartTime) {
		return nil, fmt.Errorf("%w: end time must be after start time", ErrInvalidInput)
	}

	// Validate start time is in future (allow some grace period)
	if req.StartTime.Before(time.Now().Add(-5 * time.Minute)) {
		return nil, fmt.Errorf("%w: start time must be in the future", ErrInvalidInput)
	}

	// Set defaults
	timezone := req.Timezone
	if timezone == "" {
		timezone = "UTC"
	}

	// Create appointment model
	appointment := &models.Appointment{
		BaseModel: models.BaseModel{
			ID: uuid.New(),
		},
		ProductID:   productID,
		Title:       req.Title,
		Description: req.Description,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Timezone:    timezone,
		Location:    req.Location,
		Status:      models.AppointmentStatusScheduled,
		CreatedBy:   userID,
		Metadata:    req.Metadata,
	}

	// Validate appointment
	if err := appointment.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	// Build participants list
	participants := make([]models.AppointmentParticipant, 0, len(req.Participants))
	creatorIncluded := false

	for _, p := range req.Participants {
		participant := models.AppointmentParticipant{
			BaseModel: models.BaseModel{
				ID: uuid.New(),
			},
			AppointmentID:  appointment.ID,
			ExternalUserID: p.ExternalUserID,
			Role:           p.Role,
			Status:         models.ParticipantStatusPending,
			UserMetadata:   p.UserMetadata,
		}

		// If creator is in participants, set their status to accepted
		if p.ExternalUserID == userID {
			participant.Status = models.ParticipantStatusAccepted
			creatorIncluded = true
		}

		// Validate participant
		if err := participant.Validate(); err != nil {
			return nil, fmt.Errorf("%w: participant validation failed: %v", ErrInvalidInput, err)
		}

		participants = append(participants, participant)
	}

	// If creator is not in participants list, add them as host
	if !creatorIncluded {
		hostParticipant := models.AppointmentParticipant{
			BaseModel: models.BaseModel{
				ID: uuid.New(),
			},
			AppointmentID:  appointment.ID,
			ExternalUserID: userID,
			Role:           models.ParticipantRoleHost,
			Status:         models.ParticipantStatusAccepted,
		}
		participants = append(participants, hostParticipant)
	}

	// Create appointment with participants
	if err := s.repo.Create(ctx, appointment, participants); err != nil {
		s.logger.Error("Failed to create appointment",
			zap.Error(err),
			zap.String("product_id", productID.String()),
		)
		return nil, fmt.Errorf("failed to create appointment: %w", err)
	}

	// Fetch the created appointment with participants
	created, err := s.repo.GetByID(ctx, appointment.ID)
	if err != nil {
		s.logger.Warn("Appointment created but failed to fetch",
			zap.Error(err),
			zap.String("appointment_id", appointment.ID.String()),
		)
		appointment.Participants = participants
		return appointment, nil
	}

	s.logger.Info("Appointment created",
		zap.String("appointment_id", created.ID.String()),
		zap.String("product_id", productID.String()),
		zap.String("created_by", userID),
	)

	return created, nil
}

// GetByID retrieves an appointment by ID with authorization check
func (s *AppointmentService) GetByID(ctx context.Context, productID uuid.UUID, appointmentID uuid.UUID) (*models.Appointment, error) {
	appointment, err := s.repo.GetByID(ctx, appointmentID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrAppointmentNotFound
		}
		return nil, fmt.Errorf("failed to get appointment: %w", err)
	}

	// Enforce multi-tenancy: appointment must belong to the product
	if appointment.ProductID != productID {
		s.logger.Warn("Unauthorized access attempt",
			zap.String("appointment_id", appointmentID.String()),
			zap.String("requested_product", productID.String()),
			zap.String("actual_product", appointment.ProductID.String()),
		)
		return nil, ErrAppointmentNotFound // Return not found to avoid leaking info
	}

	return appointment, nil
}

// ListByUser retrieves appointments for a user
func (s *AppointmentService) ListByUser(ctx context.Context, productID uuid.UUID, userID string, req ListAppointmentsRequest) ([]models.Appointment, error) {
	// Set defaults
	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	filters := repository.AppointmentFilters{
		Status:        req.Status,
		StartTimeFrom: req.StartTimeFrom,
		StartTimeTo:   req.StartTimeTo,
		Limit:         limit,
		Offset:        req.Offset,
	}

	appointments, err := s.repo.GetByProductAndUser(ctx, productID, userID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list appointments: %w", err)
	}

	return appointments, nil
}

// Update updates an appointment
func (s *AppointmentService) Update(ctx context.Context, productID uuid.UUID, userID string, role auth.Role, appointmentID uuid.UUID, req UpdateAppointmentRequest) (*models.Appointment, error) {
	// Get existing appointment
	appointment, err := s.GetByID(ctx, productID, appointmentID)
	if err != nil {
		return nil, err
	}

	// Check permissions: only creator, admin, or provider can update
	if appointment.CreatedBy != userID && !role.CanManageAppointments() {
		return nil, ErrForbidden
	}

	// Cannot modify cancelled or completed appointments
	if appointment.Status == models.AppointmentStatusCancelled ||
		appointment.Status == models.AppointmentStatusCompleted {
		return nil, fmt.Errorf("%w: appointment is %s", ErrInvalidStatusTransition, appointment.Status)
	}

	// Cannot modify past appointments
	if appointment.IsPast() {
		return nil, ErrAppointmentInPast
	}

	// Apply updates
	if req.Title != nil {
		appointment.Title = *req.Title
	}
	if req.Description != nil {
		appointment.Description = *req.Description
	}
	if req.StartTime != nil {
		appointment.StartTime = *req.StartTime
	}
	if req.EndTime != nil {
		appointment.EndTime = *req.EndTime
	}
	if req.Timezone != nil {
		appointment.Timezone = *req.Timezone
	}
	if req.Location != nil {
		appointment.Location = *req.Location
	}
	if req.Metadata != nil {
		appointment.Metadata = req.Metadata
	}

	// Validate updated appointment
	if err := appointment.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	// Save updates
	if err := s.repo.Update(ctx, appointment); err != nil {
		return nil, fmt.Errorf("failed to update appointment: %w", err)
	}

	s.logger.Info("Appointment updated",
		zap.String("appointment_id", appointmentID.String()),
		zap.String("updated_by", userID),
	)

	return appointment, nil
}

// Respond allows admin/provider to change appointment status (confirm, cancel, etc.)
func (s *AppointmentService) Respond(ctx context.Context, productID uuid.UUID, userID string, role auth.Role, appointmentID uuid.UUID, req RespondToAppointmentRequest) (*models.Appointment, error) {
	// Only admin or provider can respond to appointments
	if !role.CanManageAppointments() {
		return nil, ErrForbidden
	}

	// Get existing appointment
	appointment, err := s.GetByID(ctx, productID, appointmentID)
	if err != nil {
		return nil, err
	}

	// Validate status transition
	if !s.isValidStatusTransition(appointment.Status, req.Status) {
		return nil, fmt.Errorf("%w: cannot transition from %s to %s",
			ErrInvalidStatusTransition, appointment.Status, req.Status)
	}

	// Update status
	if err := s.repo.UpdateStatus(ctx, appointmentID, req.Status); err != nil {
		return nil, fmt.Errorf("failed to update appointment status: %w", err)
	}

	appointment.Status = req.Status

	s.logger.Info("Appointment status updated",
		zap.String("appointment_id", appointmentID.String()),
		zap.String("new_status", string(req.Status)),
		zap.String("updated_by", userID),
		zap.String("role", string(role)),
	)

	return appointment, nil
}

// Cancel cancels an appointment
func (s *AppointmentService) Cancel(ctx context.Context, productID uuid.UUID, userID string, role auth.Role, appointmentID uuid.UUID) (*models.Appointment, error) {
	// Get existing appointment
	appointment, err := s.GetByID(ctx, productID, appointmentID)
	if err != nil {
		return nil, err
	}

	// Check permissions: creator, admin, or provider can cancel
	if appointment.CreatedBy != userID && !role.CanManageAppointments() {
		return nil, ErrForbidden
	}

	// Cannot cancel already cancelled or completed
	if appointment.Status == models.AppointmentStatusCancelled {
		return nil, fmt.Errorf("%w: appointment already cancelled", ErrInvalidStatusTransition)
	}
	if appointment.Status == models.AppointmentStatusCompleted {
		return nil, fmt.Errorf("%w: cannot cancel completed appointment", ErrInvalidStatusTransition)
	}

	// Update status to cancelled
	if err := s.repo.UpdateStatus(ctx, appointmentID, models.AppointmentStatusCancelled); err != nil {
		return nil, fmt.Errorf("failed to cancel appointment: %w", err)
	}

	appointment.Status = models.AppointmentStatusCancelled

	s.logger.Info("Appointment cancelled",
		zap.String("appointment_id", appointmentID.String()),
		zap.String("cancelled_by", userID),
	)

	return appointment, nil
}

// Delete soft-deletes an appointment
func (s *AppointmentService) Delete(ctx context.Context, productID uuid.UUID, userID string, role auth.Role, appointmentID uuid.UUID) error {
	// Only admin can delete appointments
	if role != auth.RoleAdmin {
		return ErrForbidden
	}

	// Get existing appointment to verify ownership
	_, err := s.GetByID(ctx, productID, appointmentID)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, appointmentID); err != nil {
		return fmt.Errorf("failed to delete appointment: %w", err)
	}

	s.logger.Info("Appointment deleted",
		zap.String("appointment_id", appointmentID.String()),
		zap.String("deleted_by", userID),
	)

	return nil
}

// AddParticipant adds a participant to an appointment
func (s *AppointmentService) AddParticipant(ctx context.Context, productID uuid.UUID, userID string, role auth.Role, appointmentID uuid.UUID, req ParticipantRequest) (*models.AppointmentParticipant, error) {
	// Get existing appointment
	appointment, err := s.GetByID(ctx, productID, appointmentID)
	if err != nil {
		return nil, err
	}

	// Check permissions
	if appointment.CreatedBy != userID && !role.CanManageAppointments() {
		return nil, ErrForbidden
	}

	// Check if participant already exists
	for _, p := range appointment.Participants {
		if p.ExternalUserID == req.ExternalUserID {
			return nil, ErrDuplicateParticipant
		}
	}

	participant := &models.AppointmentParticipant{
		BaseModel: models.BaseModel{
			ID: uuid.New(),
		},
		AppointmentID:  appointmentID,
		ExternalUserID: req.ExternalUserID,
		Role:           req.Role,
		Status:         models.ParticipantStatusPending,
		UserMetadata:   req.UserMetadata,
	}

	if err := participant.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	if err := s.repo.AddParticipant(ctx, participant); err != nil {
		return nil, fmt.Errorf("failed to add participant: %w", err)
	}

	s.logger.Info("Participant added",
		zap.String("appointment_id", appointmentID.String()),
		zap.String("participant_id", req.ExternalUserID),
		zap.String("added_by", userID),
	)

	return participant, nil
}

// RemoveParticipant removes a participant from an appointment
func (s *AppointmentService) RemoveParticipant(ctx context.Context, productID uuid.UUID, userID string, role auth.Role, appointmentID uuid.UUID, participantUserID string) error {
	// Get existing appointment
	appointment, err := s.GetByID(ctx, productID, appointmentID)
	if err != nil {
		return err
	}

	// Check permissions
	if appointment.CreatedBy != userID && !role.CanManageAppointments() {
		return ErrForbidden
	}

	// Cannot remove the host
	for _, p := range appointment.Participants {
		if p.ExternalUserID == participantUserID && p.IsHost() {
			return fmt.Errorf("%w: cannot remove host from appointment", ErrInvalidInput)
		}
	}

	if err := s.repo.RemoveParticipant(ctx, appointmentID, participantUserID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrParticipantNotFound
		}
		return fmt.Errorf("failed to remove participant: %w", err)
	}

	s.logger.Info("Participant removed",
		zap.String("appointment_id", appointmentID.String()),
		zap.String("participant_id", participantUserID),
		zap.String("removed_by", userID),
	)

	return nil
}

// UpdateParticipantStatus updates a participant's response status
// Users can update their own status, admin/provider can update any
func (s *AppointmentService) UpdateParticipantStatus(ctx context.Context, productID uuid.UUID, userID string, role auth.Role, appointmentID uuid.UUID, participantUserID string, status models.ParticipantStatus) error {
	// Get existing appointment
	appointment, err := s.GetByID(ctx, productID, appointmentID)
	if err != nil {
		return err
	}

	// Check permissions: user can update own status, admin/provider can update any
	if participantUserID != userID && !role.CanManageAppointments() {
		return ErrForbidden
	}

	// Verify participant exists
	found := false
	for _, p := range appointment.Participants {
		if p.ExternalUserID == participantUserID {
			found = true
			break
		}
	}
	if !found {
		return ErrParticipantNotFound
	}

	if err := s.repo.UpdateParticipantStatus(ctx, appointmentID, participantUserID, status); err != nil {
		return fmt.Errorf("failed to update participant status: %w", err)
	}

	s.logger.Info("Participant status updated",
		zap.String("appointment_id", appointmentID.String()),
		zap.String("participant_id", participantUserID),
		zap.String("new_status", string(status)),
		zap.String("updated_by", userID),
	)

	return nil
}

// isValidStatusTransition checks if a status transition is allowed
func (s *AppointmentService) isValidStatusTransition(from, to models.AppointmentStatus) bool {
	// Define allowed transitions
	transitions := map[models.AppointmentStatus][]models.AppointmentStatus{
		models.AppointmentStatusScheduled: {
			models.AppointmentStatusConfirmed,
			models.AppointmentStatusCancelled,
		},
		models.AppointmentStatusConfirmed: {
			models.AppointmentStatusCompleted,
			models.AppointmentStatusCancelled,
			models.AppointmentStatusNoShow,
		},
	}

	allowed, exists := transitions[from]
	if !exists {
		return false
	}

	for _, status := range allowed {
		if status == to {
			return true
		}
	}

	return false
}
