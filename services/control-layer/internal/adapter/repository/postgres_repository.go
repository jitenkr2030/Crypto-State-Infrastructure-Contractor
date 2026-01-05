package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"control-layer/internal/core/domain"
	"control-layer/internal/core/ports"
)

// PostgresPolicyRepository implements ports.PolicyRepository for PostgreSQL
type PostgresPolicyRepository struct {
	db     *sql.DB
	logger ports.Logger
}

// NewPostgresPolicyRepository creates a new PostgresPolicyRepository
func NewPostgresPolicyRepository(db *sql.DB, logger ports.Logger) *PostgresPolicyRepository {
	return &PostgresPolicyRepository{
		db:     db,
		logger: logger,
	}
}

// CreatePolicy creates a new policy in the database
func (r *PostgresPolicyRepository) CreatePolicy(ctx context.Context, policy *domain.Policy) error {
	query := `
		INSERT INTO policies (
			id, name, description, effect, resources, actions, subjects,
			conditions, priority, version, is_active, metadata, created_at, updated_at, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	resourcesJSON, _ := json.Marshal(policy.Resources)
	actionsJSON, _ := json.Marshal(policy.Actions)
	subjectsJSON, _ := json.Marshal(policy.Subjects)
	conditionsJSON, _ := json.Marshal(policy.Conditions)
	metadataJSON, _ := json.Marshal(policy.Metadata)

	_, err := r.db.ExecContext(ctx, query,
		policy.ID,
		policy.Name,
		policy.Description,
		policy.Effect,
		resourcesJSON,
		actionsJSON,
		subjectsJSON,
		conditionsJSON,
		policy.Priority,
		policy.Version,
		policy.IsActive,
		metadataJSON,
		policy.CreatedAt,
		policy.UpdatedAt,
		policy.CreatedBy,
	)

	if err != nil {
		r.logger.Error("Failed to create policy", "error", err, "id", policy.ID)
		return fmt.Errorf("failed to create policy: %w", err)
	}

	return nil
}

// GetPolicy retrieves a policy by ID
func (r *PostgresPolicyRepository) GetPolicy(ctx context.Context, id string) (*domain.Policy, error) {
	query := `
		SELECT id, name, description, effect, resources, actions, subjects,
			   conditions, priority, version, is_active, metadata, created_at, updated_at, created_by
		FROM policies
		WHERE id = $1
	`

	policy := &domain.Policy{}
	var resourcesJSON, actionsJSON, subjectsJSON, conditionsJSON, metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&policy.ID,
		&policy.Name,
		&policy.Description,
		&policy.Effect,
		&resourcesJSON,
		&actionsJSON,
		&subjectsJSON,
		&conditionsJSON,
		&policy.Priority,
		&policy.Version,
		&policy.IsActive,
		&metadataJSON,
		&policy.CreatedAt,
		&policy.UpdatedAt,
		&policy.CreatedBy,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("policy not found")
	}
	if err != nil {
		r.logger.Error("Failed to get policy", "error", err, "id", id)
		return nil, fmt.Errorf("failed to get policy: %w", err)
	}

	json.Unmarshal(resourcesJSON, &policy.Resources)
	json.Unmarshal(actionsJSON, &policy.Actions)
	json.Unmarshal(subjectsJSON, &policy.Subjects)
	json.Unmarshal(conditionsJSON, &policy.Conditions)
	json.Unmarshal(metadataJSON, &policy.Metadata)

	return policy, nil
}

// UpdatePolicy updates an existing policy
func (r *PostgresPolicyRepository) UpdatePolicy(ctx context.Context, policy *domain.Policy) error {
	query := `
		UPDATE policies
		SET name = $1, description = $2, effect = $3, resources = $4, actions = $5,
			subjects = $6, conditions = $7, priority = $8, version = $9,
			is_active = $10, metadata = $11, updated_at = $12
		WHERE id = $13
	`

	resourcesJSON, _ := json.Marshal(policy.Resources)
	actionsJSON, _ := json.Marshal(policy.Actions)
	subjectsJSON, _ := json.Marshal(policy.Subjects)
	conditionsJSON, _ := json.Marshal(policy.Conditions)
	metadataJSON, _ := json.Marshal(policy.Metadata)

	result, err := r.db.ExecContext(ctx, query,
		policy.Name,
		policy.Description,
		policy.Effect,
		resourcesJSON,
		actionsJSON,
		subjectsJSON,
		conditionsJSON,
		policy.Priority,
		policy.Version,
		policy.IsActive,
		metadataJSON,
		time.Now().UTC(),
		policy.ID,
	)

	if err != nil {
		r.logger.Error("Failed to update policy", "error", err, "id", policy.ID)
		return fmt.Errorf("failed to update policy: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("policy not found")
	}

	return nil
}

// DeletePolicy deletes a policy
func (r *PostgresPolicyRepository) DeletePolicy(ctx context.Context, id string) error {
	query := `DELETE FROM policies WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("Failed to delete policy", "error", err, "id", id)
		return fmt.Errorf("failed to delete policy: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("policy not found")
	}

	return nil
}

