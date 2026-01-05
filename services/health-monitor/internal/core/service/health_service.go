package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"health-monitor/internal/core/domain"
	"health-monitor/internal/core/ports"
)

//go:generate mockgen -destination=../../../mocks/mock_health_service.go -package=mocks . HealthMonitorService

var (
	// ErrServiceNotFound is returned when a service is not found
	ErrServiceNotFound = errors.New("service not found")

	// ErrServiceAlreadyRegistered is returned when a service is already registered
	ErrServiceAlreadyRegistered = errors.New("service already registered")

	// ErrAlertNotFound is returned when an alert is not found
	ErrAlertNotFound = errors.New("alert not found")
)

// HealthMonitorServiceImpl implements the HealthMonitorService interface
type HealthMonitorServiceImpl struct {
	repo       ports.HealthMonitorRepository
	messaging  ports.MessagingClient
	httpClient *http.Client
	logger     ports.Logger
	monitoring bool
	mu         sync.RWMutex
}

// NewHealthMonitorService creates a new HealthMonitorServiceImpl
func NewHealthMonitorService(repo ports.HealthMonitorRepository, messaging ports.MessagingClient, logger ports.Logger) *HealthMonitorServiceImpl {
	return &HealthMonitorServiceImpl{
		repo: repo,
		messaging: messaging,
		logger: logger,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// RegisterService registers a new service for monitoring
func (s *HealthMonitorServiceImpl) RegisterService(ctx context.Context, request domain.ServiceRegistration) (*domain.RegisteredService, error) {
	// Check if service already exists
	services, err := s.repo.GetRegisteredServices(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing services: %w", err)
	}

	for _, svc := range services {
		if svc.ServiceName == request.ServiceName {
			return nil, ErrServiceAlreadyRegistered
		}
	}

	service := &domain.RegisteredService{
		ID:           generateServiceID(),
		ServiceName:  request.ServiceName,
		DisplayName:  request.DisplayName,
		Description:  request.Description,
		Endpoints:    request.Endpoints,
		Tags:         request.Tags,
		Metadata:     request.Metadata,
		CheckURL:     fmt.Sprintf("http://%s/health", request.ServiceName),
		CheckInterval: 30,
		IsActive:     true,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	if err := s.repo.RegisterService(ctx, service); err != nil {
		s.logger.Error("Failed to register service", "error", err, "name", request.ServiceName)
		return nil, fmt.Errorf("failed to register service: %w", err)
	}

	s.logger.Info("Service registered", "name", request.ServiceName)
	return service, nil
}

// UnregisterService unregisters a service from monitoring
func (s *HealthMonitorServiceImpl) UnregisterService(ctx context.Context, serviceName string) error {
	if err := s.repo.UnregisterService(ctx, serviceName); err != nil {
		s.logger.Error("Failed to unregister service", "error", err, "name", serviceName)
		return fmt.Errorf("failed to unregister service: %w", err)
	}

	s.logger.Info("Service unregistered", "name", serviceName)
	return nil
}

// GetRegisteredServices returns all registered services
func (s *HealthMonitorServiceImpl) GetRegisteredServices(ctx context.Context) ([]*domain.RegisteredService, error) {
	return s.repo.GetRegisteredServices(ctx)
}

// ProcessHeartbeat processes a heartbeat signal from a service
func (s *HealthMonitorServiceImpl) ProcessHeartbeat(ctx context.Context, request domain.HeartbeatRequest) error {
	health := &domain.ServiceHealth{
		ServiceName:  request.ServiceName,
		Status:       request.Status,
		LastHeartbeat: time.Now().UTC(),
		Metrics:      request.Metrics,
		Details:      request.Details,
		UpdatedAt:    time.Now().UTC(),
	}

	// Update service health
	if err := s.repo.UpdateServiceHealth(ctx, health); err != nil {
		s.logger.Error("Failed to update service health", "error", err, "name", request.ServiceName)
		return fmt.Errorf("failed to update service health: %w", err)
	}

	// Record history
	history := &domain.HistoryEntry{
		ServiceName: request.ServiceName,
		Status:      request.Status,
		Metrics:     request.Metrics,
		RecordedAt:  time.Now().UTC(),
	}

	if err := s.repo.RecordHistory(ctx, history); err != nil {
		s.logger.Warn("Failed to record history", "error", err)
	}

	// Check for alerts
	s.checkForAlerts(ctx, health)

	// Publish event
	if s.messaging != nil {
		s.messaging.PublishHealthStatus(ctx, health)
	}

	return nil
}

// GetServiceHealth returns the health status of a specific service
func (s *HealthMonitorServiceImpl) GetServiceHealth(ctx context.Context, serviceName string) (*domain.ServiceHealth, error) {
	return s.repo.GetServiceHealth(ctx, serviceName)
}

// GetDashboardSummary returns an aggregated dashboard view
func (s *HealthMonitorServiceImpl) GetDashboardSummary(ctx context.Context) (*domain.DashboardSummary, error) {
	services, err := s.repo.GetAllServiceHealth(ctx)
	if err != nil {
		s.logger.Error("Failed to get service health", "error", err)
		return nil, fmt.Errorf("failed to get dashboard summary: %w", err)
	}

	summary := &domain.DashboardSummary{
		TotalServices: len(services),
		Services:      services,
		SystemMetrics: domain.SystemMetrics{
			UptimePercentage: 100.0,
		},
		LastUpdated: time.Now().UTC(),
	}

	for _, svc := range services {
		switch svc.Status {
		case domain.StatusHealthy:
			summary.HealthyServices++
		case domain.StatusDegraded:
			summary.DegradedServices++
		case domain.StatusDown:
			summary.DownServices++
		}
	}

	// Determine overall status
	if summary.DownServices > 0 {
		summary.OverallStatus = domain.StatusDown
	} else if summary.DegradedServices > 0 {
		summary.OverallStatus = domain.StatusDegraded
	} else {
		summary.OverallStatus = domain.StatusHealthy
	}

	// Calculate uptime percentage
	if summary.TotalServices > 0 {
		summary.SystemMetrics.UptimePercentage = float64(summary.HealthyServices) / float64(summary.TotalServices) * 100
	}

	return summary, nil
}

// GetAllServiceHealth returns health status for all services
func (s *HealthMonitorServiceImpl) GetAllServiceHealth(ctx context.Context) ([]*domain.ServiceHealth, error) {
	return s.repo.GetAllServiceHealth(ctx)
}

// GetServiceHistory returns historical health data
func (s *HealthMonitorServiceImpl) GetServiceHistory(ctx context.Context, request domain.HistoryRequest) ([]*domain.HistoryEntry, error) {
	return s.repo.GetHistory(ctx, request)
}

// GetActiveAlerts returns all active alerts
func (s *HealthMonitorServiceImpl) GetActiveAlerts(ctx context.Context) ([]*domain.Alert, error) {
	return s.repo.GetActiveAlerts(ctx)
}

// CreateAlertRule creates a new alert rule
func (s *HealthMonitorServiceImpl) CreateAlertRule(ctx context.Context, rule *domain.AlertRule) error {
	rule.ID = generateAlertRuleID()
	rule.CreatedAt = time.Now().UTC()
	rule.UpdatedAt = time.Now().UTC()

	if err := s.repo.CreateAlertRule(ctx, rule); err != nil {
		s.logger.Error("Failed to create alert rule", "error", err)
		return fmt.Errorf("failed to create alert rule: %w", err)
	}

	s.logger.Info("Alert rule created", "id", rule.ID, "service", rule.ServiceName)
	return nil
}

// GetAlertRules returns all alert rules
func (s *HealthMonitorServiceImpl) GetAlertRules(ctx context.Context) ([]*domain.AlertRule, error) {
	return s.repo.GetAlertRules(ctx)
}

// StartActiveMonitoring starts active health checking
func (s *HealthMonitorServiceImpl) StartActiveMonitoring(ctx context.Context) error {
	s.mu.Lock()
	if s.monitoring {
		s.mu.Unlock()
		return nil
	}
	s.monitoring = true
	s.mu.Unlock()

	s.logger.Info("Starting active monitoring")

	// Start background checker
	go s.activeMonitorLoop(ctx)

	return nil
}

// StopActiveMonitoring stops active health checking
func (s *HealthMonitorServiceImpl) StopActiveMonitoring(ctx context.Context) error {
	s.mu.Lock()
	s.monitoring = false
	s.mu.Unlock()

	s.logger.Info("Stopping active monitoring")
	return nil
}

// HealthCheck checks the health of the service
func (s *HealthMonitorServiceImpl) HealthCheck(ctx context.Context) (bool, error) {
	if err := s.repo.HealthCheck(ctx); err != nil {
		s.logger.Error("Health check failed", "error", err)
		return false, err
	}
	return true, nil
}

// Helper functions

func generateServiceID() string {
	return fmt.Sprintf("svc_%d", time.Now().UnixNano())
}

func generateAlertRuleID() string {
	return fmt.Sprintf("rule_%d", time.Now().UnixNano())
}

// activeMonitorLoop runs active health checks
func (s *HealthMonitorServiceImpl) activeMonitorLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !s.isMonitoring() {
				return
			}

			services, err := s.repo.GetRegisteredServices(ctx)
			if err != nil {
				s.logger.Error("Failed to get registered services", "error", err)
				continue
			}

			for _, svc := range services {
				if !svc.IsActive {
					continue
				}

				health, err := s.checkServiceHealth(ctx, svc.CheckURL)
				if err != nil {
					s.logger.Warn("Health check failed", "error", err, "service", svc.ServiceName)
					continue
				}

				// Update health in repository
				s.repo.UpdateServiceHealth(ctx, health)
			}
		}
	}
}

