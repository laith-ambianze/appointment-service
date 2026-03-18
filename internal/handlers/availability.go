package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/laith-ambianze/appointment-service/internal/middleware"
	"github.com/laith-ambianze/appointment-service/internal/service"
	"go.uber.org/zap"
)

// AvailabilityHandler handles availability-related HTTP requests
type AvailabilityHandler struct {
	service *service.AvailabilityService
	logger  *zap.Logger
}

// NewAvailabilityHandler creates a new availability handler
func NewAvailabilityHandler(svc *service.AvailabilityService, logger *zap.Logger) *AvailabilityHandler {
	return &AvailabilityHandler{
		service: svc,
		logger:  logger,
	}
}

// CreateAvailabilityRule handles POST /providers/:provider_id/availability
// @Summary Create availability rule
// @Description Creates a new availability rule for a provider
// @Tags availability
// @Accept json
// @Produce json
// @Param provider_id path string true "Provider ID"
// @Param rule body service.CreateAvailabilityRuleRequest true "Availability rule data"
// @Success 201 {object} models.AvailabilityRule
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Security BearerAuth
// @Router /providers/{provider_id}/availability [post]
func (h *AvailabilityHandler) CreateAvailabilityRule(c *gin.Context) {
	productID := middleware.MustGetProductID(c)
	providerID := c.Param("provider_id")

	if providerID == "" {
		h.respondError(c, http.StatusBadRequest, "provider_id is required", nil)
		return
	}

	var req service.CreateAvailabilityRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// Override provider_id from path
	req.ProviderID = providerID

	rule, err := h.service.CreateAvailabilityRule(c.Request.Context(), productID, req)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, rule)
}

// ListAvailabilityRules handles GET /providers/:provider_id/availability
// @Summary List availability rules
// @Description Lists all availability rules for a provider
// @Tags availability
// @Produce json
// @Param provider_id path string true "Provider ID"
// @Success 200 {array} models.AvailabilityRule
// @Failure 401 {object} ErrorResponse
// @Security BearerAuth
// @Router /providers/{provider_id}/availability [get]
func (h *AvailabilityHandler) ListAvailabilityRules(c *gin.Context) {
	productID := middleware.MustGetProductID(c)
	providerID := c.Param("provider_id")

	if providerID == "" {
		h.respondError(c, http.StatusBadRequest, "provider_id is required", nil)
		return
	}

	rules, err := h.service.ListAvailabilityRules(c.Request.Context(), productID, providerID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"provider_id": providerID,
		"rules":       rules,
		"count":       len(rules),
	})
}

// GetAvailabilityRule handles GET /providers/:provider_id/availability/:rule_id
// @Summary Get availability rule
// @Description Gets a specific availability rule by ID
// @Tags availability
// @Produce json
// @Param provider_id path string true "Provider ID"
// @Param rule_id path string true "Rule ID"
// @Success 200 {object} models.AvailabilityRule
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /providers/{provider_id}/availability/{rule_id} [get]
func (h *AvailabilityHandler) GetAvailabilityRule(c *gin.Context) {
	productID := middleware.MustGetProductID(c)

	ruleID, err := uuid.Parse(c.Param("rule_id"))
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "invalid rule_id", err)
		return
	}

	rule, err := h.service.GetAvailabilityRule(c.Request.Context(), productID, ruleID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, rule)
}

// UpdateAvailabilityRule handles PATCH /providers/:provider_id/availability/:rule_id
// @Summary Update availability rule
// @Description Updates an existing availability rule
// @Tags availability
// @Accept json
// @Produce json
// @Param provider_id path string true "Provider ID"
// @Param rule_id path string true "Rule ID"
// @Param rule body service.UpdateAvailabilityRuleRequest true "Update data"
// @Success 200 {object} models.AvailabilityRule
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /providers/{provider_id}/availability/{rule_id} [patch]
func (h *AvailabilityHandler) UpdateAvailabilityRule(c *gin.Context) {
	productID := middleware.MustGetProductID(c)

	ruleID, err := uuid.Parse(c.Param("rule_id"))
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "invalid rule_id", err)
		return
	}

	var req service.UpdateAvailabilityRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	rule, err := h.service.UpdateAvailabilityRule(c.Request.Context(), productID, ruleID, req)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, rule)
}

