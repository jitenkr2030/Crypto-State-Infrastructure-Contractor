package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"health-monitor/internal/core/domain"
	"health-monitor/internal/core/ports"
)

// PostgresHealthRepository implements ports.HealthMonitorRepository for PostgreSQL
type PostgresHealthRepository struct {
	db     *sql.DB
	logger ports.Logger
}

// NewPostgresHealthRepository creates a new PostgresHealthRepository
func NewPostgresHealthRepository(db *sql.DB, logger ports.Logger) *PostgresHealthRepository {
	return &PostgresHealthRepository{
		db:     db,
		logger: logger,
	}
}

// RegisterService registers a new service
func (r *PostgresHealthRepository) RegisterService(ctx context.Context, service *domain.RegisteredService) error {
	query := `
		INSERT INTO registered_services (
			id, service_name, display_name, description, endpoints, tags,
			metadata, check_url, check_interval, is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	endpointsJSON, _ := json.Marshal(service.Endpoints)
	tagsJSON, _ := json.Marshal(service.Tags)
	metadataJSON, _ := json.Marshal(service.Metadata)

	_, err := r.db.ExecContext(ctx, query,
		service.ID,
		service.ServiceName,
		service.DisplayName,
		service.Description,
		endpointsJSON,
		tagsJSON,
		metadataJSON,
		service.CheckURL,
		service.CheckInterval,
		service.IsActive,
		service.CreatedAt,
		service.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("Failed to register service", "error", err, "name", service.ServiceName)
		return fmt.Errorf("failed to register service: %w", err)
	}

	return nil
}

// UnregisterService unregisters a service
func (r *PostgresHealthRepository) UnregisterService(ctx context.Context, serviceName string) error {
	query := `DELETE FROM registered_services WHERE service_name = $1`

	_, err := r.db.ExecContext(ctx, query, serviceName)
	if err != nil {
		r.logger.Error("Failed to unregister service", "error", err, "name", serviceName)
		return fmt.Errorf("failed to unregister service: %w", err)
	}

	return nil
}

// GetRegisteredServices returns all registered services
func (r *PostgresHealthRepository) GetRegisteredServices(ctx context.Context) ([]*domain.RegisteredService, error) {
	query := `
		SELECT id, service_name, display_name, description, endpoints, tags,
			   metadata, check_url, check_interval, is_active, created_at, updated_at
		FROM registered_services
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		r.logger.Error("Failed to get registered services", "error", err)
		return nil, fmt.Errorf("failed to get registered services: %w", err)
	}
	defer rows.Close()

	var services []*domain.RegisteredService
	for rows.Next() {
		service := &domain.RegisteredService{}
		var endpointsJSON, tagsJSON, metadataJSON []byte

		err := rows.Scan(
			&service.ID,
			&service.ServiceName,
			&service.DisplayName,
			&service.Description,
			&endpointsJSON,
			&tagsJSON,
			&metadataJSON,
			&service.CheckURL,
			&service.CheckInterval,
			&service.IsActive,
			&service.CreatedAt,
			&service.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan service", "error", err)
			continue
		}

		json.Unmarshal(endpointsJSON, &service.Endpoints)
		json.Unmarshal(tagsJSON, &service.Tags)
		json.Unmarshal(metadataJSON, &service.Metadata)
		services = append(services, service)
	}

	return services, nil
}

// UpdateServiceHealth updates the health status of a service
func (r *PostgresHealthRepository) UpdateServiceHealth(ctx context.Context, health *domain.ServiceHealth) error {
	query := `
		INSERT INTO service_health (
			service_name, status, last_heartbeat, metrics, details,
			error_message, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (service_name) DO UPDATE SET
			status = EXCLUDED.status,
			last_heartbeat = EXCLUDED.last_heartbeat,
			metrics = EXCLUDED.metrics,
			details = EXCLUDED.details,
			error_message = EXCLUDED.error_message,
			updated_at = EXCLUDED.updated_at
	`

	metricsJSON, _ := json.Marshal(health.Metrics)
	detailsJSON, _ := json.Marshal(health.Details)

	_, err := r.db.ExecContext(ctx, query,
		health.ServiceName,
		health.Status,
		health.LastHeartbeat,
		metricsJSON,
		detailsJSON,
		health.ErrorMessage,
		time.Now().UTC(),
	)

	if err != nil {
		r.logger.Error("Failed to update service health", "error", err, "name", health.ServiceName)
		return fmt.Errorf("failed to update service health: %w", err)
	}

	return nil
}

// GetServiceHealth returns the health status of a service
func (r *PostgresHealthRepository) GetServiceHealth(ctx context.Context, serviceName string) (*domain.ServiceHealth, error) {
	query := `
		SELECT service_name, status, last_heartbeat, metrics, details,
			   error_message, registered_at, updated_at
		FROM service_health
		WHERE service_name = $1
	`

	health := &domain.ServiceHealth{}
	var metricsJSON, detailsJSON []byte
	var registeredAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, serviceName).Scan(
		&health.ServiceName,
		&health.Status,
		&health.LastHeartbeat,
		&metricsJSON,
		&detailsJSON,
		&health.ErrorMessage,
		&registeredAt,
		&health.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("service not found")
	}
	if err != nil {
		r.logger.Error("Failed to get service health", "error", err, "name", serviceName)
		return nil, fmt.Errorf("failed to get service health: %w", err)
	}

	if registeredAt.Valid {
		health.RegisteredAt = registeredAt.Time
	}

	json.Unmarshal(metricsJSON, &health.Metrics)
	json.Unmarshal(detailsJSON, &health.Details)

	return health, nil
}

