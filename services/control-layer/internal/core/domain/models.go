package domain

import (
	"time"
)

// PolicyEffect defines the effect of a policy
type PolicyEffect string

const (
	PolicyEffectAllow  PolicyEffect = "allow"
	PolicyEffectDeny   PolicyEffect = "deny"
)

// Policy represents an access control policy
type Policy struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Effect      PolicyEffect           `json:"effect"`
	Resources   []string               `json:"resources"`
	Actions     []string               `json:"actions"`
	Subjects    []string               `json:"subjects,omitempty"`    // Roles, users, groups
	Conditions  map[string]interface{} `json:"conditions,omitempty"`  // Rule conditions
	Priority    int                    `json:"priority"`              // Higher priority = evaluated first
	Version     int                    `json:"version"`
	IsActive    bool                   `json:"is_active"`
	Metadata    map[string]string      `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	CreatedBy   string                 `json:"created_by"`
}

// PolicyVersion represents a historical version of a policy
type PolicyVersion struct {
	ID          string                 `json:"id"`
	PolicyID    string                 `json:"policy_id"`
	Version     int                    `json:"version"`
	PolicyData  *Policy                `json:"policy_data"`
	ChangeType  string                 `json:"change_type"` // created, updated, deleted
	ChangedBy   string                 `json:"changed_by"`
	ChangedAt   time.Time              `json:"changed_at"`
	Reason      string                 `json:"reason,omitempty"`
}

// PolicyCreateRequest represents a request to create a new policy
type PolicyCreateRequest struct {
	Name       string                 `json:"name" binding:"required"`
	Description string                `json:"description"`
	Effect     string                 `json:"effect" binding:"required,oneof=allow deny"`
	Resources  []string               `json:"resources" binding:"required"`
	Actions    []string               `json:"actions" binding:"required"`
	Subjects   []string               `json:"subjects,omitempty"`
	Conditions map[string]interface{} `json:"conditions,omitempty"`
	Priority   int                    `json:"priority"`
	Metadata   map[string]string      `json:"metadata,omitempty"`
}

// PolicyUpdateRequest represents a request to update a policy
type PolicyUpdateRequest struct {
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	Effect      string                 `json:"effect,omitempty,oneof=allow deny"`
	Resources   []string               `json:"resources,omitempty"`
	Actions     []string               `json:"actions,omitempty"`
	Subjects    []string               `json:"subjects,omitempty"`
	Conditions  map[string]interface{} `json:"conditions,omitempty"`
	Priority    int                    `json:"priority,omitempty"`
	Metadata    map[string]string      `json:"metadata,omitempty"`
	Reason      string                 `json:"reason,omitempty"`
}

// AccessCheckRequest represents an access control check request
type AccessCheckRequest struct {
	Subject   Subject  `json:"subject" binding:"required"`
	Action    string   `json:"action" binding:"required"`
	Resource  Resource `json:"resource" binding:"required"`
	Context   Context  `json:"context,omitempty"`
}

// Subject represents the entity requesting access
type Subject struct {
	ID    string   `json:"id"`
	Type  string   `json:"type"` // user, service, api_key
	Roles []string `json:"roles"`
	Attrs map[string]string `json:"attrs,omitempty"`
}

// Resource represents the resource being accessed
type Resource struct {
	Type string `json:"type"`
	ID   string `json:"id,omitempty"`
	Path string `json:"path,omitempty"`
}

// Context represents additional context for access decision
type Context struct {
	Time          *time.Time `json:"time,omitempty"`
	IPAddress     string     `json:"ip_address,omitempty"`
	UserAgent     string     `json:"user_agent,omitempty"`
	RequestMethod string     `json:"request_method,omitempty"`
	Environment   string     `json:"environment,omitempty"`
	Custom        map[string]interface{} `json:"custom,omitempty"`
}

// AccessCheckResponse represents the result of an access check
type AccessCheckResponse struct {
	Allowed    bool     `json:"allowed"`
	Reason     string   `json:"reason"`
	PolicyID   string   `json:"policy_id,omitempty"`
	PolicyName string   `json:"policy_name,omitempty"`
	MatchedOn  []string `json:"matched_on,omitempty"`
	CheckedAt  time.Time `json:"checked_at"`
	CacheHit   bool     `json:"cache_hit,omitempty"`
}

// PolicyListResponse represents a list of policies with pagination
type PolicyListResponse struct {
	Policies   []*Policy `json:"policies"`
	TotalCount int64     `json:"total_count"`
	Page       int       `json:"page"`
	PageSize   int       `json:"page_size"`
	TotalPages int       `json:"total_pages"`
}

// PolicyHistoryResponse represents policy version history
type PolicyHistoryResponse struct {
	PolicyID  string           `json:"policy_id"`
	PolicyName string          `json:"policy_name"`
	Versions  []*PolicyVersion `json:"versions"`
	TotalCount int64           `json:"total_count"`
}

// ConditionOperator defines supported condition operators
type ConditionOperator string

const (
	CondEquals       ConditionOperator = "eq"
	CondNotEquals    ConditionOperator = "neq"
	CondIn           ConditionOperator = "in"
	CondNotIn        ConditionOperator = "not_in"
	CondGreaterThan  ConditionOperator = "gt"
	CondLessThan     ConditionOperator = "lt"
	CondGreaterEqual ConditionOperator = "gte"
	CondLessEqual    ConditionOperator = "lte"
	CondContains     ConditionOperator = "contains"
	CondRegex        ConditionOperator = "regex"
	CondTimeBetween  ConditionOperator = "time_between"
	CondIPInRange    ConditionOperator = "ip_in_range"
)

// Condition represents a single policy condition
type Condition struct {
	Field    string            `json:"field"`
	Operator ConditionOperator `json:"operator"`
	Value    interface{}       `json:"value"`
}

// PolicyTemplate represents a reusable policy template
type PolicyTemplate struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Resources   []string `json:"resources"`
	Actions     []string `json:"actions"`
	Template    string   `json:"template"` // Template expression
}
