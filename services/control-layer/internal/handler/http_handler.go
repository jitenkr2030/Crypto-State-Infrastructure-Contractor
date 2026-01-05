package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"control-layer/internal/core/domain"
	"control-layer/internal/core/ports"
)

// PolicyHandler handles HTTP requests for policy operations
type PolicyHandler struct {
	policyService ports.PolicyService
	logger        ports.Logger
}

// NewPolicyHandler creates a new PolicyHandler instance
func NewPolicyHandler(policyService ports.PolicyService, logger ports.Logger) *PolicyHandler {
	return &PolicyHandler{
		policyService: policyService,
		logger:        logger,
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

// RegisterRoutes registers all policy handler routes
func (h *PolicyHandler) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api/v1/control")
	{
		// Policy management
		api.POST("/policies", h.CreatePolicy)
		api.GET("/policies", h.ListPolicies)
		api.GET("/policies/:id", h.GetPolicy)
		api.PUT("/policies/:id", h.UpdatePolicy)
		api.DELETE("/policies/:id", h.DeletePolicy)

		// Version history
		api.GET("/policies/:id/history", h.GetPolicyHistory)
		api.POST("/policies/:id/restore", h.RestorePolicyVersion)

		// Templates
		api.GET("/templates", h.GetPolicyTemplates)
		api.POST("/templates/:template_id/apply", h.ApplyPolicyTemplate)

		// Access control (main enforcement point)
		api.POST("/check", h.CheckAccess)
		api.POST("/check/bulk", h.BulkCheckAccess)
	}
}

// CreatePolicy handles creating a new policy
func (h *PolicyHandler) CreatePolicy(c *gin.Context) {
	var req domain.PolicyCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse create policy request", "error", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body",
			Details: map[string]string{
				"validation": err.Error(),
			},
		})
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		userID = "system"
	}

	policy, err := h.policyService.CreatePolicy(c.Request.Context(), req, userID)
	if err != nil {
		h.logger.Error("Failed to create policy", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to create policy",
		})
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse{
		Message: "Policy created successfully",
		Data:    policy,
	})
}

// GetPolicy retrieves a policy by ID
func (h *PolicyHandler) GetPolicy(c *gin.Context) {
	policyID := c.Param("id")
	if policyID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Policy ID is required",
		})
		return
	}

	policy, err := h.policyService.GetPolicy(c.Request.Context(), policyID)
	if err != nil {
		if err.Error() == "policy not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "Policy not found",
			})
			return
		}
		h.logger.Error("Failed to get policy", "error", err, "id", policyID)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to retrieve policy",
		})
		return
	}

	c.JSON(http.StatusOK, policy)
}

// ListPolicies lists all policies
func (h *PolicyHandler) ListPolicies(c *gin.Context) {
	activeOnly := c.Query("active_only") == "true"
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	response, err := h.policyService.ListPolicies(c.Request.Context(), activeOnly, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list policies", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to list policies",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// UpdatePolicy handles updating a policy
func (h *PolicyHandler) UpdatePolicy(c *gin.Context) {
	policyID := c.Param("id")
	if policyID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Policy ID is required",
		})
		return
	}

	var req domain.PolicyUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse update policy request", "error", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body",
			Details: map[string]string{
				"validation": err.Error(),
			},
		})
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		userID = "system"
	}

	policy, err := h.policyService.UpdatePolicy(c.Request.Context(), policyID, req, userID)
	if err != nil {
		if err.Error() == "policy not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "Policy not found",
			})
			return
		}
		h.logger.Error("Failed to update policy", "error", err, "id", policyID)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to update policy",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Policy updated successfully",
		Data:    policy,
	})
}

// DeletePolicy handles deleting a policy
func (h *PolicyHandler) DeletePolicy(c *gin.Context) {
	policyID := c.Param("id")
	if policyID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Policy ID is required",
		})
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&req)

	userID := c.GetString("user_id")
	if userID == "" {
		userID = "system"
	}

	err := h.policyService.DeletePolicy(c.Request.Context(), policyID, userID, req.Reason)
	if err != nil {
		if err.Error() == "policy not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "Policy not found",
			})
			return
		}
		h.logger.Error("Failed to delete policy", "error", err, "id", policyID)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to delete policy",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Policy deleted successfully",
	})
}

