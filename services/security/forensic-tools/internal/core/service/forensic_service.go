package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"forensic-tools/internal/core/domain"
	"forensic-tools/internal/core/ports"
)

//go:generate mockgen -destination=../../../mocks/mock_forensic_service.go -package=mocks . ForensicService

var (
	// ErrEvidenceNotFound is returned when evidence is not found
	ErrEvidenceNotFound = errors.New("evidence not found")

	// ErrAnalysisNotFound is returned when analysis is not found
	ErrAnalysisNotFound = errors.New("analysis not found")

	// ErrAnalysisInProgress is returned when analysis is still running
	ErrAnalysisInProgress = errors.New("analysis is still in progress")

	// ErrInvalidEvidenceType is returned when evidence type is invalid
	ErrInvalidEvidenceType = errors.New("invalid evidence type")

	// ErrInvalidAnalysisType is returned when analysis type is invalid
	ErrInvalidAnalysisType = errors.New("invalid analysis type")

	// ErrChainOfCustodyBroken is returned when chain of custody is broken
	ErrChainOfCustodyBroken = errors.New("chain of custody is broken or incomplete")

	// ErrEvidenceCorrupted is returned when evidence hash verification fails
	ErrEvidenceCorrupted = errors.New("evidence hash verification failed - file may be corrupted")
)

// ForensicServiceImpl implements the ForensicService interface
type ForensicServiceImpl struct {
	repo     ports.Repository
	storage  ports.StorageBackend
	messaging ports.MessagingClient
	logger   ports.Logger
}

// NewForensicService creates a new ForensicServiceImpl
func NewForensicService(repo ports.Repository, storage ports.StorageBackend, messaging ports.MessagingClient, logger ports.Logger) *ForensicServiceImpl {
	return &ForensicServiceImpl{
		repo:     repo,
		storage:  storage,
		messaging: messaging,
		logger:   logger,
	}
}

// CollectEvidence collects new evidence
func (s *ForensicServiceImpl) CollectEvidence(ctx context.Context, name, evidenceType, source, description string, tags []string, metadata map[string]string) (*domain.Evidence, error) {
	// Validate evidence type
	if !isValidEvidenceType(evidenceType) {
		return nil, ErrInvalidEvidenceType
	}

	evidence := &domain.Evidence{
		ID:            generateID(),
		Name:          name,
		Type:          domain.EvidenceType(evidenceType),
		Description:   description,
		Source:        source,
		HashAlgorithm: "SHA256",
		Tags:          tags,
		Metadata:      metadata,
		Status:        domain.EvidenceStatusCollected,
		CollectedAt:   time.Now().UTC(),
		CollectedBy:   getUserFromContext(ctx),
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	// Create evidence in repository
	if err := s.repo.CreateEvidence(ctx, evidence); err != nil {
		s.logger.Error("Failed to create evidence", "error", err, "name", name)
		return nil, fmt.Errorf("failed to create evidence: %w", err)
	}

	// Add initial custody record
	custodyRecord := &domain.CustodyRecord{
		ID:         generateID(),
		EvidenceID: evidence.ID,
		Handler:    evidence.CollectedBy,
		Action:     "COLLECTED",
		Location:   source,
		Notes:      "Initial evidence collection",
		Timestamp:  time.Now().UTC(),
		CreatedAt:  time.Now().UTC(),
	}

	if err := s.repo.AddCustodyRecord(ctx, custodyRecord); err != nil {
		s.logger.Error("Failed to add custody record", "error", err, "evidenceId", evidence.ID)
	}

	// Publish event
	if s.messaging != nil {
		if err := s.messaging.PublishEvidenceCollected(ctx, evidence); err != nil {
			s.logger.Warn("Failed to publish evidence collected event", "error", err)
		}
	}

	s.logger.Info("Evidence collected", "id", evidence.ID, "name", name, "type", evidenceType)
	return evidence, nil
}

// BatchCollectEvidence collects multiple pieces of evidence
func (s *ForensicServiceImpl) BatchCollectEvidence(ctx context.Context, requests []domain.EvidenceRequest) ([]*domain.Evidence, error) {
	results := make([]*domain.Evidence, 0, len(requests))
	errors := make([]error, 0)

	for _, req := range requests {
		evidence, err := s.CollectEvidence(ctx, req.Name, req.Type, req.Source, req.Description, req.Tags, req.Metadata)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to collect evidence %s: %w", req.Name, err))
			continue
		}
		results = append(results, evidence)
	}

	if len(errors) > 0 && len(results) == 0 {
		return nil, errors[0]
	}

	return results, nil
}

