package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/laith-ambianze/appointment-service/internal/models"
	"github.com/laith-ambianze/appointment-service/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// mockAvailabilityRepo is a mock implementation of AvailabilityRepository
type mockAvailabilityRepo struct {
	rules map[string]*models.AvailabilityRule
}

func newMockAvailabilityRepo() *mockAvailabilityRepo {
	return &mockAvailabilityRepo{
		rules: make(map[string]*models.AvailabilityRule),
	}
}

func (m *mockAvailabilityRepo) Create(ctx context.Context, rule *models.AvailabilityRule) error {
	key := rule.ProviderID + "_" + string(rune(rule.DayOfWeek))
	m.rules[key] = rule
	return nil
}

func (m *mockAvailabilityRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.AvailabilityRule, error) {
	for _, rule := range m.rules {
		if rule.ID == id {
			return rule, nil
		}
	}
	return nil, repository.ErrAvailabilityRuleNotFound
}

func (m *mockAvailabilityRepo) GetByProviderAndDay(ctx context.Context, productID uuid.UUID, providerID string, dayOfWeek models.DayOfWeek) (*models.AvailabilityRule, error) {
	key := providerID + "_" + string(rune(dayOfWeek))
	if rule, ok := m.rules[key]; ok && rule.ProductID == productID && rule.IsActive {
		return rule, nil
	}
	return nil, repository.ErrAvailabilityRuleNotFound
}

func (m *mockAvailabilityRepo) ListByProvider(ctx context.Context, productID uuid.UUID, providerID string) ([]models.AvailabilityRule, error) {
	var rules []models.AvailabilityRule
	for _, rule := range m.rules {
		if rule.ProductID == productID && rule.ProviderID == providerID {
			rules = append(rules, *rule)
		}
	}
	return rules, nil
}

func (m *mockAvailabilityRepo) Update(ctx context.Context, rule *models.AvailabilityRule) error {
	key := rule.ProviderID + "_" + string(rune(rule.DayOfWeek))
	m.rules[key] = rule
	return nil
}

func (m *mockAvailabilityRepo) Delete(ctx context.Context, id uuid.UUID) error {
	for key, rule := range m.rules {
		if rule.ID == id {
			delete(m.rules, key)
			return nil
		}
	}
	return repository.ErrAvailabilityRuleNotFound
}

func (m *mockAvailabilityRepo) DeleteByProviderAndDay(ctx context.Context, productID uuid.UUID, providerID string, dayOfWeek models.DayOfWeek) error {
	key := providerID + "_" + string(rune(dayOfWeek))
	if _, ok := m.rules[key]; ok {
		delete(m.rules, key)
		return nil
	}
	return repository.ErrAvailabilityRuleNotFound
}

// mockAppointmentRepoForAvailability is a mock implementation for appointment queries used by availability tests
type mockAppointmentRepoForAvailability struct {
	appointments []models.Appointment
}

func newMockAppointmentRepoForAvailability() *mockAppointmentRepoForAvailability {
	return &mockAppointmentRepoForAvailability{
		appointments: []models.Appointment{},
	}
}

func (m *mockAppointmentRepoForAvailability) Create(ctx context.Context, appointment *models.Appointment, participants []models.AppointmentParticipant) error {
	m.appointments = append(m.appointments, *appointment)
	return nil
}

func (m *mockAppointmentRepoForAvailability) CreateWithLock(ctx context.Context, appointment *models.Appointment, participants []models.AppointmentParticipant) error {
	// Check for overlapping appointments
	for _, existing := range m.appointments {
		if existing.ProviderID != nil && appointment.ProviderID != nil &&
			*existing.ProviderID == *appointment.ProviderID &&
			existing.Status != models.AppointmentStatusCancelled {
			// Check for overlap
			if appointment.StartTime.Before(existing.EndTime) && appointment.EndTime.After(existing.StartTime) {
				return ErrBookingConflict
			}
		}
	}
	m.appointments = append(m.appointments, *appointment)
	return nil
}

