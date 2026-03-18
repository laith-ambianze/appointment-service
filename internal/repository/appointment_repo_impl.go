package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/laith-ambianze/appointment-service/internal/models"
	"go.uber.org/zap"
)

// appointmentRepository implements AppointmentRepository interface
type appointmentRepository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

// NewAppointmentRepository creates a new appointment repository
func NewAppointmentRepository(db *pgxpool.Pool, logger *zap.Logger) AppointmentRepository {
	return &appointmentRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new appointment with participants in a transaction
func (r *appointmentRepository) Create(ctx context.Context, appointment *models.Appointment, participants []models.AppointmentParticipant) error {
	// Validate appointment
	if err := appointment.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	// Validate participants
	if len(participants) == 0 {
		return fmt.Errorf("%w: at least one participant is required", ErrInvalidInput)
	}

	for i := range participants {
		if err := participants[i].Validate(); err != nil {
			return fmt.Errorf("%w: participant %d: %v", ErrInvalidInput, i, err)
		}
	}

	// Generate ID if not provided
	if appointment.ID == uuid.Nil {
		appointment.ID = uuid.New()
	}

	// Marshal metadata
	metadataJSON, err := appointment.MetadataJSON()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	// Start transaction
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%w: failed to begin transaction: %v", ErrDatabase, err)
	}
	defer tx.Rollback(ctx)

	// Insert appointment
	appointmentQuery := `
		INSERT INTO appointments (
			id, product_id, provider_id, title, description, start_time, end_time,
			timezone, location, status, created_by, metadata,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)
	`

	now := time.Now()
	appointment.CreatedAt = now
	appointment.UpdatedAt = now

	_, err = tx.Exec(ctx, appointmentQuery,
		appointment.ID,
		appointment.ProductID,
		nullStringPtr(appointment.ProviderID),
		appointment.Title,
		appointment.Description,
		appointment.StartTime,
		appointment.EndTime,
		appointment.Timezone,
		nullString(appointment.Location),
		appointment.Status,
		appointment.CreatedBy,
		metadataJSON,
		appointment.CreatedAt,
		appointment.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("%w: failed to create appointment: %v", ErrDatabase, err)
	}

	// Insert participants
	participantQuery := `
		INSERT INTO appointment_participants (
			id, appointment_id, external_user_id, role, status,
			user_metadata, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)
	`

	for i := range participants {
		participant := &participants[i]

		// Generate ID if not provided
		if participant.ID == uuid.Nil {
			participant.ID = uuid.New()
		}

		participant.AppointmentID = appointment.ID
		participant.CreatedAt = now
		participant.UpdatedAt = now

		userMetadataJSON, err := participant.UserMetadataJSON()
		if err != nil {
			return fmt.Errorf("%w: failed to marshal participant metadata: %v", ErrInvalidInput, err)
		}

		_, err = tx.Exec(ctx, participantQuery,
			participant.ID,
			participant.AppointmentID,
			participant.ExternalUserID,
			participant.Role,
			participant.Status,
			userMetadataJSON,
			participant.CreatedAt,
			participant.UpdatedAt,
		)

		if err != nil {
			return fmt.Errorf("%w: failed to create participant: %v", ErrDatabase, err)
		}
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("%w: failed to commit transaction: %v", ErrDatabase, err)
	}

	appointment.Participants = participants

	r.logger.Info("Appointment created",
		zap.String("appointment_id", appointment.ID.String()),
		zap.String("product_id", appointment.ProductID.String()),
		zap.Int("participants", len(participants)),
	)

	return nil
}

