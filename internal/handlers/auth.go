package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/laith-ambianze/appointment-service/internal/service"
	"github.com/laith-ambianze/appointment-service/pkg/auth"
	"go.uber.org/zap"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	productService *service.ProductService
	jwtManager     *auth.JWTManager
	logger         *zap.Logger
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(productService *service.ProductService, jwtManager *auth.JWTManager, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		productService: productService,
		jwtManager:     jwtManager,
		logger:         logger,
	}
}

// GenerateTokenRequest represents the request body for generating a token
// ExternalUserID is the user identifier from the integrating product's system
type GenerateTokenRequest struct {
	APIKey         string `json:"api_key" binding:"required"`
	APISecret      string `json:"api_secret" binding:"required"`
	ExternalUserID string `json:"external_user_id" binding:"required"`
	Role           string `json:"role" binding:"required,oneof=admin user provider"`
}

// GenerateTokenResponse represents the response with JWT token
type GenerateTokenResponse struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"` // seconds
	TokenType string `json:"token_type"`
}

// GenerateToken handles POST /v1/auth/token
// @Summary Generate JWT token
// @Description Generates a JWT token for API access using API credentials
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body GenerateTokenRequest true "API credentials and user info"
// @Success 200 {object} GenerateTokenResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /v1/auth/token [post]
func (h *AuthHandler) GenerateToken(c *gin.Context) {
	var req GenerateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Bad Request",
			Message: "invalid request body",
			Details: err.Error(),
		})
		return
	}

	// Validate API credentials
	product, err := h.productService.ValidateCredentials(c.Request.Context(), req.APIKey, req.APISecret)
	if err != nil {
		h.logger.Debug("Token generation failed - invalid credentials",
			zap.String("api_key", req.APIKey),
			zap.Error(err),
		)
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Unauthorized",
			Message: "invalid API credentials",
		})
		return
	}

	// Convert role string to auth.Role
	role := auth.Role(req.Role)
	if !role.IsValid() {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Bad Request",
			Message: "invalid role, must be: admin, user, or provider",
		})
		return
	}

	// Generate JWT token
	token, err := h.jwtManager.GenerateToken(product.ID, req.ExternalUserID, role)
	if err != nil {
		h.logger.Error("Failed to generate token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: "failed to generate token",
		})
		return
	}

	h.logger.Info("JWT token generated",
		zap.String("product_id", product.ID.String()),
		zap.String("external_user_id", req.ExternalUserID),
		zap.String("role", req.Role),
	)

	c.JSON(http.StatusOK, GenerateTokenResponse{
		Token:     token,
		ExpiresIn: 86400, // 24 hours in seconds
		TokenType: "Bearer",
	})
}