func (m *mockAppointmentRepoForAvailability) GetByID(ctx context.Context, id uuid.UUID) (*models.Appointment, error) {
	return nil, nil
}

func (m *mockAppointmentRepoForAvailability) GetByProductAndUser(ctx context.Context, productID uuid.UUID, externalUserID string, filters repository.AppointmentFilters) ([]models.Appointment, error) {
	return nil, nil
}

func (m *mockAppointmentRepoForAvailability) GetByDateRange(ctx context.Context, productID uuid.UUID, startTime, endTime time.Time) ([]models.Appointment, error) {
	return nil, nil
}

func (m *mockAppointmentRepoForAvailability) GetByProviderAndDateRange(ctx context.Context, productID uuid.UUID, providerID string, startTime, endTime time.Time) ([]models.Appointment, error) {
	var results []models.Appointment
	for _, apt := range m.appointments {
		if apt.ProductID == productID &&
			apt.ProviderID != nil && *apt.ProviderID == providerID &&
			apt.Status != models.AppointmentStatusCancelled &&
			apt.StartTime.Before(endTime) && apt.EndTime.After(startTime) {
			results = append(results, apt)
		}
	}
	return results, nil
}

func (m *mockAppointmentRepoForAvailability) Update(ctx context.Context, appointment *models.Appointment) error {
	return nil
}

func (m *mockAppointmentRepoForAvailability) UpdateStatus(ctx context.Context, id uuid.UUID, status models.AppointmentStatus) error {
	return nil
}

func (m *mockAppointmentRepoForAvailability) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockAppointmentRepoForAvailability) AddParticipant(ctx context.Context, participant *models.AppointmentParticipant) error {
	return nil
}

func (m *mockAppointmentRepoForAvailability) RemoveParticipant(ctx context.Context, appointmentID uuid.UUID, externalUserID string) error {
	return nil
}

func (m *mockAppointmentRepoForAvailability) UpdateParticipantStatus(ctx context.Context, appointmentID uuid.UUID, externalUserID string, status models.ParticipantStatus) error {
	return nil
}

func (m *mockAppointmentRepoForAvailability) GetParticipants(ctx context.Context, appointmentID uuid.UUID) ([]models.AppointmentParticipant, error) {
	return nil, nil
}

func newAvailabilityTestService() (*AvailabilityService, *mockAvailabilityRepo, *mockAppointmentRepoForAvailability) {
	logger, _ := zap.NewDevelopment()
	availRepo := newMockAvailabilityRepo()
	apptRepo := newMockAppointmentRepoForAvailability()
	svc := NewAvailabilityService(availRepo, apptRepo, logger)
	return svc, availRepo, apptRepo
}

func TestSlotGeneration(t *testing.T) {
	svc, availRepo, _ := newAvailabilityTestService()
	productID := uuid.New()
	providerID := "provider123"

	// Create availability rule: 9:00 - 17:00, 30 min duration, 15 min interval
	rule := &models.AvailabilityRule{
		ID:                  uuid.New(),
		ProductID:           productID,
		ProviderID:          providerID,
		DayOfWeek:           models.Monday,
		StartTime:           models.NewLocalTime(9, 0),
		EndTime:             models.NewLocalTime(17, 0),
		DurationMinutes:     30,
		SlotIntervalMinutes: 15,
		BufferBeforeMinutes: 0,
		BufferAfterMinutes:  0,
		Timezone:            "UTC",
		IsActive:            true,
	}
	availRepo.rules[providerID+"_"+string(rune(models.Monday))] = rule

	// Find a Monday date in the future
	now := time.Now()
	date := now.AddDate(0, 0, 1)
	for date.Weekday() != time.Monday {
		date = date.AddDate(0, 0, 1)
	}

	response, err := svc.GetAvailableSlots(context.Background(), productID, GetAvailableSlotsRequest{
		ProviderID: providerID,
		Date:       date.Format("2006-01-02"),
	})

	require.NoError(t, err)
	require.NotNil(t, response)

	// With 8 hours (9:00-17:00), 30 min duration, 15 min interval
	// Slots start at: 9:00, 9:15, 9:30, ..., until 16:30 (last slot ends at 17:00)
	// That's (16:30 - 9:00) / 15 + 1 = 450/15 + 1 = 31 slots
	assert.Equal(t, 31, len(response.Slots))
	assert.Equal(t, 30, response.DurationMinutes)

	// Verify first slot
	firstSlot := response.Slots[0]
	assert.Equal(t, 30, firstSlot.Duration)

	// Verify last slot ends at 17:00
	lastSlot := response.Slots[len(response.Slots)-1]
	expectedLastEnd := time.Date(date.Year(), date.Month(), date.Day(), 17, 0, 0, 0, time.UTC)
	assert.Equal(t, expectedLastEnd.Hour(), lastSlot.EndTime.Hour())
	assert.Equal(t, expectedLastEnd.Minute(), lastSlot.EndTime.Minute())
}

