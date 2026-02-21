package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/laith-ambianze/appointment-service/internal/middleware"
	"github.com/laith-ambianze/appointment-service/internal/models"
	"github.com/laith-ambianze/appointment-service/internal/service"
	"go.uber.org/zap"
)

// ProductHandler handles product-related HTTP requests
type ProductHandler struct {
	service *service.ProductService
	logger  *zap.Logger
}

// NewProductHandler creates a new product handler
func NewProductHandler(svc *service.ProductService, logger *zap.Logger) *ProductHandler {
	return &ProductHandler{
		service: svc,
		logger:  logger,
	}
}

// Register handles POST /products/register
// @Summary Register a new product (company)
// @Description Creates a new product with generated API credentials
// @Tags products
// @Accept json
// @Produce json
// @Param product body service.RegisterProductRequest true "Product data"
// @Success 201 {object} service.RegisterProductResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /products/register [post]
func (h *ProductHandler) Register(c *gin.Context) {
	var req service.RegisterProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	response, err := h.service.Register(c.Request.Context(), req)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetCurrent handles GET /products/me
// @Summary Get current product info
// @Description Retrieves the product associated with the current JWT token
// @Tags products
// @Produce json
// @Success 200 {object} models.Product
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /products/me [get]
func (h *ProductHandler) GetCurrent(c *gin.Context) {
	productID := middleware.MustGetProductID(c)

	product, err := h.service.GetByID(c.Request.Context(), productID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, product)
}

// GetByID handles GET /products/:id
// @Summary Get product by ID (Admin only)
// @Description Retrieves a product by its ID
// @Tags products
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} models.Product
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /products/{id} [get]
func (h *ProductHandler) GetByID(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "invalid product ID", err)
		return
	}

	product, err := h.service.GetByID(c.Request.Context(), productID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, product)
}

// UpdateCurrent handles PATCH /products/me
// @Summary Update current product
// @Description Updates the product associated with the current JWT token
// @Tags products
// @Accept json
// @Produce json
// @Param product body service.UpdateProductRequest true "Update data"
// @Success 200 {object} models.Product
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /products/me [patch]
func (h *ProductHandler) UpdateCurrent(c *gin.Context) {
	productID := middleware.MustGetProductID(c)

	var req service.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	product, err := h.service.Update(c.Request.Context(), productID, req)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, product)
}

// Update handles PATCH /products/:id
// @Summary Update product by ID (Admin only)
// @Description Updates a product by its ID
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param product body service.UpdateProductRequest true "Update data"
// @Success 200 {object} models.Product
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /products/{id} [patch]
func (h *ProductHandler) Update(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "invalid product ID", err)
		return
	}

	var req service.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	product, err := h.service.Update(c.Request.Context(), productID, req)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, product)
}

// List handles GET /products
// @Summary List all products (Admin only)
// @Description Retrieves all products with optional filters
// @Tags products
// @Produce json
// @Param status query string false "Filter by status (active, inactive, suspended)"
// @Param limit query int false "Limit results (default 20, max 100)"
// @Param offset query int false "Offset for pagination"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Security BearerAuth
// @Router /products [get]
func (h *ProductHandler) List(c *gin.Context) {
	var req service.ListProductsRequest

	// Parse query parameters
	if status := c.Query("status"); status != "" {
		s := models.ProductStatus(status)
		req.Status = &s
	}

	// Limit and offset with defaults
	req.Limit = 20
	if limit := c.GetInt("limit"); limit > 0 {
		req.Limit = limit
	}
	req.Offset = c.GetInt("offset")

	products, err := h.service.List(c.Request.Context(), req)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   products,
		"count":  len(products),
		"limit":  req.Limit,
		"offset": req.Offset,
	})
}

// Delete handles DELETE /products/:id
// @Summary Delete product (Admin only)
// @Description Soft-deletes a product
// @Tags products
// @Param id path string true "Product ID"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /products/{id} [delete]
func (h *ProductHandler) Delete(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "invalid product ID", err)
		return
	}

	if err := h.service.Delete(c.Request.Context(), productID); err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// RegenerateCredentials handles POST /products/me/regenerate-credentials
// @Summary Regenerate API credentials
// @Description Generates new API key and secret for the current product
// @Tags products
// @Produce json
// @Success 200 {object} service.RegisterProductResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /products/me/regenerate-credentials [post]
func (h *ProductHandler) RegenerateCredentials(c *gin.Context) {
	productID := middleware.MustGetProductID(c)

	response, err := h.service.RegenerateCredentials(c.Request.Context(), productID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// ValidateCredentials handles POST /products/validate
// @Summary Validate API credentials
// @Description Validates API key and secret, returns product info if valid
// @Tags products
// @Accept json
// @Produce json
// @Param credentials body ValidateCredentialsRequest true "API credentials"
// @Success 200 {object} models.Product
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /products/validate [post]
func (h *ProductHandler) ValidateCredentials(c *gin.Context) {
	var req ValidateCredentialsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	product, err := h.service.ValidateCredentials(c.Request.Context(), req.APIKey, req.APISecret)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, product)
}

// ValidateCredentialsRequest represents the request body for validating credentials
type ValidateCredentialsRequest struct {
	APIKey    string `json:"api_key" binding:"required"`
	APISecret string `json:"api_secret" binding:"required"`
}

// respondError sends an error response
func (h *ProductHandler) respondError(c *gin.Context, status int, message string, err error) {
	response := ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
	}

	if err != nil && h.logger != nil {
		h.logger.Debug("Request error",
			zap.String("message", message),
			zap.Error(err),
		)
		response.Details = err.Error()
	}

	c.JSON(status, response)
}

// handleServiceError maps service errors to HTTP responses
func (h *ProductHandler) handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrProductNotFound):
		h.respondError(c, http.StatusNotFound, "product not found", nil)
	case errors.Is(err, service.ErrProductInactive):
		h.respondError(c, http.StatusForbidden, "product is inactive or suspended", nil)
	case errors.Is(err, service.ErrDuplicateAPIKey):
		h.respondError(c, http.StatusConflict, "API key already exists", nil)
	case errors.Is(err, service.ErrInvalidCredentials):
		h.respondError(c, http.StatusUnauthorized, "invalid API credentials", nil)
	case errors.Is(err, service.ErrInvalidInput):
		h.respondError(c, http.StatusBadRequest, "invalid input", err)
	default:
		h.logger.Error("Unexpected service error", zap.Error(err))
		h.respondError(c, http.StatusInternalServerError, "internal server error", nil)
	}
}
