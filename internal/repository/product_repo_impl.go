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
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/laith-ambianze/appointment-service/internal/models"
	"go.uber.org/zap"
)

// productRepository implements ProductRepository interface
type productRepository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

// NewProductRepository creates a new product repository
func NewProductRepository(db *pgxpool.Pool, logger *zap.Logger) ProductRepository {
	return &productRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new product
func (r *productRepository) Create(ctx context.Context, product *models.Product) error {
	// Validate product
	if err := product.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	// Generate ID if not provided
	if product.ID == uuid.Nil {
		product.ID = uuid.New()
	}

	// Marshal metadata
	metadataJSON, err := product.MetadataJSON()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	// Insert query
	query := `
		INSERT INTO products (
			id, name, description, api_key, api_secret_hash,
			status, webhook_url, webhook_secret, metadata,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)
	`

	now := time.Now()
	product.CreatedAt = now
	product.UpdatedAt = now

	_, err = r.db.Exec(ctx, query,
		product.ID,
		product.Name,
		product.Description,
		product.APIKey,
		product.APISecretHash,
		product.Status,
		nullString(product.WebhookURL),
		nullString(product.WebhookSecret),
		metadataJSON,
		product.CreatedAt,
		product.UpdatedAt,
	)

	if err != nil {
		// Check for duplicate key error
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return fmt.Errorf("%w: %v", ErrDuplicate, err)
		}
		return fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	r.logger.Info("Product created",
		zap.String("product_id", product.ID.String()),
		zap.String("name", product.Name),
	)

	return nil
}

// GetByID retrieves a product by ID
func (r *productRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	query := `
		SELECT 
			id, name, description, api_key, api_secret_hash,
			status, webhook_url, webhook_secret, metadata,
			created_at, updated_at, deleted_at
		FROM products
		WHERE id = $1 AND deleted_at IS NULL
	`

	var product models.Product
	var description, webhookURL, webhookSecret sql.NullString
	var metadataJSON string

	err := r.db.QueryRow(ctx, query, id).Scan(
		&product.ID,
		&product.Name,
		&description,
		&product.APIKey,
		&product.APISecretHash,
		&product.Status,
		&webhookURL,
		&webhookSecret,
		&metadataJSON,
		&product.CreatedAt,
		&product.UpdatedAt,
		&product.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	// Set nullable fields
	product.Description = description.String
	product.WebhookURL = webhookURL.String
	product.WebhookSecret = webhookSecret.String

	// Unmarshal metadata
	if metadataJSON != "" && metadataJSON != "{}" {
		if err := json.Unmarshal([]byte(metadataJSON), &product.Metadata); err != nil {
			r.logger.Warn("Failed to unmarshal product metadata",
				zap.String("product_id", id.String()),
				zap.Error(err),
			)
		}
	}

	return &product, nil
}

// GetByAPIKey retrieves a product by API key
func (r *productRepository) GetByAPIKey(ctx context.Context, apiKey string) (*models.Product, error) {
	query := `
		SELECT 
			id, name, description, api_key, api_secret_hash,
			status, webhook_url, webhook_secret, metadata,
			created_at, updated_at, deleted_at
		FROM products
		WHERE api_key = $1 AND deleted_at IS NULL
	`

	var product models.Product
	var description, webhookURL, webhookSecret sql.NullString
	var metadataJSON string

	err := r.db.QueryRow(ctx, query, apiKey).Scan(
		&product.ID,
		&product.Name,
		&description,
		&product.APIKey,
		&product.APISecretHash,
		&product.Status,
		&webhookURL,
		&webhookSecret,
		&metadataJSON,
		&product.CreatedAt,
		&product.UpdatedAt,
		&product.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	// Set nullable fields
	product.Description = description.String
	product.WebhookURL = webhookURL.String
	product.WebhookSecret = webhookSecret.String

	// Unmarshal metadata
	if metadataJSON != "" && metadataJSON != "{}" {
		if err := json.Unmarshal([]byte(metadataJSON), &product.Metadata); err != nil {
			r.logger.Warn("Failed to unmarshal product metadata",
				zap.String("api_key", apiKey),
				zap.Error(err),
			)
		}
	}

	return &product, nil
}

// List retrieves all products with optional filters
func (r *productRepository) List(ctx context.Context, filters ProductFilters) ([]models.Product, error) {
	// Build query with filters
	query := `
		SELECT 
			id, name, description, api_key, api_secret_hash,
			status, webhook_url, webhook_secret, metadata,
			created_at, updated_at, deleted_at
		FROM products
		WHERE 1=1
	`

	var conditions []string
	var args []interface{}
	argIndex := 1

	// Apply filters
	if !filters.IncludeDeleted {
		conditions = append(conditions, "deleted_at IS NULL")
	}

	if filters.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *filters.Status)
		argIndex++
	}

	// Add conditions to query
	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}

	// Add ordering
	query += " ORDER BY created_at DESC"

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
	var products []models.Product
	for rows.Next() {
		var product models.Product
		var description, webhookURL, webhookSecret sql.NullString
		var metadataJSON string

		err := rows.Scan(
			&product.ID,
			&product.Name,
			&description,
			&product.APIKey,
			&product.APISecretHash,
			&product.Status,
			&webhookURL,
			&webhookSecret,
			&metadataJSON,
			&product.CreatedAt,
			&product.UpdatedAt,
			&product.DeletedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrDatabase, err)
		}

		// Set nullable fields
		product.Description = description.String
		product.WebhookURL = webhookURL.String
		product.WebhookSecret = webhookSecret.String

		// Unmarshal metadata
		if metadataJSON != "" && metadataJSON != "{}" {
			if err := json.Unmarshal([]byte(metadataJSON), &product.Metadata); err != nil {
				r.logger.Warn("Failed to unmarshal product metadata",
					zap.String("product_id", product.ID.String()),
					zap.Error(err),
				)
			}
		}

		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	return products, nil
}

// Update updates an existing product
func (r *productRepository) Update(ctx context.Context, product *models.Product) error {
	// Validate product
	if err := product.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	// Marshal metadata
	metadataJSON, err := product.MetadataJSON()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	// Update query
	query := `
		UPDATE products SET
			name = $1,
			description = $2,
			api_key = $3,
			api_secret_hash = $4,
			status = $5,
			webhook_url = $6,
			webhook_secret = $7,
			metadata = $8,
			updated_at = $9
		WHERE id = $10 AND deleted_at IS NULL
	`

	product.UpdatedAt = time.Now()

	result, err := r.db.Exec(ctx, query,
		product.Name,
		product.Description,
		product.APIKey,
		product.APISecretHash,
		product.Status,
		nullString(product.WebhookURL),
		nullString(product.WebhookSecret),
		metadataJSON,
		product.UpdatedAt,
		product.ID,
	)

	if err != nil {
		// Check for duplicate key error
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("%w: %v", ErrDuplicate, err)
		}
		return fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	r.logger.Info("Product updated",
		zap.String("product_id", product.ID.String()),
	)

	return nil
}

// Delete soft-deletes a product
func (r *productRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE products 
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

	r.logger.Info("Product soft-deleted",
		zap.String("product_id", id.String()),
	)

	return nil
}

// HardDelete permanently deletes a product
func (r *productRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM products WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)

	if err != nil {
		return fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	r.logger.Warn("Product hard-deleted",
		zap.String("product_id", id.String()),
	)

	return nil
}

// Helper function to convert empty strings to NULL
func nullString(s string) sql.NullString {
	return sql.NullString{
		String: s,
		Valid:  s != "",
	}
}