func TestSlotGenerationWithBuffers(t *testing.T) {
	svc, availRepo, apptRepo := newAvailabilityTestService()
	productID := uuid.New()
	providerID := "provider123"

	// Create availability rule with buffers
	rule := &models.AvailabilityRule{
		ID:                  uuid.New(),
		ProductID:           productID,
		ProviderID:          providerID,
		DayOfWeek:           models.Monday,
		StartTime:           models.NewLocalTime(9, 0),
		EndTime:             models.NewLocalTime(12, 0),
		DurationMinutes:     30,
		SlotIntervalMinutes: 30,
		BufferBeforeMinutes: 5,
		BufferAfterMinutes:  10,
		Timezone:            "UTC",
		IsActive:            true,
	}
	availRepo.rules[providerID+"_"+string(rune(models.Monday))] = rule

	// Find a Monday date in the future
	now := time.Now()
	date := now.AddDate(0, 0, 1)
	for date.Weekday() != time.Monday {
		date = date.AddDate(0, 0, 1)
	}

	// Add an existing appointment at 10:00-10:30
	existingAppt := models.Appointment{
		BaseModel: models.BaseModel{
			ID: uuid.New(),
		},
		ProductID:  productID,
		ProviderID: &providerID,
		StartTime:  time.Date(date.Year(), date.Month(), date.Day(), 10, 0, 0, 0, time.UTC),
		EndTime:    time.Date(date.Year(), date.Month(), date.Day(), 10, 30, 0, 0, time.UTC),
		Status:     models.AppointmentStatusScheduled,
	}
	apptRepo.appointments = append(apptRepo.appointments, existingAppt)

	response, err := svc.GetAvailableSlots(context.Background(), productID, GetAvailableSlotsRequest{
		ProviderID: providerID,
		Date:       date.Format("2006-01-02"),
	})

	require.NoError(t, err)
	require.NotNil(t, response)

	// The 10:00-10:30 appointment with 5 min buffer before and 10 min after
	// blocks the effective range: 9:55 - 10:40
	// So slots at 9:30 and 10:00 should be blocked
	// Available slots: 9:00, 11:00, 11:30
	for _, slot := range response.Slots {
		// Verify no slot overlaps with the blocked range
		// Blocked range considering buffers: appointment_start - buffer_before to appointment_end + buffer_after
		blockedStart := existingAppt.StartTime.Add(-5 * time.Minute) // 9:55
		blockedEnd := existingAppt.EndTime.Add(10 * time.Minute)     // 10:40

		slotEffectiveStart := slot.StartTime.Add(-5 * time.Minute)
		slotEffectiveEnd := slot.EndTime.Add(10 * time.Minute)

		// Check no overlap
		overlaps := slotEffectiveStart.Before(blockedEnd) && slotEffectiveEnd.After(blockedStart)
		assert.False(t, overlaps, "Slot %v should not overlap with blocked range", slot.StartTime)
	}
}

