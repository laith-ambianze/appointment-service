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

// ParticipantStatus represents the status of a participant's response
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
		p.Role = ParticipantRoleAttendee
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

// HasAccepted returns true if the participant has accepted the invitation
func (p *AppointmentParticipant) HasAccepted() bool {
	return p.Status == ParticipantStatusAccepted
}

// HasDeclined returns true if the participant has declined the invitation
func (p *AppointmentParticipant) HasDeclined() bool {
	return p.Status == ParticipantStatusDeclined
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

// ParticipantInput represents input for creating a participant
type ParticipantInput struct {
	ExternalUserID string                 `json:"external_user_id" binding:"required,max=255"`
	Role           ParticipantRole        `json:"role" binding:"required"`
	UserMetadata   map[string]interface{} `json:"user_metadata"`
}

// UpdateParticipantInput represents input for updating a participant
type UpdateParticipantInput struct {
	Role         *ParticipantRole       `json:"role"`
	Status       *ParticipantStatus     `json:"status"`
	UserMetadata map[string]interface{} `json:"user_metadata"`
}
