package models

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// DayOfWeek represents a day of the week (0=Sunday, 6=Saturday)
type DayOfWeek int

const (
	Sunday    DayOfWeek = 0
	Monday    DayOfWeek = 1
	Tuesday   DayOfWeek = 2
	Wednesday DayOfWeek = 3
	Thursday  DayOfWeek = 4
	Friday    DayOfWeek = 5
	Saturday  DayOfWeek = 6
)

// String returns the string representation of the day
func (d DayOfWeek) String() string {
	days := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	if d < 0 || d > 6 {
		return "Invalid"
	}
	return days[d]
}

// IsValid checks if the day of week is valid
func (d DayOfWeek) IsValid() bool {
	return d >= 0 && d <= 6
}

// AvailabilityRule represents a provider's availability schedule for a specific day
type AvailabilityRule struct {
	ID                  uuid.UUID    `json:"id"`
	ProductID           uuid.UUID    `json:"product_id"`
	ProviderID          string       `json:"provider_id"`
	DayOfWeek           DayOfWeek    `json:"day_of_week"`
	StartTime           LocalTime    `json:"start_time"`
	EndTime             LocalTime    `json:"end_time"`
	DurationMinutes     int          `json:"duration_minutes"`
	SlotIntervalMinutes int          `json:"slot_interval_minutes"`
	BufferBeforeMinutes int          `json:"buffer_before_minutes"`
	BufferAfterMinutes  int          `json:"buffer_after_minutes"`
	Timezone            string       `json:"timezone"`
	IsActive            bool         `json:"is_active"`
	CreatedAt           time.Time    `json:"created_at"`
	UpdatedAt           time.Time    `json:"updated_at"`
	DeletedAt           sql.NullTime `json:"deleted_at,omitempty"`
}

// LocalTime represents a time without date (HH:MM:SS)
type LocalTime struct {
	Hour   int `json:"hour"`
	Minute int `json:"minute"`
	Second int `json:"second"`
}

// NewLocalTime creates a LocalTime from hour and minute
func NewLocalTime(hour, minute int) LocalTime {
	return LocalTime{Hour: hour, Minute: minute, Second: 0}
}

// ParseLocalTime parses a time string in format "HH:MM:SS" or "HH:MM"
func ParseLocalTime(s string) (LocalTime, error) {
	var lt LocalTime
	var n int

	// Try parsing with seconds first (HH:MM:SS)
	n, _ = fmt.Sscanf(s, "%d:%d:%d", &lt.Hour, &lt.Minute, &lt.Second)
	if n < 2 {
		// Try without seconds (HH:MM)
		n, _ = fmt.Sscanf(s, "%d:%d", &lt.Hour, &lt.Minute)
		if n < 2 {
			return LocalTime{}, fmt.Errorf("invalid time format: %s (expected HH:MM or HH:MM:SS)", s)
		}
		lt.Second = 0
	}

	if lt.Hour < 0 || lt.Hour > 23 {
		return LocalTime{}, fmt.Errorf("invalid hour: %d", lt.Hour)
	}
	if lt.Minute < 0 || lt.Minute > 59 {
		return LocalTime{}, fmt.Errorf("invalid minute: %d", lt.Minute)
	}
	if lt.Second < 0 || lt.Second > 59 {
		return LocalTime{}, fmt.Errorf("invalid second: %d", lt.Second)
	}

	return lt, nil
}

// String returns the time in HH:MM:SS format
func (lt LocalTime) String() string {
	return fmt.Sprintf("%02d:%02d:%02d", lt.Hour, lt.Minute, lt.Second)
}

// ToMinutes returns the total minutes from midnight
func (lt LocalTime) ToMinutes() int {
	return lt.Hour*60 + lt.Minute
}

// After returns true if lt is after other
func (lt LocalTime) After(other LocalTime) bool {
	return lt.ToMinutes() > other.ToMinutes()
}

// Before returns true if lt is before other
func (lt LocalTime) Before(other LocalTime) bool {
	return lt.ToMinutes() < other.ToMinutes()
}

// Equal returns true if lt equals other
func (lt LocalTime) Equal(other LocalTime) bool {
	return lt.ToMinutes() == other.ToMinutes()
}

// ToTimeOnDate converts LocalTime to time.Time on a specific date in the given timezone
func (lt LocalTime) ToTimeOnDate(date time.Time, loc *time.Location) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), lt.Hour, lt.Minute, lt.Second, 0, loc)
}