func TestOverlapDetection(t *testing.T) {
	svc, availRepo, apptRepo := newAvailabilityTestService()
	productID := uuid.New()
	providerID := "provider123"

	// Create availability rule
	rule := &models.AvailabilityRule{
		ID:                  uuid.New(),
		ProductID:           productID,
		ProviderID:          providerID,
		DayOfWeek:           models.Monday,
		StartTime:           models.NewLocalTime(9, 0),
		EndTime:             models.NewLocalTime(17, 0),
		DurationMinutes:     60,
		SlotIntervalMinutes: 60,
		BufferBeforeMinutes: 0,
		BufferAfterMinutes:  0,
		Timezone:            "UTC",
		IsActive:            true,
	}
	availRepo.rules[providerID+"_"+string(rune(models.Monday))] = rule

	// Find a Monday date in the future
	now := time.Now()
	date := now.AddDate(0, 0, 1)
	for date.Weekday() != time.Monday {
		date = date.AddDate(0, 0, 1)
	}

	// Add existing appointments
	appointments := []models.Appointment{
		{
			BaseModel:  models.BaseModel{ID: uuid.New()},
			ProductID:  productID,
			ProviderID: &providerID,
			StartTime:  time.Date(date.Year(), date.Month(), date.Day(), 10, 0, 0, 0, time.UTC),
			EndTime:    time.Date(date.Year(), date.Month(), date.Day(), 11, 0, 0, 0, time.UTC),
			Status:     models.AppointmentStatusScheduled,
		},
		{
			BaseModel:  models.BaseModel{ID: uuid.New()},
			ProductID:  productID,
			ProviderID: &providerID,
			StartTime:  time.Date(date.Year(), date.Month(), date.Day(), 14, 0, 0, 0, time.UTC),
			EndTime:    time.Date(date.Year(), date.Month(), date.Day(), 15, 0, 0, 0, time.UTC),
			Status:     models.AppointmentStatusScheduled,
		},
	}
	apptRepo.appointments = appointments

	response, err := svc.GetAvailableSlots(context.Background(), productID, GetAvailableSlotsRequest{
		ProviderID: providerID,
		Date:       date.Format("2006-01-02"),
	})

	require.NoError(t, err)
	require.NotNil(t, response)

	// Verify blocked times are not in available slots
	blockedTimes := []int{10, 14} // Hours that should be blocked
	for _, slot := range response.Slots {
		hour := slot.StartTime.Hour()
		for _, blocked := range blockedTimes {
			assert.NotEqual(t, blocked, hour, "Slot at hour %d should be blocked", blocked)
		}
	}
}

func TestNoAvailabilityForDay(t *testing.T) {
	svc, _, _ := newAvailabilityTestService()
	productID := uuid.New()
	providerID := "provider123"

	// No rules created - should return empty slots
	response, err := svc.GetAvailableSlots(context.Background(), productID, GetAvailableSlotsRequest{
		ProviderID: providerID,
		Date:       time.Now().AddDate(0, 0, 1).Format("2006-01-02"),
	})

	require.NoError(t, err)
	assert.Empty(t, response.Slots)
}

func TestPastDateRejection(t *testing.T) {
	svc, _, _ := newAvailabilityTestService()
	productID := uuid.New()

	// Try to get slots for yesterday
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	_, err := svc.GetAvailableSlots(context.Background(), productID, GetAvailableSlotsRequest{
		ProviderID: "provider123",
		Date:       yesterday,
	})

	assert.ErrorIs(t, err, ErrDateInPast)
}

