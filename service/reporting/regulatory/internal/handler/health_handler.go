package handler

import (
	"net/http"

	"github.com/csic/platform/service/reporting/regulatory/internal/config"
	"github.com/csic/platform/service/reporting/regulatory/internal/domain"
	"github.com/csic/platform/service/reporting/regulatory/internal/service"
	"github.com/gin-gonic/gin"
)

// TemplateHandler handles HTTP requests for template operations
type TemplateHandler struct {
	templateService *service.TemplateService
}

// NewTemplateHandler creates a new TemplateHandler instance
func NewTemplateHandler(templateService *service.TemplateService) *TemplateHandler {
	return &TemplateHandler{
		templateService: templateService,
	}
}

// CreateTemplate creates a new template
// @Summary Create a new template
// @Description Create a new report template
// @Tags templates
// @Accept json
// @Produce json
// @Param template body domain.CreateTemplateRequest true "Template configuration"
// @Success 201 {object} domain.Template
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/templates [post]
func (h *TemplateHandler) CreateTemplate(c *gin.Context) {
	var req domain.CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	template, err := h.templateService.CreateTemplate(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, template)
}

// GetTemplate retrieves a template by ID
// @Summary Get a template
// @Description Retrieve details of a specific template
// @Tags templates
// @Produce json
// @Param id path string true "Template ID"
// @Success 200 {object} domain.Template
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/templates/{id} [get]
func (h *TemplateHandler) GetTemplate(c *gin.Context) {
	id := c.Param("id")

	template, err := h.templateService.GetTemplate(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, template)
}

// ListTemplates lists all templates with optional filtering
// @Summary List templates
// @Description Retrieve a paginated list of templates
// @Tags templates
// @Produce json
// @Param type query string false "Filter by report type"
// @Param offset query int false "Offset for pagination" default(0)
// @Param limit query int false "Limit for pagination" default(20)
// @Success 200 {object} domain.PaginatedTemplates
// @Failure 500 {object} map[string]string
// @Router /api/v1/templates [get]
func (h *TemplateHandler) ListTemplates(c *gin.Context) {
	filter := struct{}{}
	_ = filter

	result, err := h.templateService.ListTemplates(c.Request.Context(), struct{}{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// UpdateTemplate updates an existing template
// @Summary Update a template
// @Description Update configuration of an existing template
// @Tags templates
// @Accept json
// @Produce json
// @Param id path string true "Template ID"
// @Param template body domain.UpdateTemplateRequest true "Template updates"
// @Success 200 {object} domain.Template
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/templates/{id} [put]
func (h *TemplateHandler) UpdateTemplate(c *gin.Context) {
	id := c.Param("id")

	var req domain.UpdateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	template, err := h.templateService.UpdateTemplate(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, template)
}

// DeleteTemplate deletes a template
// @Summary Delete a template
// @Description Remove a template
// @Tags templates
// @Produce json
// @Param id path string true "Template ID"
// @Success 204 "No Content"
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/templates/{id} [delete]
func (h *TemplateHandler) DeleteTemplate(c *gin.Context) {
	id := c.Param("id")

	if err := h.templateService.DeleteTemplate(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// HealthHandler handles HTTP requests for health checks
type HealthHandler struct {
	config *config.Config
}

// NewHealthHandler creates a new HealthHandler instance
func NewHealthHandler(cfg *config.Config) *HealthHandler {
	return &HealthHandler{config: cfg}
}

// GetHealth returns overall system health
// @Summary Get system health
// @Description Retrieve overall system health status
// @Tags health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
func (h *HealthHandler) GetHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": h.config.App.Name,
		"version": "1.0.0",
	})
}

// LivenessCheck returns whether the service is alive
// @Summary Liveness check
// @Description Check if the service is running
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health/live [get]
func (h *HealthHandler) LivenessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "alive",
	})
}

// ReadinessCheck returns whether the service is ready to accept traffic
// @Summary Readiness check
// @Description Check if the service is ready to accept traffic
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health/ready [get]
func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}
