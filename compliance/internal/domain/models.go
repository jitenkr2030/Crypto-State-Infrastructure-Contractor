package domain

import (
	"time"
)

// ComplianceStatus represents the result of a compliance check
type ComplianceStatus string

const (
	StatusPass    ComplianceStatus = "PASS"
	StatusFail    ComplianceStatus = "FAIL"
	StatusWarn    ComplianceStatus = "WARN"
	StatusPending ComplianceStatus = "PENDING"
	StatusError   ComplianceStatus = "ERROR"
)

// RuleType defines the type of compliance rule
type RuleType string

const (
	RuleTypeAML          RuleType = "AML"
	RuleTypeKYC          RuleType = "KYC"
	RuleTypeSanctions    RuleType = "SANCTIONS"
	RuleTypeTransaction  RuleType = "TRANSACTION"
	RuleTypeGeographic   RuleType = "GEOGRAPHIC"
	RuleTypeAmount       RuleType = "AMOUNT"
	RuleTypeFrequency    RuleType = "FREQUENCY"
	RuleTypeCustom       RuleType = "CUSTOM"
)

// Severity defines the severity level of a compliance violation
type Severity string

const (
	SeverityInfo     Severity = "INFO"
	SeverityLow      Severity = "LOW"
	SeverityMedium   Severity = "MEDIUM"
	SeverityHigh     Severity = "HIGH"
	SeverityCritical Severity = "CRITICAL"
)

// Rule represents a compliance rule in the system
type Rule struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        RuleType  `json:"type"`
	Severity    Severity  `json:"severity"`
	Enabled     bool      `json:"enabled"`
	Version     int       `json:"version"`
	Expression  string    `json:"expression"`
	Parameters  Parameters `json:"parameters"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// Parameters holds rule-specific configuration
type Parameters struct {
	Threshold        float64  `json:"threshold,omitempty"`
	WindowSeconds    int      `json:"window_seconds,omitempty"`
	MaxAttempts      int      `json:"max_attempts,omitempty"`
	AllowedCountries []string `json:"allowed_countries,omitempty"`
	BlockedCountries []string `json:"blocked_countries,omitempty"`
	MinAmount        float64  `json:"min_amount,omitempty"`
	MaxAmount        float64  `json:"max_amount,omitempty"`
	RequiredFields   []string `json:"required_fields,omitempty"`
	CustomConfig     map[string]interface{} `json:"custom_config,omitempty"`
}

// Transaction represents a transaction to be checked for compliance
type Transaction struct {
	ID              string                 `json:"id"`
	Type            string                 `json:"type"`
	SourceID        string                 `json:"source_id"`
	SourceType      string                 `json:"source_type"`
	SourceName      string                 `json:"source_name"`
	SourceAccount   string                 `json:"source_account"`
	SourceCountry   string                 `json:"source_country"`
	TargetID        string                 `json:"target_id"`
	TargetType      string                 `json:"target_type"`
	TargetName      string                 `json:"target_name"`
	TargetAccount   string                 `json:"target_account"`
	TargetCountry   string                 `json:"target_country"`
	Amount          float64                `json:"amount"`
	Currency        string                 `json:"currency"`
	AssetType       string                 `json:"asset_type"`
	AssetID         string                 `json:"asset_id"`
	Timestamp       time.Time              `json:"timestamp"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// ComplianceCheck represents a single compliance check result
type ComplianceCheck struct {
	ID         string           `json:"id"`
	RuleID     string           `json:"rule_id"`
	RuleName   string           `json:"rule_name"`
	RuleType   RuleType         `json:"rule_type"`
	Status     ComplianceStatus `json:"status"`
	Severity   Severity         `json:"severity"`
	Message    string           `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
	CheckedAt  time.Time        `json:"checked_at"`
	Duration   int64            `json:"duration_ms"`
}

// ComplianceResult represents the complete compliance check result
type ComplianceResult struct {
	TransactionID   string            `json:"transaction_id"`
	OverallStatus   ComplianceStatus  `json:"overall_status"`
	RiskScore       float64           `json:"risk_score"`
	Checks          []ComplianceCheck `json:"checks"`
	Violations      []Violation       `json:"violations,omitempty"`
	Summary         ResultSummary     `json:"summary"`
	CheckedAt       time.Time         `json:"checked_at"`
	ProcessingTime  int64             `json:"processing_time_ms"`
}

// ResultSummary provides a summary of compliance check results
type ResultSummary struct {
	TotalChecks     int            `json:"total_checks"`
	PassedChecks    int            `json:"passed_checks"`
	FailedChecks    int            `json:"failed_checks"`
	WarningChecks   int            `json:"warning_checks"`
	CriticalCount   int            `json:"critical_count"`
	HighCount       int            `json:"high_count"`
	MediumCount     int            `json:"medium_count"`
	LowCount        int            `json:"low_count"`
}

// Violation represents a compliance violation
type Violation struct {
	ID             string                 `json:"id"`
	TransactionID  string                 `json:"transaction_id"`
	RuleID         string                 `json:"rule_id"`
	RuleName       string                 `json:"rule_name"`
	Severity       Severity               `json:"severity"`
	Status         string                 `json:"status"`
	Resolution     string                 `json:"resolution,omitempty"`
	ResolvedBy     string                 `json:"resolved_by,omitempty"`
	ResolvedAt     *time.Time             `json:"resolved_at,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	Details        map[string]interface{} `json:"details,omitempty"`
}

// Ruleset represents a collection of compliance rules
type Ruleset struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Version     int       `json:"version"`
	Active      bool      `json:"active"`
	Rules       []Rule    `json:"rules"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Entity represents an entity subject to compliance checks
type Entity struct {
	ID           string            `json:"id"`
	Type         string            `json:"type"`
	Name         string            `json:"name"`
	Country      string            `json:"country"`
	RiskScore    float64           `json:"risk_score"`
	Tags         []string          `json:"tags"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Blacklisted  bool              `json:"blacklisted"`
	Watchlist    bool              `json:"watchlist"`
	KYCVerified  bool              `json:"kyc_verified"`
	AMLStatus    string            `json:"aml_status"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// ComplianceEvent represents an event in the compliance system
type ComplianceEvent struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Transaction Transaction `json:"transaction"`
	Result      ComplianceResult `json:"result"`
	Timestamp   time.Time `json:"timestamp"`
}

// WatchlistEntry represents an entry on a watchlist
type WatchlistEntry struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"`
	Name         string    `json:"name"`
	Country      string    `json:"country"`
	ListSource   string    `json:"list_source"`
	MatchScore   float64   `json:"match_score"`
	Active       bool      `json:"active"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
}
