package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	startTime time.Time
}

// NewHealthHandler creates a new health handler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{
		startTime: time.Now(),
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Service   string    `json:"service"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
	Uptime    string    `json:"uptime"`
}

// Health returns the service health status
// @Summary Health check endpoint
// @Description Returns the health status of the service
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	uptime := time.Since(h.startTime)

	c.JSON(http.StatusOK, HealthResponse{
		Status:    "ok",
		Service:   "appointment-service",
		Version:   "1.0.0",
		Timestamp: time.Now().UTC(),
		Uptime:    uptime.String(),
	})
}

// Ready returns readiness status (for Kubernetes)
// @Summary Readiness check endpoint
// @Description Returns whether the service is ready to accept traffic
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /ready [get]
func (h *HealthHandler) Ready(c *gin.Context) {
	// TODO: Add actual readiness checks (database connection, etc.)
	c.JSON(http.StatusOK, HealthResponse{
		Status:    "ready",
		Service:   "appointment-service",
		Version:   "1.0.0",
		Timestamp: time.Now().UTC(),
		Uptime:    time.Since(h.startTime).String(),
	})
}

// Live returns liveness status (for Kubernetes)
// @Summary Liveness check endpoint
// @Description Returns whether the service is alive
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /live [get]
func (h *HealthHandler) Live(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status:    "alive",
		Service:   "appointment-service",
		Version:   "1.0.0",
		Timestamp: time.Now().UTC(),
		Uptime:    time.Since(h.startTime).String(),
	})
}
