package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"forensic-tools/internal/core/ports"
	"forensic-tools/internal/core/service"
)

// ForensicHandler handles HTTP requests for forensic operations
type ForensicHandler struct {
	forensicService ports.ForensicService
	logger          ports.Logger
}

// NewForensicHandler creates a new ForensicHandler instance
func NewForensicHandler(forensicService ports.ForensicService, logger ports.Logger) *ForensicHandler {
	return &ForensicHandler{
		forensicService: forensicService,
		logger:          logger,
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

// ChainOfCustodyResponse represents chain of custody records
type ChainOfCustodyResponse struct {
	EvidenceID    string    `json:"evidence_id"`
	EvidenceType  string    `json:"evidence_type"`
	Hash          string    `json:"hash"`
	CustodyRecords []CustodyRecord `json:"custody_records"`
}

// CustodyRecord represents a single custody transfer record
type CustodyRecord struct {
	Timestamp   time.Time `json:"timestamp"`
	Handler     string    `json:"handler"`
	Action      string    `json:"action"`
	Location    string    `json:"location"`
	Notes       string    `json:"notes,omitempty"`
	DigitalSig  string    `json:"digital_signature"`
}

// AnalysisRequest represents an evidence analysis request
type AnalysisRequest struct {
	EvidenceID  string                 `json:"evidence_id" binding:"required"`
	AnalysisType string                `json:"analysis_type" binding:"required"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// AnalysisResponse represents the result of evidence analysis
type AnalysisResponse struct {
	AnalysisID    string                 `json:"analysis_id"`
	EvidenceID    string                 `json:"evidence_id"`
	AnalysisType  string                 `json:"analysis_type"`
	Status        string                 `json:"status"`
	Results       map[string]interface{} `json:"results"`
	Findings      []string               `json:"findings"`
	StartedAt     time.Time              `json:"started_at"`
	CompletedAt   *time.Time             `json:"completed_at,omitempty"`
	ProcessedBy   string                 `json:"processed_by"`
}

// SearchRequest represents an evidence search request
type SearchRequest struct {
	Query       string   `json:"query" binding:"required"`
	EvidenceTypes []string `json:"evidence_types,omitempty"`
	DateFrom    *time.Time `json:"date_from,omitempty"`
	DateTo      *time.Time `json:"date_to,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Page        int      `json:"page"`
	PageSize    int      `json:"page_size"`
}

// SearchResponse represents search results
type SearchResponse struct {
	Results     []EvidenceSummary `json:"results"`
	TotalCount  int64             `json:"total_count"`
	Page        int               `json:"page"`
	PageSize    int               `json:"page_size"`
	TotalPages  int               `json:"total_pages"`
}

// EvidenceSummary represents a brief summary of evidence
type EvidenceSummary struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Name        string    `json:"name"`
	Hash        string    `json:"hash"`
	CollectedAt time.Time `json:"collected_at"`
	Tags        []string  `json:"tags"`
}

// EvidenceRequest represents a new evidence collection request
type EvidenceRequest struct {
	Name        string            `json:"name" binding:"required"`
	Type        string            `json:"type" binding:"required"`
	Description string            `json:"description"`
	Source      string            `json:"source" binding:"required"`
	Tags        []string          `json:"tags,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// CollectionRequest represents evidence collection with file upload
type CollectionRequest struct {
	Name        string   `form:"name" binding:"required"`
	Type        string   `form:"type" binding:"required"`
	Description string   `form:"description"`
	Source      string   `form:"source" binding:"required"`
	Tags        []string `form:"tags"`
	File        *os.File `form:"file"`
}

// RegisterRoutes registers all forensic handler routes
func (h *ForensicHandler) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api/v1/forensic")
	{
		// Evidence collection endpoints
		api.POST("/evidence", h.CollectEvidence)
		api.POST("/evidence/batch", h.BatchCollectEvidence)
		api.GET("/evidence/:id", h.GetEvidence)
		api.GET("/evidence/:id/download", h.DownloadEvidence)
		api.DELETE("/evidence/:id", h.DeleteEvidence)

		// Chain of custody endpoints
		api.GET("/evidence/:id/custody", h.GetChainOfCustody)
		api.POST("/evidence/:id/custody", h.AddCustodyRecord)
		api.POST("/evidence/:id/custody/verify", h.VerifyChainOfCustody)

		// Analysis endpoints
		api.POST("/analysis", h.StartAnalysis)
		api.GET("/analysis/:id", h.GetAnalysisStatus)
		api.GET("/analysis/:id/results", h.GetAnalysisResults)
		api.GET("/analysis", h.ListAnalyses)

		// Search endpoints
		api.POST("/search", h.SearchEvidence)
		api.GET("/search", h.SearchEvidenceGet)
	}
}

// CollectEvidence handles new evidence collection
func (h *ForensicHandler) CollectEvidence(c *gin.Context) {
	var req EvidenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse evidence collection request", "error", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body",
			Details: map[string]string{
				"validation": err.Error(),
			},
		})
		return
	}

	evidence, err := h.forensicService.CollectEvidence(c.Request.Context(), req.Name, req.Type, req.Source, req.Description, req.Tags, req.Metadata)
	if err != nil {
		h.logger.Error("Failed to collect evidence", "error", err, "name", req.Name)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to collect evidence",
			Details: map[string]string{
				"service": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse{
		Message: "Evidence collected successfully",
		Data:    evidence,
	})
}

// BatchCollectEvidence handles batch evidence collection
func (h *ForensicHandler) BatchCollectEvidence(c *gin.Context) {
	var requests []EvidenceRequest
	if err := c.ShouldBindJSON(&requests); err != nil {
		h.logger.Error("Failed to parse batch collection request", "error", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body",
			Details: map[string]string{
				"validation": err.Error(),
			},
		})
		return
	}

	if len(requests) == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "No evidence requests provided",
		})
		return
	}

	if len(requests) > 100 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Batch size exceeds maximum limit of 100",
		})
		return
	}

	results, err := h.forensicService.BatchCollectEvidence(c.Request.Context(), requests)
	if err != nil {
		h.logger.Error("Failed to batch collect evidence", "error", err, "count", len(requests))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to process batch collection",
			Details: map[string]string{
				"service": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse{
		Message: fmt.Sprintf("Processed %d evidence items", len(results)),
		Data:    results,
	})
}

// GetEvidence retrieves evidence by ID
func (h *ForensicHandler) GetEvidence(c *gin.Context) {
	evidenceID := c.Param("id")
	if evidenceID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Evidence ID is required",
		})
		return
	}

	evidence, err := h.forensicService.GetEvidence(c.Request.Context(), evidenceID)
	if err != nil {
		if err == service.ErrEvidenceNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "Evidence not found",
			})
			return
		}
		h.logger.Error("Failed to get evidence", "error", err, "id", evidenceID)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to retrieve evidence",
		})
		return
	}

	c.JSON(http.StatusOK, evidence)
}

// DownloadEvidence handles evidence file download
func (h *ForensicHandler) DownloadEvidence(c *gin.Context) {
	evidenceID := c.Param("id")
	if evidenceID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Evidence ID is required",
		})
		return
	}

	// Verify chain of custody before download
	verified, err := h.forensicService.VerifyChainOfCustody(c.Request.Context(), evidenceID)
	if err != nil {
		h.logger.Error("Failed to verify chain of custody", "error", err, "id", evidenceID)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to verify chain of custody",
		})
		return
	}

	if !verified {
		c.JSON(http.StatusForbidden, ErrorResponse{
			Error: "Chain of custody verification failed - evidence integrity compromised",
		})
		return
	}

	file, metadata, err := h.forensicService.GetEvidenceFile(c.Request.Context(), evidenceID)
	if err != nil {
		if err == service.ErrEvidenceNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "Evidence not found",
			})
			return
		}
		h.logger.Error("Failed to download evidence", "error", err, "id", evidenceID)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to download evidence",
		})
		return
	}
	defer file.Close()

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", metadata.Filename))
	c.Header("Content-Type", metadata.ContentType)
	c.Header("Content-Length", strconv.FormatInt(metadata.Size, 10))
	c.Header("X-Evidence-Hash", metadata.Hash)
	c.Header("X-Evidence-ID", evidenceID)

	_, err = io.Copy(c.Writer, file)
	if err != nil {
		h.logger.Error("Failed to stream file", "error", err, "id", evidenceID)
	}
}

// DeleteEvidence handles evidence deletion (with audit trail)
func (h *ForensicHandler) DeleteEvidence(c *gin.Context) {
	evidenceID := c.Param("id")
	if evidenceID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Evidence ID is required",
		})
		return
	}

	// Check if user has permission to delete
	deletedBy := c.GetString("user_id")
	if deletedBy == "" {
		deletedBy = "system"
	}

	err := h.forensicService.DeleteEvidence(c.Request.Context(), evidenceID, deletedBy, "Deleted via API")
	if err != nil {
		if err == service.ErrEvidenceNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "Evidence not found",
			})
			return
		}
		h.logger.Error("Failed to delete evidence", "error", err, "id", evidenceID)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to delete evidence",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Evidence deleted successfully",
	})
}

// GetChainOfCustody retrieves the chain of custody for evidence
func (h *ForensicHandler) GetChainOfCustody(c *gin.Context) {
	evidenceID := c.Param("id")
	if evidenceID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Evidence ID is required",
		})
		return
	}

	custody, err := h.forensicService.GetChainOfCustody(c.Request.Context(), evidenceID)
	if err != nil {
		if err == service.ErrEvidenceNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "Evidence not found",
			})
			return
		}
		h.logger.Error("Failed to get chain of custody", "error", err, "id", evidenceID)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to retrieve chain of custody",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Chain of custody retrieved",
		Data:    custody,
	})
}

// AddCustodyRecord adds a new custody record
func (h *ForensicHandler) AddCustodyRecord(c *gin.Context) {
	evidenceID := c.Param("id")
	if evidenceID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Evidence ID is required",
		})
		return
	}

	var record CustodyRecord
	if err := c.ShouldBindJSON(&record); err != nil {
		h.logger.Error("Failed to parse custody record", "error", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body",
			Details: map[string]string{
				"validation": err.Error(),
			},
		})
		return
	}

	handler := c.GetString("user_id")
	if handler == "" {
		handler = "system"
	}

	err := h.forensicService.AddCustodyRecord(c.Request.Context(), evidenceID, handler, record.Action, record.Location, record.Notes, record.DigitalSig)
	if err != nil {
		if err == service.ErrEvidenceNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "Evidence not found",
			})
			return
		}
		h.logger.Error("Failed to add custody record", "error", err, "id", evidenceID)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to add custody record",
		})
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse{
		Message: "Custody record added successfully",
	})
}

// VerifyChainOfCustody verifies the integrity of the chain of custody
func (h *ForensicHandler) VerifyChainOfCustody(c *gin.Context) {
	evidenceID := c.Param("id")
	if evidenceID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Evidence ID is required",
		})
		return
	}

	verified, err := h.forensicService.VerifyChainOfCustody(c.Request.Context(), evidenceID)
	if err != nil {
		h.logger.Error("Failed to verify chain of custody", "error", err, "id", evidenceID)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to verify chain of custody",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Chain of custody verification complete",
		Data: map[string]interface{}{
			"verified":   verified,
			"evidenceId": evidenceID,
			"timestamp":  time.Now().UTC(),
		},
	})
}

// StartAnalysis initiates evidence analysis
func (h *ForensicHandler) StartAnalysis(c *gin.Context) {
	var req AnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse analysis request", "error", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body",
			Details: map[string]string{
				"validation": err.Error(),
			},
		})
		return
	}

	// Verify evidence exists
	exists, err := h.forensicService.EvidenceExists(c.Request.Context(), req.EvidenceID)
	if err != nil {
		h.logger.Error("Failed to check evidence existence", "error", err, "id", req.EvidenceID)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to verify evidence",
		})
		return
	}

	if !exists {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error: "Evidence not found",
		})
		return
	}

	processedBy := c.GetString("user_id")
	if processedBy == "" {
		processedBy = "system"
	}

	analysis, err := h.forensicService.StartAnalysis(c.Request.Context(), req.EvidenceID, req.AnalysisType, req.Parameters, processedBy)
	if err != nil {
		h.logger.Error("Failed to start analysis", "error", err, "evidenceId", req.EvidenceID)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to start analysis",
		})
		return
	}

	c.JSON(http.StatusAccepted, SuccessResponse{
		Message: "Analysis started successfully",
		Data:    analysis,
	})
}

// GetAnalysisStatus retrieves the status of an analysis
func (h *ForensicHandler) GetAnalysisStatus(c *gin.Context) {
	analysisID := c.Param("id")
	if analysisID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Analysis ID is required",
		})
		return
	}

	analysis, err := h.forensicService.GetAnalysis(c.Request.Context(), analysisID)
	if err != nil {
		if err == service.ErrAnalysisNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "Analysis not found",
			})
			return
		}
		h.logger.Error("Failed to get analysis status", "error", err, "id", analysisID)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to retrieve analysis status",
		})
		return
	}

	c.JSON(http.StatusOK, analysis)
}

// GetAnalysisResults retrieves the results of a completed analysis
func (h *ForensicHandler) GetAnalysisResults(c *gin.Context) {
	analysisID := c.Param("id")
	if analysisID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Analysis ID is required",
		})
		return
	}

	results, err := h.forensicService.GetAnalysisResults(c.Request.Context(), analysisID)
	if err != nil {
		if err == service.ErrAnalysisNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "Analysis not found",
			})
			return
		}
		if err == service.ErrAnalysisInProgress {
			c.JSON(http.StatusAccepted, ErrorResponse{
				Error: "Analysis is still in progress",
				Details: map[string]string{
					"status": "processing",
				},
			})
			return
		}
		h.logger.Error("Failed to get analysis results", "error", err, "id", analysisID)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to retrieve analysis results",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Analysis results retrieved",
		Data:    results,
	})
}

// ListAnalyses lists analyses with optional filters
func (h *ForensicHandler) ListAnalyses(c *gin.Context) {
	evidenceID := c.Query("evidence_id")
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	analyses, total, err := h.forensicService.ListAnalyses(c.Request.Context(), evidenceID, status, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list analyses", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to list analyses",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Analyses retrieved",
		Data: map[string]interface{}{
			"analyses":   analyses,
			"total":      total,
			"page":       page,
			"page_size":  pageSize,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// SearchEvidence searches for evidence using various criteria
func (h *ForensicHandler) SearchEvidence(c *gin.Context) {
	var req SearchRequest
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

	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	results, total, err := h.forensicService.SearchEvidence(c.Request.Context(), req.Query, req.EvidenceTypes, req.DateFrom, req.DateTo, req.Tags, req.Page, req.PageSize)
	if err != nil {
		h.logger.Error("Failed to search evidence", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to search evidence",
		})
		return
	}

	response := SearchResponse{
		Results:    results,
		TotalCount: total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: int((total + int64(req.PageSize) - 1) / int64(req.PageSize)),
	}

	c.JSON(http.StatusOK, response)
}

// SearchEvidenceGet handles GET-based search for convenience
func (h *ForensicHandler) SearchEvidenceGet(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Search query (q) is required",
		})
		return
	}

	evidenceTypes := c.QueryArray("types")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	results, total, err := h.forensicService.SearchEvidence(c.Request.Context(), query, evidenceTypes, nil, nil, nil, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to search evidence", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to search evidence",
		})
		return
	}

	response := SearchResponse{
		Results:    results,
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int((total + int64(pageSize) - 1) / int64(pageSize)),
	}

	c.JSON(http.StatusOK, response)
}

// HealthCheck handles health check endpoint
func (h *ForensicHandler) HealthCheck(c *gin.Context) {
	healthy, err := h.forensicService.HealthCheck(c.Request.Context())
	if err != nil || !healthy {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "unhealthy",
			"service": "forensic-tools",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "forensic-tools",
		"time":    time.Now().UTC(),
	})
}

// Custom JSON marshaling for ChainOfCustodyResponse
func (r ChainOfCustodyResponse) MarshalJSON() ([]byte, error) {
	type Alias ChainOfCustodyResponse
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(&r),
	})
}
