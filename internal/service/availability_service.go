package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/laith-ambianze/appointment-service/internal/models"
	"github.com/laith-ambianze/appointment-service/internal/repository"
	"go.uber.org/zap"
)

// Availability service errors
var (
	ErrProviderNotFound           = errors.New("provider not found")
	ErrAvailabilityRuleNotFound   = errors.New("availability rule not found for this day")
	ErrSlotNotAvailable           = errors.New("requested slot is not available")
	ErrInvalidSlotTime            = errors.New("invalid slot time")
	ErrDateInPast                 = errors.New("cannot book appointments in the past")
	ErrBookingConflict            = errors.New("booking conflict: time slot is already taken")
	ErrInvalidTimezone            = errors.New("invalid timezone")
	ErrAvailabilityWindowTooSmall = errors.New("availability window is too small for the requested duration")
)

// AvailabilityService handles availability and booking business logic
type AvailabilityService struct {
	availabilityRepo repository.AvailabilityRepository
	appointmentRepo  repository.AppointmentRepository
	logger           *zap.Logger
}

// NewAvailabilityService creates a new availability service
func NewAvailabilityService(
	availabilityRepo repository.AvailabilityRepository,
	appointmentRepo repository.AppointmentRepository,
	logger *zap.Logger,
) *AvailabilityService {
	return &AvailabilityService{
		availabilityRepo: availabilityRepo,
		appointmentRepo:  appointmentRepo,
		logger:           logger,
	}
}

// CreateAvailabilityRuleRequest contains data for creating an availability rule
type CreateAvailabilityRuleRequest struct {
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

// UpdateAvailabilityRuleRequest contains data for updating an availability rule
type UpdateAvailabilityRuleRequest struct {
	StartTime           *string `json:"start_time"`
	EndTime             *string `json:"end_time"`
	DurationMinutes     *int    `json:"duration_minutes" binding:"omitempty,min=1,max=480"`
	SlotIntervalMinutes *int    `json:"slot_interval_minutes" binding:"omitempty,min=1"`
	BufferBeforeMinutes *int    `json:"buffer_before_minutes" binding:"omitempty,min=0,max=120"`
	BufferAfterMinutes  *int    `json:"buffer_after_minutes" binding:"omitempty,min=0,max=120"`
	Timezone            *string `json:"timezone"`
	IsActive            *bool   `json:"is_active"`
}

// GetAvailableSlotsRequest contains data for getting available slots
type GetAvailableSlotsRequest struct {
	ProviderID string `json:"provider_id" binding:"required"`
	Date       string `json:"date" binding:"required"` // YYYY-MM-DD format
	Timezone   string `json:"timezone"`                // Optional client timezone
}

// BookAppointmentRequest contains data for booking an appointment
type BookAppointmentRequest struct {
	ProviderID   string                 `json:"provider_id" binding:"required"`
	StartTime    time.Time              `json:"start_time" binding:"required"`
	Title        string                 `json:"title" binding:"required,min=1,max=255"`
	Description  string                 `json:"description" binding:"max=2000"`
	Location     string                 `json:"location" binding:"max=500"`
	Timezone     string                 `json:"timezone"`
	Metadata     map[string]interface{} `json:"metadata"`
	Participants []ParticipantRequest   `json:"participants" binding:"required,min=1,dive"`
}

// AvailableSlotsResponse contains available slots for a provider on a date
type AvailableSlotsResponse struct {
	ProviderID      string            `json:"provider_id"`
	Date            string            `json:"date"`
	Timezone        string            `json:"timezone"`
	DurationMinutes int               `json:"duration_minutes"`
	Slots           []models.TimeSlot `json:"slots"`
}

// CreateAvailabilityRule creates a new availability rule for a provider
func (s *AvailabilityService) CreateAvailabilityRule(ctx context.Context, productID uuid.UUID, req CreateAvailabilityRuleRequest) (*models.AvailabilityRule, error) {
	s.logger.Debug("Creating availability rule",
		zap.String("product_id", productID.String()),
		zap.String("provider_id", req.ProviderID),
		zap.Int("day_of_week", req.DayOfWeek),
	)

	// Parse times
	startTime, err := models.ParseLocalTime(req.StartTime)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid start_time: %v", ErrInvalidInput, err)
	}

	endTime, err := models.ParseLocalTime(req.EndTime)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid end_time: %v", ErrInvalidInput, err)
	}

	// Default timezone
	timezone := req.Timezone
	if timezone == "" {
		timezone = "UTC"
	}

	// Validate timezone
	_, err = time.LoadLocation(timezone)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidTimezone, timezone)
	}

	rule := &models.AvailabilityRule{
		ID:                  uuid.New(),
		ProductID:           productID,
		ProviderID:          req.ProviderID,
		DayOfWeek:           models.DayOfWeek(req.DayOfWeek),
		StartTime:           startTime,
		EndTime:             endTime,
		DurationMinutes:     req.DurationMinutes,
		SlotIntervalMinutes: req.SlotIntervalMinutes,
		BufferBeforeMinutes: req.BufferBeforeMinutes,
		BufferAfterMinutes:  req.BufferAfterMinutes,
		Timezone:            timezone,
		IsActive:            true,
	}

	// Validate rule
	if err := rule.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	if err := s.availabilityRepo.Create(ctx, rule); err != nil {
		if errors.Is(err, repository.ErrDuplicateAvailabilityRule) {
			return nil, fmt.Errorf("%w: rule already exists for this provider on %s", ErrInvalidInput, rule.DayOfWeek.String())
		}
		return nil, err
	}

	return rule, nil
}

