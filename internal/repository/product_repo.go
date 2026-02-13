package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/laith-ambianze/appointment-service/internal/models"
)

// ProductRepository defines the interface for product data operations
type ProductRepository interface {
	// Create creates a new product
	Create(ctx context.Context, product *models.Product) error

	// GetByID retrieves a product by ID
	GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error)

	// GetByAPIKey retrieves a product by API key
	GetByAPIKey(ctx context.Context, apiKey string) (*models.Product, error)

	// List retrieves all products with optional filters
	List(ctx context.Context, filters ProductFilters) ([]models.Product, error)

	// Update updates an existing product
	Update(ctx context.Context, product *models.Product) error

	// Delete soft-deletes a product
	Delete(ctx context.Context, id uuid.UUID) error

	// HardDelete permanently deletes a product
	HardDelete(ctx context.Context, id uuid.UUID) error
}

// ProductFilters defines filters for listing products
type ProductFilters struct {
	Status         *models.ProductStatus
	IncludeDeleted bool
	Limit          int
	Offset         int
}
