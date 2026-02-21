package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/laith-ambianze/appointment-service/internal/models"
	"github.com/laith-ambianze/appointment-service/internal/repository"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// Product service errors
var (
	ErrProductNotFound    = errors.New("product not found")
	ErrProductInactive    = errors.New("product is inactive or suspended")
	ErrDuplicateAPIKey    = errors.New("API key already exists")
	ErrInvalidCredentials = errors.New("invalid API credentials")
)

// ProductService handles product business logic
type ProductService struct {
	repo   repository.ProductRepository
	logger *zap.Logger
}

// NewProductService creates a new product service
func NewProductService(repo repository.ProductRepository, logger *zap.Logger) *ProductService {
	return &ProductService{
		repo:   repo,
		logger: logger,
	}
}

// RegisterProductRequest contains data for registering a new product
type RegisterProductRequest struct {
	Name        string                 `json:"name" binding:"required,min=1,max=255"`
	Description string                 `json:"description" binding:"max=1000"`
	WebhookURL  string                 `json:"webhook_url" binding:"omitempty,url"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// RegisterProductResponse contains the registered product with credentials
type RegisterProductResponse struct {
	Product   *models.Product `json:"product"`
	APIKey    string          `json:"api_key"`
	APISecret string          `json:"api_secret"` // Only returned once during registration
}

// UpdateProductRequest contains data for updating a product
type UpdateProductRequest struct {
	Name          *string                `json:"name" binding:"omitempty,min=1,max=255"`
	Description   *string                `json:"description" binding:"omitempty,max=1000"`
	Status        *models.ProductStatus  `json:"status"`
	WebhookURL    *string                `json:"webhook_url" binding:"omitempty,url"`
	WebhookSecret *string                `json:"webhook_secret"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// ListProductsRequest contains filters for listing products
type ListProductsRequest struct {
	Status *models.ProductStatus `json:"status"`
	Limit  int                   `json:"limit"`
	Offset int                   `json:"offset"`
}

// Register creates a new product with generated API credentials
func (s *ProductService) Register(ctx context.Context, req RegisterProductRequest) (*RegisterProductResponse, error) {
	s.logger.Debug("Registering new product",
		zap.String("name", req.Name),
	)

	// Generate API key and secret
	apiKey, err := generateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate API key: %w", err)
	}

	apiSecret, err := generateAPISecret()
	if err != nil {
		return nil, fmt.Errorf("failed to generate API secret: %w", err)
	}

	// Hash the API secret
	secretHash, err := bcrypt.GenerateFromPassword([]byte(apiSecret), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash API secret: %w", err)
	}

	// Create product model
	product := &models.Product{
		BaseModel: models.BaseModel{
			ID: uuid.New(),
		},
		Name:          req.Name,
		Description:   req.Description,
		APIKey:        apiKey,
		APISecretHash: string(secretHash),
		Status:        models.ProductStatusActive,
		WebhookURL:    req.WebhookURL,
		Metadata:      req.Metadata,
	}

	// Validate product
	if err := product.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	// Create in repository
	if err := s.repo.Create(ctx, product); err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return nil, ErrDuplicateAPIKey
		}
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	s.logger.Info("Product registered",
		zap.String("product_id", product.ID.String()),
		zap.String("name", product.Name),
	)

	return &RegisterProductResponse{
		Product:   product,
		APIKey:    apiKey,
		APISecret: apiSecret, // Return secret only once
	}, nil
}

// GetByID retrieves a product by ID
func (s *ProductService) GetByID(ctx context.Context, productID uuid.UUID) (*models.Product, error) {
	product, err := s.repo.GetByID(ctx, productID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return product, nil
}

// GetByAPIKey retrieves a product by API key
func (s *ProductService) GetByAPIKey(ctx context.Context, apiKey string) (*models.Product, error) {
	product, err := s.repo.GetByAPIKey(ctx, apiKey)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return product, nil
}

// ValidateCredentials validates API key and secret, returns the product if valid
func (s *ProductService) ValidateCredentials(ctx context.Context, apiKey, apiSecret string) (*models.Product, error) {
	product, err := s.repo.GetByAPIKey(ctx, apiKey)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	// Verify API secret
	if err := bcrypt.CompareHashAndPassword([]byte(product.APISecretHash), []byte(apiSecret)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Check if product is active
	if !product.IsActive() {
		return nil, ErrProductInactive
	}

	return product, nil
}

// Update updates a product
func (s *ProductService) Update(ctx context.Context, productID uuid.UUID, req UpdateProductRequest) (*models.Product, error) {
	// Get existing product
	product, err := s.GetByID(ctx, productID)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if req.Name != nil {
		product.Name = *req.Name
	}
	if req.Description != nil {
		product.Description = *req.Description
	}
	if req.Status != nil {
		product.Status = *req.Status
	}
	if req.WebhookURL != nil {
		product.WebhookURL = *req.WebhookURL
	}
	if req.WebhookSecret != nil {
		product.WebhookSecret = *req.WebhookSecret
	}
	if req.Metadata != nil {
		product.Metadata = req.Metadata
	}

	// Validate updated product
	if err := product.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	// Save updates
	if err := s.repo.Update(ctx, product); err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	s.logger.Info("Product updated",
		zap.String("product_id", productID.String()),
	)

	return product, nil
}

// List retrieves products with optional filters
func (s *ProductService) List(ctx context.Context, req ListProductsRequest) ([]models.Product, error) {
	// Set defaults
	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	filters := repository.ProductFilters{
		Status: req.Status,
		Limit:  limit,
		Offset: req.Offset,
	}

	products, err := s.repo.List(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}

	return products, nil
}

// Delete soft-deletes a product
func (s *ProductService) Delete(ctx context.Context, productID uuid.UUID) error {
	// Verify product exists
	_, err := s.GetByID(ctx, productID)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, productID); err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	s.logger.Info("Product deleted",
		zap.String("product_id", productID.String()),
	)

	return nil
}

// RegenerateCredentials generates new API key and secret for a product
func (s *ProductService) RegenerateCredentials(ctx context.Context, productID uuid.UUID) (*RegisterProductResponse, error) {
	// Get existing product
	product, err := s.GetByID(ctx, productID)
	if err != nil {
		return nil, err
	}

	// Generate new API key and secret
	apiKey, err := generateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate API key: %w", err)
	}

	apiSecret, err := generateAPISecret()
	if err != nil {
		return nil, fmt.Errorf("failed to generate API secret: %w", err)
	}

	// Hash the new API secret
	secretHash, err := bcrypt.GenerateFromPassword([]byte(apiSecret), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash API secret: %w", err)
	}

	// Update product with new credentials
	product.APIKey = apiKey
	product.APISecretHash = string(secretHash)

	if err := s.repo.Update(ctx, product); err != nil {
		return nil, fmt.Errorf("failed to update product credentials: %w", err)
	}

	s.logger.Info("Product credentials regenerated",
		zap.String("product_id", productID.String()),
	)

	return &RegisterProductResponse{
		Product:   product,
		APIKey:    apiKey,
		APISecret: apiSecret,
	}, nil
}

// generateAPIKey generates a unique API key
func generateAPIKey() (string, error) {
	// Format: apk_<32 hex chars>
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "apk_" + hex.EncodeToString(bytes), nil
}

// generateAPISecret generates a secure API secret
func generateAPISecret() (string, error) {
	// Format: aps_<64 hex chars>
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "aps_" + hex.EncodeToString(bytes), nil
}
