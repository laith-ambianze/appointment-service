package database

import (
	"context"
	"time"
)

// HealthStatus represents the health status of the database
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusDegraded  HealthStatus = "degraded"
)

// HealthCheck represents the result of a database health check
type HealthCheck struct {
	Status       HealthStatus    `json:"status"`
	Message      string          `json:"message,omitempty"`
	Connections  ConnectionStats `json:"connections"`
	ResponseTime string          `json:"response_time"`
	Timestamp    time.Time       `json:"timestamp"`
}

// ConnectionStats holds connection pool statistics
type ConnectionStats struct {
	TotalConnections     int32 `json:"total_connections"`
	IdleConnections      int32 `json:"idle_connections"`
	AcquiredConnections  int32 `json:"acquired_connections"`
	MaxConnections       int32 `json:"max_connections"`
	AcquireCount         int64 `json:"acquire_count"`
	AcquireDurationMs    int64 `json:"acquire_duration_ms"`
	CanceledAcquireCount int64 `json:"canceled_acquire_count"`
	EmptyAcquireCount    int64 `json:"empty_acquire_count"`
}

// Health performs a comprehensive health check on the database connection.
// It pings the database, collects connection pool statistics, and returns
// a detailed health report.
func (db *PostgresDB) Health(ctx context.Context) HealthCheck {
	start := time.Now()

	// Try to ping the database
	err := db.Ping(ctx)
	responseTime := time.Since(start)

	if err != nil {
		return HealthCheck{
			Status:       HealthStatusUnhealthy,
			Message:      err.Error(),
			ResponseTime: responseTime.String(),
			Timestamp:    time.Now(),
		}
	}

	// Get connection pool stats
	stats := db.Stats()

	// Build connection statistics
	connStats := ConnectionStats{
		TotalConnections:     stats.TotalConns(),
		IdleConnections:      stats.IdleConns(),
		AcquiredConnections:  stats.AcquiredConns(),
		MaxConnections:       stats.MaxConns(),
		AcquireCount:         stats.AcquireCount(),
		AcquireDurationMs:    stats.AcquireDuration().Milliseconds(),
		CanceledAcquireCount: stats.CanceledAcquireCount(),
		EmptyAcquireCount:    stats.EmptyAcquireCount(),
	}

	// Check for degraded status (high connection usage)
	status := HealthStatusHealthy
	message := ""

	usagePercent := float64(stats.AcquiredConns()) / float64(stats.MaxConns()) * 100
	if usagePercent > 80 {
		status = HealthStatusDegraded
		message = "High connection pool usage (>80%)"
	}

	// Check if too many acquire operations are being canceled
	if stats.CanceledAcquireCount() > 100 {
		status = HealthStatusDegraded
		if message != "" {
			message += "; "
		}
		message += "High number of canceled acquire operations"
	}

	return HealthCheck{
		Status:       status,
		Message:      message,
		Connections:  connStats,
		ResponseTime: responseTime.String(),
		Timestamp:    time.Now(),
	}
}

// QuickHealth performs a quick health check (just ping) without collecting
// detailed statistics. Use this for frequent health checks.
func (db *PostgresDB) QuickHealth(ctx context.Context) bool {
	return db.Ping(ctx) == nil
}

// DetailedStats returns detailed connection pool statistics
func (db *PostgresDB) DetailedStats() ConnectionStats {
	stats := db.Stats()
	return ConnectionStats{
		TotalConnections:     stats.TotalConns(),
		IdleConnections:      stats.IdleConns(),
		AcquiredConnections:  stats.AcquiredConns(),
		MaxConnections:       stats.MaxConns(),
		AcquireCount:         stats.AcquireCount(),
		AcquireDurationMs:    stats.AcquireDuration().Milliseconds(),
		CanceledAcquireCount: stats.CanceledAcquireCount(),
		EmptyAcquireCount:    stats.EmptyAcquireCount(),
	}
}
