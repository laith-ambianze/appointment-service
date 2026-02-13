// Package database provides PostgreSQL database connection management
// using pgx v5 with connection pooling, health checks, and transaction support.
package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Config holds database configuration parameters
type Config struct {
	Host            string
	Port            string
	User            string
	Password        string
	Database        string
	SSLMode         string
	MaxConnections  int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// PostgresDB wraps the pgx connection pool with logging and health check capabilities
type PostgresDB struct {
	Pool   *pgxpool.Pool
	logger *zap.Logger
	config Config
}

// NewPostgresDB creates a new PostgreSQL connection pool with the provided configuration.
// It validates the connection by performing a ping and returns an error if the connection fails.
func NewPostgresDB(cfg Config, logger *zap.Logger) (*PostgresDB, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	// Build connection string
	dsn := buildDSN(cfg)

	// Configure connection pool
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Set pool configuration with sensible defaults
	poolConfig.MaxConns = int32(cfg.MaxConnections)
	if poolConfig.MaxConns <= 0 {
		poolConfig.MaxConns = 25
	}

	poolConfig.MinConns = int32(cfg.MaxIdleConns)
	if poolConfig.MinConns <= 0 {
		poolConfig.MinConns = 5
	}

	poolConfig.MaxConnLifetime = cfg.ConnMaxLifetime
	if poolConfig.MaxConnLifetime <= 0 {
		poolConfig.MaxConnLifetime = 5 * time.Minute
	}

	poolConfig.MaxConnIdleTime = 30 * time.Minute
	poolConfig.HealthCheckPeriod = 1 * time.Minute

	// Create connection pool with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database connection established",
		zap.String("host", cfg.Host),
		zap.String("port", cfg.Port),
		zap.String("database", cfg.Database),
		zap.Int("max_connections", cfg.MaxConnections),
		zap.Int("min_connections", cfg.MaxIdleConns),
	)

	return &PostgresDB{
		Pool:   pool,
		logger: logger,
		config: cfg,
	}, nil
}

// buildDSN constructs the PostgreSQL connection string from config
func buildDSN(cfg Config) string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode,
	)
}

// Close closes the database connection pool gracefully
func (db *PostgresDB) Close() {
	if db.Pool != nil {
		db.logger.Info("Closing database connection pool")
		db.Pool.Close()
	}
}

// Ping checks if the database connection is alive
func (db *PostgresDB) Ping(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}

// Stats returns connection pool statistics
func (db *PostgresDB) Stats() *pgxpool.Stat {
	return db.Pool.Stat()
}

// GetPool returns the underlying connection pool for direct access
// Use this when you need to execute queries directly
func (db *PostgresDB) GetPool() *pgxpool.Pool {
	return db.Pool
}

// IsConnected returns true if the database is reachable
func (db *PostgresDB) IsConnected(ctx context.Context) bool {
	return db.Ping(ctx) == nil
}
