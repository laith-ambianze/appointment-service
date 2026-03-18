package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/laith-ambianze/appointment-service/internal/handlers"
	"github.com/laith-ambianze/appointment-service/internal/middleware"
	"github.com/laith-ambianze/appointment-service/pkg/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// TestExternalUserJWTFlow tests the complete JWT flow for external users
// This verifies that:
// 1. JWT tokens are generated correctly for external users
// 2. External user ID is properly extracted from JWT
// 3. Product ID is properly extracted from JWT
// 4. RBAC works correctly based on roles
func TestExternalUserJWTFlow(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret-key-for-integration")
	logger, _ := zap.NewDevelopment()

	// Setup router with JWT middleware
	router := gin.New()
	router.Use(middleware.JWTAuth(middleware.JWTAuthConfig{
		JWTManager: jwtManager,
		Logger:     logger,
		SkipPaths:  []string{"/health"},
	}))

	// Test endpoint that returns the extracted claims
	router.GET("/api/v1/me", func(c *gin.Context) {
		productID := middleware.MustGetProductID(c)
		externalUserID := middleware.MustGetExternalUserID(c)
		role := middleware.MustGetRole(c)

		c.JSON(http.StatusOK, gin.H{
			"product_id":       productID.String(),
			"external_user_id": externalUserID,
			"role":             string(role),
		})
	})

	t.Run("ExternalUserCanAccessWithValidJWT", func(t *testing.T) {
		productID := uuid.New()
		externalUserID := "ext-user-12345"
		role := auth.RoleUser

		// Generate JWT for external user
		token, err := jwtManager.GenerateToken(productID, externalUserID, role)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
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
	})

	t.Run("DifferentProductsHaveDifferentContexts", func(t *testing.T) {
		product1ID := uuid.New()
		product2ID := uuid.New()
		externalUserID := "shared-user-id" // Same user ID can exist in different products

		token1, _ := jwtManager.GenerateToken(product1ID, externalUserID, auth.RoleUser)
		token2, _ := jwtManager.GenerateToken(product2ID, externalUserID, auth.RoleUser)

		// Request with product 1 token
		req1 := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
		req1.Header.Set("Authorization", "Bearer "+token1)
		w1 := httptest.NewRecorder()
		router.ServeHTTP(w1, req1)

		var resp1 map[string]string
		json.Unmarshal(w1.Body.Bytes(), &resp1)

		// Request with product 2 token
		req2 := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
		req2.Header.Set("Authorization", "Bearer "+token2)
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)

		var resp2 map[string]string
		json.Unmarshal(w2.Body.Bytes(), &resp2)

		// Same external user ID but different product contexts
		assert.Equal(t, externalUserID, resp1["external_user_id"])
		assert.Equal(t, externalUserID, resp2["external_user_id"])
		assert.NotEqual(t, resp1["product_id"], resp2["product_id"])
		assert.Equal(t, product1ID.String(), resp1["product_id"])
		assert.Equal(t, product2ID.String(), resp2["product_id"])
	})
}

