package ports

import (
	"context"

	"health-monitor/internal/core/domain"
)

// Logger defines the logging interface
type Logger interface {
	Info(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Debug(msg string, keysAndValues ...interface{})
}

// HealthMonitorRepository defines the interface for health data persistence
type HealthMonitorRepository interface {
	// Service registration
	RegisterService(ctx context.Context, service *domain.RegisteredService) error
	UnregisterService(ctx context.Context, serviceName string) error
	GetRegisteredServices(ctx context.Context) ([]*domain.RegisteredService, error)

	// Health status
	UpdateServiceHealth(ctx context.Context, health *domain.ServiceHealth) error
	GetServiceHealth(ctx context.Context, serviceName string) (*domain.ServiceHealth, error)
	GetAllServiceHealth(ctx context.Context) ([]*domain.ServiceHealth, error)

	// History
	RecordHistory(ctx context.Context, entry *domain.HistoryEntry) error
	GetHistory(ctx context.Context, request domain.HistoryRequest) ([]*domain.HistoryEntry, error)

	// Alerts
	CreateAlert(ctx context.Context, alert *domain.Alert) error
	UpdateAlert(ctx context.Context, alert *domain.Alert) error
	GetActiveAlerts(ctx context.Context) ([]*domain.Alert, error)
	GetAlertsByService(ctx context.Context, serviceName string) ([]*domain.Alert, error)

	// Alert rules
	CreateAlertRule(ctx context.Context, rule *domain.AlertRule) error
	GetAlertRules(ctx context.Context) ([]*domain.AlertRule, error)
	DeleteAlertRule(ctx context.Context, ruleID string) error

	// Health
	HealthCheck(ctx context.Context) error
}

// HealthMonitorService defines the interface for health monitoring business logic
type HealthMonitorService interface {
	// Service registration
	RegisterService(ctx context.Context, request domain.ServiceRegistration) (*domain.RegisteredService, error)
	UnregisterService(ctx context.Context, serviceName string) error
	GetRegisteredServices(ctx context.Context) ([]*domain.RegisteredService, error)

	// Heartbeat processing
	ProcessHeartbeat(ctx context.Context, request domain.HeartbeatRequest) error

	// Health queries
	GetServiceHealth(ctx context.Context, serviceName string) (*domain.ServiceHealth, error)
	GetDashboardSummary(ctx context.Context) (*domain.DashboardSummary, error)
	GetAllServiceHealth(ctx context.Context) ([]*domain.ServiceHealth, error)

	// History
	GetServiceHistory(ctx context.Context, request domain.HistoryRequest) ([]*domain.HistoryEntry, error)

	// Alerts
	GetActiveAlerts(ctx context.Context) ([]*domain.Alert, error)
	CreateAlertRule(ctx context.Context, rule *domain.AlertRule) error
	GetAlertRules(ctx context.Context) ([]*domain.AlertRule, error)

	// Active monitoring
	StartActiveMonitoring(ctx context.Context) error
	StopActiveMonitoring(ctx context.Context) error

	// Health
	HealthCheck(ctx context.Context) (bool, error)
}

// HealthChecker defines the interface for active health checking
type HealthChecker interface {
	Check(ctx context.Context, config domain.HealthCheckConfig) (*domain.ServiceHealth, error)
	Start(ctx context.Context, interval int) error
	Stop()
}

// MessagingClient defines the interface for event publishing
type MessagingClient interface {
	PublishHealthStatus(ctx context.Context, health *domain.ServiceHealth) error
	PublishAlert(ctx context.Context, alert *domain.Alert) error
	PublishMetric(ctx context.Context, serviceName string, metric string, value float64) error
	Close() error
}

// TimeSeriesClient defines the interface for time series data storage
type TimeSeriesClient interface {
	WritePoint(ctx context.Context, measurement string, tags map[string]string, fields map[string]interface{}, timestamp time.Time) error
	Query(ctx context.Context, query string) ([]map[string]interface{}, error)
	Close() error
}