// ListPolicies lists all policies with pagination
func (r *PostgresPolicyRepository) ListPolicies(ctx context.Context, activeOnly bool, page, pageSize int) ([]*domain.Policy, int64, error) {
	offset := (page - 1) * pageSize

	var whereClause string
	var args []interface{}

	if activeOnly {
		whereClause = "WHERE is_active = true"
	}

	// Get total count
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM policies %s`, whereClause)
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count policies: %w", err)
	}

	// Get policy list
	query := fmt.Sprintf(`
		SELECT id, name, description, effect, resources, actions, subjects,
			   conditions, priority, version, is_active, metadata, created_at, updated_at, created_by
		FROM policies %s
		ORDER BY priority DESC, created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, len(args)+1, len(args)+2)

	args = append(args, pageSize, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to list policies", "error", err)
		return nil, 0, fmt.Errorf("failed to list policies: %w", err)
	}
	defer rows.Close()

	var policies []*domain.Policy
	for rows.Next() {
		policy := &domain.Policy{}
		var resourcesJSON, actionsJSON, subjectsJSON, conditionsJSON, metadataJSON []byte

		err := rows.Scan(
			&policy.ID,
			&policy.Name,
			&policy.Description,
			&policy.Effect,
			&resourcesJSON,
			&actionsJSON,
			&subjectsJSON,
			&conditionsJSON,
			&policy.Priority,
			&policy.Version,
			&policy.IsActive,
			&metadataJSON,
			&policy.CreatedAt,
			&policy.UpdatedAt,
			&policy.CreatedBy,
		)
		if err != nil {
			r.logger.Error("Failed to scan policy", "error", err)
			continue
		}

		json.Unmarshal(resourcesJSON, &policy.Resources)
		json.Unmarshal(actionsJSON, &policy.Actions)
		json.Unmarshal(subjectsJSON, &policy.Subjects)
		json.Unmarshal(conditionsJSON, &policy.Conditions)
		json.Unmarshal(metadataJSON, &policy.Metadata)
		policies = append(policies, policy)
	}

	return policies, total, nil
}

// CreatePolicyVersion creates a policy version record
func (r *PostgresPolicyRepository) CreatePolicyVersion(ctx context.Context, version *domain.PolicyVersion) error {
	query := `
		INSERT INTO policy_versions (
			id, policy_id, version, policy_data, change_type, changed_by, changed_at, reason
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	policyDataJSON, _ := json.Marshal(version.PolicyData)

	_, err := r.db.ExecContext(ctx, query,
		version.ID,
		version.PolicyID,
		version.Version,
		policyDataJSON,
		version.ChangeType,
		version.ChangedBy,
		version.ChangedAt,
		version.Reason,
	)

	if err != nil {
		r.logger.Error("Failed to create policy version", "error", err, "policyId", version.PolicyID)
		return fmt.Errorf("failed to create policy version: %w", err)
	}

	return nil
}

// GetPolicyHistory retrieves policy version history
func (r *PostgresPolicyRepository) GetPolicyHistory(ctx context.Context, policyID string, page, pageSize int) ([]*domain.PolicyVersion, int64, error) {
	offset := (page - 1) * pageSize

	// Get total count
	countQuery := `SELECT COUNT(*) FROM policy_versions WHERE policy_id = $1`
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, policyID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count policy versions: %w", err)
	}

	// Get version history
	query := `
		SELECT id, policy_id, version, policy_data, change_type, changed_by, changed_at, reason
		FROM policy_versions
		WHERE policy_id = $1
		ORDER BY changed_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, policyID, pageSize, offset)
	if err != nil {
		r.logger.Error("Failed to get policy history", "error", err, "policyId", policyID)
		return nil, 0, fmt.Errorf("failed to get policy history: %w", err)
	}
	defer rows.Close()

	var versions []*domain.PolicyVersion
	for rows.Next() {
		version := &domain.PolicyVersion{}
		var policyDataJSON []byte

		err := rows.Scan(
			&version.ID,
			&version.PolicyID,
			&version.Version,
			&policyDataJSON,
			&version.ChangeType,
			&version.ChangedBy,
			&version.ChangedAt,
			&version.Reason,
		)
		if err != nil {
			r.logger.Error("Failed to scan policy version", "error", err)
			continue
		}

		json.Unmarshal(policyDataJSON, &version.PolicyData)
		versions = append(versions, version)
	}

	return versions, total, nil
}

