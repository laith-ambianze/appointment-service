package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalTime(t *testing.T) {
	t.Run("NewLocalTime creates correct time", func(t *testing.T) {
		lt := NewLocalTime(9, 30)
		assert.Equal(t, 9, lt.Hour)
		assert.Equal(t, 30, lt.Minute)
		assert.Equal(t, 0, lt.Second)
	})

	t.Run("ParseLocalTime with HH:MM format", func(t *testing.T) {
		lt, err := ParseLocalTime("09:30")
		require.NoError(t, err)
		assert.Equal(t, 9, lt.Hour)
		assert.Equal(t, 30, lt.Minute)
		assert.Equal(t, 0, lt.Second)
	})

	t.Run("ParseLocalTime with HH:MM:SS format", func(t *testing.T) {
		lt, err := ParseLocalTime("14:45:30")
		require.NoError(t, err)
		assert.Equal(t, 14, lt.Hour)
		assert.Equal(t, 45, lt.Minute)
		assert.Equal(t, 30, lt.Second)
	})

	t.Run("ParseLocalTime with invalid format", func(t *testing.T) {
		_, err := ParseLocalTime("invalid")
		assert.Error(t, err)
	})

	t.Run("ParseLocalTime with invalid hour", func(t *testing.T) {
		_, err := ParseLocalTime("25:00")
		assert.Error(t, err)
	})

	t.Run("ParseLocalTime with invalid minute", func(t *testing.T) {
		_, err := ParseLocalTime("10:60")
		assert.Error(t, err)
	})

	t.Run("String returns correct format", func(t *testing.T) {
		lt := LocalTime{Hour: 9, Minute: 5, Second: 3}
		assert.Equal(t, "09:05:03", lt.String())
	})

	t.Run("ToMinutes calculates correctly", func(t *testing.T) {
		lt := NewLocalTime(2, 30)
		assert.Equal(t, 150, lt.ToMinutes())
	})

	t.Run("After comparison", func(t *testing.T) {
		lt1 := NewLocalTime(10, 0)
		lt2 := NewLocalTime(9, 0)
		assert.True(t, lt1.After(lt2))
		assert.False(t, lt2.After(lt1))
	})

	t.Run("Before comparison", func(t *testing.T) {
		lt1 := NewLocalTime(9, 0)
		lt2 := NewLocalTime(10, 0)
		assert.True(t, lt1.Before(lt2))
		assert.False(t, lt2.Before(lt1))
	})

	t.Run("Equal comparison", func(t *testing.T) {
		lt1 := NewLocalTime(9, 30)
		lt2 := NewLocalTime(9, 30)
		lt3 := NewLocalTime(9, 31)
		assert.True(t, lt1.Equal(lt2))
		assert.False(t, lt1.Equal(lt3))
	})

	t.Run("ToTimeOnDate converts correctly", func(t *testing.T) {
		lt := NewLocalTime(14, 30)
		date := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
		loc, _ := time.LoadLocation("America/New_York")

		result := lt.ToTimeOnDate(date, loc)

		assert.Equal(t, 2024, result.Year())
		assert.Equal(t, time.March, result.Month())
		assert.Equal(t, 15, result.Day())
		assert.Equal(t, 14, result.Hour())
		assert.Equal(t, 30, result.Minute())
		assert.Equal(t, "America/New_York", result.Location().String())
	})
}

func TestDayOfWeek(t *testing.T) {
	t.Run("String returns correct day name", func(t *testing.T) {
		assert.Equal(t, "Sunday", Sunday.String())
		assert.Equal(t, "Monday", Monday.String())
		assert.Equal(t, "Tuesday", Tuesday.String())
		assert.Equal(t, "Wednesday", Wednesday.String())
		assert.Equal(t, "Thursday", Thursday.String())
		assert.Equal(t, "Friday", Friday.String())
		assert.Equal(t, "Saturday", Saturday.String())
	})

	t.Run("Invalid day returns Invalid", func(t *testing.T) {
		assert.Equal(t, "Invalid", DayOfWeek(-1).String())
		assert.Equal(t, "Invalid", DayOfWeek(7).String())
	})

	t.Run("IsValid checks correctly", func(t *testing.T) {
		assert.True(t, Sunday.IsValid())
		assert.True(t, Saturday.IsValid())
		assert.False(t, DayOfWeek(-1).IsValid())
		assert.False(t, DayOfWeek(7).IsValid())
	})
}

