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

// AppointmentHandler handles appointment-related HTTP requests
type AppointmentHandler struct {
	service *service.AppointmentService
	logger  *zap.Logger
}

// NewAppointmentHandler creates a new appointment handler
func NewAppointmentHandler(svc *service.AppointmentService, logger *zap.Logger) *AppointmentHandler {
	return &AppointmentHandler{
		service: svc,
		logger:  logger,
	}
}

// Create handles POST /appointments
// @Summary Create a new appointment
// @Description Creates a new appointment with participants
// @Tags appointments
// @Accept json
// @Produce json
// @Param appointment body service.CreateAppointmentRequest true "Appointment data"
// @Success 201 {object} models.Appointment
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /appointments [post]
func (h *AppointmentHandler) Create(c *gin.Context) {
	productID := middleware.MustGetProductID(c)
	externalUserID := middleware.MustGetExternalUserID(c)

	var req service.CreateAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	appointment, err := h.service.Create(c.Request.Context(), productID, externalUserID, req)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, appointment)
}

// GetByID handles GET /appointments/:id
// @Summary Get appointment by ID
// @Description Retrieves an appointment by its ID
// @Tags appointments
// @Produce json
// @Param id path string true "Appointment ID"
// @Success 200 {object} models.Appointment
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /appointments/{id} [get]
func (h *AppointmentHandler) GetByID(c *gin.Context) {
	productID := middleware.MustGetProductID(c)

	appointmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "invalid appointment ID", err)
		return
	}

	appointment, err := h.service.GetByID(c.Request.Context(), productID, appointmentID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, appointment)
}

// List handles GET /appointments
// @Summary List appointments
// @Description Lists appointments for the authenticated user
// @Tags appointments
// @Produce json
// @Param status query string false "Filter by status"
// @Param limit query int false "Limit results (default 20, max 100)"
// @Param offset query int false "Offset for pagination"
// @Success 200 {array} models.Appointment
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /appointments [get]
func (h *AppointmentHandler) List(c *gin.Context) {
	productID := middleware.MustGetProductID(c)
	externalUserID := middleware.MustGetExternalUserID(c)

	var req service.ListAppointmentsRequest

	// Parse query parameters
	if status := c.Query("status"); status != "" {
		s := models.AppointmentStatus(status)
		req.Status = &s
	}

	// Limit and offset with defaults
	req.Limit = 20
	if limit := c.GetInt("limit"); limit > 0 {
		req.Limit = limit
	}
	req.Offset = c.GetInt("offset")

	appointments, err := h.service.ListByUser(c.Request.Context(), productID, externalUserID, req)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   appointments,
		"count":  len(appointments),
		"limit":  req.Limit,
		"offset": req.Offset,
	})
}

// Update handles PATCH /appointments/:id
// @Summary Update an appointment
// @Description Updates an existing appointment
// @Tags appointments
// @Accept json
// @Produce json
// @Param id path string true "Appointment ID"
// @Param appointment body service.UpdateAppointmentRequest true "Update data"
// @Success 200 {object} models.Appointment
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /appointments/{id} [patch]
func (h *AppointmentHandler) Update(c *gin.Context) {
	productID := middleware.MustGetProductID(c)
	externalUserID := middleware.MustGetExternalUserID(c)
	role := middleware.MustGetRole(c)

	appointmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "invalid appointment ID", err)
		return
	}

	var req service.UpdateAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	appointment, err := h.service.Update(c.Request.Context(), productID, externalUserID, role, appointmentID, req)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, appointment)
}

// Respond handles PATCH /appointments/:id/response
// @Summary Respond to an appointment (Admin/Provider only)
// @Description Update appointment status (confirm, complete, etc.)
// @Tags appointments
// @Accept json
// @Produce json
// @Param id path string true "Appointment ID"
// @Param response body service.RespondToAppointmentRequest true "Response data"
// @Success 200 {object} models.Appointment
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /appointments/{id}/response [patch]
func (h *AppointmentHandler) Respond(c *gin.Context) {
	productID := middleware.MustGetProductID(c)
	externalUserID := middleware.MustGetExternalUserID(c)
	role := middleware.MustGetRole(c)

	appointmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "invalid appointment ID", err)
		return
	}

	var req service.RespondToAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	appointment, err := h.service.Respond(c.Request.Context(), productID, externalUserID, role, appointmentID, req)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, appointment)
}

// Cancel handles POST /appointments/:id/cancel
// @Summary Cancel an appointment
// @Description Cancels an appointment (creator, admin, or provider)
// @Tags appointments
// @Produce json
// @Param id path string true "Appointment ID"
// @Success 200 {object} models.Appointment
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /appointments/{id}/cancel [post]
func (h *AppointmentHandler) Cancel(c *gin.Context) {
	productID := middleware.MustGetProductID(c)
	externalUserID := middleware.MustGetExternalUserID(c)
	role := middleware.MustGetRole(c)

	appointmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "invalid appointment ID", err)
		return
	}

	appointment, err := h.service.Cancel(c.Request.Context(), productID, externalUserID, role, appointmentID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, appointment)
}

