package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORSConfig holds CORS middleware configuration
type CORSConfig struct {
	// AllowedOrigins is a list of origins that are allowed to access the resource.
	// Use "*" to allow all origins (not recommended for production with credentials).
	AllowedOrigins []string

	// AllowedMethods is a list of methods the client is allowed to use.
	AllowedMethods []string

	// AllowedHeaders is a list of headers the client is allowed to use.
	AllowedHeaders []string

	// ExposedHeaders is a list of headers that are safe to expose to the client.
	ExposedHeaders []string

	// AllowCredentials indicates whether the request can include user credentials.
	AllowCredentials bool

	// MaxAge indicates how long (in seconds) the results of a preflight request can be cached.
	MaxAge int
}

// DefaultCORSConfig returns a default CORS configuration suitable for development
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{
			"Origin",
			"Content-Type",
			"Content-Length",
			"Accept",
			"Accept-Encoding",
			"X-API-Key",
			"X-API-Secret",
			"X-External-User-ID",
			"X-Role",
			"X-Request-ID",
		},
		ExposedHeaders: []string{
			"Content-Length",
			"Content-Type",
			"X-Request-ID",
		},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	}
}

// CORS creates a Gin middleware for handling CORS requests
func CORS(config CORSConfig) gin.HandlerFunc {
	// Pre-compute allowed origins map for faster lookup
	allowAllOrigins := false
	originsMap := make(map[string]bool)
	for _, origin := range config.AllowedOrigins {
		if origin == "*" {
			allowAllOrigins = true
			break
		}
		originsMap[strings.ToLower(origin)] = true
	}

	// Pre-compute header strings
	allowMethods := strings.Join(config.AllowedMethods, ", ")
	allowHeaders := strings.Join(config.AllowedHeaders, ", ")
	exposeHeaders := strings.Join(config.ExposedHeaders, ", ")

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Determine if origin is allowed
		var allowedOrigin string
		if allowAllOrigins {
			if config.AllowCredentials {
				// When credentials are allowed, we can't use "*"
				// We must echo back the specific origin
				allowedOrigin = origin
			} else {
				allowedOrigin = "*"
			}
		} else if origin != "" && originsMap[strings.ToLower(origin)] {
			allowedOrigin = origin
		}

		// Set CORS headers if origin is allowed
		if allowedOrigin != "" {
			c.Header("Access-Control-Allow-Origin", allowedOrigin)
			c.Header("Access-Control-Allow-Methods", allowMethods)
			c.Header("Access-Control-Allow-Headers", allowHeaders)

			if exposeHeaders != "" {
				c.Header("Access-Control-Expose-Headers", exposeHeaders)
			}

			if config.AllowCredentials {
				c.Header("Access-Control-Allow-Credentials", "true")
			}

			if config.MaxAge > 0 {
				c.Header("Access-Control-Max-Age", strconv.Itoa(config.MaxAge))
			}
		}

		// Handle preflight request
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// CORSFromConfig creates CORS middleware from string configuration
// This is useful when loading from environment variables
func CORSFromConfig(origins, methods, headers string) gin.HandlerFunc {
	config := DefaultCORSConfig()

	if origins != "" {
		config.AllowedOrigins = splitAndTrim(origins, ",")
	}
	if methods != "" {
		config.AllowedMethods = splitAndTrim(methods, ",")
	}
	if headers != "" {
		config.AllowedHeaders = splitAndTrim(headers, ",")
	}

	return CORS(config)
}

// splitAndTrim splits a string by separator and trims whitespace from each part
func splitAndTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