// GetPolicyVersion retrieves a specific version of a policy
func (r *PostgresPolicyRepository) GetPolicyVersion(ctx context.Context, policyID string, version int) (*domain.PolicyVersion, error) {
	query := `
		SELECT id, policy_id, version, policy_data, change_type, changed_by, changed_at, reason
		FROM policy_versions
		WHERE policy_id = $1 AND version = $2
	`

	versionRecord := &domain.PolicyVersion{}
	var policyDataJSON []byte

	err := r.db.QueryRowContext(ctx, query, policyID, version).Scan(
		&versionRecord.ID,
		&versionRecord.PolicyID,
		&versionRecord.Version,
		&policyDataJSON,
		&versionRecord.ChangeType,
		&versionRecord.ChangedBy,
		&versionRecord.ChangedAt,
		&versionRecord.Reason,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("policy version not found")
	}
	if err != nil {
		r.logger.Error("Failed to get policy version", "error", err, "policyId", policyID, "version", version)
		return nil, fmt.Errorf("failed to get policy version: %w", err)
	}

	json.Unmarshal(policyDataJSON, &versionRecord.PolicyData)
	return versionRecord, nil
}

// FindApplicablePolicies finds policies applicable to a resource and action
func (r *PostgresPolicyRepository) FindApplicablePolicies(ctx context.Context, resource, action string) ([]*domain.Policy, error) {
	query := `
		SELECT id, name, description, effect, resources, actions, subjects,
			   conditions, priority, version, is_active, metadata, created_at, updated_at, created_by
		FROM policies
		WHERE is_active = true
		  AND (resources @> $1::jsonb OR resources @> '["*"]'::jsonb)
		  AND (actions @> $2::jsonb OR actions @> '["*"]'::jsonb)
		ORDER BY priority DESC
	`

	resourcesJSON, _ := json.Marshal([]string{resource})
	actionsJSON, _ := json.Marshal([]string{action})

	rows, err := r.db.QueryContext(ctx, query, resourcesJSON, actionsJSON)
	if err != nil {
		r.logger.Error("Failed to find applicable policies", "error", err)
		return nil, fmt.Errorf("failed to find applicable policies: %w", err)
	}
	defer rows.Close()

	var policies []*domain.Policy
	for rows.Next() {
		policy := &domain.Policy{}
		var resourcesJSON, actionsJSON, subjectsJSON, conditionsJSON, metadataJSON []byte

		err := rows.Scan(
			&policy.ID,
			&policy.Name,
			&policy.Description,
			&policy.Effect,
			&resourcesJSON,
			&actionsJSON,
			&subjectsJSON,
			&conditionsJSON,
			&policy.Priority,
			&policy.Version,
			&policy.IsActive,
			&metadataJSON,
			&policy.CreatedAt,
			&policy.UpdatedAt,
			&policy.CreatedBy,
		)
		if err != nil {
			r.logger.Error("Failed to scan policy", "error", err)
			continue
		}

		json.Unmarshal(resourcesJSON, &policy.Resources)
		json.Unmarshal(actionsJSON, &policy.Actions)
		json.Unmarshal(subjectsJSON, &policy.Subjects)
		json.Unmarshal(conditionsJSON, &policy.Conditions)
		json.Unmarshal(metadataJSON, &policy.Metadata)
		policies = append(policies, policy)
	}

	return policies, nil
}

// HealthCheck checks database connectivity
func (r *PostgresPolicyRepository) HealthCheck(ctx context.Context) error {
	return r.db.PingContext(ctx)
}
