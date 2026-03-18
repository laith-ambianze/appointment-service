package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/laith-ambianze/appointment-service/internal/models"
	"go.uber.org/zap"
)

// availabilityRepository implements AvailabilityRepository interface
type availabilityRepository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

// NewAvailabilityRepository creates a new availability repository
func NewAvailabilityRepository(db *pgxpool.Pool, logger *zap.Logger) AvailabilityRepository {
	return &availabilityRepository{
		db:     db,
		logger: logger,
	}
}

// ErrAvailabilityRuleNotFound indicates the availability rule was not found
var ErrAvailabilityRuleNotFound = errors.New("availability rule not found")

// ErrDuplicateAvailabilityRule indicates a duplicate availability rule
var ErrDuplicateAvailabilityRule = errors.New("availability rule already exists for this provider and day")

// Create creates a new availability rule
func (r *availabilityRepository) Create(ctx context.Context, rule *models.AvailabilityRule) error {
	// Validate rule
	if err := rule.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	// Generate ID if not provided
	if rule.ID == uuid.Nil {
		rule.ID = uuid.New()
	}

	query := `
		INSERT INTO provider_availability_rules (
			id, product_id, provider_id, day_of_week,
			start_time, end_time, duration_minutes, slot_interval_minutes,
			buffer_before_minutes, buffer_after_minutes, timezone, is_active,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)
	`

	now := time.Now()
	rule.CreatedAt = now
	rule.UpdatedAt = now

	_, err := r.db.Exec(ctx, query,
		rule.ID,
		rule.ProductID,
		rule.ProviderID,
		rule.DayOfWeek,
		rule.StartTime.String(),
		rule.EndTime.String(),
		rule.DurationMinutes,
		rule.SlotIntervalMinutes,
		rule.BufferBeforeMinutes,
		rule.BufferAfterMinutes,
		rule.Timezone,
		rule.IsActive,
		rule.CreatedAt,
		rule.UpdatedAt,
	)

	if err != nil {
		// Check for unique constraint violation
		if isUniqueViolation(err) {
			return ErrDuplicateAvailabilityRule
		}
		return fmt.Errorf("%w: failed to create availability rule: %v", ErrDatabase, err)
	}

	r.logger.Info("Availability rule created",
		zap.String("rule_id", rule.ID.String()),
		zap.String("provider_id", rule.ProviderID),
		zap.Int("day_of_week", int(rule.DayOfWeek)),
	)

	return nil
}

// GetByID retrieves an availability rule by ID
func (r *availabilityRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.AvailabilityRule, error) {
	query := `
		SELECT 
			id, product_id, provider_id, day_of_week,
			start_time, end_time, duration_minutes, slot_interval_minutes,
			buffer_before_minutes, buffer_after_minutes, timezone, is_active,
			created_at, updated_at, deleted_at
		FROM provider_availability_rules
		WHERE id = $1 AND deleted_at IS NULL
	`

	rule, err := r.scanRule(r.db.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAvailabilityRuleNotFound
		}
		return nil, err
	}

	return rule, nil
}

// GetByProviderAndDay retrieves an availability rule for a provider on a specific day
func (r *availabilityRepository) GetByProviderAndDay(ctx context.Context, productID uuid.UUID, providerID string, dayOfWeek models.DayOfWeek) (*models.AvailabilityRule, error) {
	query := `
		SELECT 
			id, product_id, provider_id, day_of_week,
			start_time, end_time, duration_minutes, slot_interval_minutes,
			buffer_before_minutes, buffer_after_minutes, timezone, is_active,
			created_at, updated_at, deleted_at
		FROM provider_availability_rules
		WHERE product_id = $1 
			AND provider_id = $2 
			AND day_of_week = $3 
			AND deleted_at IS NULL
			AND is_active = true
	`

	rule, err := r.scanRule(r.db.QueryRow(ctx, query, productID, providerID, dayOfWeek))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAvailabilityRuleNotFound
		}
		return nil, err
	}

	return rule, nil
}

// ListByProvider retrieves all availability rules for a provider
func (r *availabilityRepository) ListByProvider(ctx context.Context, productID uuid.UUID, providerID string) ([]models.AvailabilityRule, error) {
	query := `
		SELECT 
			id, product_id, provider_id, day_of_week,
			start_time, end_time, duration_minutes, slot_interval_minutes,
			buffer_before_minutes, buffer_after_minutes, timezone, is_active,
			created_at, updated_at, deleted_at
		FROM provider_availability_rules
		WHERE product_id = $1 
			AND provider_id = $2 
			AND deleted_at IS NULL
		ORDER BY day_of_week ASC
	`

	rows, err := r.db.Query(ctx, query, productID, providerID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabase, err)
	}
	defer rows.Close()

	rules := []models.AvailabilityRule{}
	for rows.Next() {
		rule, err := r.scanRuleFromRows(rows)
		if err != nil {
			return nil, err
		}
		rules = append(rules, *rule)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	return rules, nil
}