func TestTimezoneConversion(t *testing.T) {
	svc, availRepo, _ := newAvailabilityTestService()
	productID := uuid.New()
	providerID := "provider123"

	// Create rule with New York timezone
	rule := &models.AvailabilityRule{
		ID:                  uuid.New(),
		ProductID:           productID,
		ProviderID:          providerID,
		DayOfWeek:           models.Monday,
		StartTime:           models.NewLocalTime(9, 0),  // 9 AM ET
		EndTime:             models.NewLocalTime(17, 0), // 5 PM ET
		DurationMinutes:     30,
		SlotIntervalMinutes: 30,
		BufferBeforeMinutes: 0,
		BufferAfterMinutes:  0,
		Timezone:            "America/New_York",
		IsActive:            true,
	}
	availRepo.rules[providerID+"_"+string(rune(models.Monday))] = rule

	// Find a Monday date in the future
	now := time.Now()
	date := now.AddDate(0, 0, 1)
	for date.Weekday() != time.Monday {
		date = date.AddDate(0, 0, 1)
	}

	response, err := svc.GetAvailableSlots(context.Background(), productID, GetAvailableSlotsRequest{
		ProviderID: providerID,
		Date:       date.Format("2006-01-02"),
	})

	require.NoError(t, err)
	require.NotNil(t, response)
	require.NotEmpty(t, response.Slots)

	// All times should be in UTC
	for _, slot := range response.Slots {
		assert.Equal(t, "UTC", slot.StartTime.Location().String())
		assert.Equal(t, "UTC", slot.EndTime.Location().String())
	}
}

func TestBookingConflictDetection(t *testing.T) {
	svc, availRepo, _ := newAvailabilityTestService()
	productID := uuid.New()
	providerID := "provider123"
	userID := "user123"

	// Create availability rule
	rule := &models.AvailabilityRule{
		ID:                  uuid.New(),
		ProductID:           productID,
		ProviderID:          providerID,
		DayOfWeek:           models.Monday,
		StartTime:           models.NewLocalTime(9, 0),
		EndTime:             models.NewLocalTime(17, 0),
		DurationMinutes:     60,
		SlotIntervalMinutes: 60,
		BufferBeforeMinutes: 0,
		BufferAfterMinutes:  0,
		Timezone:            "UTC",
		IsActive:            true,
	}
	availRepo.rules[providerID+"_"+string(rune(models.Monday))] = rule

	// Find a Monday date in the future
	now := time.Now()
	date := now.AddDate(0, 0, 1)
	for date.Weekday() != time.Monday {
		date = date.AddDate(0, 0, 1)
	}

	startTime := time.Date(date.Year(), date.Month(), date.Day(), 10, 0, 0, 0, time.UTC)

	// Book first appointment
	_, err := svc.BookAppointment(context.Background(), productID, userID, BookAppointmentRequest{
		ProviderID: providerID,
		StartTime:  startTime,
		Title:      "First Appointment",
		Participants: []ParticipantRequest{
			{ExternalUserID: userID, Role: models.ParticipantRoleGuest},
		},
	})
	require.NoError(t, err)

	// Try to book conflicting appointment
	_, err = svc.BookAppointment(context.Background(), productID, "user456", BookAppointmentRequest{
		ProviderID: providerID,
		StartTime:  startTime,
		Title:      "Conflicting Appointment",
		Participants: []ParticipantRequest{
			{ExternalUserID: "user456", Role: models.ParticipantRoleGuest},
		},
	})

	assert.ErrorIs(t, err, ErrBookingConflict)
}

func TestBookingPastTimeRejection(t *testing.T) {
	svc, availRepo, _ := newAvailabilityTestService()
	productID := uuid.New()
	providerID := "provider123"
	userID := "user123"

	// Create rule for today's day of week
	today := time.Now()
	dayOfWeek := models.DayOfWeek(today.Weekday())

	rule := &models.AvailabilityRule{
		ID:                  uuid.New(),
		ProductID:           productID,
		ProviderID:          providerID,
		DayOfWeek:           dayOfWeek,
		StartTime:           models.NewLocalTime(0, 0),
		EndTime:             models.NewLocalTime(23, 59),
		DurationMinutes:     60,
		SlotIntervalMinutes: 60,
		BufferBeforeMinutes: 0,
		BufferAfterMinutes:  0,
		Timezone:            "UTC",
		IsActive:            true,
	}
	availRepo.rules[providerID+"_"+string(rune(dayOfWeek))] = rule

	// Try to book in the past
	pastTime := time.Now().Add(-2 * time.Hour)

	_, err := svc.BookAppointment(context.Background(), productID, userID, BookAppointmentRequest{
		ProviderID: providerID,
		StartTime:  pastTime,
		Title:      "Past Appointment",
		Participants: []ParticipantRequest{
			{ExternalUserID: userID, Role: models.ParticipantRoleGuest},
		},
	})

	assert.ErrorIs(t, err, ErrDateInPast)
}

