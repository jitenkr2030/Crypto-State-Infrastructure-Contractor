package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"audit-log/internal/core/domain"
	"audit-log/internal/core/ports"
)

//go:generate mockgen -destination=../../../mocks/mock_audit_service.go -package=mocks . AuditLogService

var (
	// ErrEntryNotFound is returned when an audit entry is not found
	ErrEntryNotFound = errors.New("audit entry not found")

	// ErrChainBroken is returned when the hash chain verification fails
	ErrChainBroken = errors.New("audit chain is broken - tampering detected")

	// ErrInvalidHash is returned when an entry has an invalid hash
	ErrInvalidHash = errors.New("invalid hash detected")

	// ErrInvalidRequest is returned when the request is invalid
	ErrInvalidRequest = errors.New("invalid request")
)

// AuditLogServiceImpl implements the AuditLogService interface
type AuditLogServiceImpl struct {
	repo     ports.AuditLogRepository
	producer ports.KafkaProducer
	logger   ports.Logger
}

// NewAuditLogService creates a new AuditLogServiceImpl
func NewAuditLogService(repo ports.AuditLogRepository, producer ports.KafkaProducer, logger ports.Logger) *AuditLogServiceImpl {
	return &AuditLogServiceImpl{
		repo:     repo,
		producer: producer,
		logger:   logger,
	}
}

// CreateEntry creates a new audit entry with hash chaining
func (s *AuditLogServiceImpl) CreateEntry(ctx context.Context, request domain.AuditEntryRequest) (*domain.AuditEntry, error) {
	// Get the previous hash for chaining
	previousHash, err := s.repo.GetLastHash(ctx)
	if err != nil && !errors.Is(err, ErrEntryNotFound) {
		s.logger.Error("Failed to get last hash", "error", err)
		return nil, fmt.Errorf("failed to create audit entry: %w", err)
	}

	// Create the entry
	entry := &domain.AuditEntry{
		ID:           generateID(),
		TraceID:      request.TraceID,
		ActorID:      request.ActorID,
		ActorType:    request.ActorType,
		Action:       request.Action,
		Resource:     request.Resource,
		ResourceID:   request.ResourceID,
		Operation:    request.Operation,
		Outcome:      request.Outcome,
		Severity:     request.Severity,
		Payload:      request.Payload,
		Metadata:     request.Metadata,
		SourceIP:     request.SourceIP,
		UserAgent:    request.UserAgent,
		Timestamp:    time.Now().UTC(),
		PreviousHash: previousHash,
		CreatedAt:    time.Now().UTC(),
	}

	// Calculate the hash
	entry.CurrentHash = calculateHash(entry)

	// Save to repository
	if err := s.repo.CreateEntry(ctx, entry); err != nil {
		s.logger.Error("Failed to create audit entry", "error", err, "id", entry.ID)
		return nil, fmt.Errorf("failed to create audit entry: %w", err)
	}

	s.logger.Info("Audit entry created", "id", entry.ID, "action", entry.Action, "actor", entry.ActorID)

	// Publish to Kafka if producer is available
	if s.producer != nil {
		if err := s.producer.PublishEntry(ctx, entry); err != nil {
			s.logger.Warn("Failed to publish audit entry to Kafka", "error", err, "id", entry.ID)
		}
	}

	return entry, nil
}

// CreateEntryWithHash creates an audit entry with a pre-calculated hash
func (s *AuditLogServiceImpl) CreateEntryWithHash(ctx context.Context, entry *domain.AuditEntry) error {
	if entry.CurrentHash == "" {
		previousHash, err := s.repo.GetLastHash(ctx)
		if err != nil && !errors.Is(err, ErrEntryNotFound) {
			return fmt.Errorf("failed to get last hash: %w", err)
		}
		entry.PreviousHash = previousHash
		entry.CurrentHash = calculateHash(entry)
	}

	if err := s.repo.CreateEntry(ctx, entry); err != nil {
		s.logger.Error("Failed to create audit entry", "error", err, "id", entry.ID)
		return fmt.Errorf("failed to create audit entry: %w", err)
	}

	s.logger.Info("Audit entry created with hash", "id", entry.ID)
	return nil
}

// ProcessEvent processes an audit event from Kafka
func (s *AuditLogServiceImpl) ProcessEvent(ctx context.Context, event domain.AuditEvent) (*domain.AuditEntry, error) {
	request := domain.AuditEntryRequest{
		TraceID:    event.TraceID,
		ActorID:    event.ActorID,
		ActorType:  event.ActorType,
		Action:     event.Action,
		Resource:   event.Resource,
		ResourceID: event.ResourceID,
		Outcome:    event.Outcome,
		Severity:   event.Severity,
		Payload:    event.Payload,
		Metadata:   event.Metadata,
		SourceIP:   event.SourceIP,
		UserAgent:  event.UserAgent,
	}

	return s.CreateEntry(ctx, request)
}

// GetEntry retrieves an audit entry by ID
func (s *AuditLogServiceImpl) GetEntry(ctx context.Context, id string) (*domain.AuditEntry, error) {
	entry, err := s.repo.GetEntry(ctx, id)
	if err != nil {
		if errors.Is(err, ErrEntryNotFound) {
			return nil, ErrEntryNotFound
		}
		s.logger.Error("Failed to get audit entry", "error", err, "id", id)
		return nil, fmt.Errorf("failed to get audit entry: %w", err)
	}

	return entry, nil
}