// Update updates an existing availability rule
func (r *availabilityRepository) Update(ctx context.Context, rule *models.AvailabilityRule) error {
	// Validate rule
	if err := rule.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	query := `
		UPDATE provider_availability_rules SET
			start_time = $1,
			end_time = $2,
			duration_minutes = $3,
			slot_interval_minutes = $4,
			buffer_before_minutes = $5,
			buffer_after_minutes = $6,
			timezone = $7,
			is_active = $8,
			updated_at = $9
		WHERE id = $10 AND deleted_at IS NULL
	`

	rule.UpdatedAt = time.Now()

	result, err := r.db.Exec(ctx, query,
		rule.StartTime.String(),
		rule.EndTime.String(),
		rule.DurationMinutes,
		rule.SlotIntervalMinutes,
		rule.BufferBeforeMinutes,
		rule.BufferAfterMinutes,
		rule.Timezone,
		rule.IsActive,
		rule.UpdatedAt,
		rule.ID,
	)

	if err != nil {
		return fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	if result.RowsAffected() == 0 {
		return ErrAvailabilityRuleNotFound
	}

	r.logger.Info("Availability rule updated",
		zap.String("rule_id", rule.ID.String()),
	)

	return nil
}

// Delete soft-deletes an availability rule
func (r *availabilityRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE provider_availability_rules 
		SET deleted_at = $1, updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL
	`

	now := time.Now()
	result, err := r.db.Exec(ctx, query, now, now, id)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	if result.RowsAffected() == 0 {
		return ErrAvailabilityRuleNotFound
	}

	r.logger.Info("Availability rule deleted",
		zap.String("rule_id", id.String()),
	)

	return nil
}

// DeleteByProviderAndDay deletes an availability rule for a provider on a specific day
func (r *availabilityRepository) DeleteByProviderAndDay(ctx context.Context, productID uuid.UUID, providerID string, dayOfWeek models.DayOfWeek) error {
	query := `
		UPDATE provider_availability_rules 
		SET deleted_at = $1, updated_at = $2
		WHERE product_id = $3 
			AND provider_id = $4 
			AND day_of_week = $5 
			AND deleted_at IS NULL
	`

	now := time.Now()
	result, err := r.db.Exec(ctx, query, now, now, productID, providerID, dayOfWeek)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	if result.RowsAffected() == 0 {
		return ErrAvailabilityRuleNotFound
	}

	r.logger.Info("Availability rule deleted by provider and day",
		zap.String("provider_id", providerID),
		zap.Int("day_of_week", int(dayOfWeek)),
	)

	return nil
}

// scanRule scans a single row into an AvailabilityRule
func (r *availabilityRepository) scanRule(row pgx.Row) (*models.AvailabilityRule, error) {
	var rule models.AvailabilityRule
	var startTimeStr, endTimeStr string

	err := row.Scan(
		&rule.ID,
		&rule.ProductID,
		&rule.ProviderID,
		&rule.DayOfWeek,
		&startTimeStr,
		&endTimeStr,
		&rule.DurationMinutes,
		&rule.SlotIntervalMinutes,
		&rule.BufferBeforeMinutes,
		&rule.BufferAfterMinutes,
		&rule.Timezone,
		&rule.IsActive,
		&rule.CreatedAt,
		&rule.UpdatedAt,
		&rule.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
		return nil, fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	// Parse time strings
	startTime, err := models.ParseLocalTime(startTimeStr)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid start_time: %v", ErrDatabase, err)
	}
	rule.StartTime = startTime

	endTime, err := models.ParseLocalTime(endTimeStr)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid end_time: %v", ErrDatabase, err)
	}
	rule.EndTime = endTime

	return &rule, nil
}

// scanRuleFromRows scans a row from rows into an AvailabilityRule
func (r *availabilityRepository) scanRuleFromRows(rows pgx.Rows) (*models.AvailabilityRule, error) {
	var rule models.AvailabilityRule
	var startTimeStr, endTimeStr string

	err := rows.Scan(
		&rule.ID,
		&rule.ProductID,
		&rule.ProviderID,
		&rule.DayOfWeek,
		&startTimeStr,
		&endTimeStr,
		&rule.DurationMinutes,
		&rule.SlotIntervalMinutes,
		&rule.BufferBeforeMinutes,
		&rule.BufferAfterMinutes,
		&rule.Timezone,
		&rule.IsActive,
		&rule.CreatedAt,
		&rule.UpdatedAt,
		&rule.DeletedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	// Parse time strings
	startTime, err := models.ParseLocalTime(startTimeStr)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid start_time: %v", ErrDatabase, err)
	}
	rule.StartTime = startTime

	endTime, err := models.ParseLocalTime(endTimeStr)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid end_time: %v", ErrDatabase, err)
	}
	rule.EndTime = endTime

	return &rule, nil
}

// isUniqueViolation checks if the error is a unique constraint violation
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	// PostgreSQL error code 23505 is unique_violation
	errStr := err.Error()
	return contains(errStr, "23505") || contains(errStr, "unique") || contains(errStr, "duplicate")
}

// contains checks if s contains substr (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && indexOf(s, substr) >= 0))
}

// indexOf returns the index of substr in s, or -1 if not found
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// Ensure sql.NullTime is imported by using it
var _ sql.NullTime
