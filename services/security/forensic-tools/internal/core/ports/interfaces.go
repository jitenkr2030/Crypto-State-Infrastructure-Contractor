package ports

import (
	"context"
	"io"
	"time"
)

// Logger defines the logging interface
type Logger interface {
	Info(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Debug(msg string, keysAndValues ...interface{})
}

// ForensicService defines the interface for forensic operations
type ForensicService interface {
	// Evidence Management
	CollectEvidence(ctx context.Context, name, evidenceType, source, description string, tags []string, metadata map[string]string) (*Evidence, error)
	BatchCollectEvidence(ctx context.Context, requests []EvidenceRequest) ([]*Evidence, error)
	GetEvidence(ctx context.Context, id string) (*Evidence, error)
	DeleteEvidence(ctx context.Context, id, deletedBy, reason string) error
	EvidenceExists(ctx context.Context, id string) (bool, error)
	GetEvidenceFile(ctx context.Context, id string) (io.ReadCloser, *FileMetadata, error)

	// Chain of Custody
	GetChainOfCustody(ctx context.Context, evidenceID string) (*ChainOfCustody, error)
	AddCustodyRecord(ctx context.Context, evidenceID, handler, action, location, notes, digitalSig string) error
	VerifyChainOfCustody(ctx context.Context, evidenceID string) (bool, error)

	// Analysis
	StartAnalysis(ctx context.Context, evidenceID, analysisType string, parameters map[string]interface{}, processedBy string) (*Analysis, error)
	GetAnalysis(ctx context.Context, id string) (*Analysis, error)
	GetAnalysisResults(ctx context.Context, id string) (*AnalysisResults, error)
	ListAnalyses(ctx context.Context, evidenceID, status string, page, pageSize int) ([]*Analysis, int64, error)

	// Search
	SearchEvidence(ctx context.Context, query string, evidenceTypes []string, dateFrom, dateTo *time.Time, tags []string, page, pageSize int) ([]EvidenceSummary, int64, error)

	// Health
	HealthCheck(ctx context.Context) (bool, error)
}

// StorageBackend defines the interface for evidence storage
type StorageBackend interface {
	StoreEvidence(ctx context.Context, evidenceID string, file io.Reader, metadata *FileMetadata) error
	RetrieveEvidence(ctx context.Context, evidenceID string) (io.ReadCloser, *FileMetadata, error)
	DeleteEvidence(ctx context.Context, evidenceID string) error
	EvidenceExists(ctx context.Context, evidenceID string) (bool, error)
	GetStorageStats(ctx context.Context) (*StorageStats, error)
}

// Repository defines the interface for data persistence
type Repository interface {
	// Evidence operations
	CreateEvidence(ctx context.Context, evidence *Evidence) error
	GetEvidence(ctx context.Context, id string) (*Evidence, error)
	UpdateEvidence(ctx context.Context, evidence *Evidence) error
	DeleteEvidence(ctx context.Context, id string) error
	ListEvidence(ctx context.Context, page, pageSize int) ([]*Evidence, int64, error)
	SearchEvidence(ctx context.Context, query string, evidenceTypes []string, dateFrom, dateTo *time.Time, tags []string, page, pageSize int) ([]EvidenceSummary, int64, error)

	// Chain of custody operations
	AddCustodyRecord(ctx context.Context, record *CustodyRecord) error
	GetChainOfCustody(ctx context.Context, evidenceID string) ([]*CustodyRecord, error)

	// Analysis operations
	CreateAnalysis(ctx context.Context, analysis *Analysis) error
	GetAnalysis(ctx context.Context, id string) (*Analysis, error)
	UpdateAnalysis(ctx context.Context, analysis *Analysis) error
	ListAnalyses(ctx context.Context, evidenceID, status string, page, pageSize int) ([]*Analysis, int64, error)

	// Health check
	HealthCheck(ctx context.Context) error
}

// MessagingClient defines the interface for event publishing
type MessagingClient interface {
	PublishEvidenceCollected(ctx context.Context, evidence *Evidence) error
	PublishAnalysisStarted(ctx context.Context, analysis *Analysis) error
	PublishAnalysisCompleted(ctx context.Context, analysis *Analysis) error
	PublishCustodyTransfer(ctx context.Context, evidenceID string, record *CustodyRecord) error
	PublishSecurityEvent(ctx context.Context, event *SecurityEvent) error
	Close() error
}

// FileMetadata contains metadata about a stored file
type FileMetadata struct {
	Filename    string
	ContentType string
	Size        int64
	Hash        string
	CreatedAt   time.Time
}

// StorageStats contains storage statistics
type StorageStats struct {
	TotalSize     int64
	FileCount     int64
	LastUpdated   time.Time
}

// SecurityEvent represents a security-related event
type SecurityEvent struct {
	EventType   string
	Severity    string
	EvidenceID  string
	AnalysisID  string
	Description string
	Timestamp   time.Time
	Source      string
	Details     map[string]interface{}
}
