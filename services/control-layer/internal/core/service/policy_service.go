package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"control-layer/internal/core/domain"
	"control-layer/internal/core/ports"
)

//go:generate mockgen -destination=../../../mocks/mock_policy_service.go -package=mocks . PolicyService

var (
	// ErrPolicyNotFound is returned when a policy is not found
	ErrPolicyNotFound = errors.New("policy not found")

	// ErrPolicyVersionNotFound is returned when a policy version is not found
	ErrPolicyVersionNotFound = errors.New("policy version not found")

	// ErrPolicyConflict is returned when there's a policy conflict
	ErrPolicyConflict = errors.New("policy conflict")

	// ErrInvalidEffect is returned when the policy effect is invalid
	ErrInvalidEffect = errors.New("invalid policy effect")

	// ErrAccessDenied is returned when access is denied
	ErrAccessDenied = errors.New("access denied")
)

// PolicyServiceImpl implements the PolicyService interface
type PolicyServiceImpl struct {
	repo     ports.PolicyRepository
	cache    ports.CacheClient
	messaging ports.MessagingClient
	logger   ports.Logger
}

// NewPolicyService creates a new PolicyServiceImpl
func NewPolicyService(repo ports.PolicyRepository, cache ports.CacheClient, messaging ports.MessagingClient, logger ports.Logger) *PolicyServiceImpl {
	return &PolicyServiceImpl{
		repo:     repo,
		cache:    cache,
		messaging: messaging,
		logger:   logger,
	}
}

// CreatePolicy creates a new policy
func (s *PolicyServiceImpl) CreatePolicy(ctx context.Context, request domain.PolicyCreateRequest, createdBy string) (*domain.Policy, error) {
	// Validate effect
	if request.Effect != "allow" && request.Effect != "deny" {
		return nil, ErrInvalidEffect
	}

	policy := &domain.Policy{
		ID:          generatePolicyID(),
		Name:        request.Name,
		Description: request.Description,
		Effect:      domain.PolicyEffect(request.Effect),
		Resources:   request.Resources,
		Actions:     request.Actions,
		Subjects:    request.Subjects,
		Conditions:  request.Conditions,
		Priority:    request.Priority,
		Version:     1,
		IsActive:    true,
		Metadata:    request.Metadata,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
		CreatedBy:   createdBy,
	}

	// Save to repository
	if err := s.repo.CreatePolicy(ctx, policy); err != nil {
		s.logger.Error("Failed to create policy", "error", err, "name", request.Name)
		return nil, fmt.Errorf("failed to create policy: %w", err)
	}

	// Create initial version history
	version := &domain.PolicyVersion{
		ID:         generateVersionID(),
		PolicyID:   policy.ID,
		Version:    1,
		PolicyData: policy,
		ChangeType: "created",
		ChangedBy:  createdBy,
		ChangedAt:  time.Now().UTC(),
		Reason:     "Initial policy creation",
	}

	if err := s.repo.CreatePolicyVersion(ctx, version); err != nil {
		s.logger.Warn("Failed to create policy version", "error", err)
	}

	// Publish event
	if s.messaging != nil {
		if err := s.messaging.PublishPolicyCreated(ctx, policy); err != nil {
			s.logger.Warn("Failed to publish policy created event", "error", err)
		}
	}

	s.logger.Info("Policy created", "id", policy.ID, "name", policy.Name, "effect", policy.Effect)
	return policy, nil
}

// UpdatePolicy updates an existing policy
func (s *PolicyServiceImpl) UpdatePolicy(ctx context.Context, id string, request domain.PolicyUpdateRequest, updatedBy string) (*domain.Policy, error) {
	// Get existing policy
	policy, err := s.repo.GetPolicy(ctx, id)
	if err != nil {
		if errors.Is(err, ErrPolicyNotFound) {
			return nil, ErrPolicyNotFound
		}
		return nil, err
	}

	// Update fields
	if request.Name != "" {
		policy.Name = request.Name
	}
	if request.Description != "" {
		policy.Description = request.Description
	}
	if request.Effect != "" {
		if request.Effect != "allow" && request.Effect != "deny" {
			return nil, ErrInvalidEffect
		}
		policy.Effect = domain.PolicyEffect(request.Effect)
	}
	if len(request.Resources) > 0 {
		policy.Resources = request.Resources
	}
	if len(request.Actions) > 0 {
		policy.Actions = request.Actions
	}
	if len(request.Subjects) > 0 {
		policy.Subjects = request.Subjects
	}
	if request.Conditions != nil {
		policy.Conditions = request.Conditions
	}
	if request.Priority != 0 {
		policy.Priority = request.Priority
	}
	if request.Metadata != nil {
		policy.Metadata = request.Metadata
	}

	// Increment version
	policy.Version++
	policy.UpdatedAt = time.Now().UTC()

	// Save to repository
	if err := s.repo.UpdatePolicy(ctx, policy); err != nil {
		s.logger.Error("Failed to update policy", "error", err, "id", id)
		return nil, fmt.Errorf("failed to update policy: %w", err)
	}

	// Create version history
	version := &domain.PolicyVersion{
		ID:         generateVersionID(),
		PolicyID:   policy.ID,
		Version:    policy.Version,
		PolicyData: policy,
		ChangeType: "updated",
		ChangedBy:  updatedBy,
		ChangedAt:  time.Now().UTC(),
		Reason:     request.Reason,
	}

	if err := s.repo.CreatePolicyVersion(ctx, version); err != nil {
		s.logger.Warn("Failed to create policy version", "error", err)
	}

	// Invalidate cache
	if s.cache != nil {
		s.cache.InvalidateByPattern(ctx, fmt.Sprintf("check:%s:*", policy.ID))
	}

	// Publish event
	if s.messaging != nil {
		if err := s.messaging.PublishPolicyUpdated(ctx, policy, policy.Version); err != nil {
			s.logger.Warn("Failed to publish policy updated event", "error", err)
		}
	}

	s.logger.Info("Policy updated", "id", policy.ID, "version", policy.Version)
	return policy, nil
}

