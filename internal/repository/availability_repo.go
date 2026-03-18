package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/laith-ambianze/appointment-service/internal/models"
)

// AvailabilityRepository defines the interface for availability rule data operations
type AvailabilityRepository interface {
	// Create creates a new availability rule
	Create(ctx context.Context, rule *models.AvailabilityRule) error

	// GetByID retrieves an availability rule by ID
	GetByID(ctx context.Context, id uuid.UUID) (*models.AvailabilityRule, error)

	// GetByProviderAndDay retrieves an availability rule for a provider on a specific day
	GetByProviderAndDay(ctx context.Context, productID uuid.UUID, providerID string, dayOfWeek models.DayOfWeek) (*models.AvailabilityRule, error)

	// ListByProvider retrieves all availability rules for a provider
	ListByProvider(ctx context.Context, productID uuid.UUID, providerID string) ([]models.AvailabilityRule, error)

	// Update updates an existing availability rule
	Update(ctx context.Context, rule *models.AvailabilityRule) error

	// Delete soft-deletes an availability rule
	Delete(ctx context.Context, id uuid.UUID) error

	// DeleteByProviderAndDay deletes an availability rule for a provider on a specific day
	DeleteByProviderAndDay(ctx context.Context, productID uuid.UUID, providerID string, dayOfWeek models.DayOfWeek) error
}
