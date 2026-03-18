package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/laith-ambianze/appointment-service/internal/config"
	"github.com/laith-ambianze/appointment-service/internal/handlers"
	"github.com/laith-ambianze/appointment-service/internal/middleware"
	"github.com/laith-ambianze/appointment-service/internal/repository"
	"github.com/laith-ambianze/appointment-service/internal/routes"
	"github.com/laith-ambianze/appointment-service/internal/service"
	"github.com/laith-ambianze/appointment-service/pkg/auth"
	"github.com/laith-ambianze/appointment-service/pkg/database"
	"github.com/laith-ambianze/appointment-service/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log, err := logger.New(cfg.LogLevel, cfg.LogFormat)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	log.Info("=== Appointment Service Starting ===")
	log.Info("Configuration",
		zap.String("environment", cfg.Env),
		zap.String("port", cfg.APIPort),
		zap.String("log_level", cfg.LogLevel),
	)

	// Initialize database connection
	dbConfig := database.Config{
		Host:            cfg.DBHost,
		Port:            cfg.DBPort,
		User:            cfg.DBUser,
		Password:        cfg.DBPassword,
		Database:        cfg.DBName,
		SSLMode:         cfg.DBSSLMode,
		MaxConnections:  cfg.DBMaxConnections,
		MaxIdleConns:    cfg.DBMaxIdleConnections,
		ConnMaxLifetime: parseDuration(cfg.DBMaxLifetime, 5*time.Minute),
	}

	db, err := database.NewPostgresDB(dbConfig, log.Logger)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()
	log.Info("Database connected",
		zap.String("host", cfg.DBHost),
		zap.String("database", cfg.DBName),
	)

	// Set Gin mode
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	router := gin.New()

	// Add middleware
	router.Use(gin.Recovery())
	router.Use(middleware.CORSFromConfig(cfg.CORSAllowedOrigins, cfg.CORSAllowedMethods, cfg.CORSAllowedHeaders))
	router.Use(loggerMiddleware(log))

	// Initialize JWT manager
	jwtManager := auth.NewJWTManager(cfg.JWTSecret)

	// Initialize repositories
	appointmentRepo := repository.NewAppointmentRepository(db.Pool, log.Logger)
	productRepo := repository.NewProductRepository(db.Pool, log.Logger)
	availabilityRepo := repository.NewAvailabilityRepository(db.Pool, log.Logger)

	// Initialize services
	appointmentService := service.NewAppointmentService(appointmentRepo, log.Logger)
	productService := service.NewProductService(productRepo, log.Logger)
	availabilityService := service.NewAvailabilityService(availabilityRepo, appointmentRepo, log.Logger)

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler()
	appointmentHandler := handlers.NewAppointmentHandler(appointmentService, log.Logger)
	productHandler := handlers.NewProductHandler(productService, log.Logger)
	authHandler := handlers.NewAuthHandler(productService, jwtManager, log.Logger)
	availabilityHandler := handlers.NewAvailabilityHandler(availabilityService, log.Logger)

	// Setup routes
	setupRoutes(router, db, jwtManager, log.Logger, healthHandler, appointmentHandler, productHandler, authHandler, availabilityHandler)
	log.Info("Routes registered successfully")

	// Create HTTP server
	addr := fmt.Sprintf("%s:%s", cfg.APIHost, cfg.APIPort)
	srv := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Info("Server listening on " + addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("=== Shutdown signal received ===")

	// Graceful shutdown with 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown", zap.Error(err))
	}

	log.Info("=== Server shutdown complete ===")
}

// setupRoutes configures all application routes
func setupRoutes(router *gin.Engine, db *database.PostgresDB, jwtManager *auth.JWTManager, zapLogger *zap.Logger, healthHandler *handlers.HealthHandler, appointmentHandler *handlers.AppointmentHandler, productHandler *handlers.ProductHandler, authHandler *handlers.AuthHandler, availabilityHandler *handlers.AvailabilityHandler) {
	// Register all routes using the routes package
	routes.RegisterRoutes(routes.Config{
		Router:              router,
		JWTManager:          jwtManager,
		Logger:              zapLogger,
		HealthHandler:       healthHandler,
		AppointmentHandler:  appointmentHandler,
		ProductHandler:      productHandler,
		AuthHandler:         authHandler,
		AvailabilityHandler: availabilityHandler,
	})

	// Ready endpoint with database health check
	router.GET("/ready", func(c *gin.Context) {
		ctx := c.Request.Context()
		dbHealth := db.Health(ctx)

		response := gin.H{
			"status": "ok",
			"checks": gin.H{
				"database": gin.H{
					"status":        dbHealth.Status,
					"response_time": dbHealth.ResponseTime,
					"connections":   dbHealth.Connections.TotalConnections,
				},
			},
		}

		if dbHealth.Status != database.HealthStatusHealthy {
			response["status"] = "degraded"
			if dbHealth.Status == database.HealthStatusUnhealthy {
				response["status"] = "unhealthy"
				c.JSON(http.StatusServiceUnavailable, response)
				return
			}
		}

		c.JSON(http.StatusOK, response)
	})
}

// loggerMiddleware creates a Gin middleware for logging
func loggerMiddleware(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		// Skip logging for health checks and metrics (reduces noise)
		if path == "/health" || path == "/metrics" || path == "/ready" {
			c.Next()
			return
		}

		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log request
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		if raw != "" {
			path = path + "?" + raw
		}

		// Format: METHOD /path -> STATUS (latency)
		fields := []zap.Field{
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("ip", clientIP),
		}
		if errorMessage != "" {
			fields = append(fields, zap.String("error", errorMessage))
		}

		log.Info(method+" "+path, fields...)
	}
}

// parseDuration parses a duration string and returns default if parsing fails
func parseDuration(s string, defaultDuration time.Duration) time.Duration {
	if s == "" {
		return defaultDuration
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return defaultDuration
	}
	return d
}