// GetAllServiceHealth returns health status for all services
func (r *PostgresHealthRepository) GetAllServiceHealth(ctx context.Context) ([]*domain.ServiceHealth, error) {
	query := `
		SELECT service_name, status, last_heartbeat, metrics, details,
			   error_message, registered_at, updated_at
		FROM service_health
		ORDER BY last_heartbeat DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		r.logger.Error("Failed to get all service health", "error", err)
		return nil, fmt.Errorf("failed to get service health: %w", err)
	}
	defer rows.Close()

	var healthList []*domain.ServiceHealth
	for rows.Next() {
		health := &domain.ServiceHealth{}
		var metricsJSON, detailsJSON []byte
		var registeredAt sql.NullTime

		err := rows.Scan(
			&health.ServiceName,
			&health.Status,
			&health.LastHeartbeat,
			&metricsJSON,
			&detailsJSON,
			&health.ErrorMessage,
			&registeredAt,
			&health.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan service health", "error", err)
			continue
		}

		if registeredAt.Valid {
			health.RegisteredAt = registeredAt.Time
		}

		json.Unmarshal(metricsJSON, &health.Metrics)
		json.Unmarshal(detailsJSON, &health.Details)
		healthList = append(healthList, health)
	}

	return healthList, nil
}

// RecordHistory records a historical health entry
func (r *PostgresHealthRepository) RecordHistory(ctx context.Context, entry *domain.HistoryEntry) error {
	query := `
		INSERT INTO health_history (service_name, status, metrics, recorded_at)
		VALUES ($1, $2, $3, $4)
	`

	metricsJSON, _ := json.Marshal(entry.Metrics)

	_, err := r.db.ExecContext(ctx, query,
		entry.ServiceName,
		entry.Status,
		metricsJSON,
		entry.RecordedAt,
	)

	if err != nil {
		r.logger.Error("Failed to record history", "error", err, "name", entry.ServiceName)
		return fmt.Errorf("failed to record history: %w", err)
	}

	return nil
}

// GetHistory returns historical health data
func (r *PostgresHealthRepository) GetHistory(ctx context.Context, request domain.HistoryRequest) ([]*domain.HistoryEntry, error) {
	query := `
		SELECT service_name, status, metrics, recorded_at
		FROM health_history
		WHERE service_name = $1
		AND recorded_at >= $2
		AND recorded_at <= $3
		ORDER BY recorded_at DESC
		LIMIT $4
	`

	startTime := time.Now().UTC().Add(-24 * time.Hour)
	endTime := time.Now().UTC()
	limit := 1000

	if request.StartTime != nil {
		startTime = *request.StartTime
	}
	if request.EndTime != nil {
		endTime = *request.EndTime
	}
	if request.Limit > 0 {
		limit = request.Limit
	}

	rows, err := r.db.QueryContext(ctx, query, request.ServiceName, startTime, endTime, limit)
	if err != nil {
		r.logger.Error("Failed to get history", "error", err, "name", request.ServiceName)
		return nil, fmt.Errorf("failed to get history: %w", err)
	}
	defer rows.Close()

	var entries []*domain.HistoryEntry
	for rows.Next() {
		entry := &domain.HistoryEntry{}
		var metricsJSON []byte

		err := rows.Scan(
			&entry.ServiceName,
			&entry.Status,
			&metricsJSON,
			&entry.RecordedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan history entry", "error", err)
			continue
		}

		json.Unmarshal(metricsJSON, &entry.Metrics)
		entries = append(entries, entry)
	}

	return entries, nil
}

// CreateAlert creates a new alert
func (r *PostgresHealthRepository) CreateAlert(ctx context.Context, alert *domain.Alert) error {
	query := `
		INSERT INTO alerts (
			id, service_name, severity, message, details,
			triggered_at, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.ExecContext(ctx, query,
		alert.ID,
		alert.ServiceName,
		alert.Severity,
		alert.Message,
		alert.Details,
		alert.TriggeredAt,
		alert.Status,
	)

	if err != nil {
		r.logger.Error("Failed to create alert", "error", err, "service", alert.ServiceName)
		return fmt.Errorf("failed to create alert: %w", err)
	}

	return nil
}

// UpdateAlert updates an alert
func (r *PostgresHealthRepository) UpdateAlert(ctx context.Context, alert *domain.Alert) error {
	query := `
		UPDATE alerts
		SET status = $1, resolved_at = $2
		WHERE id = $3
	`

	_, err := r.db.ExecContext(ctx, query,
		alert.Status,
		alert.ResolvedAt,
		alert.ID,
	)

	if err != nil {
		r.logger.Error("Failed to update alert", "error", err, "id", alert.ID)
		return fmt.Errorf("failed to update alert: %w", err)
	}

	return nil
}

// GetActiveAlerts returns all active alerts
func (r *PostgresHealthRepository) GetActiveAlerts(ctx context.Context) ([]*domain.Alert, error) {
	query := `
		SELECT id, service_name, severity, message, details,
			   triggered_at, resolved_at, status
		FROM alerts
		WHERE status IN ('active', 'acknowledged')
		ORDER BY triggered_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		r.logger.Error("Failed to get active alerts", "error", err)
		return nil, fmt.Errorf("failed to get active alerts: %w", err)
	}
	defer rows.Close()

	var alerts []*domain.Alert
	for rows.Next() {
		alert := &domain.Alert{}
		var resolvedAt sql.NullTime

		err := rows.Scan(
			&alert.ID,
			&alert.ServiceName,
			&alert.Severity,
			&alert.Message,
			&alert.Details,
			&alert.TriggeredAt,
			&resolvedAt,
			&alert.Status,
		)
		if err != nil {
			r.logger.Error("Failed to scan alert", "error", err)
			continue
		}

		if resolvedAt.Valid {
			alert.ResolvedAt = &resolvedAt.Time
		}

		alerts = append(alerts, alert)
	}

	return alerts, nil
}

// GetAlertsByService returns alerts for a specific service
func (r *PostgresHealthRepository) GetAlertsByService(ctx context.Context, serviceName string) ([]*domain.Alert, error) {
	query := `
		SELECT id, service_name, severity, message, details,
			   triggered_at, resolved_at, status
		FROM alerts
		WHERE service_name = $1
		ORDER BY triggered_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, serviceName)
	if err != nil {
		r.logger.Error("Failed to get alerts by service", "error", err, "name", serviceName)
		return nil, fmt.Errorf("failed to get alerts: %w", err)
	}
	defer rows.Close()

	var alerts []*domain.Alert
	for rows.Next() {
		alert := &domain.Alert{}
		var resolvedAt sql.NullTime

		err := rows.Scan(
			&alert.ID,
			&alert.ServiceName,
			&alert.Severity,
			&alert.Message,
			&alert.Details,
			&alert.TriggeredAt,
			&resolvedAt,
			&alert.Status,
		)
		if err != nil {
			r.logger.Error("Failed to scan alert", "error", err)
			continue
		}

		if resolvedAt.Valid {
			alert.ResolvedAt = &resolvedAt.Time
		}

		alerts = append(alerts, alert)
	}

	return alerts, nil
}