// GetPolicyHistory retrieves policy version history
func (h *PolicyHandler) GetPolicyHistory(c *gin.Context) {
	policyID := c.Param("id")
	if policyID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Policy ID is required",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	response, err := h.policyService.GetPolicyHistory(c.Request.Context(), policyID, page, pageSize)
	if err != nil {
		if err.Error() == "policy not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "Policy not found",
			})
			return
		}
		h.logger.Error("Failed to get policy history", "error", err, "id", policyID)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to get policy history",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// RestorePolicyVersion restores a policy to a previous version
func (h *PolicyHandler) RestorePolicyVersion(c *gin.Context) {
	policyID := c.Param("id")
	if policyID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Policy ID is required",
		})
		return
	}

	var req struct {
		Version int    `json:"version" binding:"required"`
		Reason  string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse restore request", "error", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body",
			Details: map[string]string{
				"validation": err.Error(),
			},
		})
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		userID = "system"
	}

	policy, err := h.policyService.RestorePolicyVersion(c.Request.Context(), policyID, req.Version, userID, req.Reason)
	if err != nil {
		h.logger.Error("Failed to restore policy version", "error", err, "id", policyID, "version", req.Version)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to restore policy version",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Policy restored successfully",
		Data:    policy,
	})
}

// GetPolicyTemplates returns available policy templates
func (h *PolicyHandler) GetPolicyTemplates(c *gin.Context) {
	templates, err := h.policyService.GetPolicyTemplates(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get policy templates", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to get policy templates",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Policy templates retrieved",
		Data:    templates,
	})
}

// ApplyPolicyTemplate applies a policy template
func (h *PolicyHandler) ApplyPolicyTemplate(c *gin.Context) {
	templateID := c.Param("template_id")
	if templateID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Template ID is required",
		})
		return
	}

	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse apply template request", "error", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body",
			Details: map[string]string{
				"validation": err.Error(),
			},
		})
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		userID = "system"
	}

	policy, err := h.policyService.ApplyPolicyTemplate(c.Request.Context(), templateID, req.Name, userID)
	if err != nil {
		h.logger.Error("Failed to apply policy template", "error", err, "templateId", templateID)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to apply policy template",
		})
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse{
		Message: "Policy created from template",
		Data:    policy,
	})
}

// CheckAccess performs an access control check
func (h *PolicyHandler) CheckAccess(c *gin.Context) {
	var req domain.AccessCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse access check request", "error", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body",
			Details: map[string]string{
				"validation": err.Error(),
			},
		})
		return
	}

	// Set context from request headers if not provided
	if req.Context.Time == nil {
		now := time.Now().UTC()
		req.Context.Time = &now
	}
	if req.Context.IPAddress == "" {
		req.Context.IPAddress = c.ClientIP()
	}
	if req.Context.UserAgent == "" {
		req.Context.UserAgent = c.GetHeader("User-Agent")
	}
	if req.Context.Environment == "" {
		req.Context.Environment = "production"
	}

	response, err := h.policyService.CheckAccess(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to check access", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to check access",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// BulkCheckAccess performs multiple access control checks
func (h *PolicyHandler) BulkCheckAccess(c *gin.Context) {
	var requests []domain.AccessCheckRequest
	if err := c.ShouldBindJSON(&requests); err != nil {
		h.logger.Error("Failed to parse bulk check request", "error", err)
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
			Error: "No requests provided",
		})
		return
	}

	if len(requests) > 100 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Batch size exceeds maximum limit of 100",
		})
		return
	}

	responses, err := h.policyService.BulkCheckAccess(c.Request.Context(), requests)
	if err != nil {
		h.logger.Error("Failed to perform bulk check", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to perform bulk check",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Bulk check completed",
		Data:    responses,
	})
}

// HealthCheck handles health check endpoint
func (h *PolicyHandler) HealthCheck(c *gin.Context) {
	healthy, err := h.policyService.HealthCheck(c.Request.Context())
	if err != nil || !healthy {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "unhealthy",
			"service": "control-layer",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "control-layer",
		"time":    time.Now().UTC(),
	})
}
