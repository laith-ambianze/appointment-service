package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

// TxFunc represents a function that executes within a transaction.
// If the function returns an error, the transaction will be rolled back.
type TxFunc func(tx pgx.Tx) error

// WithTransaction executes a function within a database transaction.
// The transaction is automatically committed if the function returns nil,
// or rolled back if the function returns an error or panics.
func (db *PostgresDB) WithTransaction(ctx context.Context, fn TxFunc) error {
	return db.WithTransactionOptions(ctx, pgx.TxOptions{}, fn)
}

// WithTransactionOptions executes a function within a database transaction
// with custom transaction options (isolation level, read-only, etc.)
func (db *PostgresDB) WithTransactionOptions(ctx context.Context, opts pgx.TxOptions, fn TxFunc) error {
	// Begin transaction
	tx, err := db.Pool.BeginTx(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Use defer to handle commit/rollback
	defer func() {
		if p := recover(); p != nil {
			// Panic occurred, attempt rollback
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				db.logger.Error("Failed to rollback transaction after panic",
					zap.Error(rbErr),
					zap.Any("panic", p),
				)
			}
			panic(p) // Re-throw panic after rollback
		}
	}()

	// Execute the function within the transaction
	if err := fn(tx); err != nil {
		// Error occurred, rollback
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			db.logger.Error("Failed to rollback transaction",
				zap.Error(rbErr),
				zap.Error(err),
			)
			return fmt.Errorf("transaction error: %w, rollback error: %v", err, rbErr)
		}
		return err
	}

	// Success, commit
	if err := tx.Commit(ctx); err != nil {
		db.logger.Error("Failed to commit transaction", zap.Error(err))
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// WithReadOnlyTransaction executes a function within a read-only transaction.
// This is useful for queries that don't modify data and can benefit from
// read replicas or optimized read paths.
func (db *PostgresDB) WithReadOnlyTransaction(ctx context.Context, fn TxFunc) error {
	return db.WithTransactionOptions(ctx, pgx.TxOptions{
		AccessMode: pgx.ReadOnly,
	}, fn)
}

// WithSerializableTransaction executes a function within a serializable transaction.
// This provides the highest isolation level but may have performance implications.
func (db *PostgresDB) WithSerializableTransaction(ctx context.Context, fn TxFunc) error {
	return db.WithTransactionOptions(ctx, pgx.TxOptions{
		IsoLevel: pgx.Serializable,
	}, fn)
}