func TestAvailabilityRuleValidate(t *testing.T) {
	productID := uuid.New()

	tests := []struct {
		name        string
		rule        *AvailabilityRule
		wantErr     bool
		errContains string
	}{
		{
			name: "valid rule",
			rule: &AvailabilityRule{
				ProductID:           productID,
				ProviderID:          "provider123",
				DayOfWeek:           Monday,
				StartTime:           NewLocalTime(9, 0),
				EndTime:             NewLocalTime(17, 0),
				DurationMinutes:     30,
				SlotIntervalMinutes: 15,
				BufferBeforeMinutes: 5,
				BufferAfterMinutes:  10,
				Timezone:            "UTC",
			},
			wantErr: false,
		},
		{
			name: "missing product ID",
			rule: &AvailabilityRule{
				ProviderID:          "provider123",
				DayOfWeek:           Monday,
				StartTime:           NewLocalTime(9, 0),
				EndTime:             NewLocalTime(17, 0),
				DurationMinutes:     30,
				SlotIntervalMinutes: 15,
				Timezone:            "UTC",
			},
			wantErr:     true,
			errContains: "product ID is required",
		},
		{
			name: "missing provider ID",
			rule: &AvailabilityRule{
				ProductID:           productID,
				DayOfWeek:           Monday,
				StartTime:           NewLocalTime(9, 0),
				EndTime:             NewLocalTime(17, 0),
				DurationMinutes:     30,
				SlotIntervalMinutes: 15,
				Timezone:            "UTC",
			},
			wantErr:     true,
			errContains: "provider ID is required",
		},
		{
			name: "provider ID too long",
			rule: &AvailabilityRule{
				ProductID:           productID,
				ProviderID:          string(make([]byte, 256)),
				DayOfWeek:           Monday,
				StartTime:           NewLocalTime(9, 0),
				EndTime:             NewLocalTime(17, 0),
				DurationMinutes:     30,
				SlotIntervalMinutes: 15,
				Timezone:            "UTC",
			},
			wantErr:     true,
			errContains: "must be less than 255 characters",
		},
		{
			name: "invalid day of week",
			rule: &AvailabilityRule{
				ProductID:           productID,
				ProviderID:          "provider123",
				DayOfWeek:           DayOfWeek(7),
				StartTime:           NewLocalTime(9, 0),
				EndTime:             NewLocalTime(17, 0),
				DurationMinutes:     30,
				SlotIntervalMinutes: 15,
				Timezone:            "UTC",
			},
			wantErr:     true,
			errContains: "invalid day of week",
		},
		{
			name: "end time before start time",
			rule: &AvailabilityRule{
				ProductID:           productID,
				ProviderID:          "provider123",
				DayOfWeek:           Monday,
				StartTime:           NewLocalTime(17, 0),
				EndTime:             NewLocalTime(9, 0),
				DurationMinutes:     30,
				SlotIntervalMinutes: 15,
				Timezone:            "UTC",
			},
			wantErr:     true,
			errContains: "end time must be after start time",
		},
		{
			name: "zero duration",
			rule: &AvailabilityRule{
				ProductID:           productID,
				ProviderID:          "provider123",
				DayOfWeek:           Monday,
				StartTime:           NewLocalTime(9, 0),
				EndTime:             NewLocalTime(17, 0),
				DurationMinutes:     0,
				SlotIntervalMinutes: 15,
				Timezone:            "UTC",
			},
			wantErr:     true,
			errContains: "duration must be greater than 0",
		},
		{
			name: "duration too long",
			rule: &AvailabilityRule{
				ProductID:           productID,
				ProviderID:          "provider123",
				DayOfWeek:           Monday,
				StartTime:           NewLocalTime(9, 0),
				EndTime:             NewLocalTime(17, 0),
				DurationMinutes:     500,
				SlotIntervalMinutes: 15,
				Timezone:            "UTC",
			},
			wantErr:     true,
			errContains: "duration must not exceed 480 minutes",
		},
		{
			name: "zero slot interval",
			rule: &AvailabilityRule{
				ProductID:           productID,
				ProviderID:          "provider123",
				DayOfWeek:           Monday,
				StartTime:           NewLocalTime(9, 0),
				EndTime:             NewLocalTime(17, 0),
				DurationMinutes:     30,
				SlotIntervalMinutes: 0,
				Timezone:            "UTC",
			},
			wantErr:     true,
			errContains: "slot interval must be greater than 0",
		},
		{
			name: "slot interval exceeds duration",
			rule: &AvailabilityRule{
				ProductID:           productID,
				ProviderID:          "provider123",
				DayOfWeek:           Monday,
				StartTime:           NewLocalTime(9, 0),
				EndTime:             NewLocalTime(17, 0),
				DurationMinutes:     30,
				SlotIntervalMinutes: 60,
				Timezone:            "UTC",
			},
			wantErr:     true,
			errContains: "slot interval cannot exceed duration",
		},
		{
			name: "negative buffer before",
			rule: &AvailabilityRule{
				ProductID:           productID,
				ProviderID:          "provider123",
				DayOfWeek:           Monday,
				StartTime:           NewLocalTime(9, 0),
				EndTime:             NewLocalTime(17, 0),
				DurationMinutes:     30,
				SlotIntervalMinutes: 15,
				BufferBeforeMinutes: -5,
				Timezone:            "UTC",
			},
			wantErr:     true,
			errContains: "buffer before must be non-negative",
		},
		{
			name: "buffer before too large",
			rule: &AvailabilityRule{
				ProductID:           productID,
				ProviderID:          "provider123",
				DayOfWeek:           Monday,
				StartTime:           NewLocalTime(9, 0),
				EndTime:             NewLocalTime(17, 0),
				DurationMinutes:     30,
				SlotIntervalMinutes: 15,
				BufferBeforeMinutes: 150,
				Timezone:            "UTC",
			},
			wantErr:     true,
			errContains: "buffer before must not exceed 120 minutes",
		},
		{
			name: "negative buffer after",
			rule: &AvailabilityRule{
				ProductID:           productID,
				ProviderID:          "provider123",
				DayOfWeek:           Monday,
				StartTime:           NewLocalTime(9, 0),
				EndTime:             NewLocalTime(17, 0),
				DurationMinutes:     30,
				SlotIntervalMinutes: 15,
				BufferAfterMinutes:  -5,
				Timezone:            "UTC",
			},
			wantErr:     true,
			errContains: "buffer after must be non-negative",
		},
		{
			name: "invalid timezone",
			rule: &AvailabilityRule{
				ProductID:           productID,
				ProviderID:          "provider123",
				DayOfWeek:           Monday,
				StartTime:           NewLocalTime(9, 0),
				EndTime:             NewLocalTime(17, 0),
				DurationMinutes:     30,
				SlotIntervalMinutes: 15,
				Timezone:            "Invalid/Timezone",
			},
			wantErr:     true,
			errContains: "invalid timezone",
		},
		{
			name: "availability window too small for slot with buffers",
			rule: &AvailabilityRule{
				ProductID:           productID,
				ProviderID:          "provider123",
				DayOfWeek:           Monday,
				StartTime:           NewLocalTime(9, 0),
				EndTime:             NewLocalTime(9, 30),
				DurationMinutes:     30,
				SlotIntervalMinutes: 15,
				BufferBeforeMinutes: 15,
				BufferAfterMinutes:  15,
				Timezone:            "UTC",
			},
			wantErr:     true,
			errContains: "availability window",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rule.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAvailabilityRuleDefaultTimezone(t *testing.T) {
	rule := &AvailabilityRule{
		ProductID:           uuid.New(),
		ProviderID:          "provider123",
		DayOfWeek:           Monday,
		StartTime:           NewLocalTime(9, 0),
		EndTime:             NewLocalTime(17, 0),
		DurationMinutes:     30,
		SlotIntervalMinutes: 15,
		Timezone:            "",
	}

	err := rule.Validate()
	require.NoError(t, err)
	assert.Equal(t, "UTC", rule.Timezone)
}

func TestTimeSlot(t *testing.T) {
	start := time.Date(2024, 3, 15, 9, 0, 0, 0, time.UTC)
	end := time.Date(2024, 3, 15, 9, 30, 0, 0, time.UTC)

	slot := TimeSlot{
		StartTime: start,
		EndTime:   end,
		Duration:  30,
	}

	assert.Equal(t, start, slot.StartTime)
	assert.Equal(t, end, slot.EndTime)
	assert.Equal(t, 30, slot.Duration)
}
