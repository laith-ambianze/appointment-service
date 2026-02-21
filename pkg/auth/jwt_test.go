package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRole_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		role     Role
		expected bool
	}{
		{"admin is valid", RoleAdmin, true},
		{"user is valid", RoleUser, true},
		{"provider is valid", RoleProvider, true},
		{"empty is invalid", Role(""), false},
		{"unknown is invalid", Role("unknown"), false},
		{"guest is invalid", Role("guest"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.role.IsValid())
		})
	}
}

func TestRole_CanManageAppointments(t *testing.T) {
	tests := []struct {
		name     string
		role     Role
		expected bool
	}{
		{"admin can manage", RoleAdmin, true},
		{"provider can manage", RoleProvider, true},
		{"user cannot manage", RoleUser, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.role.CanManageAppointments())
		})
	}
}

func TestClaims_Validate(t *testing.T) {
	validProductID := uuid.New()

	tests := []struct {
		name        string
		claims      Claims
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid claims",
			claims: Claims{
				ProductID: validProductID,
				UserID:    "user-123",
				Role:      RoleUser,
			},
			expectError: false,
		},
		{
			name: "missing product_id",
			claims: Claims{
				ProductID: uuid.Nil,
				UserID:    "user-123",
				Role:      RoleUser,
			},
			expectError: true,
			errorMsg:    "product_id is required",
		},
		{
			name: "missing user_id",
			claims: Claims{
				ProductID: validProductID,
				UserID:    "",
				Role:      RoleUser,
			},
			expectError: true,
			errorMsg:    "user_id is required",
		},
		{
			name: "invalid role",
			claims: Claims{
				ProductID: validProductID,
				UserID:    "user-123",
				Role:      Role("invalid"),
			},
			expectError: true,
			errorMsg:    "invalid role",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.claims.Validate()
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestJWTManager_GenerateAndValidateToken(t *testing.T) {
	manager := NewJWTManager("test-secret-key-12345")
	productID := uuid.New()
	userID := "user-123"
	role := RoleAdmin

	// Generate token
	token, err := manager.GenerateToken(productID, userID, role)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validate token
	claims, err := manager.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, productID, claims.ProductID)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, role, claims.Role)
	assert.Equal(t, userID, claims.Subject)
	assert.Equal(t, "appointment-service", claims.Issuer)
}

func TestJWTManager_ValidateToken_Expired(t *testing.T) {
	manager := NewJWTManager("test-secret-key-12345")
	productID := uuid.New()
	userID := "user-123"
	role := RoleUser

	// Generate token with past expiration
	expiry := time.Now().Add(-1 * time.Hour)
	token, err := manager.GenerateTokenWithExpiry(productID, userID, role, expiry)
	require.NoError(t, err)

	// Validate should fail
	_, err = manager.ValidateToken(token)
	assert.Error(t, err)
	assert.True(t, IsExpiredError(err))
}

func TestJWTManager_ValidateToken_InvalidSignature(t *testing.T) {
	manager1 := NewJWTManager("secret-1")
	manager2 := NewJWTManager("secret-2")

	productID := uuid.New()
	userID := "user-123"
	role := RoleUser

	// Generate token with first manager
	token, err := manager1.GenerateToken(productID, userID, role)
	require.NoError(t, err)

	// Validate with different secret should fail
	_, err = manager2.ValidateToken(token)
	assert.Error(t, err)
	assert.True(t, IsInvalidError(err))
}

func TestJWTManager_ValidateToken_MalformedToken(t *testing.T) {
	manager := NewJWTManager("test-secret")

	tests := []struct {
		name  string
		token string
	}{
		{"empty token", ""},
		{"random string", "not-a-jwt-token"},
		{"partial token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"},
		{"malformed json", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.notjson.signature"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := manager.ValidateToken(tt.token)
			assert.Error(t, err)
		})
	}
}

func TestJWTManager_GenerateTokenWithExpiry(t *testing.T) {
	manager := NewJWTManager("test-secret")
	productID := uuid.New()
	userID := "user-123"
	role := RoleProvider

	// Generate token with custom expiry
	customExpiry := time.Now().Add(48 * time.Hour)
	token, err := manager.GenerateTokenWithExpiry(productID, userID, role, customExpiry)
	require.NoError(t, err)

	// Validate and check expiry
	claims, err := manager.ValidateToken(token)
	require.NoError(t, err)
	
	// Allow 1 second tolerance for time comparison
	assert.WithinDuration(t, customExpiry, claims.ExpiresAt.Time, time.Second)
}

func TestNewJWTManagerWithConfig(t *testing.T) {
	config := JWTConfig{
		Secret:          "custom-secret",
		Issuer:          "custom-issuer",
		ExpirationHours: 48,
		SigningMethod:   DefaultJWTConfig().SigningMethod,
	}

	manager := NewJWTManagerWithConfig(config)
	productID := uuid.New()
	userID := "user-123"
	role := RoleAdmin

	token, err := manager.GenerateToken(productID, userID, role)
	require.NoError(t, err)

	claims, err := manager.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, "custom-issuer", claims.Issuer)
}

func TestErrorHelpers(t *testing.T) {
	t.Run("IsExpiredError", func(t *testing.T) {
		assert.True(t, IsExpiredError(ErrExpiredToken))
		assert.False(t, IsExpiredError(ErrInvalidToken))
		assert.False(t, IsExpiredError(nil))
	})

	t.Run("IsInvalidError", func(t *testing.T) {
		assert.True(t, IsInvalidError(ErrInvalidToken))
		assert.True(t, IsInvalidError(ErrMalformedToken))
		assert.True(t, IsInvalidError(ErrInvalidSignature))
		assert.True(t, IsInvalidError(ErrInvalidClaims))
		assert.False(t, IsInvalidError(ErrExpiredToken))
		assert.False(t, IsInvalidError(nil))
	})
}