func TestSlotValidation(t *testing.T) {
	svc, availRepo, _ := newAvailabilityTestService()
	productID := uuid.New()
	providerID := "provider123"
	userID := "user123"

	// Create availability rule: 9:00-17:00, 60 min duration, 60 min interval
	rule := &models.AvailabilityRule{
		ID:                  uuid.New(),
		ProductID:           productID,
		ProviderID:          providerID,
		DayOfWeek:           models.Monday,
		StartTime:           models.NewLocalTime(9, 0),
		EndTime:             models.NewLocalTime(17, 0),
		DurationMinutes:     60,
		SlotIntervalMinutes: 60,
		BufferBeforeMinutes: 0,
		BufferAfterMinutes:  0,
		Timezone:            "UTC",
		IsActive:            true,
	}
	availRepo.rules[providerID+"_"+string(rune(models.Monday))] = rule

	// Find a Monday date in the future
	now := time.Now()
	date := now.AddDate(0, 0, 1)
	for date.Weekday() != time.Monday {
		date = date.AddDate(0, 0, 1)
	}

	tests := []struct {
		name      string
		startTime time.Time
		wantErr   error
	}{
		{
			name:      "valid slot at 9:00",
			startTime: time.Date(date.Year(), date.Month(), date.Day(), 9, 0, 0, 0, time.UTC),
			wantErr:   nil,
		},
		{
			name:      "slot before availability window",
			startTime: time.Date(date.Year(), date.Month(), date.Day(), 8, 0, 0, 0, time.UTC),
			wantErr:   ErrSlotNotAvailable,
		},
		{
			name:      "slot after availability window",
			startTime: time.Date(date.Year(), date.Month(), date.Day(), 17, 0, 0, 0, time.UTC),
			wantErr:   ErrSlotNotAvailable,
		},
		{
			name:      "slot not aligned with interval",
			startTime: time.Date(date.Year(), date.Month(), date.Day(), 9, 30, 0, 0, time.UTC),
			wantErr:   ErrInvalidSlotTime,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.BookAppointment(context.Background(), productID, userID, BookAppointmentRequest{
				ProviderID: providerID,
				StartTime:  tt.startTime,
				Title:      "Test Appointment",
				Participants: []ParticipantRequest{
					{ExternalUserID: userID, Role: models.ParticipantRoleGuest},
				},
			})

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCancelledAppointmentsIgnored(t *testing.T) {
	svc, availRepo, apptRepo := newAvailabilityTestService()
	productID := uuid.New()
	providerID := "provider123"

	// Create availability rule
	rule := &models.AvailabilityRule{
		ID:                  uuid.New(),
		ProductID:           productID,
		ProviderID:          providerID,
		DayOfWeek:           models.Monday,
		StartTime:           models.NewLocalTime(9, 0),
		EndTime:             models.NewLocalTime(17, 0),
		DurationMinutes:     60,
		SlotIntervalMinutes: 60,
		BufferBeforeMinutes: 0,
		BufferAfterMinutes:  0,
		Timezone:            "UTC",
		IsActive:            true,
	}
	availRepo.rules[providerID+"_"+string(rune(models.Monday))] = rule

	// Find a Monday date in the future
	now := time.Now()
	date := now.AddDate(0, 0, 1)
	for date.Weekday() != time.Monday {
		date = date.AddDate(0, 0, 1)
	}

	// Add a CANCELLED appointment at 10:00
	cancelledAppt := models.Appointment{
		BaseModel:  models.BaseModel{ID: uuid.New()},
		ProductID:  productID,
		ProviderID: &providerID,
		StartTime:  time.Date(date.Year(), date.Month(), date.Day(), 10, 0, 0, 0, time.UTC),
		EndTime:    time.Date(date.Year(), date.Month(), date.Day(), 11, 0, 0, 0, time.UTC),
		Status:     models.AppointmentStatusCancelled, // Cancelled!
	}
	apptRepo.appointments = append(apptRepo.appointments, cancelledAppt)

	response, err := svc.GetAvailableSlots(context.Background(), productID, GetAvailableSlotsRequest{
		ProviderID: providerID,
		Date:       date.Format("2006-01-02"),
	})

	require.NoError(t, err)

	// The 10:00 slot should be available since the appointment is cancelled
	has10AM := false
	for _, slot := range response.Slots {
		if slot.StartTime.Hour() == 10 && slot.StartTime.Minute() == 0 {
			has10AM = true
			break
		}
	}
	assert.True(t, has10AM, "10:00 slot should be available since appointment is cancelled")
}

func TestSortSlotsByStartTime(t *testing.T) {
	slots := []models.TimeSlot{
		{StartTime: time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC)},
		{StartTime: time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC)},
		{StartTime: time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC)},
	}

	SortSlotsByStartTime(slots)

	assert.Equal(t, 9, slots[0].StartTime.Hour())
	assert.Equal(t, 11, slots[1].StartTime.Hour())
	assert.Equal(t, 14, slots[2].StartTime.Hour())
}