// DeletePolicy deletes a policy
func (s *PolicyServiceImpl) DeletePolicy(ctx context.Context, id string, deletedBy, reason string) error {
	policy, err := s.repo.GetPolicy(ctx, id)
	if err != nil {
		return err
	}

	// Delete from repository
	if err := s.repo.DeletePolicy(ctx, id); err != nil {
		s.logger.Error("Failed to delete policy", "error", err, "id", id)
		return fmt.Errorf("failed to delete policy: %w", err)
	}

	// Create version history for deletion
	version := &domain.PolicyVersion{
		ID:         generateVersionID(),
		PolicyID:   id,
		Version:    policy.Version + 1,
		PolicyData: policy,
		ChangeType: "deleted",
		ChangedBy:  deletedBy,
		ChangedAt:  time.Now().UTC(),
		Reason:     reason,
	}

	if err := s.repo.CreatePolicyVersion(ctx, version); err != nil {
		s.logger.Warn("Failed to create policy version", "error", err)
	}

	// Invalidate cache
	if s.cache != nil {
		s.cache.InvalidateByPattern(ctx, fmt.Sprintf("check:%s:*", id))
	}

	// Publish event
	if s.messaging != nil {
		if err := s.messaging.PublishPolicyDeleted(ctx, id, policy.Name); err != nil {
			s.logger.Warn("Failed to publish policy deleted event", "error", err)
		}
	}

	s.logger.Info("Policy deleted", "id", id, "deletedBy", deletedBy)
	return nil
}

// GetPolicy retrieves a policy by ID
func (s *PolicyServiceImpl) GetPolicy(ctx context.Context, id string) (*domain.Policy, error) {
	return s.repo.GetPolicy(ctx, id)
}

