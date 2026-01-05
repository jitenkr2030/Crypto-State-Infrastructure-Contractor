package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"health-monitor/internal/core/domain"
	"health-monitor/internal/core/ports"
)

// HealthMonitorHandler handles HTTP requests for health monitoring
type HealthMonitorHandler struct {
	healthService ports.HealthMonitorService
	logger        ports.Logger
}

// NewHealthMonitorHandler creates a new HealthMonitorHandler instance
func NewHealthMonitorHandler(healthService ports.HealthMonitorService, logger ports.Logger) *HealthMonitorHandler {
	return &HealthMonitorHandler{
		healthService: healthService,
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

// RegisterRoutes registers all health monitor handler routes
func (h *HealthMonitorHandler) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api/v1/monitor")
	{
		// Service registration
		api.POST("/services", h.RegisterService)
		api.GET("/services", h.GetRegisteredServices)
		api.DELETE("/services/:name", h.UnregisterService)

		// Heartbeat
		api.POST("/heartbeat", h.ProcessHeartbeat)

		// Health status
		api.GET("/health", h.GetAllHealth)
		api.GET("/health/:name", h.GetServiceHealth)

		// Dashboard
		api.GET("/dashboard", h.GetDashboard)

		// History
		api.POST("/history", h.GetServiceHistory)

		// Alerts
		api.GET("/alerts", h.GetActiveAlerts)
		api.POST("/alerts/rules", h.CreateAlertRule)
		api.GET("/alerts/rules", h.GetAlertRules)

		// Active monitoring
		api.POST("/monitoring/start", h.StartMonitoring)
		api.POST("/monitoring/stop", h.StopMonitoring)
	}

	// Prometheus metrics endpoint
	router.GET("/metrics", h.GetMetrics)
}

// RegisterService registers a new service for monitoring
func (h *HealthMonitorHandler) RegisterService(c *gin.Context) {
	var request domain.ServiceRegistration
	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Error("Failed to parse registration request", "error", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body",
			Details: map[string]string{
				"validation": err.Error(),
			},
		})
		return
	}

	service, err := h.healthService.RegisterService(c.Request.Context(), request)
	if err != nil {
		if err.Error() == "service already registered" {
			c.JSON(http.StatusConflict, ErrorResponse{
				Error: "Service already registered",
			})
			return
		}
		h.logger.Error("Failed to register service", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to register service",
		})
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse{
		Message: "Service registered successfully",
		Data:    service,
	})
}

// GetRegisteredServices returns all registered services
func (h *HealthMonitorHandler) GetRegisteredServices(c *gin.Context) {
	services, err := h.healthService.GetRegisteredServices(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get registered services", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to get registered services",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Services retrieved",
		Data:    services,
	})
}

// UnregisterService unregisters a service
func (h *HealthMonitorHandler) UnregisterService(c *gin.Context) {
	serviceName := c.Param("name")
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Service name is required",
		})
		return
	}

	err := h.healthService.UnregisterService(c.Request.Context(), serviceName)
	if err != nil {
		h.logger.Error("Failed to unregister service", "error", err, "name", serviceName)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to unregister service",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Service unregistered successfully",
	})
}

// ProcessHeartbeat processes a heartbeat signal
func (h *HealthMonitorHandler) ProcessHeartbeat(c *gin.Context) {
	var request domain.HeartbeatRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Error("Failed to parse heartbeat request", "error", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body",
			Details: map[string]string{
				"validation": err.Error(),
			},
		})
		return
	}

	if err := h.healthService.ProcessHeartbeat(c.Request.Context(), request); err != nil {
		h.logger.Error("Failed to process heartbeat", "error", err, "service", request.ServiceName)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to process heartbeat",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Heartbeat processed",
	})
}

// GetServiceHealth returns the health status of a specific service
func (h *HealthMonitorHandler) GetServiceHealth(c *gin.Context) {
	serviceName := c.Param("name")
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Service name is required",
		})
		return
	}

	health, err := h.healthService.GetServiceHealth(c.Request.Context(), serviceName)
	if err != nil {
		h.logger.Error("Failed to get service health", "error", err, "name", serviceName)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to get service health",
		})
		return
	}

	c.JSON(http.StatusOK, health)
}

// GetAllHealth returns health status for all services
func (h *HealthMonitorHandler) GetAllHealth(c *gin.Context) {
	services, err := h.healthService.GetAllServiceHealth(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get all service health", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to get service health",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Health status retrieved",
		Data:    services,
	})
}

// GetDashboard returns the dashboard summary
func (h *HealthMonitorHandler) GetDashboard(c *gin.Context) {
	summary, err := h.healthService.GetDashboardSummary(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get dashboard", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to get dashboard",
		})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// GetServiceHistory returns historical health data