// GetAvailabilityRule gets an availability rule by ID
func (s *AvailabilityService) GetAvailabilityRule(ctx context.Context, productID uuid.UUID, ruleID uuid.UUID) (*models.AvailabilityRule, error) {
	rule, err := s.availabilityRepo.GetByID(ctx, ruleID)
	if err != nil {
		if errors.Is(err, repository.ErrAvailabilityRuleNotFound) {
			return nil, ErrAvailabilityRuleNotFound
		}
		return nil, err
	}

	// Verify product ownership
	if rule.ProductID != productID {
		return nil, ErrAvailabilityRuleNotFound
	}

	return rule, nil
}

// ListAvailabilityRules lists all availability rules for a provider
func (s *AvailabilityService) ListAvailabilityRules(ctx context.Context, productID uuid.UUID, providerID string) ([]models.AvailabilityRule, error) {
	return s.availabilityRepo.ListByProvider(ctx, productID, providerID)
}

// UpdateAvailabilityRule updates an existing availability rule
func (s *AvailabilityService) UpdateAvailabilityRule(ctx context.Context, productID uuid.UUID, ruleID uuid.UUID, req UpdateAvailabilityRuleRequest) (*models.AvailabilityRule, error) {
	rule, err := s.availabilityRepo.GetByID(ctx, ruleID)
	if err != nil {
		if errors.Is(err, repository.ErrAvailabilityRuleNotFound) {
			return nil, ErrAvailabilityRuleNotFound
		}
		return nil, err
	}

	// Verify product ownership
	if rule.ProductID != productID {
		return nil, ErrAvailabilityRuleNotFound
	}

	// Apply updates
	if req.StartTime != nil {
		startTime, err := models.ParseLocalTime(*req.StartTime)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid start_time: %v", ErrInvalidInput, err)
		}
		rule.StartTime = startTime
	}

	if req.EndTime != nil {
		endTime, err := models.ParseLocalTime(*req.EndTime)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid end_time: %v", ErrInvalidInput, err)
		}
		rule.EndTime = endTime
	}

	if req.DurationMinutes != nil {
		rule.DurationMinutes = *req.DurationMinutes
	}

	if req.SlotIntervalMinutes != nil {
		rule.SlotIntervalMinutes = *req.SlotIntervalMinutes
	}

	if req.BufferBeforeMinutes != nil {
		rule.BufferBeforeMinutes = *req.BufferBeforeMinutes
	}

	if req.BufferAfterMinutes != nil {
		rule.BufferAfterMinutes = *req.BufferAfterMinutes
	}

	if req.Timezone != nil {
		_, err = time.LoadLocation(*req.Timezone)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrInvalidTimezone, *req.Timezone)
		}
		rule.Timezone = *req.Timezone
	}

	if req.IsActive != nil {
		rule.IsActive = *req.IsActive
	}

	// Validate updated rule
	if err := rule.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	if err := s.availabilityRepo.Update(ctx, rule); err != nil {
		return nil, err
	}

	return rule, nil
}

