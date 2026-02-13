package models

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParticipantValidate(t *testing.T) {
	appointmentID := uuid.New()

	tests := []struct {
		name        string
		participant *AppointmentParticipant
		wantErr     bool
		errContains string
	}{
		{
			name: "valid participant",
			participant: &AppointmentParticipant{
				AppointmentID:  appointmentID,
				ExternalUserID: "user123",
				Role:           ParticipantRoleAttendee,
				Status:         ParticipantStatusPending,
			},
			wantErr: false,
		},
		{
			name: "missing appointment ID",
			participant: &AppointmentParticipant{
				ExternalUserID: "user123",
				Role:           ParticipantRoleAttendee,
			},
			wantErr:     true,
			errContains: "appointment ID is required",
		},
		{
			name: "missing external user ID",
			participant: &AppointmentParticipant{
				AppointmentID: appointmentID,
				Role:          ParticipantRoleAttendee,
			},
			wantErr:     true,
			errContains: "external user ID is required",
		},
		{
			name: "external user ID too long",
			participant: &AppointmentParticipant{
				AppointmentID:  appointmentID,
				ExternalUserID: string(make([]byte, 256)),
				Role:           ParticipantRoleAttendee,
			},
			wantErr:     true,
			errContains: "must be less than 255 characters",
		},
		{
			name: "invalid role",
			participant: &AppointmentParticipant{
				AppointmentID:  appointmentID,
				ExternalUserID: "user123",
				Role:           "invalid_role",
			},
			wantErr:     true,
			errContains: "invalid participant role",
		},
		{
			name: "invalid status",
			participant: &AppointmentParticipant{
				AppointmentID:  appointmentID,
				ExternalUserID: "user123",
				Role:           ParticipantRoleAttendee,
				Status:         "invalid_status",
			},
			wantErr:     true,
			errContains: "invalid participant status",
		},
		{
			name: "sets default role",
			participant: &AppointmentParticipant{
				AppointmentID:  appointmentID,
				ExternalUserID: "user123",
			},
			wantErr: false,
		},
		{
			name: "sets default status",
			participant: &AppointmentParticipant{
				AppointmentID:  appointmentID,
				ExternalUserID: "user123",
				Role:           ParticipantRoleGuest,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.participant.Validate()

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
				// Check defaults
				if tt.participant.Role == "" {
					assert.Equal(t, ParticipantRoleAttendee, tt.participant.Role)
				}
				if tt.participant.Status == "" {
					assert.Equal(t, ParticipantStatusPending, tt.participant.Status)
				}
			}
		})
	}
}

func TestParticipantIsHost(t *testing.T) {
	tests := []struct {
		name        string
		participant *AppointmentParticipant
		want        bool
	}{
		{
			name: "is host",
			participant: &AppointmentParticipant{
				Role: ParticipantRoleHost,
			},
			want: true,
		},
		{
			name: "is not host - guest",
			participant: &AppointmentParticipant{
				Role: ParticipantRoleGuest,
			},
			want: false,
		},
		{
			name: "is not host - attendee",
			participant: &AppointmentParticipant{
				Role: ParticipantRoleAttendee,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.participant.IsHost()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParticipantHasAccepted(t *testing.T) {
	tests := []struct {
		name        string
		participant *AppointmentParticipant
		want        bool
	}{
		{
			name: "has accepted",
			participant: &AppointmentParticipant{
				Status: ParticipantStatusAccepted,
			},
			want: true,
		},
		{
			name: "has not accepted - pending",
			participant: &AppointmentParticipant{
				Status: ParticipantStatusPending,
			},
			want: false,
		},
		{
			name: "has not accepted - declined",
			participant: &AppointmentParticipant{
				Status: ParticipantStatusDeclined,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.participant.HasAccepted()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParticipantHasDeclined(t *testing.T) {
	tests := []struct {
		name        string
		participant *AppointmentParticipant
		want        bool
	}{
		{
			name: "has declined",
			participant: &AppointmentParticipant{
				Status: ParticipantStatusDeclined,
			},
			want: true,
		},
		{
			name: "has not declined - accepted",
			participant: &AppointmentParticipant{
				Status: ParticipantStatusAccepted,
			},
			want: false,
		},
		{
			name: "has not declined - pending",
			participant: &AppointmentParticipant{
				Status: ParticipantStatusPending,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.participant.HasDeclined()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParticipantUserMetadataJSON(t *testing.T) {
	tests := []struct {
		name        string
		participant *AppointmentParticipant
		want        string
		wantErr     bool
	}{
		{
			name: "nil user metadata",
			participant: &AppointmentParticipant{
				UserMetadata: nil,
			},
			want:    "{}",
			wantErr: false,
		},
		{
			name: "empty user metadata",
			participant: &AppointmentParticipant{
				UserMetadata: map[string]interface{}{},
			},
			want:    "{}",
			wantErr: false,
		},
		{
			name: "with user metadata",
			participant: &AppointmentParticipant{
				UserMetadata: map[string]interface{}{
					"name":  "John Doe",
					"email": "john@example.com",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.participant.UserMetadataJSON()

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.name == "nil user metadata" || tt.name == "empty user metadata" {
					assert.Equal(t, tt.want, got)
				} else {
					assert.Contains(t, got, "name")
					assert.Contains(t, got, "John Doe")
				}
			}
		})
	}
}