// --- Break overlap / filterBreakSlots tests ---

func TestFilterBreakSlots(t *testing.T) {
	svc, _, _ := newAvailabilityTestService()

	// Base date: some future Monday used to anchor slot times
	date := time.Date(2030, 1, 7, 0, 0, 0, 0, time.UTC) // Monday
	loc := time.UTC

	makeSlot := func(startHour, startMin, endHour, endMin int) models.TimeSlot {
		return models.TimeSlot{
			StartTime: time.Date(date.Year(), date.Month(), date.Day(), startHour, startMin, 0, 0, loc),
			EndTime:   time.Date(date.Year(), date.Month(), date.Day(), endHour, endMin, 0, 0, loc),
			Duration:  30,
		}
	}

	breakWindow := []models.AvailabilityBreak{
		{
			StartTime: models.NewLocalTime(11, 0),
			EndTime:   models.NewLocalTime(12, 0),
		},
	}

	tests := []struct {
		name     string
		slot     models.TimeSlot
		wantKeep bool
	}{
		{
			name:     "slot entirely before break",
			slot:     makeSlot(9, 0, 9, 30),
			wantKeep: true,
		},
		{
			name:     "slot entirely after break",
			slot:     makeSlot(12, 0, 12, 30),
			wantKeep: true,
		},
		{
			name:     "slot fully inside break",
			slot:     makeSlot(11, 15, 11, 45),
			wantKeep: false,
		},
		{
			name:     "slot partially overlapping break start",
			slot:     makeSlot(10, 30, 11, 30),
			wantKeep: false,
		},
		{
			name:     "slot partially overlapping break end",
			slot:     makeSlot(11, 30, 12, 30),
			wantKeep: false,
		},
		{
			name:     "slot spanning entire break",
			slot:     makeSlot(10, 0, 13, 0),
			wantKeep: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.filterBreakSlots([]models.TimeSlot{tt.slot}, breakWindow, date, loc)
			if tt.wantKeep {
				assert.Len(t, result, 1, "slot should be kept")
			} else {
				assert.Len(t, result, 0, "slot should be filtered out")
			}
		})
	}

	t.Run("no breaks returns all slots", func(t *testing.T) {
		slots := []models.TimeSlot{makeSlot(9, 0, 9, 30), makeSlot(11, 0, 11, 30)}
		result := svc.filterBreakSlots(slots, nil, date, loc)
		assert.Len(t, result, 2)
	})
}