func (h *HealthMonitorHandler) GetServiceHistory(c *gin.Context) {
	var request domain.HistoryRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Error("Failed to parse history request", "error", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body",
			Details: map[string]string{
				"validation": err.Error(),
			},
		})
		return
	}

	entries, err := h.healthService.GetServiceHistory(c.Request.Context(), request)
	if err != nil {
		h.logger.Error("Failed to get service history", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to get service history",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "History retrieved",
		Data:    entries,
	})
}

// GetActiveAlerts returns all active alerts
func (h *HealthMonitorHandler) GetActiveAlerts(c *gin.Context) {
	alerts, err := h.healthService.GetActiveAlerts(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get active alerts", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to get alerts",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Alerts retrieved",
		Data:    alerts,
	})
}

// CreateAlertRule creates a new alert rule
func (h *HealthMonitorHandler) CreateAlertRule(c *gin.Context) {
	var rule domain.AlertRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		h.logger.Error("Failed to parse alert rule", "error", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body",
			Details: map[string]string{
				"validation": err.Error(),
			},
		})
		return
	}

	if err := h.healthService.CreateAlertRule(c.Request.Context(), &rule); err != nil {
		h.logger.Error("Failed to create alert rule", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to create alert rule",
		})
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse{
		Message: "Alert rule created",
		Data:    rule,
	})
}

// GetAlertRules returns all alert rules
func (h *HealthMonitorHandler) GetAlertRules(c *gin.Context) {
	rules, err := h.healthService.GetAlertRules(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get alert rules", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to get alert rules",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Alert rules retrieved",
		Data:    rules,
	})
}

// StartMonitoring starts active monitoring
func (h *HealthMonitorHandler) StartMonitoring(c *gin.Context) {
	if err := h.healthService.StartActiveMonitoring(c.Request.Context()); err != nil {
		h.logger.Error("Failed to start monitoring", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to start monitoring",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Active monitoring started",
	})
}

// StopMonitoring stops active monitoring
func (h *HealthMonitorHandler) StopMonitoring(c *gin.Context) {
	if err := h.healthService.StopActiveMonitoring(c.Request.Context()); err != nil {
		h.logger.Error("Failed to stop monitoring", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to stop monitoring",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Active monitoring stopped",
	})
}

// GetMetrics returns Prometheus-compatible metrics
func (h *HealthMonitorHandler) GetMetrics(c *gin.Context) {
	summary, err := h.healthService.GetDashboardSummary(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get metrics", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to get metrics",
		})
		return
	}

	// Generate Prometheus metrics format
	metrics := "# HELP csic_services_total Total number of services\n"
	metrics += "# TYPE csic_services_total gauge\n"
	metrics += fmt.Sprintf("csic_services_total %d\n\n", summary.TotalServices)

	metrics += "# HELP csic_services_healthy Number of healthy services\n"
	metrics += "# TYPE csic_services_healthy gauge\n"
	metrics += fmt.Sprintf("csic_services_healthy %d\n\n", summary.HealthyServices)

	metrics += "# HELP csic_services_degraded Number of degraded services\n"
	metrics += "# TYPE csic_services_degraded gauge\n"
	metrics += fmt.Sprintf("csic_services_degraded %d\n\n", summary.DegradedServices)

	metrics += "# HELP csic_services_down Number of down services\n"
	metrics += "# TYPE csic_services_down gauge\n"
	metrics += fmt.Sprintf("csic_services_down %d\n\n", summary.DownServices)

	metrics += "# HELP csic_uptime_percentage System uptime percentage\n"
	metrics += "# TYPE csic_uptime_percentage gauge\n"
	metrics += fmt.Sprintf("csic_uptime_percentage %.2f\n\n", summary.SystemMetrics.UptimePercentage)

	// Per-service metrics
	for _, service := range summary.Services {
		metrics += fmt.Sprintf("# HELP csic_service_status Service %s status\n", service.ServiceName)
		metrics += fmt.Sprintf("# TYPE csic_service_status gauge\n")
		statusValue := 0
		switch service.Status {
		case domain.StatusHealthy:
			statusValue = 1
		case domain.StatusDegraded:
			statusValue = 0.5
		case domain.StatusDown:
			statusValue = 0
		}
		metrics += fmt.Sprintf("csic_service_status{service=\"%s\"} %.1f\n", service.ServiceName, statusValue)
	}

	c.String(http.StatusOK, metrics)
}

// HealthCheck handles health check endpoint
func (h *HealthMonitorHandler) HealthCheck(c *gin.Context) {
	healthy, err := h.healthService.HealthCheck(c.Request.Context())
	if err != nil || !healthy {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "unhealthy",
			"service": "health-monitor",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "health-monitor",
		"time":    time.Now().UTC(),
	})
}