func (s *HealthMonitorServiceImpl) isMonitoring() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.monitoring
}

// checkServiceHealth performs an active health check
func (s *HealthMonitorServiceImpl) checkServiceHealth(ctx context.Context, url string) (*domain.ServiceHealth, error) {
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return &domain.ServiceHealth{
			Status:       domain.StatusDown,
			LastHeartbeat: time.Now().UTC(),
			ErrorMessage: err.Error(),
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return &domain.ServiceHealth{
			Status:        domain.StatusHealthy,
			LastHeartbeat: time.Now().UTC(),
		}, nil
	}

	return &domain.ServiceHealth{
		Status:        domain.StatusDegraded,
		LastHeartbeat: time.Now().UTC(),
		ErrorMessage: fmt.Sprintf("HTTP %d", resp.StatusCode),
	}, nil
}

// checkForAlerts checks if any alerts should be triggered
func (s *HealthMonitorServiceImpl) checkForAlerts(ctx context.Context, health *domain.ServiceHealth) {
	rules, err := s.repo.GetAlertRules(ctx)
	if err != nil {
		s.logger.Warn("Failed to get alert rules", "error", err)
		return
	}

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		if rule.ServiceName != "" && rule.ServiceName != health.ServiceName {
			continue
		}

		if s.evaluateAlertCondition(rule.Condition, health) {
			s.triggerAlert(ctx, health, rule)
		}
	}
}