func TestSlotGenerationWithBreak(t *testing.T) {
	svc, availRepo, _ := newAvailabilityTestService()
	productID := uuid.New()
	providerID := "provider-breaks"

	// Rule: 09:00-17:00, 30 min duration, 30 min interval, break 11:00-12:00
	rule := &models.AvailabilityRule{
		ID:                  uuid.New(),
		ProductID:           productID,
		ProviderID:          providerID,
		DayOfWeek:           models.Monday,
		StartTime:           models.NewLocalTime(9, 0),
		EndTime:             models.NewLocalTime(17, 0),
		DurationMinutes:     30,
		SlotIntervalMinutes: 30,
		BufferBeforeMinutes: 0,
		BufferAfterMinutes:  0,
		Timezone:            "UTC",
		IsActive:            true,
		Breaks: []models.AvailabilityBreak{
			{
				StartTime: models.NewLocalTime(11, 0),
				EndTime:   models.NewLocalTime(12, 0),
			},
		},
	}
	availRepo.rules[providerID+"_"+string(rune(models.Monday))] = rule

	// Find a future Monday
	date := time.Now().AddDate(0, 0, 1)
	for date.Weekday() != time.Monday {
		date = date.AddDate(0, 0, 1)
	}

	response, err := svc.GetAvailableSlots(context.Background(), productID, GetAvailableSlotsRequest{
		ProviderID: providerID,
		Date:       date.Format("2006-01-02"),
	})

	require.NoError(t, err)
	require.NotNil(t, response)

	// No slot should start or end inside the 11:00-12:00 break
	breakStart := time.Date(date.Year(), date.Month(), date.Day(), 11, 0, 0, 0, time.UTC)
	breakEnd := time.Date(date.Year(), date.Month(), date.Day(), 12, 0, 0, 0, time.UTC)

	for _, slot := range response.Slots {
		overlaps := slot.StartTime.Before(breakEnd) && slot.EndTime.After(breakStart)
		assert.False(t, overlaps, "slot %v-%v must not overlap break 11:00-12:00", slot.StartTime, slot.EndTime)
	}

	// Slots at 11:00-11:30, 11:30-12:00 should be absent; 12:00-12:30 should be present
	has1100 := false
	has1200 := false
	for _, slot := range response.Slots {
		if slot.StartTime.Equal(time.Date(date.Year(), date.Month(), date.Day(), 11, 0, 0, 0, time.UTC)) {
			has1100 = true
		}
		if slot.StartTime.Equal(time.Date(date.Year(), date.Month(), date.Day(), 12, 0, 0, 0, time.UTC)) {
			has1200 = true
		}
	}
	assert.False(t, has1100, "11:00 slot should not exist during break")
	assert.True(t, has1200, "12:00 slot should exist after break")
}

func TestParseAndValidateBreaks(t *testing.T) {
	ruleStart := models.NewLocalTime(9, 0)
	ruleEnd := models.NewLocalTime(17, 0)

	tests := []struct {
		name    string
		reqs    []BreakRequest
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid single break",
			reqs:    []BreakRequest{{StartTime: "11:00", EndTime: "12:00"}},
			wantErr: false,
		},
		{
			name:    "no breaks is valid",
			reqs:    nil,
			wantErr: false,
		},
		{
			name:    "end before start",
			reqs:    []BreakRequest{{StartTime: "12:00", EndTime: "11:00"}},
			wantErr: true,
		},
		{
			name:    "break starts before rule start",
			reqs:    []BreakRequest{{StartTime: "08:00", EndTime: "09:30"}},
			wantErr: true,
		},
		{
			name:    "break ends after rule end",
			reqs:    []BreakRequest{{StartTime: "16:30", EndTime: "17:30"}},
			wantErr: true,
		},
		{
			name: "overlapping breaks",
			reqs: []BreakRequest{
				{StartTime: "11:00", EndTime: "12:00"},
				{StartTime: "11:30", EndTime: "13:00"},
			},
			wantErr: true,
		},
		{
			name:    "more than 10 breaks",
			reqs:    make([]BreakRequest, 11),
			wantErr: true,
		},
		{
			name:    "invalid start_time format",
			reqs:    []BreakRequest{{StartTime: "bad", EndTime: "12:00"}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseAndValidateBreaks(tt.reqs, ruleStart, ruleEnd)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