// DeleteAvailabilityRule deletes an availability rule
func (s *AvailabilityService) DeleteAvailabilityRule(ctx context.Context, productID uuid.UUID, ruleID uuid.UUID) error {
	rule, err := s.availabilityRepo.GetByID(ctx, ruleID)
	if err != nil {
		if errors.Is(err, repository.ErrAvailabilityRuleNotFound) {
			return ErrAvailabilityRuleNotFound
		}
		return err
	}

	// Verify product ownership
	if rule.ProductID != productID {
		return ErrAvailabilityRuleNotFound
	}

	return s.availabilityRepo.Delete(ctx, ruleID)
}

// GetAvailableSlots generates available time slots for a provider on a specific date
func (s *AvailabilityService) GetAvailableSlots(ctx context.Context, productID uuid.UUID, req GetAvailableSlotsRequest) (*AvailableSlotsResponse, error) {
	s.logger.Debug("Getting available slots",
		zap.String("product_id", productID.String()),
		zap.String("provider_id", req.ProviderID),
		zap.String("date", req.Date),
	)

	// Parse date
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid date format (expected YYYY-MM-DD): %v", ErrInvalidInput, err)
	}

	// Check if date is in the past
	today := time.Now().Truncate(24 * time.Hour)
	if date.Before(today) {
		return nil, ErrDateInPast
	}

	// Determine the day of week (Go: Sunday=0)
	dayOfWeek := models.DayOfWeek(date.Weekday())

	// Get availability rule for this day
	rule, err := s.availabilityRepo.GetByProviderAndDay(ctx, productID, req.ProviderID, dayOfWeek)
	if err != nil {
		if errors.Is(err, repository.ErrAvailabilityRuleNotFound) {
			return &AvailableSlotsResponse{
				ProviderID:      req.ProviderID,
				Date:            req.Date,
				Timezone:        "UTC",
				DurationMinutes: 0,
				Slots:           []models.TimeSlot{},
			}, nil
		}
		return nil, err
	}

	// Load timezone
	loc, err := time.LoadLocation(rule.Timezone)
	if err != nil {
		s.logger.Error("Invalid timezone in rule", zap.String("timezone", rule.Timezone), zap.Error(err))
		loc = time.UTC
	}

	// Generate all possible slots
	slots := s.generateSlots(date, rule, loc)

	// Get existing appointments for this provider on this date
	dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, loc).UTC()
	dayEnd := dayStart.Add(24 * time.Hour)

	existingAppointments, err := s.appointmentRepo.GetByProviderAndDateRange(ctx, productID, req.ProviderID, dayStart, dayEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing appointments: %w", err)
	}

	// Filter out slots that overlap with existing appointments
	availableSlots := s.filterAvailableSlots(slots, existingAppointments, rule)

	// Determine response timezone
	responseTimezone := rule.Timezone
	if req.Timezone != "" {
		if _, err := time.LoadLocation(req.Timezone); err == nil {
			responseTimezone = req.Timezone
		}
	}

	return &AvailableSlotsResponse{
		ProviderID:      req.ProviderID,
		Date:            req.Date,
		Timezone:        responseTimezone,
		DurationMinutes: rule.DurationMinutes,
		Slots:           availableSlots,
	}, nil
}

// generateSlots generates all possible time slots for a given date and rule
func (s *AvailabilityService) generateSlots(date time.Time, rule *models.AvailabilityRule, loc *time.Location) []models.TimeSlot {
	slots := []models.TimeSlot{}

	// Convert rule times to actual timestamps on this date
	availabilityStart := rule.StartTime.ToTimeOnDate(date, loc)
	availabilityEnd := rule.EndTime.ToTimeOnDate(date, loc)

	// Handle rules that span midnight
	if availabilityEnd.Before(availabilityStart) {
		availabilityEnd = availabilityEnd.Add(24 * time.Hour)
	}

	// Generate slots at each interval
	currentSlotStart := availabilityStart
	duration := time.Duration(rule.DurationMinutes) * time.Minute
	interval := time.Duration(rule.SlotIntervalMinutes) * time.Minute

	for {
		slotEnd := currentSlotStart.Add(duration)

		// Check if slot fits within availability window
		if slotEnd.After(availabilityEnd) {
			break
		}

		slots = append(slots, models.TimeSlot{
			StartTime: currentSlotStart.UTC(),
			EndTime:   slotEnd.UTC(),
			Duration:  rule.DurationMinutes,
		})

		currentSlotStart = currentSlotStart.Add(interval)
	}

	return slots
}

