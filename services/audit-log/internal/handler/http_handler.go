package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"audit-log/internal/core/domain"
	"audit-log/internal/core/ports"
)

// AuditHandler handles HTTP requests for audit log operations
type AuditHandler struct {
	auditService ports.AuditLogService
	logger       ports.Logger
}

// NewAuditHandler creates a new AuditHandler instance
func NewAuditHandler(auditService ports.AuditLogService, logger ports.Logger) *AuditHandler {
	return &AuditHandler{
		auditService: auditService,
		logger:       logger,
	}
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string            `json:"error"`
	Details map[string]string `json:"details,omitempty"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// CreateEntryRequest represents a request to create an audit entry
type CreateEntryRequest struct {
	TraceID    string                 `json:"trace_id" binding:"required"`
	ActorID    string                 `json:"actor_id" binding:"required"`
	ActorType  string                 `json:"actor_type" binding:"required"`
	Action     string                 `json:"action" binding:"required"`
	Resource   string                 `json:"resource" binding:"required"`
	ResourceID string                 `json:"resource_id"`
	Operation  string                 `json:"operation"`
	Outcome    string                 `json:"outcome" binding:"required,oneof=success failure partial"`
	Severity   string                 `json:"severity" binding:"required,oneof=info warning error critical"`
	Payload    map[string]interface{} `json:"payload,omitempty"`
	Metadata   map[string]string      `json:"metadata,omitempty"`
}

// RegisterRoutes registers all audit handler routes
func (h *AuditHandler) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api/v1/audit")
	{
		// Entry management
		api.POST("/entries", h.CreateEntry)
		api.GET("/entries/:id", h.GetEntry)
		api.POST("/entries/search", h.SearchEntries)
		api.GET("/entries/search", h.SearchEntriesGet)

		// Verification
		api.GET("/entries/:id/verify", h.VerifyEntry)
		api.GET("/verify", h.VerifyChain)

		// Chain information
		api.GET("/chain/summary", h.GetChainSummary)

		// Reports
		api.POST("/reports/compliance", h.GenerateComplianceReport)
		api.POST("/reports/activity", h.GenerateActivityReport)
	}

	// Direct trace endpoint
	router.GET("/api/v1/audit/trace/:trace_id", h.GetByTraceID)
}

// CreateEntry handles creating a new audit entry
func (h *AuditHandler) CreateEntry(c *gin.Context) {
	var req CreateEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse create entry request", "error", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body",
			Details: map[string]string{
				"validation": err.Error(),
			},
		})
		return
	}

	// Extract source info from request
	sourceIP := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	request := domain.AuditEntryRequest{
		TraceID:    req.TraceID,
		ActorID:    req.ActorID,
		ActorType:  req.ActorType,
		Action:     req.Action,
		Resource:   req.Resource,
		ResourceID: req.ResourceID,
		Operation:  req.Operation,
		Outcome:    req.Outcome,
		Severity:   req.Severity,
		Payload:    req.Payload,
		Metadata:   req.Metadata,
		SourceIP:   sourceIP,
		UserAgent:  userAgent,
	}

	entry, err := h.auditService.CreateEntry(c.Request.Context(), request)
	if err != nil {
		h.logger.Error("Failed to create audit entry", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to create audit entry",
		})
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse{
		Message: "Audit entry created successfully",
		Data:    entry,
	})
}

// GetEntry retrieves an audit entry by ID
func (h *AuditHandler) GetEntry(c *gin.Context) {
	entryID := c.Param("id")
	if entryID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Entry ID is required",
		})
		return
	}

	entry, err := h.auditService.GetEntry(c.Request.Context(), entryID)
	if err != nil {
		if err.Error() == "audit entry not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "Audit entry not found",
			})
			return
		}
		h.logger.Error("Failed to get audit entry", "error", err, "id", entryID)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to retrieve audit entry",
		})
		return
	}

	c.JSON(http.StatusOK, entry)
}

// GetByTraceID retrieves all audit entries for a trace ID
func (h *AuditHandler) GetByTraceID(c *gin.Context) {
	traceID := c.Param("trace_id")
	if traceID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Trace ID is required",
		})
		return
	}

	entries, err := h.auditService.GetEntryByTraceID(c.Request.Context(), traceID)
	if err != nil {
		h.logger.Error("Failed to get audit entries by trace ID", "error", err, "traceID", traceID)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to retrieve audit entries",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Audit entries retrieved",
		Data:    entries,
	})
}

// SearchEntries searches for audit entries
func (h *AuditHandler) SearchEntries(c *gin.Context) {
	var req domain.AuditSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse search request", "error", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body",
			Details: map[string]string{
				"validation": err.Error(),
			},
		})
		return
	}

	response, err := h.auditService.SearchEntries(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to search audit entries", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to search audit entries",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// SearchEntriesGet handles GET-based search for convenience
func (h *AuditHandler) SearchEntriesGet(c *gin.Context) {
	req := domain.AuditSearchRequest{
		TraceID:   c.Query("trace_id"),
		ActorID:   c.Query("actor_id"),
		ActorType: c.Query("actor_type"),
		Action:    c.Query("action"),
		Resource:  c.Query("resource"),
		Outcome:   c.Query("outcome"),
		Severity:  c.Query("severity"),
		SourceIP:  c.Query("source_ip"),
		Page:      1,
		PageSize:  20,
	}

	// Parse pagination
	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil {
			req.Page = p
		}
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil {
			req.PageSize = ps
		}
	}

	// Parse date range
	if startTime := c.Query("start_time"); startTime != "" {
		if t, err := time.Parse(time.RFC3339, startTime); err == nil {
			req.StartTime = &t
		}
	}
	if endTime := c.Query("end_time"); endTime != "" {
		if t, err := time.Parse(time.RFC3339, endTime); err == nil {
			req.EndTime = &t
		}
	}

	response, err := h.auditService.SearchEntries(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to search audit entries", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to search audit entries",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// VerifyEntry verifies a single audit entry's integrity
func (h *AuditHandler) VerifyEntry(c *gin.Context) {
	entryID := c.Param("id")
	if entryID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Entry ID is required",
		})
		return
	}

	result, err := h.auditService.VerifyEntry(c.Request.Context(), entryID)
	if err != nil {
		if err.Error() == "audit entry not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "Audit entry not found",
			})
			return
		}
		h.logger.Error("Failed to verify audit entry", "error", err, "id", entryID)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to verify audit entry",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Verification complete",
		Data:    result,
	})
}

// VerifyChain verifies the hash chain
func (h *AuditHandler) VerifyChain(c *gin.Context) {
	startID := c.Query("start_id")
	if startID == "" {
		// Get the most recent entry ID
		summary, err := h.auditService.GetChainSummary(c.Request.Context())
		if err != nil {
			h.logger.Error("Failed to get chain summary", "error", err)
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: "Failed to get chain summary",
			})
			return
		}
		if summary.TotalEntries == 0 {
			c.JSON(http.StatusOK, SuccessResponse{
				Message: "Chain is empty",
				Data: map[string]interface{}{
					"valid":  true,
					"reason": "no entries to verify",
				},
			})
			return
		}
	}

	limit := 100
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	result, err := h.auditService.VerifyChain(c.Request.Context(), startID, limit)
	if err != nil {
		h.logger.Error("Failed to verify chain", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to verify chain",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Chain verification complete",
		Data:    result,
	})
}

// GetChainSummary returns statistics about the audit chain
func (h *AuditHandler) GetChainSummary(c *gin.Context) {
	summary, err := h.auditService.GetChainSummary(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get chain summary", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to get chain summary",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Chain summary retrieved",
		Data:    summary,
	})
}

// GenerateComplianceReport generates a compliance report
func (h *AuditHandler) GenerateComplianceReport(c *gin.Context) {
	var req struct {
		StartDate time.Time `json:"start_date" binding:"required"`
		EndDate   time.Time `json:"end_date" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse report request", "error", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body",
			Details: map[string]string{
				"validation": err.Error(),
			},
		})
		return
	}

	// For now, return a simple report
	searchReq := domain.AuditSearchRequest{
		StartTime: &req.StartDate,
		EndTime:   &req.EndDate,
		Page:      1,
		PageSize:  1000,
	}

	response, err := h.auditService.SearchEntries(c.Request.Context(), searchReq)
	if err != nil {
		h.logger.Error("Failed to generate compliance report", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to generate compliance report",
		})
		return
	}

	// Calculate statistics
	stats := map[string]interface{}{
		"period": map[string]interface{}{
			"start": req.StartDate,
			"end":   req.EndDate,
		},
		"total_entries":    response.TotalCount,
		"entries":          response.Entries,
		"generated_at":     time.Now().UTC(),
		"report_type":      "compliance",
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Compliance report generated",
		Data:    stats,
	})
}