// ListPolicies lists all policies
func (s *PolicyServiceImpl) ListPolicies(ctx context.Context, activeOnly bool, page, pageSize int) (*domain.PolicyListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	policies, total, err := s.repo.ListPolicies(ctx, activeOnly, page, pageSize)
	if err != nil {
		s.logger.Error("Failed to list policies", "error", err)
		return nil, fmt.Errorf("failed to list policies: %w", err)
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &domain.PolicyListResponse{
		Policies:   policies,
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetPolicyHistory retrieves policy version history
func (s *PolicyServiceImpl) GetPolicyHistory(ctx context.Context, id string, page, pageSize int) (*domain.PolicyHistoryResponse, error) {
	policy, err := s.repo.GetPolicy(ctx, id)
	if err != nil {
		return nil, err
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	versions, total, err := s.repo.GetPolicyHistory(ctx, id, page, pageSize)
	if err != nil {
		s.logger.Error("Failed to get policy history", "error", err, "id", id)
		return nil, fmt.Errorf("failed to get policy history: %w", err)
	}

	return &domain.PolicyHistoryResponse{
		PolicyID:   id,
		PolicyName: policy.Name,
		Versions:   versions,
		TotalCount: total,
	}, nil
}

// RestorePolicyVersion restores a policy to a previous version
func (s *PolicyServiceImpl) RestorePolicyVersion(ctx context.Context, policyID string, version int, restoredBy, reason string) (*domain.Policy, error) {
	// Get the version to restore
	oldVersion, err := s.repo.GetPolicyVersion(ctx, policyID, version)
	if err != nil {
		return nil, err
	}

	if oldVersion.PolicyData == nil {
		return nil, ErrPolicyVersionNotFound
	}

	// Create update request with old version data
	updateReq := domain.PolicyUpdateRequest{
		Name:        oldVersion.PolicyData.Name,
		Description: oldVersion.PolicyData.Description,
		Effect:      string(oldVersion.PolicyData.Effect),
		Resources:   oldVersion.PolicyData.Resources,
		Actions:     oldVersion.PolicyData.Actions,
		Subjects:    oldVersion.PolicyData.Subjects,
		Conditions:  oldVersion.PolicyData.Conditions,
		Priority:    oldVersion.PolicyData.Priority,
		Metadata:    oldVersion.PolicyData.Metadata,
		Reason:      fmt.Sprintf("Restored from version %d: %s", version, reason),
	}

	return s.UpdatePolicy(ctx, policyID, updateReq, restoredBy)
}

// CheckAccess performs an access control check
func (s *PolicyServiceImpl) CheckAccess(ctx context.Context, request domain.AccessCheckRequest) (*domain.AccessCheckResponse, error) {
	// Generate cache key
	cacheKey := generateCacheKey(request)
	if s.cache != nil {
		if cached, found := s.cache.Get(ctx, cacheKey); found {
			cached.CacheHit = true
			s.logger.Debug("Cache hit", "key", cacheKey)
			return cached, nil
		}
	}

	// Find applicable policies
	policies, err := s.repo.FindApplicablePolicies(ctx, request.Resource.Type, request.Action)
	if err != nil {
		s.logger.Error("Failed to find applicable policies", "error", err)
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	if len(policies) == 0 {
		// Default deny if no policies match
		response := &domain.AccessCheckResponse{
			Allowed:   false,
			Reason:    "No applicable policies found - access denied by default",
			CheckedAt: time.Now().UTC(),
		}

		// Cache negative result
		if s.cache != nil {
			s.cache.Set(ctx, cacheKey, response, 60)
		}

		return response, nil
	}

	// Sort by priority (higher first)
	sortPolicies(policies)

	// Evaluate policies
	var matchedPolicy *domain.Policy
	var matchedOn []string

	for _, policy := range policies {
		if !policy.IsActive {
			continue
		}

		// Check if policy applies to this subject
		if !s.matchesSubject(policy, request.Subject) {
			continue
		}

		// Check conditions
		if !s.evaluateConditions(policy.Conditions, request) {
			continue
		}

		matchedPolicy = policy
		matchedOn = []string{
			fmt.Sprintf("resource:%s", request.Resource.Type),
			fmt.Sprintf("action:%s", request.Action),
		}

		break
	}

	response := &domain.AccessCheckResponse{
		CheckedAt: time.Now().UTC(),
	}

	if matchedPolicy != nil {
		response.Allowed = matchedPolicy.Effect == domain.PolicyEffectAllow
		response.Reason = fmt.Sprintf("Matched policy: %s (version %d)", matchedPolicy.Name, matchedPolicy.Version)
		response.PolicyID = matchedPolicy.ID
		response.PolicyName = matchedPolicy.Name
		response.MatchedOn = matchedOn
	} else {
		response.Allowed = false
		response.Reason = "No matching policy found - access denied by default"
	}

	// Cache the result
	if s.cache != nil {
		s.cache.Set(ctx, cacheKey, response, 60)
	}

	// Publish access decision event
	if s.messaging != nil {
		s.messaging.PublishAccessDecision(ctx, request, response)
	}

	s.logger.Info("Access check completed",
		"allowed", response.Allowed,
		"policy", response.PolicyName,
		"resource", request.Resource.Type,
		"action", request.Action,
		"subject", request.Subject.ID,
	)

	return response, nil
}

// BulkCheckAccess performs multiple access checks
func (s *PolicyServiceImpl) BulkCheckAccess(ctx context.Context, requests []domain.AccessCheckRequest) ([]*domain.AccessCheckResponse, error) {
	responses := make([]*domain.AccessCheckResponse, len(requests))

	for i, request := range requests {
		response, err := s.CheckAccess(ctx, request)
		if err != nil {
			return nil, fmt.Errorf("bulk check failed at index %d: %w", i, err)
		}
		responses[i] = response
	}

	return responses, nil
}

// GetPolicyTemplates returns available policy templates
func (s *PolicyServiceImpl) GetPolicyTemplates(ctx context.Context) ([]*domain.PolicyTemplate, error) {
	templates := []*domain.PolicyTemplate{
		{
			ID:          "read-only",
			Name:        "Read-Only Access",
			Description: "Grants read-only access to resources",
			Resources:   []string{"*"},
			Actions:     []string{"read", "list"},
			Template:    "allow read access to all resources",
		},
		{
			ID:          "admin-access",
			Name:        "Full Admin Access",
			Description: "Grants full admin access",
			Resources:   []string{"*"},
			Actions:     []string{"*"},
			Template:    "allow all actions for admin role",
		},
		{
			ID:          "restricted-delete",
			Name:        "Restricted Delete",
			Description: "Allows delete only during business hours",
			Resources:   []string{"*"},
			Actions:     []string{"delete"},
			Template:    "allow delete only when time is between 9:00 and 17:00",
		},
	}

	return templates, nil
}

// ApplyPolicyTemplate applies a policy template
func (s *PolicyServiceImpl) ApplyPolicyTemplate(ctx context.Context, templateID string, name string, createdBy string) (*domain.Policy, error) {
	templates, _ := s.GetPolicyTemplates(ctx)

	var template *domain.PolicyTemplate
	for _, t := range templates {
		if t.ID == templateID {
			template = t
			break
		}
	}

	if template == nil {
		return nil, fmt.Errorf("template not found: %s", templateID)
	}

	request := domain.PolicyCreateRequest{
		Name:        name,
		Description: fmt.Sprintf("Created from template: %s", template.Name),
		Effect:      "allow",
		Resources:   template.Resources,
		Actions:     template.Actions,
		Priority:    100,
	}

	return s.CreatePolicy(ctx, request, createdBy)
}

// HealthCheck checks the health of the service
func (s *PolicyServiceImpl) HealthCheck(ctx context.Context) (bool, error) {
	if err := s.repo.HealthCheck(ctx); err != nil {
		s.logger.Error("Health check failed", "error", err)
		return false, err
	}
	return true, nil
}

// Helper functions

func generatePolicyID() string {
	return fmt.Sprintf("pol_%d", time.Now().UnixNano())
}

func generateVersionID() string {
	return fmt.Sprintf("ver_%d", time.Now().UnixNano())
}

func generateCacheKey(request domain.AccessCheckRequest) string {
	// Create deterministic cache key
	data := struct {
		SubjectID  string
		SubjectType string
		Roles      []string
		Resource   string
		ResourceID string
		Action     string
	}{
		SubjectID:  request.Subject.ID,
		SubjectType: request.Subject.Type,
		Roles:      request.Subject.Roles,
		Resource:   request.Resource.Type,
		ResourceID: request.Resource.ID,
		Action:     request.Action,
	}
	jsonBytes, _ := json.Marshal(data)
	return fmt.Sprintf("check:%s", string(jsonBytes))
}

func sortPolicies(policies []*domain.Policy) {
	// Sort by priority descending
	for i := 0; i < len(policies); i++ {
		for j := i + 1; j < len(policies); j++ {
			if policies[j].Priority > policies[i].Priority {
				policies[i], policies[j] = policies[j], policies[i]
			}
		}
	}
}

func (s *PolicyServiceImpl) matchesSubject(policy *domain.Policy, subject domain.Subject) bool {
	if len(policy.Subjects) == 0 {
		return true // No subject restriction
	}

	for _, role := range policy.Subjects {
		for _, subjectRole := range subject.Roles {
			if strings.EqualFold(role, subjectRole) {
				return true
			}
		}
	}

	return false
}

func (s *PolicyServiceImpl) evaluateConditions(conditions map[string]interface{}, request domain.AccessCheckRequest) bool {
	if len(conditions) == 0 {
		return true
	}

	for field, value := range conditions {
		if !s.evaluateCondition(field, value, request) {
			return false
		}
	}

	return true
}

func (s *PolicyServiceImpl) evaluateCondition(field string, value interface{}, request domain.AccessCheckRequest) bool {
	switch field {
	case "time_start":
		// Check if current time is after start time
		if request.Context.Time != nil {
			startTime, _ := time.Parse("15:04", value.(string))
			currentTime := request.Context.Time.Format("15:04")
			if currentTime < startTime.Format("15:04") {
				return false
			}
		}
	case "time_end":
		// Check if current time is before end time
		if request.Context.Time != nil {
			endTime, _ := time.Parse("15:04", value.(string))
			currentTime := request.Context.Time.Format("15:04")
			if currentTime > endTime.Format("15:04") {
				return false
			}
		}
	case "ip_whitelist":
		// Check if IP is in whitelist
		if request.Context.IPAddress != "" {
			whitelist := value.(string)
			if !strings.Contains(whitelist, request.Context.IPAddress) {
				return false
			}
		}
	case "environment":
		// Check environment match
		if request.Context.Environment != "" && request.Context.Environment != value.(string) {
			return false
		}
	case "user_agent_pattern":
		// Regex match on user agent
		if request.Context.UserAgent != "" {
			pattern := value.(string)
			matched, _ := regexp.MatchString(pattern, request.Context.UserAgent)
			if !matched {
				return false
			}
		}
	}

	return true
}