// evaluateAlertCondition evaluates an alert condition
func (s *HealthMonitorServiceImpl) evaluateAlertCondition(condition string, health *domain.ServiceHealth) bool {
	switch condition {
	case "status == down":
		return health.Status == domain.StatusDown
	case "status == degraded":
		return health.Status == domain.StatusDegraded
	case "status != healthy":
		return health.Status != domain.StatusHealthy
	default:
		return false
	}
}

// triggerAlert creates and publishes an alert
func (s *HealthMonitorServiceImpl) triggerAlert(ctx context.Context, health *domain.ServiceHealth, rule *domain.AlertRule) {
	alert := &domain.Alert{
		ID:          fmt.Sprintf("alert_%d", time.Now().UnixNano()),
		ServiceName: health.ServiceName,
		Severity:    rule.Severity,
		Message:     fmt.Sprintf("Alert triggered: %s", rule.Condition),
		TriggeredAt: time.Now().UTC(),
		Status:      domain.AlertActive,
	}

	if err := s.repo.CreateAlert(ctx, alert); err != nil {
		s.logger.Warn("Failed to create alert", "error", err)
		return
	}

	// Publish alert event
	if s.messaging != nil {
		s.messaging.PublishAlert(ctx, alert)
	}

	s.logger.Info("Alert triggered", "service", health.ServiceName, "severity", rule.Severity)
}
