package domain

import (
	"time"
)

// ReportStatus represents the status of a report
type ReportStatus string

const (
	ReportStatusPending    ReportStatus = "pending"
	ReportStatusGenerating ReportStatus = "generating"
	ReportStatusCompleted  ReportStatus = "completed"
	ReportStatusFailed     ReportStatus = "failed"
	ReportStatusExpired    ReportStatus = "expired"
)

// ReportType represents the type of regulatory report
type ReportType string

const (
	ReportTypeAMLMonitoring       ReportType = "aml_monitoring"
	ReportTypeCTFReport           ReportType = "ctf_report"
	ReportTypeKYCReport           ReportType = "kyc_report"
	ReportTypeSAR                 ReportType = "sar"
	ReportTypeCTR                 ReportType = "ctr"
	ReportTypeCMIR                ReportType = "cmir"
	ReportTypeComplianceSummary   ReportType = "compliance_summary"
	ReportTypeRiskAssessment      ReportType = "risk_assessment"
	ReportTypeTransactionHistory  ReportType = "transaction_history"
	ReportTypeAuditTrail          ReportType = "audit_trail"
	ReportTypeRegulatoryFiling    ReportType = "regulatory_filing"
)

// ReportFormat represents the output format of a report
type ReportFormat string

const (
	ReportFormatPDF  ReportFormat = "pdf"
	ReportFormatXLSX ReportFormat = "xlsx"
	ReportFormatCSV  ReportFormat = "csv"
	ReportFormatJSON ReportFormat = "json"
)

// Report represents a regulatory report
type Report struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Type        ReportType        `json:"type"`
	Format      ReportFormat      `json:"format"`
	Status      ReportStatus      `json:"status"`
	Description string            `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
	Filters     ReportFilters     `json:"filters"`
	Result      *ReportResult     `json:"result"`
	Metadata    map[string]string `json:"metadata"`
	ScheduledID string            `json:"scheduled_id"`
	GeneratedAt *time.Time        `json:"generated_at"`
	ExpiresAt   *time.Time        `json:"expires_at"`
	FilePath    string            `json:"file_path"`
	FileSize    int64             `json:"file_size"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// ReportFilters represents filters for report data
type ReportFilters struct {
	StartDate    *time.Time `json:"start_date"`
	EndDate      *time.Time `json:"end_date"`
	Networks     []string   `json:"networks"`
	Addresses    []string   `json:"addresses"`
	TransactionTypes []string `json:"transaction_types"`
	RiskLevels   []string   `json:"risk_levels"`
	Statuses     []string   `json:"statuses"`
}

// ReportResult contains the result of report generation
type ReportResult struct {
	TotalRecords   int                    `json:"total_records"`
	FilteredRecords int                   `json:"filtered_records"`
	Summary        map[string]interface{} `json:"summary"`
	Data           []map[string]interface{} `json:"data"`
	Errors         []ReportError          `json:"errors"`
}

// ReportError represents an error during report generation
type ReportError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details"`
}

// ReportTypeInfo contains metadata about a report type
type ReportTypeInfo struct {
	Type          ReportType `json:"type"`
	Name          string     `json:"name"`
	Description   string     `json:"description"`
	Category      string     `json:"category"`
	Formats       []ReportFormat `json:"formats"`
	Parameters    []ReportParameter `json:"parameters"`
	RegulatoryFramework string   `json:"regulatory_framework"`
	RetentionDays int        `json:"retention_days"`
}

// ReportParameter represents a parameter for report generation
type ReportParameter struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Required    bool        `json:"required"`
	Description string      `json:"description"`
	Default     interface{} `json:"default"`
	Options     []string    `json:"options"`
}

// Schedule represents a scheduled report configuration
type Schedule struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	ReportType  ReportType        `json:"report_type"`
	Format      ReportFormat      `json:"format"`
	Cron        string            `json:"cron"`
	Enabled     bool              `json:"enabled"`
	Parameters  map[string]interface{} `json:"parameters"`
	Filters     ReportFilters     `json:"filters"`
	Recipients  []string          `json:"recipients"`
	Metadata    map[string]string `json:"metadata"`
	LastRun     *time.Time        `json:"last_run"`
	NextRun     *time.Time        `json:"next_run"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// Template represents a report template
type Template struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Type        ReportType        `json:"type"`
	Content     string            `json:"content"`
	Parameters  []ReportParameter `json:"parameters"`
	Variables   []TemplateVariable `json:"variables"`
	Metadata    map[string]string `json:"metadata"`
	Version     int               `json:"version"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// TemplateVariable represents a variable in a template
type TemplateVariable struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

// CreateReportRequest represents a request to create a new report
type CreateReportRequest struct {
	Name        string                 `json:"name" binding:"required"`
	Type        ReportType             `json:"type" binding:"required"`
	Format      ReportFormat           `json:"format"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
	Filters     ReportFilters          `json:"filters"`
}

// UpdateReportRequest represents a request to update a report
type UpdateReportRequest struct {
	Name        *string                 `json:"name"`
	Description *string                 `json:"description"`
	Parameters  map[string]interface{}  `json:"parameters"`
	Filters     *ReportFilters          `json:"filters"`
}

// CreateScheduleRequest represents a request to create a new schedule
type CreateScheduleRequest struct {
	Name       string                 `json:"name" binding:"required"`
	ReportType ReportType             `json:"report_type" binding:"required"`
	Format     ReportFormat           `json:"format"`
	Cron       string                 `json:"cron" binding:"required"`
	Enabled    bool                   `json:"enabled"`
	Parameters map[string]interface{} `json:"parameters"`
	Filters    ReportFilters          `json:"filters"`
	Recipients []string               `json:"recipients"`
}

// UpdateScheduleRequest represents a request to update a schedule
type UpdateScheduleRequest struct {
	Name       *string                 `json:"name"`
	Cron       *string                 `json:"cron"`
	Enabled    *bool                   `json:"enabled"`
	Parameters map[string]interface{}  `json:"parameters"`
	Filters    *ReportFilters          `json:"filters"`
	Recipients []string                `json:"recipients"`
}

// CreateTemplateRequest represents a request to create a new template
type CreateTemplateRequest struct {
	Name       string                 `json:"name" binding:"required"`
	Type       ReportType             `json:"type" binding:"required"`
	Content    string                 `json:"content"`
	Parameters []ReportParameter      `json:"parameters"`
	Variables  []TemplateVariable     `json:"variables"`
}

// UpdateTemplateRequest represents a request to update a template
type UpdateTemplateRequest struct {
	Name       *string                `json:"name"`
	Content    *string                `json:"content"`
	Parameters []ReportParameter      `json:"parameters"`
	Variables  []TemplateVariable     `json:"variables"`
}

// PaginatedReports represents a paginated list of reports
type PaginatedReports struct {
	Reports  []*Report `json:"reports"`
	Total    int       `json:"total"`
	Offset   int       `json:"offset"`
	Limit    int       `json:"limit"`
}

// PaginatedSchedules represents a paginated list of schedules
type PaginatedSchedules struct {
	Schedules []*Schedule `json:"schedules"`
	Total     int         `json:"total"`
	Offset    int         `json:"offset"`
	Limit     int         `json:"limit"`
}

// PaginatedTemplates represents a paginated list of templates
type PaginatedTemplates struct {
	Templates []*Template `json:"templates"`
	Total     int         `json:"total"`
	Offset    int         `json:"offset"`
	Limit     int         `json:"limit"`
}
