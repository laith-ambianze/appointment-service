package database

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Common database error codes from PostgreSQL
const (
	// Class 23 - Integrity Constraint Violation
	UniqueViolationCode     = "23505"
	ForeignKeyViolationCode = "23503"
	NotNullViolationCode    = "23502"
	CheckViolationCode      = "23514"

	// Class 42 - Syntax Error or Access Rule Violation
	UndefinedTableCode  = "42P01"
	UndefinedColumnCode = "42703"

	// Class 53 - Insufficient Resources
	DiskFullCode           = "53100"
	OutOfMemoryCode        = "53200"
	TooManyConnectionsCode = "53300"
)

// Common database errors
var (
	ErrNotFound          = errors.New("record not found")
	ErrDuplicate         = errors.New("duplicate record")
	ErrForeignKey        = errors.New("foreign key constraint violation")
	ErrNotNull           = errors.New("not null constraint violation")
	ErrCheckConstraint   = errors.New("check constraint violation")
	ErrInvalidInput      = errors.New("invalid input")
	ErrDatabase          = errors.New("database error")
	ErrConnectionFailed  = errors.New("database connection failed")
	ErrTransactionFailed = errors.New("transaction failed")
)

// IsNotFound returns true if the error indicates no rows were found
func IsNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows) || errors.Is(err, ErrNotFound)
}

// IsUniqueViolation returns true if the error is a unique constraint violation
func IsUniqueViolation(err error) bool {
	return isPgErrorCode(err, UniqueViolationCode)
}

// IsForeignKeyViolation returns true if the error is a foreign key constraint violation
func IsForeignKeyViolation(err error) bool {
	return isPgErrorCode(err, ForeignKeyViolationCode)
}

// IsNotNullViolation returns true if the error is a not-null constraint violation
func IsNotNullViolation(err error) bool {
	return isPgErrorCode(err, NotNullViolationCode)
}

// IsCheckViolation returns true if the error is a check constraint violation
func IsCheckViolation(err error) bool {
	return isPgErrorCode(err, CheckViolationCode)
}

// IsTooManyConnections returns true if the error indicates too many database connections
func IsTooManyConnections(err error) bool {
	return isPgErrorCode(err, TooManyConnectionsCode)
}

// isPgErrorCode checks if the error is a PostgreSQL error with the given code
func isPgErrorCode(err error, code string) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == code
	}
	return false
}

// GetPgErrorCode returns the PostgreSQL error code if available
func GetPgErrorCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}

// GetPgErrorMessage returns a human-readable message for PostgreSQL errors
func GetPgErrorMessage(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Message
	}
	return err.Error()
}

// GetPgErrorDetail returns detailed information about PostgreSQL errors
func GetPgErrorDetail(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Detail
	}
	return ""
}

// WrapDatabaseError wraps a database error with context
func WrapDatabaseError(err error, operation string) error {
	if err == nil {
		return nil
	}

	// Check for specific error types
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}

	if IsUniqueViolation(err) {
		return ErrDuplicate
	}

	if IsForeignKeyViolation(err) {
		return ErrForeignKey
	}

	if IsNotNullViolation(err) {
		return ErrNotNull
	}

	if IsCheckViolation(err) {
		return ErrCheckConstraint
	}

	// Return generic database error
	return errors.Join(ErrDatabase, err)
}