// CreateAlertRule creates a new alert rule
func (r *PostgresHealthRepository) CreateAlertRule(ctx context.Context, rule *domain.AlertRule) error {
	query := `
		INSERT INTO alert_rules (
			id, service_name, condition, severity, enabled,
			cooldown, notify_via, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	notifyViaJSON, _ := json.Marshal(rule.NotifyVia)

	_, err := r.db.ExecContext(ctx, query,
		rule.ID,
		rule.ServiceName,
		rule.Condition,
		rule.Severity,
		rule.Enabled,
		rule.Cooldown,
		notifyViaJSON,
		rule.CreatedAt,
		rule.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("Failed to create alert rule", "error", err)
		return fmt.Errorf("failed to create alert rule: %w", err)
	}

	return nil
}

// GetAlertRules returns all alert rules
func (r *PostgresHealthRepository) GetAlertRules(ctx context.Context) ([]*domain.AlertRule, error) {
	query := `
		SELECT id, service_name, condition, severity, enabled,
			   cooldown, notify_via, created_at, updated_at
		FROM alert_rules
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		r.logger.Error("Failed to get alert rules", "error", err)
		return nil, fmt.Errorf("failed to get alert rules: %w", err)
	}
	defer rows.Close()

	var rules []*domain.AlertRule
	for rows.Next() {
		rule := &domain.AlertRule{}
		var notifyViaJSON []byte

		err := rows.Scan(
			&rule.ID,
			&rule.ServiceName,
			&rule.Condition,
			&rule.Severity,
			&rule.Enabled,
			&rule.Cooldown,
			&notifyViaJSON,
			&rule.CreatedAt,
			&rule.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan alert rule", "error", err)
			continue
		}

		json.Unmarshal(notifyViaJSON, &rule.NotifyVia)
		rules = append(rules, rule)
	}

	return rules, nil
}

// DeleteAlertRule deletes an alert rule
func (r *PostgresHealthRepository) DeleteAlertRule(ctx context.Context, ruleID string) error {
	query := `DELETE FROM alert_rules WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, ruleID)
	if err != nil {
		r.logger.Error("Failed to delete alert rule", "error", err, "id", ruleID)
		return fmt.Errorf("failed to delete alert rule: %w", err)
	}

	return nil
}

// HealthCheck checks database connectivity
func (r *PostgresHealthRepository) HealthCheck(ctx context.Context) error {
	return r.db.PingContext(ctx)
}