// Delete handles DELETE /appointments/:id
// @Summary Delete an appointment (Admin only)
// @Description Soft-deletes an appointment
// @Tags appointments
// @Param id path string true "Appointment ID"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /appointments/{id} [delete]
func (h *AppointmentHandler) Delete(c *gin.Context) {
	productID := middleware.MustGetProductID(c)
	externalUserID := middleware.MustGetExternalUserID(c)
	role := middleware.MustGetRole(c)

	appointmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "invalid appointment ID", err)
		return
	}

	if err := h.service.Delete(c.Request.Context(), productID, externalUserID, role, appointmentID); err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// AddParticipant handles POST /appointments/:id/participants
// @Summary Add a participant to an appointment
// @Description Adds a new participant to an existing appointment
// @Tags appointments
// @Accept json
// @Produce json
// @Param id path string true "Appointment ID"
// @Param participant body service.ParticipantRequest true "Participant data"
// @Success 201 {object} models.AppointmentParticipant
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /appointments/{id}/participants [post]
func (h *AppointmentHandler) AddParticipant(c *gin.Context) {
	productID := middleware.MustGetProductID(c)
	externalUserID := middleware.MustGetExternalUserID(c)
	role := middleware.MustGetRole(c)

	appointmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "invalid appointment ID", err)
		return
	}

	var req service.ParticipantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	participant, err := h.service.AddParticipant(c.Request.Context(), productID, externalUserID, role, appointmentID, req)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, participant)
}

// RemoveParticipant handles DELETE /appointments/:id/participants/:user_id
// @Summary Remove a participant from an appointment
// @Description Removes a participant from an existing appointment
// @Tags appointments
// @Param id path string true "Appointment ID"
// @Param user_id path string true "Participant's external user ID"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /appointments/{id}/participants/{user_id} [delete]
func (h *AppointmentHandler) RemoveParticipant(c *gin.Context) {
	productID := middleware.MustGetProductID(c)
	externalUserID := middleware.MustGetExternalUserID(c)
	role := middleware.MustGetRole(c)

	appointmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "invalid appointment ID", err)
		return
	}

	participantUserID := c.Param("user_id")
	if participantUserID == "" {
		h.respondError(c, http.StatusBadRequest, "participant user ID is required", nil)
		return
	}

	if err := h.service.RemoveParticipant(c.Request.Context(), productID, externalUserID, role, appointmentID, participantUserID); err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// UpdateParticipantStatus handles PATCH /appointments/:id/participants/:user_id/status
// @Summary Update participant status
// @Description Updates a participant's response status (accept, decline, etc.)
// @Tags appointments
// @Accept json
// @Produce json
// @Param id path string true "Appointment ID"
// @Param user_id path string true "Participant's external user ID"
// @Param status body UpdateParticipantStatusRequest true "Status data"
// @Success 200 {object} gin.H
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /appointments/{id}/participants/{user_id}/status [patch]
func (h *AppointmentHandler) UpdateParticipantStatus(c *gin.Context) {
	productID := middleware.MustGetProductID(c)
	externalUserID := middleware.MustGetExternalUserID(c)
	role := middleware.MustGetRole(c)

	appointmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "invalid appointment ID", err)
		return
	}

	participantUserID := c.Param("user_id")
	if participantUserID == "" {
		h.respondError(c, http.StatusBadRequest, "participant user ID is required", nil)
		return
	}

	var req UpdateParticipantStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	if err := h.service.UpdateParticipantStatus(c.Request.Context(), productID, externalUserID, role, appointmentID, participantUserID, req.Status); err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "participant status updated",
		"status":  req.Status,
	})
}

// UpdateParticipantStatusRequest represents the request body for updating participant status
type UpdateParticipantStatusRequest struct {
	Status models.ParticipantStatus `json:"status" binding:"required"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Details string `json:"details,omitempty"`
}

// respondError sends an error response
func (h *AppointmentHandler) respondError(c *gin.Context, status int, message string, err error) {
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
func (h *AppointmentHandler) handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrAppointmentNotFound):
		h.respondError(c, http.StatusNotFound, "appointment not found", nil)
	case errors.Is(err, service.ErrParticipantNotFound):
		h.respondError(c, http.StatusNotFound, "participant not found", nil)
	case errors.Is(err, service.ErrUnauthorized):
		h.respondError(c, http.StatusUnauthorized, "unauthorized", nil)
	case errors.Is(err, service.ErrForbidden):
		h.respondError(c, http.StatusForbidden, "insufficient permissions", nil)
	case errors.Is(err, service.ErrInvalidInput):
		h.respondError(c, http.StatusBadRequest, "invalid input", err)
	case errors.Is(err, service.ErrInvalidStatusTransition):
		h.respondError(c, http.StatusBadRequest, "invalid status transition", err)
	case errors.Is(err, service.ErrAppointmentInPast):
		h.respondError(c, http.StatusBadRequest, "cannot modify past appointment", nil)
	case errors.Is(err, service.ErrDuplicateParticipant):
		h.respondError(c, http.StatusConflict, "participant already exists", nil)
	default:
		h.logger.Error("Unexpected service error", zap.Error(err))
		h.respondError(c, http.StatusInternalServerError, "internal server error", nil)
	}
}
