package ports

import (
	"context"

	"control-layer/internal/core/domain"
)

// Logger defines the logging interface
type Logger interface {
	Info(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Debug(msg string, keysAndValues ...interface{})
}

// PolicyRepository defines the interface for policy persistence
type PolicyRepository interface {
	// Policy CRUD
	CreatePolicy(ctx context.Context, policy *domain.Policy) error
	GetPolicy(ctx context.Context, id string) (*domain.Policy, error)
	UpdatePolicy(ctx context.Context, policy *domain.Policy) error
	DeletePolicy(ctx context.Context, id string) error
	ListPolicies(ctx context.Context, activeOnly bool, page, pageSize int) ([]*domain.Policy, int64, error)

	// Version history
	CreatePolicyVersion(ctx context.Context, version *domain.PolicyVersion) error
	GetPolicyHistory(ctx context.Context, policyID string, page, pageSize int) ([]*domain.PolicyVersion, int64, error)
	GetPolicyVersion(ctx context.Context, policyID string, version int) (*domain.PolicyVersion, error)

	// Query
	FindApplicablePolicies(ctx context.Context, resource, action string) ([]*domain.Policy, error)

	// Health
	HealthCheck(ctx context.Context) error
}

// PolicyService defines the interface for policy business logic
type PolicyService interface {
	// Policy management
	CreatePolicy(ctx context.Context, request domain.PolicyCreateRequest, createdBy string) (*domain.Policy, error)
	UpdatePolicy(ctx context.Context, id string, request domain.PolicyUpdateRequest, updatedBy string) (*domain.Policy, error)
	DeletePolicy(ctx context.Context, id string, deletedBy, reason string) error
	GetPolicy(ctx context.Context, id string) (*domain.Policy, error)
	ListPolicies(ctx context.Context, activeOnly bool, page, pageSize int) (*domain.PolicyListResponse, error)

	// Version management
	GetPolicyHistory(ctx context.Context, id string, page, pageSize int) (*domain.PolicyHistoryResponse, error)
	RestorePolicyVersion(ctx context.Context, policyID string, version int, restoredBy, reason string) (*domain.Policy, error)

	// Access control
	CheckAccess(ctx context.Context, request domain.AccessCheckRequest) (*domain.AccessCheckResponse, error)
	BulkCheckAccess(ctx context.Context, requests []domain.AccessCheckRequest) ([]*domain.AccessCheckResponse, error)

	// Templates
	GetPolicyTemplates(ctx context.Context) ([]*domain.PolicyTemplate, error)
	ApplyPolicyTemplate(ctx context.Context, templateID string, name string, createdBy string) (*domain.Policy, error)

	// Health
	HealthCheck(ctx context.Context) (bool, error)
}

// CacheClient defines the interface for caching access decisions
type CacheClient interface {
	Get(ctx context.Context, key string) (*domain.AccessCheckResponse, bool)
	Set(ctx context.Context, key string, response *domain.AccessCheckResponse, ttl int) error
	Delete(ctx context.Context, key string) error
	InvalidateByPattern(ctx context.Context, pattern string) error
	Close() error
}

// MessagingClient defines the interface for event publishing
type MessagingClient interface {
	PublishPolicyCreated(ctx context.Context, policy *domain.Policy) error
	PublishPolicyUpdated(ctx context.Context, policy *domain.Policy, version int) error
	PublishPolicyDeleted(ctx context.Context, policyID string, name string) error
	PublishAccessDecision(ctx context.Context, request domain.AccessCheckRequest, response *domain.AccessCheckResponse) error
	Close() error
}