// GetEvidence retrieves evidence by ID
func (s *ForensicServiceImpl) GetEvidence(ctx context.Context, id string) (*domain.Evidence, error) {
	evidence, err := s.repo.GetEvidence(ctx, id)
	if err != nil {
		if errors.Is(err, ErrEvidenceNotFound) {
			return nil, ErrEvidenceNotFound
		}
		s.logger.Error("Failed to get evidence", "error", err, "id", id)
		return nil, fmt.Errorf("failed to get evidence: %w", err)
	}

	return evidence, nil
}

// DeleteEvidence deletes evidence with audit trail
func (s *ForensicServiceImpl) DeleteEvidence(ctx context.Context, id, deletedBy, reason string) error {
	// Get evidence first to ensure it exists
	evidence, err := s.GetEvidence(ctx, id)
	if err != nil {
		return err
	}

	// Add custody record for deletion
	custodyRecord := &domain.CustodyRecord{
		ID:         generateID(),
		EvidenceID: id,
		Handler:    deletedBy,
		Action:     "DELETED",
		Location:   "System",
		Notes:      reason,
		Timestamp:  time.Now().UTC(),
		CreatedAt:  time.Now().UTC(),
	}

	if err := s.repo.AddCustodyRecord(ctx, custodyRecord); err != nil {
		s.logger.Error("Failed to add deletion custody record", "error", err)
	}

	// Update status to deleted
	evidence.Status = domain.EvidenceStatusDeleted
	evidence.UpdatedAt = time.Now().UTC()

	if err := s.repo.UpdateEvidence(ctx, evidence); err != nil {
		s.logger.Error("Failed to update evidence status", "error", err, "id", id)
		return fmt.Errorf("failed to delete evidence: %w", err)
	}

	// Delete from storage
	if err := s.storage.DeleteEvidence(ctx, id); err != nil {
		s.logger.Warn("Failed to delete evidence from storage", "error", err, "id", id)
	}

	s.logger.Info("Evidence deleted", "id", id, "deletedBy", deletedBy, "reason", reason)
	return nil
}

// EvidenceExists checks if evidence exists
func (s *ForensicServiceImpl) EvidenceExists(ctx context.Context, id string) (bool, error) {
	return s.storage.EvidenceExists(ctx, id)
}

// GetEvidenceFile retrieves the evidence file
func (s *ForensicServiceImpl) GetEvidenceFile(ctx context.Context, id string) (io.ReadCloser, *ports.FileMetadata, error) {
	return s.storage.RetrieveEvidence(ctx, id)
}

// GetChainOfCustody retrieves the chain of custody for evidence
func (s *ForensicServiceImpl) GetChainOfCustody(ctx context.Context, evidenceID string) (*domain.ChainOfCustody, error) {
	evidence, err := s.GetEvidence(ctx, evidenceID)
	if err != nil {
		return nil, err
	}

	records, err := s.repo.GetChainOfCustody(ctx, evidenceID)
	if err != nil {
		s.logger.Error("Failed to get chain of custody", "error", err, "evidenceId", evidenceID)
		return nil, fmt.Errorf("failed to get chain of custody: %w", err)
	}

	if len(records) == 0 {
		return &domain.ChainOfCustody{
			EvidenceID:     evidenceID,
			EvidenceType:   evidence.Type,
			Hash:           evidence.Hash,
			CustodyRecords: []*domain.CustodyRecord{},
			Complete:       false,
		}, nil
	}

	// Calculate if chain is complete
	complete := verifyChainIntegrity(records)

	return &domain.ChainOfCustody{
		EvidenceID:     evidenceID,
		EvidenceType:   evidence.Type,
		Hash:           evidence.Hash,
		CustodyRecords: records,
		FirstCustody:   records[0].Timestamp,
		LastCustody:    records[len(records)-1].Timestamp,
		Complete:       complete,
	}, nil
}

// AddCustodyRecord adds a new custody record
func (s *ForensicServiceImpl) AddCustodyRecord(ctx context.Context, evidenceID, handler, action, location, notes, digitalSig string) error {
	// Get current chain of custody
	records, err := s.repo.GetChainOfCustody(ctx, evidenceID)
	if err != nil {
		return fmt.Errorf("failed to get chain of custody: %w", err)
	}

	var prevHash string
	if len(records) > 0 {
		prevHash = records[len(records)-1].RecordHash
	}

	record := &domain.CustodyRecord{
		ID:         generateID(),
		EvidenceID: evidenceID,
		Handler:    handler,
		Action:     action,
		Location:   location,
		Notes:      notes,
		DigitalSig: digitalSig,
		PrevHash:   prevHash,
		Timestamp:  time.Now().UTC(),
		CreatedAt:  time.Now().UTC(),
	}

	// Calculate record hash
	record.RecordHash = calculateRecordHash(record)

	if err := s.repo.AddCustodyRecord(ctx, record); err != nil {
		s.logger.Error("Failed to add custody record", "error", err, "evidenceId", evidenceID)
		return fmt.Errorf("failed to add custody record: %w", err)
	}

	// Publish event
	if s.messaging != nil {
		if err := s.messaging.PublishCustodyTransfer(ctx, evidenceID, record); err != nil {
			s.logger.Warn("Failed to publish custody transfer event", "error", err)
		}
	}

	s.logger.Info("Custody record added", "evidenceId", evidenceID, "handler", handler, "action", action)
	return nil
}

