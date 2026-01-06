package handler

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/csic/platform/service/reporting/regulatory/internal/domain"
	"github.com/csic/platform/service/reporting/regulatory/internal/service"
	"github.com/gin-gonic/gin"
)

// ReportHandler handles HTTP requests for report operations
type ReportHandler struct {
	reportService *service.ReportService
}

// NewReportHandler creates a new ReportHandler instance
func NewReportHandler(reportService *service.ReportService) *ReportHandler {
	return &ReportHandler{
		reportService: reportService,
	}
}

// CreateReport creates a new report
// @Summary Create a new report
// @Description Request creation of a new regulatory report
// @Tags reports
// @Accept json
// @Produce json
// @Param report body domain.CreateReportRequest true "Report configuration"
// @Success 202 {object} domain.Report
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/reports [post]
func (h *ReportHandler) CreateReport(c *gin.Context) {
	var req domain.CreateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	report, err := h.reportService.CreateReport(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, report)
}

// GetReport retrieves a report by ID
// @Summary Get a report
// @Description Retrieve details of a specific report
// @Tags reports
// @Produce json
// @Param id path string true "Report ID"
// @Success 200 {object} domain.Report
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/reports/{id} [get]
func (h *ReportHandler) GetReport(c *gin.Context) {
	id := c.Param("id")

	report, err := h.reportService.GetReport(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// ListReports lists all reports with optional filtering
// @Summary List reports
// @Description Retrieve a paginated list of reports
// @Tags reports
// @Produce json
// @Param type query string false "Filter by report type"
// @Param status query string false "Filter by status"
// @Param offset query int false "Offset for pagination" default(0)
// @Param limit query int false "Limit for pagination" default(20)
// @Success 200 {object} domain.PaginatedReports
// @Failure 500 {object} map[string]string
// @Router /api/v1/reports [get]
func (h *ReportHandler) ListReports(c *gin.Context) {
	// Parse filter from query parameters
	// For now, use empty filter
	filter := struct{}{}
	_ = filter

	result, err := h.reportService.ListReports(c.Request.Context(), struct{}{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GenerateReport generates a report synchronously
// @Summary Generate a report
// @Description Trigger report generation and wait for completion
// @Tags reports
// @Produce json
// @Param id path string true "Report ID"
// @Success 200 {object} domain.Report
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/reports/{id}/generate [post]
func (h *ReportHandler) GenerateReport(c *gin.Context) {
	id := c.Param("id")

	report, err := h.reportService.GenerateReport(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// DownloadReport downloads a generated report
// @Summary Download a report
// @Description Download the generated report file
// @Tags reports
// @Produce octet-stream
// @Param id path string true "Report ID"
// @Success 200 {file} binary
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/reports/{id}/download [get]
func (h *ReportHandler) DownloadReport(c *gin.Context) {
	id := c.Param("id")

	report, err := h.reportService.GetReport(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if report.FilePath == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "report not generated yet"})
		return
	}

	// Serve the file
	filename := filepath.Base(report.FilePath)
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/octet-stream")
	c.File(report.FilePath)
}

// DeleteReport deletes a report
// @Summary Delete a report
// @Description Remove a report
// @Tags reports
// @Produce json
// @Param id path string true "Report ID"
// @Success 204 "No Content"
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/reports/{id} [delete]
func (h *ReportHandler) DeleteReport(c *gin.Context) {
	id := c.Param("id")

	if err := h.reportService.DeleteReport(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Delete file if exists
	report, _ := h.reportService.GetReport(c.Request.Context(), id)
	if report != nil && report.FilePath != "" {
		os.Remove(report.FilePath)
	}

	c.Status(http.StatusNoContent)
}

// ListReportTypes lists available report types
// @Summary List report types
// @Description Get all available report types
// @Tags report-types
// @Produce json
// @Success 200 {array} domain.ReportTypeInfo
// @Router /api/v1/report-types [get]
func (h *ReportHandler) ListReportTypes(c *gin.Context) {
	types := h.reportService.ListReportTypes(c.Request.Context())
	c.JSON(http.StatusOK, types)
}

// GetReportType returns information about a specific report type
// @Summary Get report type
// @Description Get details of a specific report type
// @Tags report-types
// @Produce json
// @Param type path string true "Report type"
// @Success 200 {object} domain.ReportTypeInfo
// @Failure 404 {object} map[string]string
// @Router /api/v1/report-types/{type} [get]
func (h *ReportHandler) GetReportType(c *gin.Context) {
	reportType := domain.ReportType(c.Param("type"))

	info, err := h.reportService.GetReportType(c.Request.Context(), reportType)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, info)
}
