package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/csic/oversight/internal/core/domain"
	"github.com/csic/oversight/internal/core/ports"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// HTTPHandler handles REST API requests for the oversight service
type HTTPHandler struct {
	alertRepo   ports.AlertRepository
	ruleRepo    ports.RuleRepository
	healthRepo  ports.HealthScorerService // Interface for health service
	exchangeRepo ports.ExchangeRepository
	logger      *zap.Logger
}

// NewHTTPHandler creates a new HTTPHandler
func NewHTTPHandler(
	alertRepo ports.AlertRepository,
	ruleRepo ports.RuleRepository,
	healthScorer ports.HealthScorerService,
	exchangeRepo ports.ExchangeRepository,
	logger *zap.Logger,
) *HTTPHandler {
	return &HTTPHandler{
		alertRepo:    alertRepo,
		ruleRepo:     ruleRepo,
		healthRepo:   healthScorer,
		exchangeRepo: exchangeRepo,
		logger:       logger,
	}
}

// RegisterRoutes registers all API routes
func (h *HTTPHandler) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api/v1")
	{
		// Health and status endpoints
		api.GET("/health", h.GetHealthStatus)
		api.GET("/ready", h.GetReadiness)

		// Alert endpoints
		alerts := api.Group("/alerts")
		{
			alerts.GET("", h.ListAlerts)
			alerts.GET("/:id", h.GetAlert)
			alerts.PUT("/:id/status", h.UpdateAlertStatus)
			alerts.GET("/stats", h.GetAlertStats)
			alerts.GET("/exchange/:exchange_id", h.GetAlertsByExchange)
		}

		// Rule endpoints
		rules := api.Group("/rules")
		{
			rules.GET("", h.ListRules)
			rules.GET("/:id", h.GetRule)
			rules.POST("", h.CreateRule)
			rules.PUT("/:id", h.UpdateRule)
			rules.DELETE("/:id", h.DeleteRule)
		}

		// Exchange health endpoints
		exchanges := api.Group("/exchanges")
		{
			exchanges.GET("", h.ListExchanges)
			exchanges.GET("/:id", h.GetExchange)
			exchanges.GET("/:id/health", h.GetExchangeHealth)
			exchanges.GET("/health", h.ListExchangeHealth)
			exchanges.POST("/:id/throttle", h.ThrottleExchange)
		}

		// Analytics endpoints
		analytics := api.Group("/analytics")
		{
			analytics.GET("/volume", h.GetVolumeStats)
			analytics.GET("/trades/count", h.GetTradeCount)
		}
	}

	// Metrics endpoint (for Prometheus)
	router.GET("/metrics", h.GetMetrics)
}

// GetHealthStatus returns the service health status
func (h *HTTPHandler) GetHealthStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"service":   "oversight",
		"version":   "1.0.0",
	})
}

// GetReadiness returns the service readiness status
func (h *HTTPHandler) GetReadiness(c *gin.Context) {
	// Check database connectivity
	if err := h.alertRepo.GetAlerts(c.Request.Context(), ports.AlertFilter{Limit: 1}); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"error":  "database unavailable",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}

// ListAlerts returns a list of alerts with optional filtering
func (h *HTTPHandler) ListAlerts(c *gin.Context) {
	filter := h.parseAlertFilter(c)
	alerts, err := h.alertRepo.GetAlerts(c.Request.Context(), filter)
	if err != nil {
		h.logger.Error("Failed to list alerts", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve alerts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"alerts":      alerts,
		"total_count": len(alerts),
		"filter":      filter,
	})
}

