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
	ProductHandler     *handlers.ProductHandler // Optional: can be nil if not needed
	AuthHandler        *handlers.AuthHandler    // Optional: can be nil if not needed
}

// RegisterRoutes registers all application routes
func RegisterRoutes(cfg Config) {
	// Public health check endpoints (no auth required)
	cfg.Router.GET("/health", cfg.HealthHandler.Health)
	cfg.Router.GET("/live", cfg.HealthHandler.Live)

	// Public product endpoints (no JWT required)
	if cfg.ProductHandler != nil {
		// Product registration - public endpoint
		cfg.Router.POST("/v1/products/register", cfg.ProductHandler.Register)
		// Credential validation - public endpoint
		cfg.Router.POST("/v1/products/validate", cfg.ProductHandler.ValidateCredentials)
	}

	// Public auth endpoint (no JWT required - uses API key/secret instead)
	if cfg.AuthHandler != nil {
		cfg.Router.POST("/v1/auth/token", cfg.AuthHandler.GenerateToken)
	}

	// JWT Auth middleware config
	authConfig := middleware.JWTAuthConfig{
		JWTManager: cfg.JWTManager,
		Logger:     cfg.Logger,
		SkipPaths: []string{
			"/health",
			"/live",
			"/ready",
			"/v1/docs/*",
			"/v1/products/register",
			"/v1/products/validate",
			"/v1/auth/token",
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

		// Products - Protected endpoints
		if cfg.ProductHandler != nil {
			products := v1.Group("/products")
			{
				// Get current product (from JWT) - any authenticated user
				products.GET("/me", cfg.ProductHandler.GetCurrent)

				// Update current product - any authenticated user
				products.PATCH("/me", cfg.ProductHandler.UpdateCurrent)

				// Regenerate credentials for current product - any authenticated user
				products.POST("/me/regenerate-credentials", cfg.ProductHandler.RegenerateCredentials)

				// Admin-only product management
				// List all products - admin only
				products.GET("", middleware.RequireAdmin(), cfg.ProductHandler.List)

				// Get product by ID - admin only
				products.GET("/:id", middleware.RequireAdmin(), cfg.ProductHandler.GetByID)

				// Update product by ID - admin only
				products.PATCH("/:id", middleware.RequireAdmin(), cfg.ProductHandler.Update)

				// Delete product - admin only
				products.DELETE("/:id", middleware.RequireAdmin(), cfg.ProductHandler.Delete)
			}
		}
	}
}

// RegisterHealthRoutes registers only health check routes (useful for testing)
func RegisterHealthRoutes(router *gin.Engine, healthHandler *handlers.HealthHandler) {
	router.GET("/health", healthHandler.Health)
	router.GET("/live", healthHandler.Live)
}