// VerifyChainOfCustody verifies the integrity of the chain of custody
func (s *ForensicServiceImpl) VerifyChainOfCustody(ctx context.Context, evidenceID string) (bool, error) {
	records, err := s.repo.GetChainOfCustody(ctx, evidenceID)
	if err != nil {
		return false, fmt.Errorf("failed to get chain of custody: %w", err)
	}

	if len(records) == 0 {
		return false, ErrChainOfCustodyBroken
	}

	return verifyChainIntegrity(records), nil
}

// StartAnalysis initiates evidence analysis
func (s *ForensicServiceImpl) StartAnalysis(ctx context.Context, evidenceID, analysisType string, parameters map[string]interface{}, processedBy string) (*domain.Analysis, error) {
	// Validate analysis type
	if !isValidAnalysisType(analysisType) {
		return nil, ErrInvalidAnalysisType
	}

	// Check evidence exists
	exists, err := s.EvidenceExists(ctx, evidenceID)
	if err != nil {
		return nil, fmt.Errorf("failed to check evidence existence: %w", err)
	}
	if !exists {
		return nil, ErrEvidenceNotFound
	}

	evidence, err := s.GetEvidence(ctx, evidenceID)
	if err != nil {
		return nil, err
	}

	analysis := &domain.Analysis{
		ID:           generateID(),
		EvidenceID:   evidenceID,
		EvidenceName: evidence.Name,
		AnalysisType: domain.AnalysisType(analysisType),
		Status:       domain.AnalysisStatusPending,
		Parameters:   parameters,
		ProcessedBy:  processedBy,
		StartedAt:    time.Now().UTC(),
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	if err := s.repo.CreateAnalysis(ctx, analysis); err != nil {
		s.logger.Error("Failed to create analysis", "error", err, "evidenceId", evidenceID)
		return nil, fmt.Errorf("failed to create analysis: %w", err)
	}

	// Publish event
	if s.messaging != nil {
		if err := s.messaging.PublishAnalysisStarted(ctx, analysis); err != nil {
			s.logger.Warn("Failed to publish analysis started event", "error", err)
		}
	}

	// In a real implementation, analysis would be queued for async processing
	// For now, we'll simulate a successful start
	go s.runAnalysis(analysis.ID)

	s.logger.Info("Analysis started", "id", analysis.ID, "evidenceId", evidenceID, "type", analysisType)
	return analysis, nil
}

// runAnalysis simulates running analysis (in production, this would be a worker)
func (s *ForensicServiceImpl) runAnalysis(analysisID string) {
	// Simulate async processing
	time.Sleep(2 * time.Second)

	ctx := context.Background()

	analysis, err := s.GetAnalysis(ctx, analysisID)
	if err != nil {
		s.logger.Error("Failed to get analysis for processing", "error", err, "id", analysisID)
		return
	}

	// Update status to running
	analysis.Status = domain.AnalysisStatusRunning
	analysis.UpdatedAt = time.Now().UTC()

	if err := s.repo.UpdateAnalysis(ctx, analysis); err != nil {
		s.logger.Error("Failed to update analysis status", "error", err, "id", analysisID)
	}

	// Simulate analysis completion
	now := time.Now().UTC()
	analysis.Status = domain.AnalysisStatusCompleted
	analysis.CompletedAt = &now
	analysis.Results = map[string]interface{}{
		"summary":     "Analysis completed successfully",
		"findings":    []string{"No threats detected", "System appears clean"},
		"scan_time":   "2.5s",
		"files_scanned": 150,
	}
	analysis.Findings = []string{"No threats detected", "System appears clean"}
	analysis.UpdatedAt = time.Now().UTC()

	if err := s.repo.UpdateAnalysis(ctx, analysis); err != nil {
		s.logger.Error("Failed to update completed analysis", "error", err, "id", analysisID)
	}

	// Publish completion event
	if s.messaging != nil {
		if err := s.messaging.PublishAnalysisCompleted(ctx, analysis); err != nil {
			s.logger.Warn("Failed to publish analysis completed event", "error", err)
		}
	}

	s.logger.Info("Analysis completed", "id", analysisID)
}

// GetAnalysis retrieves analysis by ID
func (s *ForensicServiceImpl) GetAnalysis(ctx context.Context, id string) (*domain.Analysis, error) {
	analysis, err := s.repo.GetAnalysis(ctx, id)
	if err != nil {
		if errors.Is(err, ErrAnalysisNotFound) {
			return nil, ErrAnalysisNotFound
		}
		s.logger.Error("Failed to get analysis", "error", err, "id", id)
		return nil, fmt.Errorf("failed to get analysis: %w", err)
	}

	return analysis, nil
}

// GetAnalysisResults retrieves the results of a completed analysis
func (s *ForensicServiceImpl) GetAnalysisResults(ctx context.Context, id string) (*domain.AnalysisResults, error) {
	analysis, err := s.GetAnalysis(ctx, id)
	if err != nil {
		return nil, err
	}

	if analysis.Status == domain.AnalysisStatusPending || analysis.Status == domain.AnalysisStatusRunning {
		return nil, ErrAnalysisInProgress
	}

	return &domain.AnalysisResults{
		AnalysisID:   analysis.ID,
		EvidenceID:   analysis.EvidenceID,
		AnalysisType: analysis.AnalysisType,
		Status:       analysis.Status,
		Results:      analysis.Results,
		Findings:     analysis.Findings,
		StartedAt:    analysis.StartedAt,
		CompletedAt:  *analysis.CompletedAt,
		Duration:     analysis.CompletedAt.Sub(analysis.StartedAt),
		ProcessedBy:  analysis.ProcessedBy,
	}, nil
}

// ListAnalyses lists analyses with optional filters
func (s *ForensicServiceImpl) ListAnalyses(ctx context.Context, evidenceID, status string, page, pageSize int) ([]*domain.Analysis, int64, error) {
	return s.repo.ListAnalyses(ctx, evidenceID, status, page, pageSize)
}

// SearchEvidence searches for evidence
func (s *ForensicServiceImpl) SearchEvidence(ctx context.Context, query string, evidenceTypes []string, dateFrom, dateTo *time.Time, tags []string, page, pageSize int) ([]domain.EvidenceSummary, int64, error) {
	return s.repo.SearchEvidence(ctx, query, evidenceTypes, dateFrom, dateTo, tags, page, pageSize)
}

// HealthCheck checks the health of the service
func (s *ForensicServiceImpl) HealthCheck(ctx context.Context) (bool, error) {
	if err := s.repo.HealthCheck(ctx); err != nil {
		s.logger.Error("Health check failed", "error", err)
		return false, err
	}
	return true, nil
}

// Helper functions

func generateID() string {
	// Generate a unique ID (in production, use UUID)
	hash := sha256.Sum256([]byte(time.Now().UTC().String() + fmt.Sprintf("%d", time.Now().UnixNano())))
	return hex.EncodeToString(hash[:16])
}

func getUserFromContext(ctx context.Context) string {
	// In a real implementation, extract user from context
	return "system"
}

func isValidEvidenceType(evidenceType string) bool {
	validTypes := []string{
		"disk_image", "memory_dump", "network_dump", "log_file",
		"registry", "file", "database", "email", "mobile", "cloud", "other",
	}
	for _, t := range validTypes {
		if strings.EqualFold(evidenceType, t) {
			return true
		}
	}
	return false
}

func isValidAnalysisType(analysisType string) bool {
	validTypes := []string{
		"hash_verification", "file_carving", "timeline_analysis",
		"malware_analysis", "network_analysis", "memory_analysis",
		"registry_analysis", "string_extraction", "metadata_analysis",
		"hash_lookup", "yara_scan", "custom",
	}
	for _, t := range validTypes {
		if strings.EqualFold(analysisType, t) {
			return true
		}
	}
	return false
}

func verifyChainIntegrity(records []*domain.CustodyRecord) bool {
	if len(records) == 0 {
		return false
	}

	for i := 1; i < len(records); i++ {
		// Verify the chain hash
		expectedHash := calculateRecordHash(records[i-1])
		if records[i].PrevHash != expectedHash {
			return false
		}

		// Verify timestamp order
		if !records[i].Timestamp.After(records[i-1].Timestamp) {
			return false
		}
	}

	return true
}

func calculateRecordHash(record *domain.CustodyRecord) string {
	data := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s",
		record.ID,
		record.EvidenceID,
		record.Handler,
		record.Action,
		record.Location,
		record.Timestamp.Format(time.RFC3339),
		record.PrevHash,
	)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}