// DeleteAvailabilityRule handles DELETE /providers/:provider_id/availability/:rule_id
// @Summary Delete availability rule
// @Description Deletes an availability rule
// @Tags availability
// @Param provider_id path string true "Provider ID"
// @Param rule_id path string true "Rule ID"
// @Success 204
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /providers/{provider_id}/availability/{rule_id} [delete]
func (h *AvailabilityHandler) DeleteAvailabilityRule(c *gin.Context) {
	productID := middleware.MustGetProductID(c)

	ruleID, err := uuid.Parse(c.Param("rule_id"))
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "invalid rule_id", err)
		return
	}

	err = h.service.DeleteAvailabilityRule(c.Request.Context(), productID, ruleID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// GetAvailableSlots handles GET /availability
// @Summary Get available slots
// @Description Gets available time slots for a provider on a specific date
// @Tags availability
// @Produce json
// @Param provider_id query string true "Provider ID"
// @Param date query string true "Date in YYYY-MM-DD format"
// @Param timezone query string false "Client timezone for display"
// @Success 200 {object} service.AvailableSlotsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Security BearerAuth
// @Router /availability [get]
func (h *AvailabilityHandler) GetAvailableSlots(c *gin.Context) {
	productID := middleware.MustGetProductID(c)

	providerID := c.Query("provider_id")
	if providerID == "" {
		h.respondError(c, http.StatusBadRequest, "provider_id is required", nil)
		return
	}

	date := c.Query("date")
	if date == "" {
		h.respondError(c, http.StatusBadRequest, "date is required (format: YYYY-MM-DD)", nil)
		return
	}

	timezone := c.Query("timezone")

	response, err := h.service.GetAvailableSlots(c.Request.Context(), productID, service.GetAvailableSlotsRequest{
		ProviderID: providerID,
		Date:       date,
		Timezone:   timezone,
	})
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// BookAppointment handles POST /appointments/book
// @Summary Book an appointment
// @Description Books an appointment with a provider (concurrency safe)
// @Tags appointments
// @Accept json
// @Produce json
// @Param appointment body service.BookAppointmentRequest true "Booking data"
// @Success 201 {object} models.Appointment
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse "Booking conflict"
// @Security BearerAuth
// @Router /appointments/book [post]
func (h *AvailabilityHandler) BookAppointment(c *gin.Context) {
	productID := middleware.MustGetProductID(c)
	externalUserID := middleware.MustGetExternalUserID(c)

	var req service.BookAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	appointment, err := h.service.BookAppointment(c.Request.Context(), productID, externalUserID, req)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, appointment)
}

// BulkCreateAvailabilityRules handles POST /providers/:provider_id/availability/bulk
// @Summary Bulk create availability rules
// @Description Creates multiple availability rules for a provider
// @Tags availability
// @Accept json
// @Produce json
// @Param provider_id path string true "Provider ID"
// @Param rules body []service.CreateAvailabilityRuleRequest true "Array of availability rules"
// @Success 201 {array} models.AvailabilityRule
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Security BearerAuth
// @Router /providers/{provider_id}/availability/bulk [post]
func (h *AvailabilityHandler) BulkCreateAvailabilityRules(c *gin.Context) {
	productID := middleware.MustGetProductID(c)
	providerID := c.Param("provider_id")

	if providerID == "" {
		h.respondError(c, http.StatusBadRequest, "provider_id is required", nil)
		return
	}

	var rules []service.CreateAvailabilityRuleRequest
	if err := c.ShouldBindJSON(&rules); err != nil {
		h.respondError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// Set provider_id for all rules
	for i := range rules {
		rules[i].ProviderID = providerID
	}

	createdRules, err := h.service.BulkCreateAvailabilityRules(c.Request.Context(), productID, rules)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"provider_id": providerID,
		"rules":       createdRules,
		"count":       len(createdRules),
	})
}

// respondError sends an error response
func (h *AvailabilityHandler) respondError(c *gin.Context, status int, message string, err error) {
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
func (h *AvailabilityHandler) handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrAvailabilityRuleNotFound):
		h.respondError(c, http.StatusNotFound, "availability rule not found", nil)
	case errors.Is(err, service.ErrProviderNotFound):
		h.respondError(c, http.StatusNotFound, "provider not found", nil)
	case errors.Is(err, service.ErrSlotNotAvailable):
		h.respondError(c, http.StatusBadRequest, "requested slot is not available", err)
	case errors.Is(err, service.ErrInvalidSlotTime):
		h.respondError(c, http.StatusBadRequest, "invalid slot time", err)
	case errors.Is(err, service.ErrDateInPast):
		h.respondError(c, http.StatusBadRequest, "cannot book appointments in the past", nil)
	case errors.Is(err, service.ErrBookingConflict):
		h.respondError(c, http.StatusConflict, "booking conflict: time slot is already taken", nil)
	case errors.Is(err, service.ErrInvalidTimezone):
		h.respondError(c, http.StatusBadRequest, "invalid timezone", err)
	case errors.Is(err, service.ErrInvalidInput):
		h.respondError(c, http.StatusBadRequest, "invalid input", err)
	case errors.Is(err, service.ErrAvailabilityWindowTooSmall):
		h.respondError(c, http.StatusBadRequest, "availability window is too small for the requested duration", nil)
	default:
		h.logger.Error("Unexpected service error", zap.Error(err))
		h.respondError(c, http.StatusInternalServerError, "internal server error", nil)
	}
}
