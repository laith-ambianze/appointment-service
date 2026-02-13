package database

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// getTestConfig returns a database config for testing.
// Uses environment variables or defaults to local docker-compose settings.
func getTestConfig() Config {
	host := os.Getenv("TEST_DB_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("TEST_DB_PORT")
	if port == "" {
		port = "5433"
	}

	user := os.Getenv("TEST_DB_USER")
	if user == "" {
		user = "appointments"
	}

	password := os.Getenv("TEST_DB_PASSWORD")
	if password == "" {
		password = "secure_password"
	}

	database := os.Getenv("TEST_DB_NAME")
	if database == "" {
		database = "appointments_dev"
	}

	return Config{
		Host:            host,
		Port:            port,
		User:            user,
		Password:        password,
		Database:        database,
		SSLMode:         "disable",
		MaxConnections:  5,
		MaxIdleConns:    2,
		ConnMaxLifetime: 5 * time.Minute,
	}
}

// skipIfNoDatabase skips the test if database is not available
func skipIfNoDatabase(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}
	if os.Getenv("SKIP_DB_TESTS") == "true" {
		t.Skip("Skipping database tests (SKIP_DB_TESTS=true)")
	}
}

func TestNewPostgresDB_NilLogger(t *testing.T) {
	cfg := getTestConfig()

	db, err := NewPostgresDB(cfg, nil)
	assert.Error(t, err)
	assert.Nil(t, db)
	assert.Contains(t, err.Error(), "logger is required")
}

func TestNewPostgresDB_Integration(t *testing.T) {
	skipIfNoDatabase(t)

	cfg := getTestConfig()
	logger := zap.NewNop()

	db, err := NewPostgresDB(cfg, logger)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer db.Close()

	// Test ping
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Verify connection
	assert.True(t, db.IsConnected(ctx))

	err = db.Ping(ctx)
	assert.NoError(t, err)
}

func TestPostgresDB_Stats_Integration(t *testing.T) {
	skipIfNoDatabase(t)

	cfg := getTestConfig()
	logger := zap.NewNop()

	db, err := NewPostgresDB(cfg, logger)
	require.NoError(t, err)
	defer db.Close()

	stats := db.Stats()
	assert.NotNil(t, stats)
	assert.GreaterOrEqual(t, stats.TotalConns(), int32(0))
}

func TestPostgresDB_Health_Integration(t *testing.T) {
	skipIfNoDatabase(t)

	cfg := getTestConfig()
	logger := zap.NewNop()

	db, err := NewPostgresDB(cfg, logger)
	require.NoError(t, err)
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	health := db.Health(ctx)
	assert.Equal(t, HealthStatusHealthy, health.Status)
	assert.NotEmpty(t, health.ResponseTime)
}

func TestPostgresDB_QuickHealth_Integration(t *testing.T) {
	skipIfNoDatabase(t)

	cfg := getTestConfig()
	logger := zap.NewNop()

	db, err := NewPostgresDB(cfg, logger)
	require.NoError(t, err)
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	isHealthy := db.QuickHealth(ctx)
	assert.True(t, isHealthy)
}

func TestPostgresDB_Transaction_Integration(t *testing.T) {
	skipIfNoDatabase(t)

	cfg := getTestConfig()
	logger := zap.NewNop()

	db, err := NewPostgresDB(cfg, logger)
	require.NoError(t, err)
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test successful transaction
	err = db.WithTransaction(ctx, func(tx pgx.Tx) error {
		// Execute a simple query within transaction
		_, err := tx.Exec(ctx, "SELECT 1")
		return err
	})
	assert.NoError(t, err)
}

func TestPostgresDB_TransactionRollback_Integration(t *testing.T) {
	skipIfNoDatabase(t)

	cfg := getTestConfig()
	logger := zap.NewNop()

	db, err := NewPostgresDB(cfg, logger)
	require.NoError(t, err)
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test transaction rollback on error
	customErr := assert.AnError
	err = db.WithTransaction(ctx, func(tx pgx.Tx) error {
		return customErr
	})
	assert.ErrorIs(t, err, customErr)
}

func TestPostgresDB_ReadOnlyTransaction_Integration(t *testing.T) {
	skipIfNoDatabase(t)

	cfg := getTestConfig()
	logger := zap.NewNop()

	db, err := NewPostgresDB(cfg, logger)
	require.NoError(t, err)
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test read-only transaction
	err = db.WithReadOnlyTransaction(ctx, func(tx pgx.Tx) error {
		// Read operations should work
		_, err := tx.Exec(ctx, "SELECT 1")
		return err
	})
	assert.NoError(t, err)
}

// Error helper tests
func TestIsNotFound(t *testing.T) {
	assert.True(t, IsNotFound(ErrNotFound))
	assert.False(t, IsNotFound(ErrDuplicate))
	assert.False(t, IsNotFound(nil))
}

func TestIsUniqueViolation(t *testing.T) {
	// Create a mock PostgreSQL unique violation error
	pgErr := &pgconn.PgError{Code: UniqueViolationCode}
	assert.True(t, IsUniqueViolation(pgErr))
	assert.False(t, IsUniqueViolation(ErrNotFound))
	assert.False(t, IsUniqueViolation(nil))
}

func TestIsForeignKeyViolation(t *testing.T) {
	// Create a mock PostgreSQL foreign key violation error
	pgErr := &pgconn.PgError{Code: ForeignKeyViolationCode}
	assert.True(t, IsForeignKeyViolation(pgErr))
	assert.False(t, IsForeignKeyViolation(ErrNotFound))
	assert.False(t, IsForeignKeyViolation(nil))
}

func TestWrapDatabaseError(t *testing.T) {
	// Test wrapping nil error
	assert.Nil(t, WrapDatabaseError(nil, "test"))

	// Test wrapping not found error
	err := WrapDatabaseError(pgx.ErrNoRows, "test")
	assert.ErrorIs(t, err, ErrNotFound)

	// Test wrapping unique violation
	pgErr := &pgconn.PgError{Code: UniqueViolationCode}
	err = WrapDatabaseError(pgErr, "test")
	assert.ErrorIs(t, err, ErrDuplicate)

	// Test wrapping foreign key violation
	pgErr = &pgconn.PgError{Code: ForeignKeyViolationCode}
	err = WrapDatabaseError(pgErr, "test")
	assert.ErrorIs(t, err, ErrForeignKey)
}

// Benchmark tests
func BenchmarkPostgresDB_Ping(b *testing.B) {
	if os.Getenv("SKIP_DB_TESTS") == "true" {
		b.Skip("Skipping database benchmark (SKIP_DB_TESTS=true)")
	}

	cfg := getTestConfig()
	logger := zap.NewNop()

	db, err := NewPostgresDB(cfg, logger)
	if err != nil {
		b.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = db.Ping(ctx)
	}
}

func BenchmarkPostgresDB_Health(b *testing.B) {
	if os.Getenv("SKIP_DB_TESTS") == "true" {
		b.Skip("Skipping database benchmark (SKIP_DB_TESTS=true)")
	}

	cfg := getTestConfig()
	logger := zap.NewNop()

	db, err := NewPostgresDB(cfg, logger)
	if err != nil {
		b.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = db.Health(ctx)
	}
}
