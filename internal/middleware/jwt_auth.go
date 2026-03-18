package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/laith-ambianze/appointment-service/pkg/auth"
	"go.uber.org/zap"
)

// Context keys for storing JWT claims data
const (
	// ContextKeyProductID stores the product UUID from JWT
	ContextKeyProductID = "product_id"
	// ContextKeyExternalUserID stores the external user ID from JWT
	// This is the user identifier from the integrating product's system
	ContextKeyExternalUserID = "external_user_id"
	// ContextKeyRole stores the user role from JWT
	ContextKeyRole = "role"
	// ContextKeyClaims stores the full claims object
	ContextKeyClaims = "claims"
)

// JWTAuthConfig holds configuration for JWT middleware
type JWTAuthConfig struct {
	// JWTManager handles token validation
	JWTManager *auth.JWTManager
	// Logger for logging authentication events
	Logger *zap.Logger
	// SkipPaths are paths that don't require authentication
	SkipPaths []string
	// AllowedRoles restricts access to specific roles (empty = all roles allowed)
	AllowedRoles []auth.Role
}

// JWTAuth creates a Gin middleware for JWT authentication
func JWTAuth(config JWTAuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if path should be skipped
		if shouldSkipPath(c.Request.URL.Path, config.SkipPaths) {
			c.Next()
			return
		}

		// Extract token from Authorization header
		tokenString, err := extractBearerToken(c.GetHeader("Authorization"))
		if err != nil {
			config.Logger.Debug("Missing or invalid authorization header",
				zap.String("path", c.Request.URL.Path),
				zap.Error(err),
			)
			abortWithAuthError(c, err)
			return
		}

		// Validate token
		claims, err := config.JWTManager.ValidateToken(tokenString)
		if err != nil {
			config.Logger.Debug("Token validation failed",
				zap.String("path", c.Request.URL.Path),
				zap.Error(err),
			)
			abortWithAuthError(c, err)
			return
		}

		// Check role restrictions if configured
		if len(config.AllowedRoles) > 0 && !isRoleAllowed(claims.Role, config.AllowedRoles) {
			config.Logger.Debug("Role not allowed",
				zap.String("path", c.Request.URL.Path),
				zap.String("role", string(claims.Role)),
				zap.Any("allowed_roles", config.AllowedRoles),
			)
			abortWithForbidden(c, "insufficient permissions")
			return
		}

		// Store claims in context for handlers to use
		c.Set(ContextKeyProductID, claims.ProductID)
		c.Set(ContextKeyExternalUserID, claims.ExternalUserID)
		c.Set(ContextKeyRole, claims.Role)
		c.Set(ContextKeyClaims, claims)

		// Log successful authentication (only in debug mode)
		config.Logger.Debug("Request authenticated",
			zap.String("external_user_id", claims.ExternalUserID),
			zap.String("product_id", claims.ProductID.String()),
			zap.String("role", string(claims.Role)),
		)

		c.Next()
	}
}

// extractBearerToken extracts the token from "Bearer <token>" format
func extractBearerToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", auth.ErrMissingToken
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", auth.ErrMalformedToken
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", auth.ErrMissingToken
	}

	return token, nil
}

// shouldSkipPath checks if the request path should skip authentication
func shouldSkipPath(path string, skipPaths []string) bool {
	for _, skip := range skipPaths {
		// Exact match
		if path == skip {
			return true
		}
		// Prefix match (for paths ending with *)
		if strings.HasSuffix(skip, "*") {
			prefix := strings.TrimSuffix(skip, "*")
			if strings.HasPrefix(path, prefix) {
				return true
			}
		}
	}
	return false
}

// isRoleAllowed checks if the role is in the allowed roles list
func isRoleAllowed(role auth.Role, allowedRoles []auth.Role) bool {
	for _, allowed := range allowedRoles {
		if role == allowed {
			return true
		}
	}
	return false
}

// abortWithAuthError responds with appropriate auth error
func abortWithAuthError(c *gin.Context, err error) {
	var message string
	var statusCode int

	switch {
	case auth.IsExpiredError(err):
		message = "token has expired"
		statusCode = http.StatusUnauthorized
	case err == auth.ErrMissingToken:
		message = "authorization token required"
		statusCode = http.StatusUnauthorized
	default:
		message = "invalid authorization token"
		statusCode = http.StatusUnauthorized
	}

	c.AbortWithStatusJSON(statusCode, gin.H{
		"error":   "unauthorized",
		"message": message,
	})
}

// abortWithForbidden responds with 403 Forbidden
func abortWithForbidden(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
		"error":   "forbidden",
		"message": message,
	})
}

// Helper functions to extract claims from context

// GetProductIDFromContext extracts product_id from Gin context
func GetProductIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	val, exists := c.Get(ContextKeyProductID)
	if !exists {
		return uuid.Nil, false
	}
	productID, ok := val.(uuid.UUID)
	return productID, ok
}

// GetExternalUserIDFromContext extracts external_user_id from Gin context
func GetExternalUserIDFromContext(c *gin.Context) (string, bool) {
	val, exists := c.Get(ContextKeyExternalUserID)
	if !exists {
		return "", false
	}
	externalUserID, ok := val.(string)
	return externalUserID, ok
}

// GetRoleFromContext extracts role from Gin context
func GetRoleFromContext(c *gin.Context) (auth.Role, bool) {
	val, exists := c.Get(ContextKeyRole)
	if !exists {
		return "", false
	}
	role, ok := val.(auth.Role)
	return role, ok
}

// GetClaimsFromContext extracts full claims from Gin context
func GetClaimsFromContext(c *gin.Context) (*auth.Claims, bool) {
	val, exists := c.Get(ContextKeyClaims)
	if !exists {
		return nil, false
	}
	claims, ok := val.(*auth.Claims)
	return claims, ok
}

// MustGetProductID extracts product_id or panics (use only when auth middleware is guaranteed)
func MustGetProductID(c *gin.Context) uuid.UUID {
	productID, ok := GetProductIDFromContext(c)
	if !ok {
		panic("product_id not found in context - ensure JWTAuth middleware is applied")
	}
	return productID
}

// MustGetExternalUserID extracts external_user_id or panics (use only when auth middleware is guaranteed)
func MustGetExternalUserID(c *gin.Context) string {
	externalUserID, ok := GetExternalUserIDFromContext(c)
	if !ok {
		panic("external_user_id not found in context - ensure JWTAuth middleware is applied")
	}
	return externalUserID
}

// MustGetRole extracts role or panics (use only when auth middleware is guaranteed)
func MustGetRole(c *gin.Context) auth.Role {
	role, ok := GetRoleFromContext(c)
	if !ok {
		panic("role not found in context - ensure JWTAuth middleware is applied")
	}
	return role
}
