package ports

import (
	"context"
	"io"
	"time"

	"audit-log/internal/core/domain"
)

// Logger defines the logging interface
type Logger interface {
	Info(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Debug(msg string, keysAndValues ...interface{})
}

// AuditLogRepository defines the interface for audit log persistence
type AuditLogRepository interface {
	// Create operations
	CreateEntry(ctx context.Context, entry *domain.AuditEntry) error
	GetLastHash(ctx context.Context) (string, error)

	// Read operations
	GetEntry(ctx context.Context, id string) (*domain.AuditEntry, error)
	GetEntryByTraceID(ctx context.Context, traceID string) ([]*domain.AuditEntry, error)
	SearchEntries(ctx context.Context, request domain.AuditSearchRequest) ([]*domain.AuditEntry, int64, error)
	GetChainSummary(ctx context.Context) (*domain.AuditChainSummary, error)

	// Verification
	VerifyChain(ctx context.Context, startID string, limit int) (*domain.VerificationResult, error)

	// Health
	HealthCheck(ctx context.Context) error
}

// AuditLogService defines the interface for audit log business logic
type AuditLogService interface {
	// Entry creation
	CreateEntry(ctx context.Context, request domain.AuditEntryRequest) (*domain.AuditEntry, error)
	CreateEntryWithHash(ctx context.Context, entry *domain.AuditEntry) error
	ProcessEvent(ctx context.Context, event domain.AuditEvent) (*domain.AuditEntry, error)

	// Query operations
	GetEntry(ctx context.Context, id string) (*domain.AuditEntry, error)
	GetEntryByTraceID(ctx context.Context, traceID string) ([]*domain.AuditEntry, error)
	SearchEntries(ctx context.Context, request domain.AuditSearchRequest) (*domain.AuditSearchResponse, error)
	GetChainSummary(ctx context.Context) (*domain.AuditChainSummary, error)

	// Verification
	VerifyEntry(ctx context.Context, id string) (*domain.VerificationResult, error)
	VerifyChain(ctx context.Context, startID string, limit int) (*domain.VerificationResult, error)

	// Health
	HealthCheck(ctx context.Context) (bool, error)
}

// KafkaConsumer defines the interface for consuming audit events
type KafkaConsumer interface {
	Consume(ctx context.Context) error
	Close() error
}

// KafkaProducer defines the interface for publishing audit events
type KafkaProducer interface {
	PublishEntry(ctx context.Context, entry *domain.AuditEntry) error
	PublishVerificationEvent(ctx context.Context, entryID string, result *domain.VerificationResult) error
	Close() error
}

// StorageBackend defines the interface for blob storage
type StorageBackend interface {
	Store(ctx context.Context, key string, data io.Reader, contentType string) error
	Retrieve(ctx context.Context, key string) (io.Reader, string, error)
	Delete(ctx context.Context, key string) error
}

// ReportGenerator defines the interface for generating compliance reports
type ReportGenerator interface {
	GenerateComplianceReport(ctx context.Context, startDate, endDate time.Time) ([]byte, error)
	GenerateActivityReport(ctx context.Context, actorID string, startDate, endDate time.Time) ([]byte, error)
}
