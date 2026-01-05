package domain

import (
	"time"
)

// ServiceStatus represents the health status of a service
type ServiceStatus string

const (
	StatusHealthy   ServiceStatus = "healthy"
	StatusDegraded  ServiceStatus = "degraded"
	StatusDown      ServiceStatus = "down"
	StatusUnknown   ServiceStatus = "unknown"
	StatusChecking  ServiceStatus = "checking"
)

// ServiceHealth represents the health information for a service
type ServiceHealth struct {
	ServiceName   string            `json:"service_name"`
	Status        ServiceStatus     `json:"status"`
	LastHeartbeat time.Time         `json:"last_heartbeat"`
	Uptime        time.Duration     `json:"uptime"`
	Metrics       map[string]float64 `json:"metrics,omitempty"`
	Details       map[string]string `json:"details,omitempty"`
	ErrorMessage  string            `json:"error_message,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	RegisteredAt  time.Time         `json:"registered_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

// HeartbeatRequest represents a heartbeat signal from a service
type HeartbeatRequest struct {
	ServiceName string            `json:"service_name" binding:"required"`
	Status      ServiceStatus     `json:"status" binding:"required"`
	Metrics     map[string]float64 `json:"metrics,omitempty"`
	Details     map[string]string `json:"details,omitempty"`
	Version     string            `json:"version,omitempty"`
	Environment string            `json:"environment,omitempty"`
}

// ServiceRegistration represents a service registration request
type ServiceRegistration struct {
	ServiceName string   `json:"service_name" binding:"required"`
	DisplayName string   `json:"display_name"`
	Description string   `json:"description"`
	Endpoints   []string `json:"endpoints"`
	Tags        []string `json:"tags"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// RegisteredService represents a registered service in the monitor
type RegisteredService struct {
	ID          string            `json:"id"`
	ServiceName string            `json:"service_name"`
	DisplayName string            `json:"display_name"`
	Description string            `json:"description"`
	Endpoints   []string          `json:"endpoints"`
	Tags        []string          `json:"tags"`
	Metadata    map[string]string `json:"metadata"`
	CheckURL    string            `json:"check_url"`
	CheckInterval int             `json:"check_interval"` // in seconds
	IsActive    bool              `json:"is_active"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// DashboardSummary provides an aggregated view of system health
type DashboardSummary struct {
	TotalServices     int                  `json:"total_services"`
	HealthyServices   int                  `json:"healthy_services"`
	DegradedServices  int                  `json:"degraded_services"`
	DownServices      int                  `json:"down_services"`
	OverallStatus     ServiceStatus        `json:"overall_status"`
	Services          []*ServiceHealth     `json:"services"`
	SystemMetrics     SystemMetrics        `json:"system_metrics"`
	LastUpdated       time.Time            `json:"last_updated"`
}

// SystemMetrics contains system-level metrics
type SystemMetrics struct {
	TotalRequests    int64   `json:"total_requests"`
	AverageLatencyMs float64 `json:"average_latency_ms"`
	ErrorsPerMin     int     `json:"errors_per_minute"`
	UptimePercentage float64 `json:"uptime_percentage"`
	MemoryUsageMB    float64 `json:"memory_usage_mb"`
	CPUUsagePercent  float64 `json:"cpu_usage_percent"`
}

// Alert represents a health alert
type Alert struct {
	ID          string        `json:"id"`
	ServiceName string        `json:"service_name"`
	Severity    AlertSeverity `json:"severity"` // critical, warning, info
	Message     string        `json:"message"`
	Details     string        `json:"details,omitempty"`
	TriggeredAt time.Time     `json:"triggered_at"`
	ResolvedAt  *time.Time    `json:"resolved_at,omitempty"`
	Status      AlertStatus   `json:"status"` // active, acknowledged, resolved
}

// AlertSeverity defines alert severity levels
type AlertSeverity string

const (
	AlertCritical AlertSeverity = "critical"
	AlertWarning  AlertSeverity = "warning"
	AlertInfo     AlertSeverity = "info"
)

// AlertStatus defines alert status
type AlertStatus string

const (
	AlertActive       AlertStatus = "active"
	AlertAcknowledged AlertStatus = "acknowledged"
	AlertResolved     AlertStatus = "resolved"
)

// AlertRule defines conditions for triggering alerts
type AlertRule struct {
	ID           string        `json:"id"`
	ServiceName  string        `json:"service_name"`
	Condition    string        `json:"condition"` // e.g., "status == down", "latency > 1000ms"
	Severity     AlertSeverity `json:"severity"`
	Enabled      bool          `json:"enabled"`
	Cooldown     int           `json:"cooldown"` // in seconds
	NotifyVia    []string      `json:"notify_via"` // webhook, email, slack
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

// HealthCheckConfig defines how to check a service's health
type HealthCheckConfig struct {
	ServiceName string `json:"service_name"`
	CheckURL    string `json:"check_url"`
	Method      string `json:"method"` // GET, POST
	Timeout     int    `json:"timeout"` // in milliseconds
	Headers     map[string]string `json:"headers,omitempty"`
	ExpectedStatus int `json:"expected_status"` // expected HTTP status
}

// HistoryEntry represents a historical health record
type HistoryEntry struct {
	ServiceName string        `json:"service_name"`
	Status      ServiceStatus `json:"status"`
	Metrics     map[string]float64 `json:"metrics,omitempty"`
	RecordedAt  time.Time     `json:"recorded_at"`
}

// HistoryRequest represents a request for historical health data
type HistoryRequest struct {
	ServiceName string     `json:"service_name"`
	StartTime   *time.Time `json:"start_time"`
	EndTime     *time.Time `json:"end_time"`
	Interval    string     `json:"interval"` // minute, hour, day
	Limit       int        `json:"limit"`
}
