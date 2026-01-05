package domain

import (
	"time"
)

// AuditEntry represents a single audit log entry
type AuditEntry struct {
	ID            string                 `json:"id"`
	TraceID       string                 `json:"trace_id"`
	ActorID       string                 `json:"actor_id"`
	ActorType     string                 `json:"actor_type"`
	Action        string                 `json:"action"`
	Resource      string                 `json:"resource"`
	ResourceID    string                 `json:"resource_id"`
	Operation     string                 `json:"operation"`
	Outcome       string                 `json:"outcome"` // success, failure, partial
	Severity      string                 `json:"severity"` // info, warning, error, critical
	Payload       map[string]interface{} `json:"payload,omitempty"`
	EncryptedData string                 `json:"encrypted_data,omitempty"`
	Metadata      map[string]string      `json:"metadata,omitempty"`
	SourceIP      string                 `json:"source_ip"`
	UserAgent     string                 `json:"user_agent"`
	Timestamp     time.Time              `json:"timestamp"`
	PreviousHash  string                 `json:"previous_hash"`
	CurrentHash   string                 `json:"current_hash"`
	CreatedAt     time.Time              `json:"created_at"`
}

// AuditAction defines standard audit action types
type AuditAction string

const (
	// Authentication actions
	ActionLogin          AuditAction = "LOGIN"
	ActionLogout         AuditAction = "LOGOUT"
	ActionLoginFailed    AuditAction = "LOGIN_FAILED"
	ActionPasswordChange AuditAction = "PASSWORD_CHANGE"
	ActionMFAEnable      AuditAction = "MFA_ENABLE"
	ActionMFADisable     AuditAction = "MFA_DISABLE"
	ActionSessionCreate  AuditAction = "SESSION_CREATE"
	ActionSessionRevoke  AuditAction = "SESSION_REVOKE"

	// Data operations
	ActionCreate    AuditAction = "CREATE"
	ActionRead      AuditAction = "READ"
	ActionUpdate    AuditAction = "UPDATE"
	ActionDelete    AuditAction = "DELETE"
	ActionExport    AuditAction = "EXPORT"
	ActionImport    AuditAction = "IMPORT"
	ActionSearch    AuditAction = "SEARCH"
	ActionDownload  AuditAction = "DOWNLOAD"
	ActionUpload    AuditAction = "UPLOAD"

	// System operations
	ActionConfigChange  AuditAction = "CONFIG_CHANGE"
	ActionPermissionGrant AuditAction = "PERMISSION_GRANT"
	ActionPermissionRevoke AuditAction = "PERMISSION_REVOKE"
	ActionRoleAssign   AuditAction = "ROLE_ASSIGN"
	ActionRoleRemove   AuditAction = "ROLE_REMOVE"

	// Security events
	ActionThreatDetected AuditAction = "THREAT_DETECTED"
	ActionIntrusionAttempt AuditAction = "INTRUSION_ATTEMPT"
	ActionAccessDenied   AuditAction = "ACCESS_DENIED"
	ActionPolicyViolation AuditAction = "POLICY_VIOLATION"

	// API Gateway events
	ActionAPIRequest   AuditAction = "API_REQUEST"
	ActionRateLimitHit AuditAction = "RATE_LIMIT_HIT"
	ActionAuthRequired AuditAction = "AUTH_REQUIRED"

	// Administrative
	ActionServiceStart  AuditAction = "SERVICE_START"
	ActionServiceStop   AuditAction = "SERVICE_STOP"
	ActionBackupCreate  AuditAction = "BACKUP_CREATE"
	ActionBackupRestore AuditAction = "BACKUP_RESTORE"
)

// AuditEntryRequest represents a request to create an audit entry
type AuditEntryRequest struct {
	TraceID    string                 `json:"trace_id"`
	ActorID    string                 `json:"actor_id"`
	ActorType  string                 `json:"actor_type"`
	Action     string                 `json:"action"`
	Resource   string                 `json:"resource"`
	ResourceID string                 `json:"resource_id"`
	Operation  string                 `json:"operation"`
	Outcome    string                 `json:"outcome"`
	Severity   string                 `json:"severity"`
	Payload    map[string]interface{} `json:"payload,omitempty"`
	Metadata   map[string]string      `json:"metadata,omitempty"`
}

// AuditEntryResponse represents an audit entry response
type AuditEntryResponse struct {
	ID           string    `json:"id"`
	TraceID      string    `json:"trace_id"`
	ActorID      string    `json:"actor_id"`
	Action       string    `json:"action"`
	Resource     string    `json:"resource"`
	Outcome      string    `json:"outcome"`
	Severity     string    `json:"severity"`
	Timestamp    time.Time `json:"timestamp"`
	Verification *VerificationResult `json:"verification,omitempty"`
}

// VerificationResult represents the result of hash chain verification
type VerificationResult struct {
	Valid       bool      `json:"valid"`
	BlockNumber int64     `json:"block_number"`
	Timestamp   time.Time `json:"verified_at"`
	Message     string    `json:"message"`
}

// AuditSearchRequest represents search parameters for audit entries
type AuditSearchRequest struct {
	TraceID     string     `json:"trace_id,omitempty"`
	ActorID     string     `json:"actor_id,omitempty"`
	ActorType   string     `json:"actor_type,omitempty"`
	Action      string     `json:"action,omitempty"`
	Resource    string     `json:"resource,omitempty"`
	Outcome     string     `json:"outcome,omitempty"`
	Severity    string     `json:"severity,omitempty"`
	StartTime   *time.Time `json:"start_time,omitempty"`
	EndTime     *time.Time `json:"end_time,omitempty"`
	SourceIP    string     `json:"source_ip,omitempty"`
	Page        int        `json:"page"`
	PageSize    int        `json:"page_size"`
	SortBy      string     `json:"sort_by"`
	SortOrder   string     `json:"sort_order"`
}

// AuditSearchResponse represents paginated search results
type AuditSearchResponse struct {
	Entries     []AuditEntryResponse `json:"entries"`
	TotalCount  int64                `json:"total_count"`
	Page        int                  `json:"page"`
	PageSize    int                  `json:"page_size"`
	TotalPages  int                  `json:"total_pages"`
}

// AuditChainSummary provides statistics about the audit chain
type AuditChainSummary struct {
	TotalEntries    int64      `json:"total_entries"`
	FirstEntryTime  time.Time  `json:"first_entry_time"`
	LastEntryTime   time.Time  `json:"last_entry_time"`
	ChainIntegrity  string     `json:"chain_integrity"` // valid, broken, unknown
	RecentActivity  int64      `json="recent_activity_24h"`
}

// AuditEvent represents an event received from Kafka
type AuditEvent struct {
	EventType   string                 `json:"event_type"`
	TraceID     string                 `json:"trace_id"`
	ActorID     string                 `json:"actor_id"`
	ActorType   string                 `json:"actor_type"`
	Action      string                 `json:"action"`
	Resource    string                 `json:"resource"`
	ResourceID  string                 `json:"resource_id"`
	Outcome     string                 `json="outcome"`
	Severity    string                 `json="severity"`
	Payload     map[string]interface{} `json="payload,omitempty"`
	Metadata    map[string]string      `json="metadata,omitempty"`
	SourceIP    string                 `json="source_ip"`
	UserAgent   string                 `json="user_agent"`
	Timestamp   time.Time              `json="timestamp"`
}
