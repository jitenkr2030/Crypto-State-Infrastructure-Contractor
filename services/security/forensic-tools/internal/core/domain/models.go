package domain

import (
	"time"
)

// EvidenceType represents the type of forensic evidence
type EvidenceType string

const (
	EvidenceTypeDiskImage   EvidenceType = "disk_image"
	EvidenceTypeMemoryDump  EvidenceType = "memory_dump"
	EvidenceTypeNetworkDump EvidenceType = "network_dump"
	EvidenceTypeLogFile     EvidenceType = "log_file"
	EvidenceTypeRegistry    EvidenceType = "registry"
	EvidenceTypeFile        EvidenceType = "file"
	EvidenceTypeDatabase    EvidenceType = "database"
	EvidenceTypeEmail       EvidenceType = "email"
	EvidenceTypeMobile      EvidenceType = "mobile"
	EvidenceTypeCloud       EvidenceType = "cloud"
	EvidenceTypeOther       EvidenceType = "other"
)

// EvidenceStatus represents the status of evidence
type EvidenceStatus string

const (
	EvidenceStatusCollected   EvidenceStatus = "collected"
	EvidenceStatusVerifying   EvidenceStatus = "verifying"
	EvidenceStatusVerified    EvidenceStatus = "verified"
	EvidenceStatusAnalyzing   EvidenceStatus = "analyzing"
	EvidenceStatusArchived    EvidenceStatus = "archived"
	EvidenceStatusDeleted     EvidenceStatus = "deleted"
	EvidenceStatusDamaged     EvidenceStatus = "damaged"
)

// AnalysisType represents the type of forensic analysis
type AnalysisType string

const (
	AnalysisTypeHashVerification AnalysisType = "hash_verification"
	AnalysisTypeFileCarving      AnalysisType = "file_carving"
	AnalysisTypeTimelineAnalysis AnalysisType = "timeline_analysis"
	AnalysisTypeMalwareAnalysis  AnalysisType = "malware_analysis"
	AnalysisTypeNetworkAnalysis  AnalysisType = "network_analysis"
	AnalysisTypeMemoryAnalysis   AnalysisType = "memory_analysis"
	AnalysisTypeRegistryAnalysis AnalysisType = "registry_analysis"
	AnalysisTypeStringExtraction AnalysisType = "string_extraction"
	AnalysisTypeMetadataAnalysis AnalysisType = "metadata_analysis"
	AnalysisTypeHashLookup       AnalysisType = "hash_lookup"
	AnalysisTypeYaraScan         AnalysisType = "yara_scan"
	AnalysisTypeCustom           AnalysisType = "custom"
)

// AnalysisStatus represents the status of an analysis
type AnalysisStatus string

const (
	AnalysisStatusPending    AnalysisStatus = "pending"
	AnalysisStatusRunning    AnalysisStatus = "running"
	AnalysisStatusCompleted  AnalysisStatus = "completed"
	AnalysisStatusFailed     AnalysisStatus = "failed"
	AnalysisStatusCancelled  AnalysisStatus = "cancelled"
)

// Evidence represents a piece of forensic evidence
type Evidence struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Type        EvidenceType   `json:"type"`
	Description string         `json:"description"`
	Source      string         `json:"source"`
	Hash        string         `json:"hash"`
	HashAlgorithm string       `json:"hash_algorithm"`
	Size        int64          `json:"size"`
	Location    string         `json:"location"`
	Tags        []string       `json:"tags"`
	Metadata    map[string]string `json:"metadata"`
	Status      EvidenceStatus `json:"status"`
	CollectedAt time.Time      `json:"collected_at"`
	CollectedBy string         `json:"collected_by"`
	VerifiedAt  *time.Time     `json:"verified_at,omitempty"`
	VerifiedBy  string         `json:"verified_by,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// EvidenceRequest represents a request to collect evidence
type EvidenceRequest struct {
	Name        string            `json:"name" binding:"required"`
	Type        string            `json:"type" binding:"required"`
	Description string            `json:"description"`
	Source      string            `json:"source" binding:"required"`
	Tags        []string          `json:"tags,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// ChainOfCustody represents the chain of custody for evidence
type ChainOfCustody struct {
	EvidenceID     string          `json:"evidence_id"`
	EvidenceType   EvidenceType    `json:"evidence_type"`
	Hash           string          `json:"hash"`
	CustodyRecords []*CustodyRecord `json:"custody_records"`
	FirstCustody   time.Time       `json:"first_custody"`
	LastCustody    time.Time       `json:"last_custody"`
	Complete       bool            `json:"complete"`
}

// CustodyRecord represents a single custody transfer record
type CustodyRecord struct {
	ID          string    `json:"id"`
	EvidenceID  string    `json:"evidence_id"`
	Handler     string    `json:"handler"`
	Action      string    `json:"action"`
	Location    string    `json:"location"`
	Notes       string    `json:"notes,omitempty"`
	DigitalSig  string    `json:"digital_signature"`
	PrevHash    string    `json:"previous_hash"`
	RecordHash  string    `json:"record_hash"`
	Timestamp   time.Time `json:"timestamp"`
	CreatedAt   time.Time `json:"created_at"`
}

// Analysis represents a forensic analysis task
type Analysis struct {
	ID           string         `json:"id"`
	EvidenceID   string         `json:"evidence_id"`
	EvidenceName string         `json:"evidence_name"`
	AnalysisType AnalysisType   `json:"analysis_type"`
	Status       AnalysisStatus `json:"status"`
	Parameters   map[string]interface{} `json:"parameters,omitempty"`
	Results      map[string]interface{} `json:"results,omitempty"`
	Findings     []string       `json:"findings,omitempty"`
	ProcessedBy  string         `json:"processed_by"`
	StartedAt    time.Time      `json:"started_at"`
	CompletedAt  *time.Time     `json:"completed_at,omitempty"`
	ErrorMessage string         `json:"error_message,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// AnalysisResults represents the complete results of an analysis
type AnalysisResults struct {
	AnalysisID   string                 `json:"analysis_id"`
	EvidenceID   string                 `json:"evidence_id"`
	AnalysisType AnalysisType           `json:"analysis_type"`
	Status       AnalysisStatus         `json:"status"`
	Results      map[string]interface{} `json:"results"`
	Findings     []string               `json:"findings"`
	Statistics   *AnalysisStatistics    `json:"statistics"`
	StartedAt    time.Time              `json:"started_at"`
	CompletedAt  time.Time              `json:"completed_at"`
	Duration     time.Duration          `json:"duration"`
	ProcessedBy  string                 `json:"processed_by"`
}

// AnalysisStatistics contains statistics about the analysis
type AnalysisStatistics struct {
	FilesProcessed   int64   `json:"files_processed"`
	BytesAnalyzed    int64   `json:"bytes_analyzed"`
	HitsFound        int     `json:"hits_found"`
	ErrorsCount      int     `json:"errors_count"`
	WarningsCount    int     `json:"warnings_count"`
}

// EvidenceSummary represents a brief summary of evidence for search results
type EvidenceSummary struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Type        EvidenceType `json:"type"`
	Hash        string       `json:"hash"`
	Description string       `json:"description"`
	Source      string       `json:"source"`
	Tags        []string     `json:"tags"`
	CollectedAt time.Time    `json:"collected_at"`
	Status      EvidenceStatus `json:"status"`
}
