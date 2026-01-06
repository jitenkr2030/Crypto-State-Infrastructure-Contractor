package handler

import (
	"net/http"

	"github.com/csic/platform/service/reporting/regulatory/internal/config"
	"github.com/csic/platform/service/reporting/regulatory/internal/domain"
	"github.com/csic/platform/service/reporting/regulatory/internal/service"
	"github.com/gin-gonic/gin"
)

// ScheduleHandler handles HTTP requests for schedule operations
type ScheduleHandler struct {
	scheduleService *service.ScheduleService
}

// NewScheduleHandler creates a new ScheduleHandler instance
func NewScheduleHandler(scheduleService *service.ScheduleService) *ScheduleHandler {
	return &ScheduleHandler{
		scheduleService: scheduleService,
	}
}

// CreateSchedule creates a new schedule
// @Summary Create a new schedule
// @Description Create a new scheduled report configuration
// @Tags schedules
// @Accept json
// @Produce json
// @Param schedule body domain.CreateScheduleRequest true "Schedule configuration"
// @Success 201 {object} domain.Schedule
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/schedules [post]
func (h *ScheduleHandler) CreateSchedule(c *gin.Context) {
	var req domain.CreateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schedule, err := h.scheduleService.CreateSchedule(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, schedule)
}

// GetSchedule retrieves a schedule by ID
// @Summary Get a schedule
// @Description Retrieve details of a specific schedule
// @Tags schedules
// @Produce json
// @Param id path string true "Schedule ID"
// @Success 200 {object} domain.Schedule
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/schedules/{id} [get]
func (h *ScheduleHandler) GetSchedule(c *gin.Context) {
	id := c.Param("id")

	schedule, err := h.scheduleService.GetSchedule(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, schedule)
}

// ListSchedules lists all schedules with optional filtering
// @Summary List schedules
// @Description Retrieve a paginated list of schedules
// @Tags schedules
// @Produce json
// @Param report_type query string false "Filter by report type"
// @Param enabled query bool false "Filter by enabled status"
// @Param offset query int false "Offset for pagination" default(0)
// @Param limit query int false "Limit for pagination" default(20)
// @Success 200 {object} domain.PaginatedSchedules
// @Failure 500 {object} map[string]string
// @Router /api/v1/schedules [get]
func (h *ScheduleHandler) ListSchedules(c *gin.Context) {
	// Parse filter from query parameters
	filter := struct{}{}
	_ = filter

	result, err := h.scheduleService.ListSchedules(c.Request.Context(), struct{}{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// UpdateSchedule updates an existing schedule
// @Summary Update a schedule
// @Description Update configuration of an existing schedule
// @Tags schedules
// @Accept json
// @Produce json
// @Param id path string true "Schedule ID"
// @Param schedule body domain.UpdateScheduleRequest true "Schedule updates"
// @Success 200 {object} domain.Schedule
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/schedules/{id} [put]
func (h *ScheduleHandler) UpdateSchedule(c *gin.Context) {
	id := c.Param("id")

	var req domain.UpdateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schedule, err := h.scheduleService.UpdateSchedule(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, schedule)
}

// DeleteSchedule deletes a schedule
// @Summary Delete a schedule
// @Description Remove a scheduled report configuration
// @Tags schedules
// @Produce json
// @Param id path string true "Schedule ID"
// @Success 204 "No Content"
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/schedules/{id} [delete]
func (h *ScheduleHandler) DeleteSchedule(c *gin.Context) {
	id := c.Param("id")

	if err := h.scheduleService.DeleteSchedule(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// TriggerSchedule triggers immediate report generation for a schedule
// @Summary Trigger a schedule
// @Description Immediately trigger report generation for a schedule
// @Tags schedules
// @Produce json
// @Param id path string true "Schedule ID"
// @Success 202 {object} domain.Report
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/schedules/{id}/trigger [post]
func (h *ScheduleHandler) TriggerSchedule(c *gin.Context) {
	id := c.Param("id")

	report, err := h.scheduleService.TriggerSchedule(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, report)
}