// TestRBACEnforcement tests role-based access control
func TestRBACEnforcement(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret")
	logger, _ := zap.NewDevelopment()

	router := gin.New()
	router.Use(middleware.JWTAuth(middleware.JWTAuthConfig{
		JWTManager: jwtManager,
		Logger:     logger,
	}))

	// Public endpoint (after auth)
	router.GET("/api/v1/public", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"access": "public"})
	})

	// Admin-only endpoint
	router.GET("/api/v1/admin", middleware.RequireAdmin(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"access": "admin"})
	})

	// Provider endpoint
	router.GET("/api/v1/provider", middleware.RequireProvider(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"access": "provider"})
	})

	// Admin or Provider endpoint
	router.GET("/api/v1/management", middleware.RequireAdminOrProvider(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"access": "management"})
	})

	productID := uuid.New()

	tests := []struct {
		name           string
		role           auth.Role
		endpoint       string
		expectedStatus int
	}{
		// Public endpoint - all roles allowed
		{"user can access public", auth.RoleUser, "/api/v1/public", http.StatusOK},
		{"provider can access public", auth.RoleProvider, "/api/v1/public", http.StatusOK},
		{"admin can access public", auth.RoleAdmin, "/api/v1/public", http.StatusOK},

		// Admin-only endpoint
		{"user cannot access admin", auth.RoleUser, "/api/v1/admin", http.StatusForbidden},
		{"provider cannot access admin", auth.RoleProvider, "/api/v1/admin", http.StatusForbidden},
		{"admin can access admin", auth.RoleAdmin, "/api/v1/admin", http.StatusOK},

		// Provider-only endpoint
		{"user cannot access provider", auth.RoleUser, "/api/v1/provider", http.StatusForbidden},
		{"provider can access provider", auth.RoleProvider, "/api/v1/provider", http.StatusOK},
		{"admin cannot access provider-only", auth.RoleAdmin, "/api/v1/provider", http.StatusForbidden},

		// Management endpoint (admin or provider)
		{"user cannot access management", auth.RoleUser, "/api/v1/management", http.StatusForbidden},
		{"provider can access management", auth.RoleProvider, "/api/v1/management", http.StatusOK},
		{"admin can access management", auth.RoleAdmin, "/api/v1/management", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, _ := jwtManager.GenerateToken(productID, "test-user", tt.role)

			req := httptest.NewRequest(http.MethodGet, tt.endpoint, nil)
			req.Header.Set("Authorization", "Bearer "+token)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestMultiTenantDataIsolation tests that data is properly isolated between products
func TestMultiTenantDataIsolation(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret")
	logger, _ := zap.NewDevelopment()

	// Simulated data store for testing
	appointmentsByProduct := make(map[uuid.UUID][]map[string]interface{})

	router := gin.New()
	router.Use(middleware.JWTAuth(middleware.JWTAuthConfig{
		JWTManager: jwtManager,
		Logger:     logger,
	}))

	// Create appointment (stores in product's namespace)
	router.POST("/api/v1/appointments", func(c *gin.Context) {
		productID := middleware.MustGetProductID(c)
		externalUserID := middleware.MustGetExternalUserID(c)

		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		appointment := map[string]interface{}{
			"id":               uuid.New().String(),
			"product_id":       productID.String(),
			"created_by":       externalUserID,
			"title":            req["title"],
			"external_user_id": externalUserID,
		}

		appointmentsByProduct[productID] = append(appointmentsByProduct[productID], appointment)
		c.JSON(http.StatusCreated, appointment)
	})

	// List appointments (returns only product's appointments)
	router.GET("/api/v1/appointments", func(c *gin.Context) {
		productID := middleware.MustGetProductID(c)
		appointments := appointmentsByProduct[productID]
		if appointments == nil {
			appointments = []map[string]interface{}{}
		}
		c.JSON(http.StatusOK, gin.H{"data": appointments, "count": len(appointments)})
	})

	product1ID := uuid.New()
	product2ID := uuid.New()

	token1, _ := jwtManager.GenerateToken(product1ID, "user-from-product-1", auth.RoleUser)
	token2, _ := jwtManager.GenerateToken(product2ID, "user-from-product-2", auth.RoleUser)

	// Create appointment for product 1
	createReq := map[string]interface{}{"title": "Product 1 Appointment"}
	body, _ := json.Marshal(createReq)

	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/appointments", bytes.NewBuffer(body))
	req1.Header.Set("Authorization", "Bearer "+token1)
	req1.Header.Set("Content-Type", "application/json")

	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusCreated, w1.Code)

	// Create appointment for product 2
	createReq2 := map[string]interface{}{"title": "Product 2 Appointment"}
	body2, _ := json.Marshal(createReq2)

	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/appointments", bytes.NewBuffer(body2))
	req2.Header.Set("Authorization", "Bearer "+token2)
	req2.Header.Set("Content-Type", "application/json")

	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusCreated, w2.Code)

	// List appointments for product 1 - should only see product 1's appointments
	listReq1 := httptest.NewRequest(http.MethodGet, "/api/v1/appointments", nil)
	listReq1.Header.Set("Authorization", "Bearer "+token1)

	listW1 := httptest.NewRecorder()
	router.ServeHTTP(listW1, listReq1)

	var listResp1 map[string]interface{}
	json.Unmarshal(listW1.Body.Bytes(), &listResp1)

	assert.Equal(t, float64(1), listResp1["count"])
	appointments1 := listResp1["data"].([]interface{})
	assert.Len(t, appointments1, 1)
	appt1 := appointments1[0].(map[string]interface{})
	assert.Equal(t, product1ID.String(), appt1["product_id"])
	assert.Equal(t, "Product 1 Appointment", appt1["title"])

	// List appointments for product 2 - should only see product 2's appointments
	listReq2 := httptest.NewRequest(http.MethodGet, "/api/v1/appointments", nil)
	listReq2.Header.Set("Authorization", "Bearer "+token2)

	listW2 := httptest.NewRecorder()
	router.ServeHTTP(listW2, listReq2)

	var listResp2 map[string]interface{}
	json.Unmarshal(listW2.Body.Bytes(), &listResp2)

	assert.Equal(t, float64(1), listResp2["count"])
	appointments2 := listResp2["data"].([]interface{})
	assert.Len(t, appointments2, 1)
	appt2 := appointments2[0].(map[string]interface{})
	assert.Equal(t, product2ID.String(), appt2["product_id"])
	assert.Equal(t, "Product 2 Appointment", appt2["title"])
}

