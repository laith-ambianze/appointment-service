package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/laith-ambianze/appointment-service/internal/handlers"
	"github.com/laith-ambianze/appointment-service/internal/middleware"
	"github.com/laith-ambianze/appointment-service/pkg/auth"
	"go.uber.org/zap"
)

// Config holds dependencies for route registration
type Config struct {
	Router             *gin.Engine
	JWTManager         *auth.JWTManager
	Logger             *zap.Logger
	HealthHandler      *handlers.HealthHandler
	AppointmentHandler *handlers.AppointmentHandler
}

// RegisterRoutes registers all application routes
func RegisterRoutes(cfg Config) {
	// Public health check endpoints (no auth required)
	cfg.Router.GET("/health", cfg.HealthHandler.Health)
	cfg.Router.GET("/live", cfg.HealthHandler.Live)

	// JWT Auth middleware config
	authConfig := middleware.JWTAuthConfig{
		JWTManager: cfg.JWTManager,
		Logger:     cfg.Logger,
		SkipPaths: []string{
			"/health",
			"/live",
			"/ready",
			"/v1/docs/*",
		},
	}

	// API v1 routes - all require JWT authentication
	v1 := cfg.Router.Group("/v1")
	v1.Use(middleware.JWTAuth(authConfig))
	{
		// Appointments - Any authenticated user can access
		appointments := v1.Group("/appointments")
		{
			// Create appointment - any authenticated user
			appointments.POST("", cfg.AppointmentHandler.Create)

			// List user's appointments - any authenticated user
			appointments.GET("", cfg.AppointmentHandler.List)

			// Get single appointment - any authenticated user (multi-tenancy enforced in service)
			appointments.GET("/:id", cfg.AppointmentHandler.GetByID)

			// Update appointment - creator, admin, or provider
			appointments.PATCH("/:id", cfg.AppointmentHandler.Update)

			// Delete appointment - admin only
			appointments.DELETE("/:id", middleware.RequireAdmin(), cfg.AppointmentHandler.Delete)

			// Cancel appointment - creator, admin, or provider
			appointments.POST("/:id/cancel", cfg.AppointmentHandler.Cancel)

			// Respond to appointment (confirm/complete/etc.) - admin or provider only
			appointments.PATCH("/:id/response", middleware.RequireAdminOrProvider(), cfg.AppointmentHandler.Respond)

			// Participant management
			participants := appointments.Group("/:id/participants")
			{
				// Add participant - creator, admin, or provider
				participants.POST("", cfg.AppointmentHandler.AddParticipant)

				// Remove participant - creator, admin, or provider
				participants.DELETE("/:user_id", cfg.AppointmentHandler.RemoveParticipant)

				// Update participant status (accept/decline) - user can update own, admin/provider can update any
				participants.PATCH("/:user_id/status", cfg.AppointmentHandler.UpdateParticipantStatus)
			}
		}
	}
}

// RegisterHealthRoutes registers only health check routes (useful for testing)
func RegisterHealthRoutes(router *gin.Engine, healthHandler *handlers.HealthHandler) {
	router.GET("/health", healthHandler.Health)
	router.GET("/live", healthHandler.Live)
}
