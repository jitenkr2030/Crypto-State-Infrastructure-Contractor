package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/csic/platform/compliance/internal/domain"
	"github.com/csic/platform/compliance/internal/service"
	"go.uber.org/zap"
)

// EventHandler handles Kafka events for the compliance service
type EventHandler struct {
	complianceService *service.ComplianceService
	logger            *zap.Logger
}

// NewEventHandler creates a new event handler
func NewEventHandler(complianceService *service.ComplianceService, logger *zap.Logger) *EventHandler {
	return &EventHandler{
		complianceService: complianceService,
		logger:            logger,
	}
}

// HandleTransaction handles incoming transaction events from Kafka
func (h *EventHandler) HandleTransaction(ctx context.Context, tx *domain.Transaction) error {
	if tx == nil || tx.ID == "" {
		return fmt.Errorf("invalid transaction: nil or empty ID")
	}

	h.logger.Info("Processing transaction event",
		zap.String("transaction_id", tx.ID),
		zap.String("transaction_type", tx.Type),
		zap.Float64("amount", tx.Amount),
		zap.String("source", tx.SourceName),
		zap.String("target", tx.TargetName))

	// Perform compliance check
	result, err := h.complianceService.CheckCompliance(ctx, tx)
	if err != nil {
		h.logger.Error("Compliance check failed",
			zap.Error(err),
			zap.String("transaction_id", tx.ID))
		return fmt.Errorf("compliance check failed: %w", err)
	}

	// Log result
	h.logger.Info("Compliance check completed",
		zap.String("transaction_id", tx.ID),
		zap.String("overall_status", result.OverallStatus),
		zap.Float64("risk_score", result.RiskScore),
		zap.Int("checks_passed", result.Summary.PassedChecks),
		zap.Int("checks_failed", result.Summary.FailedChecks),
		zap.Int("violations", len(result.Violations)))

	// If there are violations, they are already sent to Kafka by the service

	return nil
}

// HandleBatchTransactions handles a batch of transactions
func (h *EventHandler) HandleBatchTransactions(ctx context.Context, transactions []*domain.Transaction) error {
	h.logger.Info("Processing batch of transactions",
		zap.Int("count", len(transactions)))

	for _, tx := range transactions {
		if err := h.HandleTransaction(ctx, tx); err != nil {
			h.logger.Error("Failed to process transaction in batch",
				zap.Error(err),
				zap.String("transaction_id", tx.ID))
			// Continue processing other transactions
			continue
		}
	}

	return nil
}

// ParseTransactionEvent parses a raw Kafka message into a transaction event
func (h *EventHandler) ParseTransactionEvent(data []byte) (*domain.Transaction, error) {
	var tx domain.Transaction
	if err := json.Unmarshal(data, &tx); err != nil {
		return nil, fmt.Errorf("failed to parse transaction event: %w", err)
	}

	// Validate required fields
	if tx.ID == "" {
		return nil, fmt.Errorf("transaction ID is required")
	}

	if tx.Amount <= 0 {
		return nil, fmt.Errorf("transaction amount must be positive")
	}

	return &tx, nil
}

// ViolationEvent represents a violation event published to Kafka
type ViolationEvent struct {
	Violation     domain.Violation `json:"violation"`
	TransactionID string           `json:"transaction_id"`
	RuleID        string           `json:"rule_id"`
	RuleName      string           `json:"rule_name"`
	Timestamp     string           `json:"timestamp"`
	Severity      string           `json:"severity"`
}

// CreateViolationEvent creates a violation event from a violation
func (h *EventHandler) CreateViolationEvent(violation *domain.Violation, txID string) *ViolationEvent {
	return &ViolationEvent{
		Violation:     *violation,
		TransactionID: txID,
		RuleID:        violation.RuleID,
		RuleName:      violation.RuleName,
		Timestamp:     violation.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Severity:      string(violation.Severity),
	}
}

// ComplianceResultEvent represents a compliance result event published to Kafka
type ComplianceResultEvent struct {
	TransactionID   string  `json:"transaction_id"`
	OverallStatus   string  `json:"overall_status"`
	RiskScore       float64 `json:"risk_score"`
	ChecksPassed    int     `json:"checks_passed"`
	ChecksFailed    int     `json:"checks_failed"`
	ViolationsCount int     `json:"violations_count"`
	ProcessingTime  int64   `json:"processing_time_ms"`
	Timestamp       string  `json:"timestamp"`
}

// CreateResultEvent creates a result event from a compliance result
func (h *EventHandler) CreateResultEvent(result *domain.ComplianceResult) *ComplianceResultEvent {
	return &ComplianceResultEvent{
		TransactionID:   result.TransactionID,
		OverallStatus:   string(result.OverallStatus),
		RiskScore:       result.RiskScore,
		ChecksPassed:    result.Summary.PassedChecks,
		ChecksFailed:    result.Summary.FailedChecks,
		ViolationsCount: len(result.Violations),
		ProcessingTime:  result.ProcessingTime,
		Timestamp:       result.CheckedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// MetricsEvent represents a metrics event published to Kafka
type MetricsEvent struct {
	MetricName  string            `json:"metric_name"`
	Value       float64           `json:"value"`
	Labels      map[string]string `json:"labels"`
	Timestamp   string            `json:"timestamp"`
}