// Validate validates the availability rule fields
func (r *AvailabilityRule) Validate() error {
	if r.ProductID == uuid.Nil {
		return fmt.Errorf("product ID is required")
	}

	if r.ProviderID == "" {
		return fmt.Errorf("provider ID is required")
	}

	if len(r.ProviderID) > 255 {
		return fmt.Errorf("provider ID must be less than 255 characters")
	}

	if !r.DayOfWeek.IsValid() {
		return fmt.Errorf("invalid day of week: %d (must be 0-6)", r.DayOfWeek)
	}

	if r.EndTime.Before(r.StartTime) || r.EndTime.Equal(r.StartTime) {
		return fmt.Errorf("end time must be after start time")
	}

	if r.DurationMinutes <= 0 {
		return fmt.Errorf("duration must be greater than 0")
	}
	if r.DurationMinutes > 480 {
		return fmt.Errorf("duration must not exceed 480 minutes (8 hours)")
	}

	if r.SlotIntervalMinutes <= 0 {
		return fmt.Errorf("slot interval must be greater than 0")
	}
	if r.SlotIntervalMinutes > r.DurationMinutes {
		return fmt.Errorf("slot interval cannot exceed duration")
	}

	if r.BufferBeforeMinutes < 0 {
		return fmt.Errorf("buffer before must be non-negative")
	}
	if r.BufferBeforeMinutes > 120 {
		return fmt.Errorf("buffer before must not exceed 120 minutes")
	}

	if r.BufferAfterMinutes < 0 {
		return fmt.Errorf("buffer after must be non-negative")
	}
	if r.BufferAfterMinutes > 120 {
		return fmt.Errorf("buffer after must not exceed 120 minutes")
	}

	if r.Timezone == "" {
		r.Timezone = "UTC"
	}

	// Validate timezone
	_, err := time.LoadLocation(r.Timezone)
	if err != nil {
		return fmt.Errorf("invalid timezone: %s", r.Timezone)
	}

	// Check if availability window is long enough for at least one slot
	availableMinutes := r.EndTime.ToMinutes() - r.StartTime.ToMinutes()
	totalSlotDuration := r.DurationMinutes + r.BufferBeforeMinutes + r.BufferAfterMinutes
	if availableMinutes < totalSlotDuration {
		return fmt.Errorf("availability window (%d minutes) is too short for a single slot with buffers (%d minutes)",
			availableMinutes, totalSlotDuration)
	}

	return nil
}

// IsDeleted returns true if the rule is soft-deleted
func (r *AvailabilityRule) IsDeleted() bool {
	return r.DeletedAt.Valid
}

// TimeSlot represents an available time slot for booking
type TimeSlot struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Duration  int       `json:"duration_minutes"`
}

// CreateAvailabilityRuleInput represents input for creating an availability rule
type CreateAvailabilityRuleInput struct {
	ProviderID          string `json:"provider_id" binding:"required,max=255"`
	DayOfWeek           int    `json:"day_of_week" binding:"min=0,max=6"`
	StartTime           string `json:"start_time" binding:"required"`
	EndTime             string `json:"end_time" binding:"required"`
	DurationMinutes     int    `json:"duration_minutes" binding:"required,min=1,max=480"`
	SlotIntervalMinutes int    `json:"slot_interval_minutes" binding:"required,min=1"`
	BufferBeforeMinutes int    `json:"buffer_before_minutes" binding:"min=0,max=120"`
	BufferAfterMinutes  int    `json:"buffer_after_minutes" binding:"min=0,max=120"`
	Timezone            string `json:"timezone"`
}

// UpdateAvailabilityRuleInput represents input for updating an availability rule
type UpdateAvailabilityRuleInput struct {
	StartTime           *string `json:"start_time"`
	EndTime             *string `json:"end_time"`
	DurationMinutes     *int    `json:"duration_minutes" binding:"omitempty,min=1,max=480"`
	SlotIntervalMinutes *int    `json:"slot_interval_minutes" binding:"omitempty,min=1"`
	BufferBeforeMinutes *int    `json:"buffer_before_minutes" binding:"omitempty,min=0,max=120"`
	BufferAfterMinutes  *int    `json:"buffer_after_minutes" binding:"omitempty,min=0,max=120"`
	Timezone            *string `json:"timezone"`
	IsActive            *bool   `json:"is_active"`
}

// GetAvailableSlotsInput represents input for getting available slots
type GetAvailableSlotsInput struct {
	ProviderID string    `json:"provider_id" binding:"required"`
	Date       time.Time `json:"date" binding:"required"`
	Timezone   string    `json:"timezone"` // Optional: client's timezone for display
}
