package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/laith-ambianze/appointment-service/pkg/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupTestRouter(jwtManager *auth.JWTManager, skipPaths []string, allowedRoles []auth.Role) *gin.Engine {
	logger, _ := zap.NewDevelopment()
	router := gin.New()

	router.Use(JWTAuth(JWTAuthConfig{
		JWTManager:   jwtManager,
		Logger:       logger,
		SkipPaths:    skipPaths,
		AllowedRoles: allowedRoles,
	}))

	router.GET("/protected", func(c *gin.Context) {
		productID, _ := GetProductIDFromContext(c)
		externalUserID, _ := GetExternalUserIDFromContext(c)
		role, _ := GetRoleFromContext(c)

		c.JSON(http.StatusOK, gin.H{
			"product_id":       productID.String(),
			"external_user_id": externalUserID,
			"role":             role,
		})
	})

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	return router
}

func TestJWTAuth_ValidToken(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret")
	router := setupTestRouter(jwtManager, nil, nil)

	productID := uuid.New()
	externalUserID := "user-123"
	role := auth.RoleUser

	token, err := jwtManager.GenerateToken(productID, externalUserID, role)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, productID.String(), response["product_id"])
	assert.Equal(t, externalUserID, response["external_user_id"])
	assert.Equal(t, string(role), response["role"])
}

func TestJWTAuth_MissingToken(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret")
	router := setupTestRouter(jwtManager, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "unauthorized", response["error"])
	assert.Contains(t, response["message"], "authorization token required")
}

func TestJWTAuth_InvalidToken(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret")
	router := setupTestRouter(jwtManager, nil, nil)

	tests := []struct {
		name   string
		header string
	}{
		{"invalid format", "InvalidToken"},
		{"wrong prefix", "Basic sometoken"},
		{"empty bearer", "Bearer "},
		{"malformed token", "Bearer not.a.valid.jwt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			req.Header.Set("Authorization", tt.header)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}

func TestJWTAuth_WrongSecret(t *testing.T) {
	jwtManager1 := auth.NewJWTManager("secret-1")
	jwtManager2 := auth.NewJWTManager("secret-2")

	router := setupTestRouter(jwtManager1, nil, nil)

	// Generate token with different secret
	token, err := jwtManager2.GenerateToken(uuid.New(), "user", auth.RoleUser)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestJWTAuth_SkipPaths(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret")
	router := setupTestRouter(jwtManager, []string{"/health", "/public/*"}, nil)

	// Health endpoint should be accessible without token
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestJWTAuth_AllowedRoles(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret")

	// Only allow admin and provider
	router := setupTestRouter(jwtManager, nil, []auth.Role{auth.RoleAdmin, auth.RoleProvider})

	tests := []struct {
		name       string
		role       auth.Role
		expectCode int
	}{
		{"admin allowed", auth.RoleAdmin, http.StatusOK},
		{"provider allowed", auth.RoleProvider, http.StatusOK},
		{"user forbidden", auth.RoleUser, http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := jwtManager.GenerateToken(uuid.New(), "user-123", tt.role)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			req.Header.Set("Authorization", "Bearer "+token)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectCode, w.Code)
		})
	}
}

func TestRequireRole(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret")
	logger, _ := zap.NewDevelopment()

	router := gin.New()
	router.Use(JWTAuth(JWTAuthConfig{
		JWTManager: jwtManager,
		Logger:     logger,
	}))

	// Route requiring admin
	router.GET("/admin-only", RequireAdmin(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "admin access granted"})
	})

	// Route requiring admin or provider
	router.GET("/manage", RequireAdminOrProvider(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "management access granted"})
	})

	tests := []struct {
		name       string
		path       string
		role       auth.Role
		expectCode int
	}{
		{"admin can access admin-only", "/admin-only", auth.RoleAdmin, http.StatusOK},
		{"user cannot access admin-only", "/admin-only", auth.RoleUser, http.StatusForbidden},
		{"provider cannot access admin-only", "/admin-only", auth.RoleProvider, http.StatusForbidden},
		{"admin can access manage", "/manage", auth.RoleAdmin, http.StatusOK},
		{"provider can access manage", "/manage", auth.RoleProvider, http.StatusOK},
		{"user cannot access manage", "/manage", auth.RoleUser, http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := jwtManager.GenerateToken(uuid.New(), "user-123", tt.role)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			req.Header.Set("Authorization", "Bearer "+token)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectCode, w.Code)
		})
	}
}

func TestExtractBearerToken(t *testing.T) {
	tests := []struct {
		name        string
		header      string
		expectToken string
		expectError bool
	}{
		{"valid bearer", "Bearer token123", "token123", false},
		{"bearer lowercase", "bearer token123", "token123", false},
		{"bearer with spaces", "Bearer   token123  ", "token123", false},
		{"empty header", "", "", true},
		{"no bearer prefix", "token123", "", true},
		{"wrong prefix", "Basic token123", "", true},
		{"bearer only", "Bearer ", "", true},
		{"bearer no token", "Bearer", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := extractBearerToken(tt.header)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectToken, token)
			}
		})
	}
}

func TestShouldSkipPath(t *testing.T) {
	skipPaths := []string{"/health", "/public/*", "/api/v1/docs"}

	tests := []struct {
		path     string
		expected bool
	}{
		{"/health", true},
		{"/public/file.js", true},
		{"/public/nested/file.js", true},
		{"/api/v1/docs", true},
		{"/api/v1/appointments", false},
		{"/protected", false},
		{"/healthcheck", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := shouldSkipPath(tt.path, skipPaths)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContextHelpers(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret")
	logger, _ := zap.NewDevelopment()

	productID := uuid.New()
	externalUserID := "user-456"
	role := auth.RoleProvider

	router := gin.New()
	router.Use(JWTAuth(JWTAuthConfig{
		JWTManager: jwtManager,
		Logger:     logger,
	}))

	router.GET("/test", func(c *gin.Context) {
		// Test helper functions
		gotProductID, ok := GetProductIDFromContext(c)
		assert.True(t, ok)
		assert.Equal(t, productID, gotProductID)

		gotExternalUserID, ok := GetExternalUserIDFromContext(c)
		assert.True(t, ok)
		assert.Equal(t, externalUserID, gotExternalUserID)

		gotRole, ok := GetRoleFromContext(c)
		assert.True(t, ok)
		assert.Equal(t, role, gotRole)

		claims, ok := GetClaimsFromContext(c)
		assert.True(t, ok)
		assert.Equal(t, productID, claims.ProductID)

		// Test Must* functions (should not panic)
		assert.Equal(t, productID, MustGetProductID(c))
		assert.Equal(t, externalUserID, MustGetExternalUserID(c))
		assert.Equal(t, role, MustGetRole(c))

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	token, err := jwtManager.GenerateToken(productID, externalUserID, role)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
