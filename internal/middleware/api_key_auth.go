package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/laith-ambianze/appointment-service/internal/models"
	"github.com/laith-ambianze/appointment-service/pkg/auth"
	"go.uber.org/zap"
)

// ProductCredentialsValidator interface for validating API credentials
// This is used to avoid circular dependency with service package
type ProductCredentialsValidator interface {
	ValidateCredentials(ctx context.Context, apiKey, apiSecret string) (*models.Product, error)
}

// APIKeyAuthConfig holds configuration for API Key authentication middleware
type APIKeyAuthConfig struct {
	// ProductValidator validates API credentials
	ProductValidator ProductCredentialsValidator
	// Logger for logging authentication events
	Logger *zap.Logger
	// SkipPaths are paths that don't require authentication
	SkipPaths []string
}

// APIKeyAuth creates a Gin middleware for API Key/Secret authentication
// This replaces JWT authentication - credentials are validated on every request
func APIKeyAuth(config APIKeyAuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if path should be skipped
		if shouldSkipPath(c.Request.URL.Path, config.SkipPaths) {
			c.Next()
			return
		}

		// Extract headers
		apiKey := c.GetHeader("X-API-Key")
		apiSecret := c.GetHeader("X-API-Secret")
		externalUserID := c.GetHeader("X-External-User-ID")
		roleStr := c.GetHeader("X-Role")

		// Validate required credentials
		if apiKey == "" || apiSecret == "" {
			config.Logger.Debug("Missing API credentials",
				zap.String("path", c.Request.URL.Path),
			)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "X-API-Key and X-API-Secret headers are required",
			})
			return
		}

		// Validate required user identification
		if externalUserID == "" {
			config.Logger.Debug("Missing external user ID",
				zap.String("path", c.Request.URL.Path),
			)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error":   "Bad Request",
				"message": "X-External-User-ID header is required",
			})
			return
		}

		// Validate role header
		if roleStr == "" {
			config.Logger.Debug("Missing role",
				zap.String("path", c.Request.URL.Path),
			)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error":   "Bad Request",
				"message": "X-Role header is required (admin, user, or provider)",
			})
			return
		}

		// Validate role value
		role := auth.Role(roleStr)
		if !role.IsValid() {
			config.Logger.Debug("Invalid role",
				zap.String("path", c.Request.URL.Path),
				zap.String("role", roleStr),
			)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error":   "Bad Request",
				"message": "X-Role must be one of: admin, user, provider",
			})
			return
		}

		// Validate API credentials against database
		product, err := config.ProductValidator.ValidateCredentials(c.Request.Context(), apiKey, apiSecret)
		if err != nil {
			config.Logger.Debug("Invalid API credentials",
				zap.String("path", c.Request.URL.Path),
				zap.String("api_key", apiKey),
				zap.Error(err),
			)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Invalid API credentials",
			})
			return
		}

		// Store authentication data in context (same keys as JWT middleware for compatibility)
		c.Set(ContextKeyProductID, product.ID)
		c.Set(ContextKeyExternalUserID, externalUserID)
		c.Set(ContextKeyRole, role)

		// Log successful authentication
		config.Logger.Debug("Request authenticated via API key",
			zap.String("product_id", product.ID.String()),
			zap.String("product_name", product.Name),
			zap.String("external_user_id", externalUserID),
			zap.String("role", roleStr),
		)

		c.Next()
	}
}