// GetAlert returns a specific alert by ID
func (h *HTTPHandler) GetAlert(c *gin.Context) {
	alertID := c.Param("id")
	alert, err := h.alertRepo.GetAlertByID(c.Request.Context(), alertID)
	if err != nil {
		h.logger.Error("Failed to get alert", zap.String("alert_id", alertID), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Alert not found"})
		return
	}

	c.JSON(http.StatusOK, alert)
}

// UpdateAlertStatus updates the status of an alert
func (h *HTTPHandler) UpdateAlertStatus(c *gin.Context) {
	alertID := c.Param("id")

	var req struct {
		Status    domain.AlertStatus `json:"status"`
		Resolution string           `json:"resolution"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.alertRepo.UpdateAlertStatus(c.Request.Context(), alertID, req.Status, req.Resolution); err != nil {
		h.logger.Error("Failed to update alert status",
			zap.String("alert_id", alertID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update alert"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Alert status updated successfully",
		"alert_id":  alertID,
		"new_status": req.Status,
	})
}

// GetAlertStats returns alert statistics
func (h *HTTPHandler) GetAlertStats(c *gin.Context) {
	startStr := c.Query("start_time")
	endStr := c.Query("end_time")

	var start, end time.Time
	var err error

	if startStr != "" {
		start, err = time.Parse(time.RFC3339, startStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_time format"})
			return
		}
	} else {
		start = time.Now().UTC().Add(-24 * time.Hour)
	}

	if endStr != "" {
		end, err = time.Parse(time.RFC3339, endStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_time format"})
			return
		}
	} else {
		end = time.Now().UTC()
	}

	stats, err := h.alertRepo.GetAlertStats(c.Request.Context(), start, end)
	if err != nil {
		h.logger.Error("Failed to get alert stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve alert statistics"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetAlertsByExchange returns alerts for a specific exchange
func (h *HTTPHandler) GetAlertsByExchange(c *gin.Context) {
	exchangeID := c.Param("exchange_id")
	limitStr := c.DefaultQuery("limit", "100")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 100
	}

	alerts, err := h.alertRepo.GetAlertsByExchange(c.Request.Context(), exchangeID, limit)
	if err != nil {
		h.logger.Error("Failed to get alerts by exchange",
			zap.String("exchange_id", exchangeID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve alerts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"exchange_id": exchangeID,
		"alerts":      alerts,
		"count":       len(alerts),
	})
}

// ListRules returns all detection rules
func (h *HTTPHandler) ListRules(c *gin.Context) {
	rules, err := h.ruleRepo.GetAllRules(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to list rules", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve rules"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"rules": rules,
		"count": len(rules),
	})
}

// GetRule returns a specific rule by ID
func (h *HTTPHandler) GetRule(c *gin.Context) {
	ruleID := c.Param("id")
	rule, err := h.ruleRepo.GetRuleByID(c.Request.Context(), ruleID)
	if err != nil {
		h.logger.Error("Failed to get rule", zap.String("rule_id", ruleID), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Rule not found"})
		return
	}

	c.JSON(http.StatusOK, rule)
}

// CreateRule creates a new detection rule
func (h *HTTPHandler) CreateRule(c *gin.Context) {
	var rule domain.DetectionRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.ruleRepo.SaveRule(c.Request.Context(), rule); err != nil {
		h.logger.Error("Failed to create rule", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create rule"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Rule created successfully",
		"rule_id": rule.ID,
	})
}

// UpdateRule updates an existing detection rule
func (h *HTTPHandler) UpdateRule(c *gin.Context) {
	ruleID := c.Param("id")

	var rule domain.DetectionRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	rule.ID = ruleID
	if err := h.ruleRepo.UpdateRule(c.Request.Context(), rule); err != nil {
		h.logger.Error("Failed to update rule", zap.String("rule_id", ruleID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update rule"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Rule updated successfully",
		"rule_id": ruleID,
	})
}

// DeleteRule deletes a detection rule
func (h *HTTPHandler) DeleteRule(c *gin.Context) {
	ruleID := c.Param("id")

	if err := h.ruleRepo.DeleteRule(c.Request.Context(), ruleID); err != nil {
		h.logger.Error("Failed to delete rule", zap.String("rule_id", ruleID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete rule"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Rule deleted successfully",
		"rule_id": ruleID,
	})
}

// ListExchanges returns all registered exchanges
func (h *HTTPHandler) ListExchanges(c *gin.Context) {
	exchanges, err := h.exchangeRepo.GetAllExchanges(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to list exchanges", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve exchanges"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"exchanges": exchanges,
		"count":     len(exchanges),
	})
}

// GetExchange returns a specific exchange by ID
func (h *HTTPHandler) GetExchange(c *gin.Context) {
	exchangeID := c.Param("id")
	exchange, err := h.exchangeRepo.GetExchangeByID(c.Request.Context(), exchangeID)
	if err != nil {
		h.logger.Error("Failed to get exchange", zap.String("exchange_id", exchangeID), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Exchange not found"})
		return
	}

	c.JSON(http.StatusOK, exchange)
}

// GetExchangeHealth returns the health status of a specific exchange
func (h *HTTPHandler) GetExchangeHealth(c *gin.Context) {
	exchangeID := c.Param("id")
	health, err := h.healthRepo.GetExchangeHealth(c.Request.Context(), exchangeID)
	if err != nil {
		h.logger.Error("Failed to get exchange health",
			zap.String("exchange_id", exchangeID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve health status"})
		return
	}

	c.JSON(http.StatusOK, health)
}

// ListExchangeHealth returns health status for all exchanges
func (h *HTTPHandler) ListExchangeHealth(c *gin.Context) {
	healthRecords, err := h.healthRepo.GetAllExchangeHealth(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to list exchange health", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve health status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"health_records": healthRecords,
		"count":          len(healthRecords),
	})
}

// ThrottleExchange initiates throttling for an exchange
func (h *HTTPHandler) ThrottleExchange(c *gin.Context) {
	exchangeID := c.Param("id")

	var req struct {
		TargetRatePct float64 `json:"target_rate_percent"`
		DurationSecs  int     `json:"duration_secs"`
		Reason        string  `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	cmd := domain.NewThrottleCommand(
		exchangeID,
		domain.ThrottleActionLimit,
		req.Reason,
		req.TargetRatePct,
		req.DurationSecs,
	)

	// In a real implementation, this would publish to the throttle command publisher
	h.logger.Info("Throttle command issued via API",
		zap.String("exchange_id", exchangeID),
		zap.Float64("target_rate_pct", req.TargetRatePct),
		zap.Int("duration_secs", req.DurationSecs),
	)

	c.JSON(http.StatusOK, gin.H{
		"message":      "Throttle command issued",
		"exchange_id":  exchangeID,
		"command":      cmd,
	})
}

// GetVolumeStats returns volume statistics
func (h *HTTPHandler) GetVolumeStats(c *gin.Context) {
	// This would integrate with the analytics engine
	c.JSON(http.StatusOK, gin.H{
		"message": "Volume statistics endpoint - requires analytics integration",
	})
}

// GetTradeCount returns trade count
func (h *HTTPHandler) GetTradeCount(c *gin.Context) {
	// This would integrate with the analytics engine
	c.JSON(http.StatusOK, gin.H{
		"message": "Trade count endpoint - requires analytics integration",
	})
}

// GetMetrics returns Prometheus metrics
func (h *HTTPHandler) GetMetrics(c *gin.Context) {
	c.String(http.StatusOK, "# Prometheus metrics endpoint\n# Implementation requires Prometheus client library")
}

// parseAlertFilter parses query parameters into an AlertFilter
func (h *HTTPHandler) parseAlertFilter(c *gin.Context) ports.AlertFilter {
	filter := ports.AlertFilter{}

	if exchangeID := c.Query("exchange_id"); exchangeID != "" {
		filter.ExchangeID = exchangeID
	}
	if alertType := c.Query("alert_type"); alertType != "" {
		filter.AlertType = domain.AlertType(alertType)
	}
	if severity := c.Query("severity"); severity != "" {
		filter.Severity = domain.AlertSeverity(severity)
	}
	if status := c.Query("status"); status != "" {
		filter.Status = domain.AlertStatus(status)
	}
	if tradingPair := c.Query("trading_pair"); tradingPair != "" {
		filter.TradingPair = tradingPair
	}
	if userID := c.Query("user_id"); userID != "" {
		filter.UserID = userID
	}
	if limitStr := c.DefaultQuery("limit", "100"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = limit
		}
	}
	if offsetStr := c.DefaultQuery("offset", "0"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = offset
		}
	}

	return filter
}