// filterAvailableSlots removes slots that conflict with existing appointments
func (s *AvailabilityService) filterAvailableSlots(slots []models.TimeSlot, appointments []models.Appointment, rule *models.AvailabilityRule) []models.TimeSlot {
	if len(appointments) == 0 {
		return slots
	}

	now := time.Now()
	bufferBefore := time.Duration(rule.BufferBeforeMinutes) * time.Minute
	bufferAfter := time.Duration(rule.BufferAfterMinutes) * time.Minute

	availableSlots := []models.TimeSlot{}

	for _, slot := range slots {
		// Skip slots that are in the past
		if slot.StartTime.Before(now) {
			continue
		}

		// Calculate effective slot boundaries with buffers
		effectiveSlotStart := slot.StartTime.Add(-bufferBefore)
		effectiveSlotEnd := slot.EndTime.Add(bufferAfter)

		// Check for conflicts with existing appointments
		hasConflict := false
		for _, appt := range appointments {
			// Calculate effective appointment boundaries with buffers
			effectiveApptStart := appt.StartTime.Add(-bufferBefore)
			effectiveApptEnd := appt.EndTime.Add(bufferAfter)

			// Check for overlap
			if effectiveSlotStart.Before(effectiveApptEnd) && effectiveSlotEnd.After(effectiveApptStart) {
				hasConflict = true
				break
			}
		}

		if !hasConflict {
			availableSlots = append(availableSlots, slot)
		}
	}

	return availableSlots
}

// BookAppointment books an appointment with concurrency safety
func (s *AvailabilityService) BookAppointment(ctx context.Context, productID uuid.UUID, userID string, req BookAppointmentRequest) (*models.Appointment, error) {
	s.logger.Debug("Booking appointment",
		zap.String("product_id", productID.String()),
		zap.String("user_id", userID),
		zap.String("provider_id", req.ProviderID),
		zap.Time("start_time", req.StartTime),
	)

	// Validate start time is in the future
	if req.StartTime.Before(time.Now()) {
		return nil, ErrDateInPast
	}

	// Get the day of week for the requested time
	dayOfWeek := models.DayOfWeek(req.StartTime.Weekday())

	// Get availability rule for this provider and day
	rule, err := s.availabilityRepo.GetByProviderAndDay(ctx, productID, req.ProviderID, dayOfWeek)
	if err != nil {
		if errors.Is(err, repository.ErrAvailabilityRuleNotFound) {
			return nil, ErrAvailabilityRuleNotFound
		}
		return nil, err
	}

	// Load timezone
	loc, err := time.LoadLocation(rule.Timezone)
	if err != nil {
		s.logger.Error("Invalid timezone in rule", zap.String("timezone", rule.Timezone), zap.Error(err))
		loc = time.UTC
	}

	// Calculate end time based on duration
	endTime := req.StartTime.Add(time.Duration(rule.DurationMinutes) * time.Minute)

	// Validate the requested slot
	if err := s.validateSlot(req.StartTime, endTime, rule, loc); err != nil {
		return nil, err
	}

	// Default timezone
	timezone := req.Timezone
	if timezone == "" {
		timezone = rule.Timezone
	}

	// Build appointment
	appointment := &models.Appointment{
		BaseModel: models.BaseModel{
			ID: uuid.New(),
		},
		ProductID:   productID,
		ProviderID:  &req.ProviderID,
		Title:       req.Title,
		Description: req.Description,
		StartTime:   req.StartTime.UTC(),
		EndTime:     endTime.UTC(),
		Timezone:    timezone,
		Location:    req.Location,
		Status:      models.AppointmentStatusScheduled,
		CreatedBy:   userID,
		Metadata:    req.Metadata,
	}

	// Build participants list
	participants := make([]models.AppointmentParticipant, 0, len(req.Participants))
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

		// Set creator's status to accepted
		if p.ExternalUserID == userID {
			participant.Status = models.ParticipantStatusAccepted
		}

		participants = append(participants, participant)
	}

	// Create appointment with locking (concurrency safe)
	if err := s.appointmentRepo.CreateWithLock(ctx, appointment, participants); err != nil {
		if errors.Is(err, repository.ErrBookingConflict) {
			return nil, ErrBookingConflict
		}
		return nil, err
	}

	appointment.Participants = participants

	s.logger.Info("Appointment booked successfully",
		zap.String("appointment_id", appointment.ID.String()),
		zap.String("provider_id", req.ProviderID),
		zap.Time("start_time", appointment.StartTime),
	)

	return appointment, nil
}