// TestJWTTokenStructure verifies the JWT token contains all required fields
func TestJWTTokenStructure(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret")

	productID := uuid.New()
	externalUserID := "ext-user-abc123"
	role := auth.RoleProvider

	token, err := jwtManager.GenerateToken(productID, externalUserID, role)
	require.NoError(t, err)

	// Validate and extract claims
	claims, err := jwtManager.ValidateToken(token)
	require.NoError(t, err)

	// Verify all required fields
	t.Run("has product_id", func(t *testing.T) {
		assert.Equal(t, productID, claims.ProductID)
		assert.NotEqual(t, uuid.Nil, claims.ProductID)
	})

	t.Run("has external_user_id", func(t *testing.T) {
		assert.Equal(t, externalUserID, claims.ExternalUserID)
		assert.NotEmpty(t, claims.ExternalUserID)
	})

	t.Run("has role", func(t *testing.T) {
		assert.Equal(t, role, claims.Role)
		assert.True(t, claims.Role.IsValid())
	})

	t.Run("has expiration", func(t *testing.T) {
		assert.NotNil(t, claims.ExpiresAt)
		assert.True(t, claims.ExpiresAt.After(claims.IssuedAt.Time))
	})

	t.Run("has issuer", func(t *testing.T) {
		assert.Equal(t, "appointment-service", claims.Issuer)
	})

	t.Run("subject matches external_user_id", func(t *testing.T) {
		assert.Equal(t, externalUserID, claims.Subject)
	})
}

// TestNoInternalUserStorage verifies the system works without internal user storage
func TestNoInternalUserStorage(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret")
	logger, _ := zap.NewDevelopment()

	router := gin.New()
	router.Use(middleware.JWTAuth(middleware.JWTAuthConfig{
		JWTManager: jwtManager,
		Logger:     logger,
	}))

	// Endpoint that would typically require user lookup - but doesn't
	router.GET("/api/v1/user/profile", func(c *gin.Context) {
		// All user info comes from JWT - no database lookup needed
		productID := middleware.MustGetProductID(c)
		externalUserID := middleware.MustGetExternalUserID(c)
		role := middleware.MustGetRole(c)

		// User profile is constructed from JWT claims only
		c.JSON(http.StatusOK, gin.H{
			"external_user_id": externalUserID,
			"product_id":       productID.String(),
			"role":             string(role),
			"source":           "jwt_claims", // Indicates no DB lookup
		})
	})

	productID := uuid.New()
	externalUserID := "completely-external-user-999"

	token, _ := jwtManager.GenerateToken(productID, externalUserID, auth.RoleUser)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user/profile", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, externalUserID, response["external_user_id"])
	assert.Equal(t, productID.String(), response["product_id"])
	assert.Equal(t, "jwt_claims", response["source"])
}

// ErrorResponse for testing
type ErrorResponse = handlers.ErrorResponse