// GetByID retrieves an appointment by ID with participants
func (r *appointmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Appointment, error) {
	query := `
		SELECT 
			id, product_id, provider_id, title, description, start_time, end_time,
			timezone, location, status, created_by, metadata,
			created_at, updated_at, deleted_at
		FROM appointments
		WHERE id = $1 AND deleted_at IS NULL
	`

	var appointment models.Appointment
	var providerID, description, location sql.NullString
	var metadataJSON string

	err := r.db.QueryRow(ctx, query, id).Scan(
		&appointment.ID,
		&appointment.ProductID,
		&providerID,
		&appointment.Title,
		&description,
		&appointment.StartTime,
		&appointment.EndTime,
		&appointment.Timezone,
		&location,
		&appointment.Status,
		&appointment.CreatedBy,
		&metadataJSON,
		&appointment.CreatedAt,
		&appointment.UpdatedAt,
		&appointment.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	// Set nullable fields
	if providerID.Valid {
		appointment.ProviderID = &providerID.String
	}
	appointment.Description = description.String
	appointment.Location = location.String

	// Unmarshal metadata
	if metadataJSON != "" && metadataJSON != "{}" {
		if err := json.Unmarshal([]byte(metadataJSON), &appointment.Metadata); err != nil {
			r.logger.Warn("Failed to unmarshal appointment metadata",
				zap.String("appointment_id", id.String()),
				zap.Error(err),
			)
		}
	}

	// Load participants
	participants, err := r.GetParticipants(ctx, id)
	if err != nil {
		return nil, err
	}
	appointment.Participants = participants

	return &appointment, nil
}

// GetByProductAndUser retrieves appointments for a user in a product
func (r *appointmentRepository) GetByProductAndUser(ctx context.Context, productID uuid.UUID, externalUserID string, filters AppointmentFilters) ([]models.Appointment, error) {
	query := `
		SELECT DISTINCT
			a.id, a.product_id, a.provider_id, a.title, a.description, a.start_time, a.end_time,
			a.timezone, a.location, a.status, a.created_by, a.metadata,
			a.created_at, a.updated_at, a.deleted_at
		FROM appointments a
		INNER JOIN appointment_participants ap ON a.id = ap.appointment_id
		WHERE a.product_id = $1
			AND ap.external_user_id = $2
	`

	var conditions []string
	args := []interface{}{productID, externalUserID}
	argIndex := 3

	// Apply filters
	if !filters.IncludeDeleted {
		conditions = append(conditions, "a.deleted_at IS NULL")
	}

	if filters.Status != nil {
		conditions = append(conditions, fmt.Sprintf("a.status = $%d", argIndex))
		args = append(args, *filters.Status)
		argIndex++
	}

	if filters.StartTimeFrom != nil {
		conditions = append(conditions, fmt.Sprintf("a.start_time >= $%d", argIndex))
		args = append(args, *filters.StartTimeFrom)
		argIndex++
	}

	if filters.StartTimeTo != nil {
		conditions = append(conditions, fmt.Sprintf("a.start_time <= $%d", argIndex))
		args = append(args, *filters.StartTimeTo)
		argIndex++
	}

	// Add conditions to query
	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}

	// Add ordering
	query += " ORDER BY a.start_time DESC"

	// Add pagination
	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filters.Limit)
		argIndex++
	}

	if filters.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filters.Offset)
	}

	// Execute query
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabase, err)
	}
	defer rows.Close()

	// Scan results
	appointments := []models.Appointment{}
	for rows.Next() {
		appointment, err := r.scanAppointment(rows)
		if err != nil {
			return nil, err
		}

		// Load participants
		participants, err := r.GetParticipants(ctx, appointment.ID)
		if err != nil {
			r.logger.Warn("Failed to load participants",
				zap.String("appointment_id", appointment.ID.String()),
				zap.Error(err),
			)
		} else {
			appointment.Participants = participants
		}

		appointments = append(appointments, *appointment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	return appointments, nil
}

// GetByDateRange retrieves appointments within a date range
func (r *appointmentRepository) GetByDateRange(ctx context.Context, productID uuid.UUID, startTime, endTime time.Time) ([]models.Appointment, error) {
	query := `
		SELECT 
			id, product_id, provider_id, title, description, start_time, end_time,
			timezone, location, status, created_by, metadata,
			created_at, updated_at, deleted_at
		FROM appointments
		WHERE product_id = $1
			AND start_time >= $2
			AND end_time <= $3
			AND deleted_at IS NULL
		ORDER BY start_time ASC
	`

	rows, err := r.db.Query(ctx, query, productID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabase, err)
	}
	defer rows.Close()

	appointments := []models.Appointment{}
	for rows.Next() {
		appointment, err := r.scanAppointment(rows)
		if err != nil {
			return nil, err
		}

		// Load participants
		participants, err := r.GetParticipants(ctx, appointment.ID)
		if err != nil {
			r.logger.Warn("Failed to load participants",
				zap.String("appointment_id", appointment.ID.String()),
				zap.Error(err),
			)
		} else {
			appointment.Participants = participants
		}

		appointments = append(appointments, *appointment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	return appointments, nil
}

// Update updates an existing appointment
func (r *appointmentRepository) Update(ctx context.Context, appointment *models.Appointment) error {
	// Validate appointment
	if err := appointment.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	// Marshal metadata
	metadataJSON, err := appointment.MetadataJSON()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	query := `
		UPDATE appointments SET
			title = $1,
			description = $2,
			start_time = $3,
			end_time = $4,
			timezone = $5,
			location = $6,
			status = $7,
			metadata = $8,
			updated_at = $9
		WHERE id = $10 AND deleted_at IS NULL
	`

	appointment.UpdatedAt = time.Now()

	result, err := r.db.Exec(ctx, query,
		appointment.Title,
		appointment.Description,
		appointment.StartTime,
		appointment.EndTime,
		appointment.Timezone,
		nullString(appointment.Location),
		appointment.Status,
		metadataJSON,
		appointment.UpdatedAt,
		appointment.ID,
	)

	if err != nil {
		return fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	r.logger.Info("Appointment updated",
		zap.String("appointment_id", appointment.ID.String()),
	)

	return nil
}

// UpdateStatus updates appointment status
func (r *appointmentRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.AppointmentStatus) error {
	query := `
		UPDATE appointments 
		SET status = $1, updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	r.logger.Info("Appointment status updated",
		zap.String("appointment_id", id.String()),
		zap.String("status", string(status)),
	)

	return nil
}

// Delete soft-deletes an appointment
func (r *appointmentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE appointments 
		SET deleted_at = $1, updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL
	`

	now := time.Now()
	result, err := r.db.Exec(ctx, query, now, now, id)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	r.logger.Info("Appointment soft-deleted",
		zap.String("appointment_id", id.String()),
	)

	return nil
}

// Participant operations

// AddParticipant adds a participant to an appointment
func (r *appointmentRepository) AddParticipant(ctx context.Context, participant *models.AppointmentParticipant) error {
	if err := participant.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	if participant.ID == uuid.Nil {
		participant.ID = uuid.New()
	}

	userMetadataJSON, err := participant.UserMetadataJSON()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	query := `
		INSERT INTO appointment_participants (
			id, appointment_id, external_user_id, role, status,
			user_metadata, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)
	`

	now := time.Now()
	participant.CreatedAt = now
	participant.UpdatedAt = now

	_, err = r.db.Exec(ctx, query,
		participant.ID,
		participant.AppointmentID,
		participant.ExternalUserID,
		participant.Role,
		participant.Status,
		userMetadataJSON,
		participant.CreatedAt,
		participant.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	return nil
}

// RemoveParticipant removes a participant from an appointment
func (r *appointmentRepository) RemoveParticipant(ctx context.Context, appointmentID uuid.UUID, externalUserID string) error {
	query := `
		DELETE FROM appointment_participants
		WHERE appointment_id = $1 AND external_user_id = $2
	`

	result, err := r.db.Exec(ctx, query, appointmentID, externalUserID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// UpdateParticipantStatus updates a participant's status
func (r *appointmentRepository) UpdateParticipantStatus(ctx context.Context, appointmentID uuid.UUID, externalUserID string, status models.ParticipantStatus) error {
	query := `
		UPDATE appointment_participants
		SET status = $1, updated_at = $2
		WHERE appointment_id = $3 AND external_user_id = $4
	`

	result, err := r.db.Exec(ctx, query, status, time.Now(), appointmentID, externalUserID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// GetParticipants retrieves all participants for an appointment
func (r *appointmentRepository) GetParticipants(ctx context.Context, appointmentID uuid.UUID) ([]models.AppointmentParticipant, error) {
	query := `
		SELECT 
			id, appointment_id, external_user_id, role, status,
			user_metadata, created_at, updated_at
		FROM appointment_participants
		WHERE appointment_id = $1
		ORDER BY CASE role 
			WHEN 'host' THEN 1
			WHEN 'guest' THEN 2
			WHEN 'attendee' THEN 3
			ELSE 4
		END, created_at ASC
	`

	rows, err := r.db.Query(ctx, query, appointmentID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabase, err)
	}
	defer rows.Close()

	participants := []models.AppointmentParticipant{}
	for rows.Next() {
		var participant models.AppointmentParticipant
		var userMetadataJSON string

		err := rows.Scan(
			&participant.ID,
			&participant.AppointmentID,
			&participant.ExternalUserID,
			&participant.Role,
			&participant.Status,
			&userMetadataJSON,
			&participant.CreatedAt,
			&participant.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrDatabase, err)
		}

		// Unmarshal user metadata
		if userMetadataJSON != "" && userMetadataJSON != "{}" {
			if err := json.Unmarshal([]byte(userMetadataJSON), &participant.UserMetadata); err != nil {
				r.logger.Warn("Failed to unmarshal participant user metadata",
					zap.String("participant_id", participant.ID.String()),
					zap.Error(err),
				)
			}
		}

		participants = append(participants, participant)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	return participants, nil
}

// Helper function to scan appointment from rows
func (r *appointmentRepository) scanAppointment(rows pgx.Rows) (*models.Appointment, error) {
	var appointment models.Appointment
	var providerID, description, location sql.NullString
	var metadataJSON string

	err := rows.Scan(
		&appointment.ID,
		&appointment.ProductID,
		&providerID,
		&appointment.Title,
		&description,
		&appointment.StartTime,
		&appointment.EndTime,
		&appointment.Timezone,
		&location,
		&appointment.Status,
		&appointment.CreatedBy,
		&metadataJSON,
		&appointment.CreatedAt,
		&appointment.UpdatedAt,
		&appointment.DeletedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	// Set nullable fields
	if providerID.Valid {
		appointment.ProviderID = &providerID.String
	}
	appointment.Description = description.String
	appointment.Location = location.String

	// Unmarshal metadata
	if metadataJSON != "" && metadataJSON != "{}" {
		if err := json.Unmarshal([]byte(metadataJSON), &appointment.Metadata); err != nil {
			r.logger.Warn("Failed to unmarshal appointment metadata",
				zap.String("appointment_id", appointment.ID.String()),
				zap.Error(err),
			)
		}
	}

	return &appointment, nil
}

// nullStringPtr converts a string pointer to sql.NullString
func nullStringPtr(s *string) sql.NullString {
	if s == nil || *s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *s, Valid: true}
}

// GetByProviderAndDateRange retrieves appointments for a provider within a date range
// Excludes cancelled and deleted appointments
func (r *appointmentRepository) GetByProviderAndDateRange(ctx context.Context, productID uuid.UUID, providerID string, startTime, endTime time.Time) ([]models.Appointment, error) {
	query := `
		SELECT 
			id, product_id, provider_id, title, description, start_time, end_time,
			timezone, location, status, created_by, metadata,
			created_at, updated_at, deleted_at
		FROM appointments
		WHERE product_id = $1
			AND provider_id = $2
			AND start_time < $4
			AND end_time > $3
			AND status != 'cancelled'
			AND deleted_at IS NULL
		ORDER BY start_time ASC
	`

	rows, err := r.db.Query(ctx, query, productID, providerID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabase, err)
	}
	defer rows.Close()

	appointments := []models.Appointment{}
	for rows.Next() {
		appointment, err := r.scanAppointment(rows)
		if err != nil {
			return nil, err
		}
		appointments = append(appointments, *appointment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	return appointments, nil
}

// ErrBookingConflict indicates a booking conflict (double booking)
var ErrBookingConflict = errors.New("booking conflict: time slot is not available")

// CreateWithLock creates a new appointment with transaction locking for concurrency safety
// This method uses REPEATABLE READ isolation level and SELECT FOR UPDATE to prevent race conditions
func (r *appointmentRepository) CreateWithLock(ctx context.Context, appointment *models.Appointment, participants []models.AppointmentParticipant) error {
	// Validate appointment
	if err := appointment.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	// Provider ID is required for booking
	if appointment.ProviderID == nil || *appointment.ProviderID == "" {
		return fmt.Errorf("%w: provider ID is required for booking", ErrInvalidInput)
	}

	// Validate participants
	if len(participants) == 0 {
		return fmt.Errorf("%w: at least one participant is required", ErrInvalidInput)
	}

	for i := range participants {
		if err := participants[i].Validate(); err != nil {
			return fmt.Errorf("%w: participant %d: %v", ErrInvalidInput, i, err)
		}
	}

	// Generate ID if not provided
	if appointment.ID == uuid.Nil {
		appointment.ID = uuid.New()
	}

	// Marshal metadata
	metadataJSON, err := appointment.MetadataJSON()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	// Start transaction with REPEATABLE READ isolation
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.RepeatableRead,
	})
	if err != nil {
		return fmt.Errorf("%w: failed to begin transaction: %v", ErrDatabase, err)
	}
	defer tx.Rollback(ctx)

	// Lock relevant rows to check for conflicts
	// This prevents other concurrent transactions from inserting overlapping appointments
	lockQuery := `
		SELECT id FROM appointments
		WHERE product_id = $1
			AND provider_id = $2
			AND start_time < $4
			AND end_time > $3
			AND status != 'cancelled'
			AND deleted_at IS NULL
		FOR UPDATE
	`

	rows, err := tx.Query(ctx, lockQuery,
		appointment.ProductID,
		*appointment.ProviderID,
		appointment.StartTime,
		appointment.EndTime,
	)
	if err != nil {
		return fmt.Errorf("%w: failed to lock appointments: %v", ErrDatabase, err)
	}

	// Check if any conflicting appointments exist
	hasConflict := false
	for rows.Next() {
		hasConflict = true
		break
	}
	rows.Close()

	if hasConflict {
		r.logger.Warn("Booking conflict detected",
			zap.String("provider_id", *appointment.ProviderID),
			zap.Time("start_time", appointment.StartTime),
			zap.Time("end_time", appointment.EndTime),
		)
		return ErrBookingConflict
	}

	// Insert appointment
	appointmentQuery := `
		INSERT INTO appointments (
			id, product_id, provider_id, title, description, start_time, end_time,
			timezone, location, status, created_by, metadata,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)
	`

	now := time.Now()
	appointment.CreatedAt = now
	appointment.UpdatedAt = now

	_, err = tx.Exec(ctx, appointmentQuery,
		appointment.ID,
		appointment.ProductID,
		*appointment.ProviderID,
		appointment.Title,
		appointment.Description,
		appointment.StartTime,
		appointment.EndTime,
		appointment.Timezone,
		nullString(appointment.Location),
		appointment.Status,
		appointment.CreatedBy,
		metadataJSON,
		appointment.CreatedAt,
		appointment.UpdatedAt,
	)

	if err != nil {
		// Check if the error is from the exclusion constraint
		if strings.Contains(err.Error(), "appointments_no_provider_overlap") {
			r.logger.Warn("Booking conflict detected by exclusion constraint",
				zap.String("provider_id", *appointment.ProviderID),
				zap.Time("start_time", appointment.StartTime),
				zap.Time("end_time", appointment.EndTime),
			)
			return ErrBookingConflict
		}
		return fmt.Errorf("%w: failed to create appointment: %v", ErrDatabase, err)
	}

	// Insert participants
	participantQuery := `
		INSERT INTO appointment_participants (
			id, appointment_id, external_user_id, role, status,
			user_metadata, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)
	`

	for i := range participants {
		participant := &participants[i]

		// Generate ID if not provided
		if participant.ID == uuid.Nil {
			participant.ID = uuid.New()
		}

		participant.AppointmentID = appointment.ID
		participant.CreatedAt = now
		participant.UpdatedAt = now

		userMetadataJSON, err := participant.UserMetadataJSON()
		if err != nil {
			return fmt.Errorf("%w: failed to marshal participant metadata: %v", ErrInvalidInput, err)
		}

		_, err = tx.Exec(ctx, participantQuery,
			participant.ID,
			participant.AppointmentID,
			participant.ExternalUserID,
			participant.Role,
			participant.Status,
			userMetadataJSON,
			participant.CreatedAt,
			participant.UpdatedAt,
		)

		if err != nil {
			return fmt.Errorf("%w: failed to create participant: %v", ErrDatabase, err)
		}
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		// Check for exclusion constraint violation on commit
		if strings.Contains(err.Error(), "appointments_no_provider_overlap") {
			return ErrBookingConflict
		}
		return fmt.Errorf("%w: failed to commit transaction: %v", ErrDatabase, err)
	}

	appointment.Participants = participants

	r.logger.Info("Appointment created with lock",
		zap.String("appointment_id", appointment.ID.String()),
		zap.String("product_id", appointment.ProductID.String()),
		zap.String("provider_id", *appointment.ProviderID),
		zap.Int("participants", len(participants)),
	)

	return nil
}
