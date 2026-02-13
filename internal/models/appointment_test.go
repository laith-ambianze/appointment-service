package models

import (
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppointmentValidate(t *testing.T) {
	now := time.Now()
	productID := uuid.New()

	tests := []struct {
		name        string
		appointment *Appointment
		wantErr     bool
		errContains string
	}{
		{
			name: "valid appointment",
			appointment: &Appointment{
				ProductID: productID,
				Title:     "Test Meeting",
				StartTime: now,
				EndTime:   now.Add(1 * time.Hour),
				CreatedBy: "user123",
			},
			wantErr: false,
		},
		{
			name: "missing product ID",
			appointment: &Appointment{
				Title:     "Test Meeting",
				StartTime: now,
				EndTime:   now.Add(1 * time.Hour),
				CreatedBy: "user123",
			},
			wantErr:     true,
			errContains: "product ID is required",
		},
		{
			name: "missing title",
			appointment: &Appointment{
				ProductID: productID,
				StartTime: now,
				EndTime:   now.Add(1 * time.Hour),
				CreatedBy: "user123",
			},
			wantErr:     true,
			errContains: "title is required",
		},
		{
			name: "title too long",
			appointment: &Appointment{
				ProductID: productID,
				Title:     string(make([]byte, 256)),
				StartTime: now,
				EndTime:   now.Add(1 * time.Hour),
				CreatedBy: "user123",
			},
			wantErr:     true,
			errContains: "must be less than 255 characters",
		},
		{
			name: "missing start time",
			appointment: &Appointment{
				ProductID: productID,
				Title:     "Test Meeting",
				EndTime:   now.Add(1 * time.Hour),
				CreatedBy: "user123",
			},
			wantErr:     true,
			errContains: "start time is required",
		},
		{
			name: "missing end time",
			appointment: &Appointment{
				ProductID: productID,
				Title:     "Test Meeting",
				StartTime: now,
				CreatedBy: "user123",
			},
			wantErr:     true,
			errContains: "end time is required",
		},
		{
			name: "end time before start time",
			appointment: &Appointment{
				ProductID: productID,
				Title:     "Test Meeting",
				StartTime: now.Add(1 * time.Hour),
				EndTime:   now,
				CreatedBy: "user123",
			},
			wantErr:     true,
			errContains: "end time must be after start time",
		},
		{
			name: "end time equals start time",
			appointment: &Appointment{
				ProductID: productID,
				Title:     "Test Meeting",
				StartTime: now,
				EndTime:   now,
				CreatedBy: "user123",
			},
			wantErr:     true,
			errContains: "end time must be after start time",
		},
		{
			name: "missing created by",
			appointment: &Appointment{
				ProductID: productID,
				Title:     "Test Meeting",
				StartTime: now,
				EndTime:   now.Add(1 * time.Hour),
			},
			wantErr:     true,
			errContains: "created_by is required",
		},
		{
			name: "invalid status",
			appointment: &Appointment{
				ProductID: productID,
				Title:     "Test Meeting",
				StartTime: now,
				EndTime:   now.Add(1 * time.Hour),
				CreatedBy: "user123",
				Status:    "invalid_status",
			},
			wantErr:     true,
			errContains: "invalid appointment status",
		},
		{
			name: "sets default timezone",
			appointment: &Appointment{
				ProductID: productID,
				Title:     "Test Meeting",
				StartTime: now,
				EndTime:   now.Add(1 * time.Hour),
				CreatedBy: "user123",
			},
			wantErr: false,
		},
		{
			name: "sets default status",
			appointment: &Appointment{
				ProductID: productID,
				Title:     "Test Meeting",
				StartTime: now,
				EndTime:   now.Add(1 * time.Hour),
				CreatedBy: "user123",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.appointment.Validate()

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
				// Check defaults
				if tt.appointment.Timezone == "" {
					assert.Equal(t, "UTC", tt.appointment.Timezone)
				}
				if tt.appointment.Status == "" {
					assert.Equal(t, AppointmentStatusScheduled, tt.appointment.Status)
				}
			}
		})
	}
}

func TestAppointmentIsActive(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		appointment *Appointment
		want        bool
	}{
		{
			name: "active appointment",
			appointment: &Appointment{
				BaseModel: BaseModel{
					DeletedAt: sql.NullTime{Valid: false},
				},
				Status: AppointmentStatusScheduled,
			},
			want: true,
		},
		{
			name: "cancelled appointment",
			appointment: &Appointment{
				BaseModel: BaseModel{
					DeletedAt: sql.NullTime{Valid: false},
				},
				Status: AppointmentStatusCancelled,
			},
			want: false,
		},
		{
			name: "deleted appointment",
			appointment: &Appointment{
				BaseModel: BaseModel{
					DeletedAt: sql.NullTime{
						Time:  now,
						Valid: true,
					},
				},
				Status: AppointmentStatusScheduled,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.appointment.IsActive()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAppointmentIsPast(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		appointment *Appointment
		want        bool
	}{
		{
			name: "past appointment",
			appointment: &Appointment{
				EndTime: now.Add(-1 * time.Hour),
			},
			want: true,
		},
		{
			name: "future appointment",
			appointment: &Appointment{
				EndTime: now.Add(1 * time.Hour),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.appointment.IsPast()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAppointmentIsFuture(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		appointment *Appointment
		want        bool
	}{
		{
			name: "future appointment",
			appointment: &Appointment{
				StartTime: now.Add(1 * time.Hour),
			},
			want: true,
		},
		{
			name: "past appointment",
			appointment: &Appointment{
				StartTime: now.Add(-1 * time.Hour),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.appointment.IsFuture()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAppointmentDuration(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		appointment *Appointment
		want        time.Duration
	}{
		{
			name: "1 hour appointment",
			appointment: &Appointment{
				StartTime: now,
				EndTime:   now.Add(1 * time.Hour),
			},
			want: 1 * time.Hour,
		},
		{
			name: "30 minute appointment",
			appointment: &Appointment{
				StartTime: now,
				EndTime:   now.Add(30 * time.Minute),
			},
			want: 30 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.appointment.Duration()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAppointmentMetadataJSON(t *testing.T) {
	tests := []struct {
		name        string
		appointment *Appointment
		want        string
		wantErr     bool
	}{
		{
			name: "nil metadata",
			appointment: &Appointment{
				Metadata: nil,
			},
			want:    "{}",
			wantErr: false,
		},
		{
			name: "empty metadata",
			appointment: &Appointment{
				Metadata: map[string]interface{}{},
			},
			want:    "{}",
			wantErr: false,
		},
		{
			name: "with metadata",
			appointment: &Appointment{
				Metadata: map[string]interface{}{
					"key1": "value1",
					"key2": 42,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.appointment.MetadataJSON()

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.name == "nil metadata" || tt.name == "empty metadata" {
					assert.Equal(t, tt.want, got)
				} else {
					assert.Contains(t, got, "key1")
					assert.Contains(t, got, "value1")
				}
			}
		})
	}
}