// GetEntryByTraceID retrieves all audit entries for a trace ID
func (s *AuditLogServiceImpl) GetEntryByTraceID(ctx context.Context, traceID string) ([]*domain.AuditEntry, error) {
	entries, err := s.repo.GetEntryByTraceID(ctx, traceID)
	if err != nil {
		s.logger.Error("Failed to get audit entries by trace ID", "error", err, "traceID", traceID)
		return nil, fmt.Errorf("failed to get audit entries: %w", err)
	}

	return entries, nil
}

// SearchEntries searches for audit entries based on criteria
func (s *AuditLogServiceImpl) SearchEntries(ctx context.Context, request domain.AuditSearchRequest) (*domain.AuditSearchResponse, error) {
	// Set defaults
	if request.Page < 1 {
		request.Page = 1
	}
	if request.PageSize < 1 || request.PageSize > 100 {
		request.PageSize = 20
	}
	if request.SortBy == "" {
		request.SortBy = "timestamp"
	}
	if request.SortOrder == "" {
		request.SortOrder = "desc"
	}

	entries, total, err := s.repo.SearchEntries(ctx, request)
	if err != nil {
		s.logger.Error("Failed to search audit entries", "error", err)
		return nil, fmt.Errorf("failed to search audit entries: %w", err)
	}

	// Convert to response format
	responseEntries := make([]domain.AuditEntryResponse, len(entries))
	for i, entry := range entries {
		responseEntries[i] = domain.AuditEntryResponse{
			ID:        entry.ID,
			TraceID:   entry.TraceID,
			ActorID:   entry.ActorID,
			Action:    entry.Action,
			Resource:  entry.Resource,
			Outcome:   entry.Outcome,
			Severity:  entry.Severity,
			Timestamp: entry.Timestamp,
		}
	}

	totalPages := int(total) / request.PageSize
	if int(total)%request.PageSize > 0 {
		totalPages++
	}

	return &domain.AuditSearchResponse{
		Entries:    responseEntries,
		TotalCount: total,
		Page:       request.Page,
		PageSize:   request.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetChainSummary returns statistics about the audit chain
func (s *AuditLogServiceImpl) GetChainSummary(ctx context.Context) (*domain.AuditChainSummary, error) {
	return s.repo.GetChainSummary(ctx)
}

// VerifyEntry verifies a single audit entry's integrity
func (s *AuditLogServiceImpl) VerifyEntry(ctx context.Context, id string) (*domain.VerificationResult, error) {
	result, err := s.repo.VerifyChain(ctx, id, 1)
	if err != nil {
		if errors.Is(err, ErrEntryNotFound) {
			return nil, ErrEntryNotFound
		}
		return nil, err
	}

	return result, nil
}

// VerifyChain verifies the hash chain starting from an entry
func (s *AuditLogServiceImpl) VerifyChain(ctx context.Context, startID string, limit int) (*domain.VerificationResult, error) {
	if limit <= 0 {
		limit = 100
	}

	return s.repo.VerifyChain(ctx, startID, limit)
}

// HealthCheck checks the health of the service
func (s *AuditLogServiceImpl) HealthCheck(ctx context.Context) (bool, error) {
	if err := s.repo.HealthCheck(ctx); err != nil {
		s.logger.Error("Health check failed", "error", err)
		return false, err
	}
	return true, nil
}

// Helper functions

func generateID() string {
	timestamp := time.Now().UnixNano()
	data := fmt.Sprintf("%d_%s", timestamp, randomString(16))
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:16])
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
		time.Sleep(time.Nanosecond)
	}
	return string(b)
}

func calculateHash(entry *domain.AuditEntry) string {
	// Create deterministic JSON representation of the entry without CurrentHash
	data := struct {
		ID           string                 `json:"id"`
		TraceID      string                 `json:"trace_id"`
		ActorID      string                 `json:"actor_id"`
		ActorType    string                 `json:"actor_type"`
		Action       string                 `json:"action"`
		Resource     string                 `json:"resource"`
		ResourceID   string                 `json:"resource_id"`
		Operation    string                 `json:"operation"`
		Outcome      string                 `json:"outcome"`
		Severity     string                 `json:"severity"`
		Timestamp    time.Time              `json:"timestamp"`
		PreviousHash string                 `json:"previous_hash"`
		Payload      map[string]interface{} `json:"payload,omitempty"`
	}{
		ID:           entry.ID,
		TraceID:      entry.TraceID,
		ActorID:      entry.ActorID,
		ActorType:    entry.ActorType,
		Action:       entry.Action,
		Resource:     entry.Resource,
		ResourceID:   entry.ResourceID,
		Operation:    entry.Operation,
		Outcome:      entry.Outcome,
		Severity:     entry.Severity,
		Timestamp:    entry.Timestamp,
		PreviousHash: entry.PreviousHash,
		Payload:      entry.Payload,
	}

	jsonData, _ := json.Marshal(data)
	hashInput := string(jsonData)

	// Include previous hash for chaining
	if entry.PreviousHash != "" {
		hashInput = entry.PreviousHash + "|" + hashInput
	}

	hash := sha256.Sum256([]byte(hashInput))
	return hex.EncodeToString(hash[:])
}
