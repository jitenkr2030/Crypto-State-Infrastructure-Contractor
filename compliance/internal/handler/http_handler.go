package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/csic/platform/compliance/internal/domain"
	"github.com/csic/platform/compliance/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// HTTPHandler handles HTTP requests for the compliance service
type HTTPHandler struct {
	complianceService *service.ComplianceService
	logger            *zap.Logger
}

// NewHTTPHandler creates a new HTTP handler
func NewHTTPHandler(complianceService *service.ComplianceService, logger *zap.Logger) *HTTPHandler {
	return &HTTPHandler{
		complianceService: complianceService,
		logger:            logger,
	}
}

// RegisterRoutes registers all HTTP routes
func (h *HTTPHandler) RegisterRoutes(router *gin.Engine) {
	// Health check
	router.GET("/health", h.HealthCheck)
	router.GET("/ready", h.ReadinessCheck)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Compliance check endpoints
		v1.POST("/check", h.CheckCompliance)
		v1.GET("/results/:transaction_id", h.GetComplianceResult)

		// Rule management endpoints
		rules := v1.Group("/rules")
		{
			rules.POST("", h.CreateRule)
			rules.GET("", h.ListRules)
			rules.GET("/:id", h.GetRule)
			rules.PUT("/:id", h.UpdateRule)
			rules.GET("/type/:type", h.GetRulesByType)
		}

		// Entity management endpoints
		entities := v1.Group("/entities")
		{
			entities.POST("", h.CreateEntity)
			entities.GET("/:id", h.GetEntity)
		}

		// Violation management endpoints
		violations := v1.Group("/violations")
		{
			violations.GET("", h.ListViolations)
			violations.POST("/:id/resolve", h.ResolveViolation)
		}

		// Watchlist endpoints
		watchlist := v1.Group("/watchlist")
		{
			watchlist.POST("", h.AddToWatchlist)
			watchlist.GET("/search", h.SearchWatchlist)
		}

		// Compliance results endpoints
		results := v1.Group("/results")
		{
			results.GET("", h.ListComplianceResults)
		}
	}
}

// HealthCheck handles health check requests
func (h *HTTPHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "compliance-service",
		"timestamp": time.Now().UTC(),
	})
}