// validateSlot validates that the requested slot is valid according to the availability rule
func (s *AvailabilityService) validateSlot(startTime, endTime time.Time, rule *models.AvailabilityRule, loc *time.Location) error {
	// Convert to provider's timezone
	startInLoc := startTime.In(loc)
	endInLoc := endTime.In(loc)

	// Get the date in provider's timezone
	date := time.Date(startInLoc.Year(), startInLoc.Month(), startInLoc.Day(), 0, 0, 0, 0, loc)

	// Calculate availability boundaries for this date
	availabilityStart := rule.StartTime.ToTimeOnDate(date, loc)
	availabilityEnd := rule.EndTime.ToTimeOnDate(date, loc)

	// Handle rules that span midnight
	if availabilityEnd.Before(availabilityStart) {
		availabilityEnd = availabilityEnd.Add(24 * time.Hour)
	}

	// Check if slot starts before availability window
	if startInLoc.Before(availabilityStart) {
		return fmt.Errorf("%w: slot starts before availability window (%s)", ErrSlotNotAvailable, rule.StartTime.String())
	}

	// Check if slot ends after availability window
	if endInLoc.After(availabilityEnd) {
		return fmt.Errorf("%w: slot ends after availability window (%s)", ErrSlotNotAvailable, rule.EndTime.String())
	}

	// Validate slot aligns with interval
	minutesSinceStart := int(startInLoc.Sub(availabilityStart).Minutes())
	if minutesSinceStart%rule.SlotIntervalMinutes != 0 {
		return fmt.Errorf("%w: slot does not align with interval of %d minutes", ErrInvalidSlotTime, rule.SlotIntervalMinutes)
	}

	// Validate duration matches rule
	slotDuration := int(endTime.Sub(startTime).Minutes())
	if slotDuration != rule.DurationMinutes {
		return fmt.Errorf("%w: slot duration (%d min) does not match rule duration (%d min)", ErrInvalidSlotTime, slotDuration, rule.DurationMinutes)
	}

	return nil
}

// GetProviderAppointments gets appointments for a provider within a date range
func (s *AvailabilityService) GetProviderAppointments(ctx context.Context, productID uuid.UUID, providerID string, startDate, endDate time.Time) ([]models.Appointment, error) {
	return s.appointmentRepo.GetByProviderAndDateRange(ctx, productID, providerID, startDate, endDate)
}

// BulkCreateAvailabilityRules creates multiple availability rules for a provider
func (s *AvailabilityService) BulkCreateAvailabilityRules(ctx context.Context, productID uuid.UUID, rules []CreateAvailabilityRuleRequest) ([]models.AvailabilityRule, error) {
	createdRules := make([]models.AvailabilityRule, 0, len(rules))

	for _, req := range rules {
		rule, err := s.CreateAvailabilityRule(ctx, productID, req)
		if err != nil {
			// Log error but continue with other rules
			s.logger.Warn("Failed to create availability rule",
				zap.String("provider_id", req.ProviderID),
				zap.Int("day_of_week", req.DayOfWeek),
				zap.Error(err),
			)
			continue
		}
		createdRules = append(createdRules, *rule)
	}

	return createdRules, nil
}

// GetAvailableSlotsForDateRange gets available slots for multiple days
func (s *AvailabilityService) GetAvailableSlotsForDateRange(ctx context.Context, productID uuid.UUID, providerID string, startDate, endDate time.Time, timezone string) (map[string][]models.TimeSlot, error) {
	result := make(map[string][]models.TimeSlot)

	// Iterate through each day
	current := startDate
	for !current.After(endDate) {
		dateStr := current.Format("2006-01-02")

		resp, err := s.GetAvailableSlots(ctx, productID, GetAvailableSlotsRequest{
			ProviderID: providerID,
			Date:       dateStr,
			Timezone:   timezone,
		})

		if err != nil {
			// Log error but continue with other dates
			if !errors.Is(err, ErrDateInPast) {
				s.logger.Warn("Failed to get slots for date",
					zap.String("date", dateStr),
					zap.Error(err),
				)
			}
		} else {
			result[dateStr] = resp.Slots
		}

		current = current.Add(24 * time.Hour)
	}

	return result, nil
}

// SortSlotsByStartTime sorts time slots by start time
func SortSlotsByStartTime(slots []models.TimeSlot) {
	sort.Slice(slots, func(i, j int) bool {
		return slots[i].StartTime.Before(slots[j].StartTime)
	})
}
