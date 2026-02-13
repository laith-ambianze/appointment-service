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
	ProductID    uuid.UUID                `json:"product_id"`
	Title        string                   `json:"title"`
	Description  string                   `json:"description,omitempty"`
	StartTime    time.Time                `json:"start_time"`
	EndTime      time.Time                `json:"end_time"`
	Timezone     string                   `json:"timezone"`
	Location     string                   `json:"location,omitempty"`
	Status       AppointmentStatus        `json:"status"`
	CreatedBy    string                   `json:"created_by"`
	Metadata     map[string]interface{}   `json:"metadata,omitempty"`
	Participants []AppointmentParticipant `json:"participants,omitempty"`
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
	ProductID    uuid.UUID              `json:"product_id" binding:"required"`
	Title        string                 `json:"title" binding:"required,max=255"`
	Description  string                 `json:"description" binding:"max=2000"`
	StartTime    time.Time              `json:"start_time" binding:"required"`
	EndTime      time.Time              `json:"end_time" binding:"required,gtfield=StartTime"`
	Timezone     string                 `json:"timezone" binding:"required"`
	Location     string                 `json:"location" binding:"max=500"`
	CreatedBy    string                 `json:"created_by" binding:"required"`
	Metadata     map[string]interface{} `json:"metadata"`
	Participants []ParticipantInput     `json:"participants" binding:"required,min=1,dive"`
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