// GenerateActivityReport generates an activity report for a specific actor
func (h *AuditHandler) GenerateActivityReport(c *gin.Context) {
	var req struct {
		ActorID   string    `json:"actor_id" binding:"required"`
		StartDate time.Time `json:"start_date" binding:"required"`
		EndDate   time.Time `json:"end_date" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse report request", "error", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body",
			Details: map[string]string{
				"validation": err.Error(),
			},
		})
		return
	}

	searchReq := domain.AuditSearchRequest{
		ActorID:   req.ActorID,
		StartTime: &req.StartDate,
		EndTime:   &req.EndDate,
		Page:      1,
		PageSize:  1000,
	}

	response, err := h.auditService.SearchEntries(c.Request.Context(), searchReq)
	if err != nil {
		h.logger.Error("Failed to generate activity report", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to generate activity report",
		})
		return
	}

	stats := map[string]interface{}{
		"actor_id": req.ActorID,
		"period": map[string]interface{}{
			"start": req.StartDate,
			"end":   req.EndDate,
		},
		"total_actions":   response.TotalCount,
		"actions":         response.Entries,
		"generated_at":    time.Now().UTC(),
		"report_type":     "activity",
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Activity report generated",
		Data:    stats,
	})
}

// HealthCheck handles health check endpoint
func (h *AuditHandler) HealthCheck(c *gin.Context) {
	healthy, err := h.auditService.HealthCheck(c.Request.Context())
	if err != nil || !healthy {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "unhealthy",
			"service": "audit-log",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "audit-log",
		"time":    time.Now().UTC(),
	})
}
