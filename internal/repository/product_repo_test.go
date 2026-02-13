package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/laith-ambianze/appointment-service/internal/models"
	"github.com/laith-ambianze/appointment-service/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// TestProductRepository_Create tests creating a product
func TestProductRepository_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup
	db := setupTestDB(t)
	defer db.Close()

	repo := NewProductRepository(db.GetPool(), zap.NewNop())
	ctx := context.Background()

	// Create test product
	apiSecretHash, _ := bcrypt.GenerateFromPassword([]byte("secret"), 10)
	product := &models.Product{
		Name:          "Test Product",
		Description:   "A test product",
		APIKey:        "test_key_" + uuid.NewString()[:8],
		APISecretHash: string(apiSecretHash),
		Status:        models.ProductStatusActive,
		Metadata: map[string]interface{}{
			"test": true,
		},
	}

	// Test
	err := repo.Create(ctx, product)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, product.ID)
	assert.NotZero(t, product.CreatedAt)

	// Cleanup
	_ = repo.HardDelete(ctx, product.ID)
}

// TestProductRepository_GetByID tests retrieving a product by ID
func TestProductRepository_GetByID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup
	db := setupTestDB(t)
	defer db.Close()

	repo := NewProductRepository(db.GetPool(), zap.NewNop())
	ctx := context.Background()

	// Create test product
	product := createTestProduct(t, repo, ctx)
	defer repo.HardDelete(ctx, product.ID)

	// Test
	retrieved, err := repo.GetByID(ctx, product.ID)
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	assert.Equal(t, product.Name, retrieved.Name)
	assert.Equal(t, product.APIKey, retrieved.APIKey)

	// Test not found
	_, err = repo.GetByID(ctx, uuid.New())
	assert.ErrorIs(t, err, ErrNotFound)
}

// TestProductRepository_GetByAPIKey tests retrieving a product by API key
func TestProductRepository_GetByAPIKey(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup
	db := setupTestDB(t)
	defer db.Close()

	repo := NewProductRepository(db.GetPool(), zap.NewNop())
	ctx := context.Background()

	// Create test product
	product := createTestProduct(t, repo, ctx)
	defer repo.HardDelete(ctx, product.ID)

	// Test
	retrieved, err := repo.GetByAPIKey(ctx, product.APIKey)
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	assert.Equal(t, product.ID, retrieved.ID)
}

// TestProductRepository_List tests listing products with filters
func TestProductRepository_List(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup
	db := setupTestDB(t)
	defer db.Close()

	repo := NewProductRepository(db.GetPool(), zap.NewNop())
	ctx := context.Background()

	// Create test products
	product1 := createTestProduct(t, repo, ctx)
	defer repo.HardDelete(ctx, product1.ID)

	product2 := createTestProduct(t, repo, ctx)
	product2.Status = models.ProductStatusInactive
	repo.Update(ctx, product2)
	defer repo.HardDelete(ctx, product2.ID)

	// Test list all
	products, err := repo.List(ctx, ProductFilters{})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(products), 2)

	// Test list with status filter
	activeStatus := models.ProductStatusActive
	products, err = repo.List(ctx, ProductFilters{Status: &activeStatus})
	require.NoError(t, err)
	for _, p := range products {
		assert.Equal(t, models.ProductStatusActive, p.Status)
	}

	// Test list with pagination
	products, err = repo.List(ctx, ProductFilters{Limit: 1})
	require.NoError(t, err)
	assert.Equal(t, 1, len(products))
}

// TestProductRepository_Update tests updating a product
func TestProductRepository_Update(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup
	db := setupTestDB(t)
	defer db.Close()

	repo := NewProductRepository(db.GetPool(), zap.NewNop())
	ctx := context.Background()

	// Create test product
	product := createTestProduct(t, repo, ctx)
	defer repo.HardDelete(ctx, product.ID)

	// Update
	product.Name = "Updated Name"
	product.Status = models.ProductStatusInactive
	err := repo.Update(ctx, product)
	require.NoError(t, err)

	// Verify
	retrieved, _ := repo.GetByID(ctx, product.ID)
	assert.Equal(t, "Updated Name", retrieved.Name)
	assert.Equal(t, models.ProductStatusInactive, retrieved.Status)
}

// TestProductRepository_Delete tests soft-deleting a product
func TestProductRepository_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup
	db := setupTestDB(t)
	defer db.Close()

	repo := NewProductRepository(db.GetPool(), zap.NewNop())
	ctx := context.Background()

	// Create test product
	product := createTestProduct(t, repo, ctx)
	defer repo.HardDelete(ctx, product.ID)

	// Delete
	err := repo.Delete(ctx, product.ID)
	require.NoError(t, err)

	// Verify not found
	_, err = repo.GetByID(ctx, product.ID)
	assert.ErrorIs(t, err, ErrNotFound)
}

// Helper functions

func setupTestDB(t *testing.T) *database.PostgresDB {
	cfg := database.Config{
		Host:            "127.0.0.1",
		Port:            "5433",
		User:            "appointments",
		Password:        "password123",
		Database:        "appointments_dev",
		SSLMode:         "disable",
		MaxConnections:  10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
	}

	db, err := database.NewPostgresDB(cfg, zap.NewNop())
	require.NoError(t, err)

	// Wait for database to be ready
	time.Sleep(1 * time.Second)

	return db
}

func createTestProduct(t *testing.T, repo ProductRepository, ctx context.Context) *models.Product {
	apiSecretHash, _ := bcrypt.GenerateFromPassword([]byte("secret"), 10)
	product := &models.Product{
		Name:          "Test Product " + uuid.NewString()[:8],
		Description:   "A test product",
		APIKey:        "test_key_" + uuid.NewString()[:8],
		APISecretHash: string(apiSecretHash),
		Status:        models.ProductStatusActive,
		Metadata: map[string]interface{}{
			"test": true,
		},
	}

	err := repo.Create(ctx, product)
	require.NoError(t, err)

	return product
}