// ReadinessCheck handles readiness check requests
func (h *HTTPHandler) ReadinessCheck(c *gin.Context) {
	ctx := c.Request.Context()

	if err := h.complianceService.HealthCheck(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}

// CheckComplianceRequest represents a compliance check request
type CheckComplianceRequest struct {
	ID             string                 `json:"id" binding:"required"`
	Type           string                 `json:"type" binding:"required"`
	SourceID       string                 `json:"source_id"`
	SourceType     string                 `json:"source_type"`
	SourceName     string                 `json:"source_name"`
	SourceAccount  string                 `json:"source_account"`
	SourceCountry  string                 `json:"source_country"`
	TargetID       string                 `json:"target_id"`
	TargetType     string                 `json:"target_type"`
	TargetName     string                 `json:"target_name"`
	TargetAccount  string                 `json:"target_account"`
	TargetCountry  string                 `json:"target_country"`
	Amount         float64                `json:"amount" binding:"required,gt=0"`
	Currency       string                 `json:"currency"`
	AssetType      string                 `json:"asset_type"`
	AssetID        string                 `json:"asset_id"`
	Timestamp      time.Time              `json:"timestamp"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// CheckCompliance handles compliance check requests
func (h *HTTPHandler) CheckCompliance(c *gin.Context) {
	var req CheckComplianceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid compliance check request",
			zap.Error(err),
			zap.String("path", c.Request.URL.Path))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request",
			"details": err.Error(),
		})
		return
	}

	// Convert request to transaction
	tx := &domain.Transaction{
		ID:             req.ID,
		Type:           req.Type,
		SourceID:       req.SourceID,
		SourceType:     req.SourceType,
		SourceName:     req.SourceName,
		SourceAccount:  req.SourceAccount,
		SourceCountry:  req.SourceCountry,
		TargetID:       req.TargetID,
		TargetType:     req.TargetType,
		TargetName:     req.TargetName,
		TargetAccount:  req.TargetAccount,
		TargetCountry:  req.TargetCountry,
		Amount:         req.Amount,
		Currency:       req.Currency,
		AssetType:      req.AssetType,
		AssetID:        req.AssetID,
		Timestamp:      req.Timestamp,
		Metadata:       req.Metadata,
	}

	if tx.Timestamp.IsZero() {
		tx.Timestamp = time.Now()
	}

	// Perform compliance check
	result, err := h.complianceService.CheckCompliance(c.Request.Context(), tx)
	if err != nil {
		h.logger.Error("Compliance check failed",
			zap.Error(err),
			zap.String("transaction_id", tx.ID))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "compliance check failed",
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetComplianceResult retrieves a cached compliance result
func (h *HTTPHandler) GetComplianceResult(c *gin.Context) {
	txID := c.Param("transaction_id")

	result, err := h.complianceService.GetCachedResult(c.Request.Context(), txID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "result not found",
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// CreateRuleRequest represents a rule creation request
type CreateRuleRequest struct {
	Name        string                 `json:"name" binding:"required"`
	Description string                 `json:"description"`
	Type        string                 `json:"type" binding:"required"`
	Severity    string                 `json:"severity" binding:"required"`
	Enabled     bool                   `json:"enabled"`
	Expression  string                 `json:"expression"`
	Parameters  map[string]interface{} `json:"parameters"`
	ExpiresAt   *time.Time             `json:"expires_at"`
}

// CreateRule handles rule creation requests
func (h *HTTPHandler) CreateRule(c *gin.Context) {
	var req CreateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid rule creation request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request",
			"details": err.Error(),
		})
		return
	}

	// Convert parameters
	params := domain.Parameters{}
	if req.Parameters != nil {
		if threshold, ok := req.Parameters["threshold"].(float64); ok {
			params.Threshold = threshold
		}
		if windowSec, ok := req.Parameters["window_seconds"].(float64); ok {
			params.WindowSeconds = int(windowSec)
		}
		if maxAttempts, ok := req.Parameters["max_attempts"].(float64); ok {
			params.MaxAttempts = int(maxAttempts)
		}
		if minAmount, ok := req.Parameters["min_amount"].(float64); ok {
			params.MinAmount = minAmount
		}
		if maxAmount, ok := req.Parameters["max_amount"].(float64); ok {
			params.MaxAmount = maxAmount
		}
		if allowedCountries, ok := req.Parameters["allowed_countries"].([]interface{}); ok {
			params.AllowedCountries = make([]string, len(allowedCountries))
			for i, c := range allowedCountries {
				params.AllowedCountries[i] = c.(string)
			}
		}
		if blockedCountries, ok := req.Parameters["blocked_countries"].([]interface{}); ok {
			params.BlockedCountries = make([]string, len(blockedCountries))
			for i, c := range blockedCountries {
				params.BlockedCountries[i] = c.(string)
			}
		}
		if requiredFields, ok := req.Parameters["required_fields"].([]interface{}); ok {
			params.RequiredFields = make([]string, len(requiredFields))
			for i, f := range requiredFields {
				params.RequiredFields[i] = f.(string)
			}
		}
	}

	rule := &domain.Rule{
		Name:        req.Name,
		Description: req.Description,
		Type:        domain.RuleType(req.Type),
		Severity:    domain.Severity(req.Severity),
		Enabled:     req.Enabled,
		Expression:  req.Expression,
		Parameters:  params,
		ExpiresAt:   req.ExpiresAt,
	}

	if err := h.complianceService.CreateRule(c.Request.Context(), rule); err != nil {
		h.logger.Error("Failed to create rule", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create rule",
		})
		return
	}

	c.JSON(http.StatusCreated, rule)
}

// ListRules handles rule listing requests
func (h *HTTPHandler) ListRules(c *gin.Context) {
	rules, err := h.complianceService.GetRulesByType(c.Request.Context(), "")
	if err != nil {
		h.logger.Error("Failed to list rules", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to list rules",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"rules": rules,
		"count": len(rules),
	})
}

// GetRule handles rule retrieval requests
func (h *HTTPHandler) GetRule(c *gin.Context) {
	id := c.Param("id")

	rule, err := h.complianceService.GetRule(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get rule", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get rule",
		})
		return
	}

	if rule == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "rule not found",
		})
		return
	}

	c.JSON(http.StatusOK, rule)
}

// UpdateRule handles rule update requests
func (h *HTTPHandler) UpdateRule(c *gin.Context) {
	id := c.Param("id")

	// Get existing rule
	rule, err := h.complianceService.GetRule(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get rule", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get rule",
		})
		return
	}

	if rule == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "rule not found",
		})
		return
	}

	// Bind update request
	var req CreateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid rule update request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request",
			"details": err.Error(),
		})
		return
	}

	// Update rule fields
	rule.Name = req.Name
	rule.Description = req.Description
	rule.Type = domain.RuleType(req.Type)
	rule.Severity = domain.Severity(req.Severity)
	rule.Enabled = req.Enabled
	rule.Expression = req.Expression

	if err := h.complianceService.UpdateRule(c.Request.Context(), rule); err != nil {
		h.logger.Error("Failed to update rule", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to update rule",
		})
		return
	}

	c.JSON(http.StatusOK, rule)
}

// GetRulesByType handles rule retrieval by type requests
func (h *HTTPHandler) GetRulesByType(c *gin.Context) {
	ruleType := domain.RuleType(c.Param("type"))

	rules, err := h.complianceService.GetRulesByType(c.Request.Context(), ruleType)
	if err != nil {
		h.logger.Error("Failed to get rules by type", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get rules",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"rules": rules,
		"count": len(rules),
		"type":  ruleType,
	})
}

// CreateEntityRequest represents an entity creation request
type CreateEntityRequest struct {
	Type        string                 `json:"type" binding:"required"`
	Name        string                 `json:"name" binding:"required"`
	Country     string                 `json:"country"`
	RiskScore   float64                `json:"risk_score"`
	Tags        []string               `json:"tags"`
	Metadata    map[string]interface{} `json:"metadata"`
	Blacklisted bool                   `json:"blacklisted"`
	Watchlist   bool                   `json:"watchlist"`
	KYCVerified bool                   `json:"kyc_verified"`
	AMLStatus   string                 `json:"aml_status"`
}

// CreateEntity handles entity creation requests
func (h *HTTPHandler) CreateEntity(c *gin.Context) {
	var req CreateEntityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid entity creation request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request",
			"details": err.Error(),
		})
		return
	}

	entity := &domain.Entity{
		Type:        req.Type,
		Name:        req.Name,
		Country:     req.Country,
		RiskScore:   req.RiskScore,
		Tags:        req.Tags,
		Metadata:    req.Metadata,
		Blacklisted: req.Blacklisted,
		Watchlist:   req.Watchlist,
		KYCVerified: req.KYCVerified,
		AMLStatus:   req.AMLStatus,
	}

	if err := h.complianceService.CreateEntity(c.Request.Context(), entity); err != nil {
		h.logger.Error("Failed to create entity", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create entity",
		})
		return
	}

	c.JSON(http.StatusCreated, entity)
}

// GetEntity handles entity retrieval requests
func (h *HTTPHandler) GetEntity(c *gin.Context) {
	id := c.Param("id")

	entity, err := h.complianceService.GetEntity(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get entity", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get entity",
		})
		return
	}

	if entity == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "entity not found",
		})
		return
	}

	c.JSON(http.StatusOK, entity)
}

// ListViolations handles violation listing requests
func (h *HTTPHandler) ListViolations(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	violations, err := h.complianceService.GetOpenViolations(c.Request.Context(), limit, offset)
	if err != nil {
		h.logger.Error("Failed to list violations", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to list violations",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"violations": violations,
		"count":      len(violations),
	})
}

// ResolveViolationRequest represents a violation resolution request
type ResolveViolationRequest struct {
	Resolution string `json:"resolution" binding:"required"`
	ResolvedBy string `json:"resolved_by" binding:"required"`
}

// ResolveViolation handles violation resolution requests
func (h *HTTPHandler) ResolveViolation(c *gin.Context) {
	id := c.Param("id")

	var req ResolveViolationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid violation resolution request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request",
			"details": err.Error(),
		})
		return
	}

	if err := h.complianceService.ResolveViolation(c.Request.Context(), id, req.Resolution, req.ResolvedBy); err != nil {
		h.logger.Error("Failed to resolve violation", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to resolve violation",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "violation resolved",
		"id":      id,
	})
}

// AddToWatchlistRequest represents a watchlist addition request
type AddToWatchlistRequest struct {
	Type       string     `json:"type" binding:"required"`
	Name       string     `json:"name" binding:"required"`
	Country    string     `json:"country"`
	ListSource string     `json:"list_source" binding:"required"`
	MatchScore float64    `json:"match_score"`
	ExpiresAt  *time.Time `json:"expires_at"`
}

// AddToWatchlist handles watchlist addition requests
func (h *HTTPHandler) AddToWatchlist(c *gin.Context) {
	var req AddToWatchlistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid watchlist addition request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request",
			"details": err.Error(),
		})
		return
	}

	entry := &domain.WatchlistEntry{
		Type:       req.Type,
		Name:       req.Name,
		Country:    req.Country,
		ListSource: req.ListSource,
		MatchScore: req.MatchScore,
		ExpiresAt:  req.ExpiresAt,
		Active:     true,
	}

	if err := h.complianceService.AddToWatchlist(c.Request.Context(), entry); err != nil {
		h.logger.Error("Failed to add to watchlist", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to add to watchlist",
		})
		return
	}

	c.JSON(http.StatusCreated, entry)
}

// SearchWatchlist handles watchlist search requests
func (h *HTTPHandler) SearchWatchlist(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "name parameter is required",
		})
		return
	}

	entries, err := h.complianceService.SearchWatchlist(c.Request.Context(), name)
	if err != nil {
		h.logger.Error("Failed to search watchlist", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to search watchlist",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"entries": entries,
		"count":   len(entries),
	})
}

// ListComplianceResults handles compliance result listing requests
func (h *HTTPHandler) ListComplianceResults(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	results, err := h.complianceService.GetComplianceResults(c.Request.Context(), limit, offset)
	if err != nil {
		h.logger.Error("Failed to list compliance results", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to list compliance results",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results": results,
		"count":   len(results),
	})
}

// CORSMiddleware adds CORS headers
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
