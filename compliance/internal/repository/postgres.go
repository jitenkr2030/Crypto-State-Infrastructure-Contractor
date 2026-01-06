package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/csic/platform/compliance/internal/domain"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host            string
	Port            int
	Username        string
	Password        string
	Database        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int
}

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	Host      string
	Port      int
	Password  string
	DB        int
	KeyPrefix string
	PoolSize  int
}

// PostgresRepository handles database operations
type PostgresRepository struct {
	db          *sql.DB
	redisClient *RedisClient
	logger      *zap.Logger
	keyPrefix   string
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(cfg DatabaseConfig, logger *zap.Logger) (*PostgresRepository, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Database, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Connected to PostgreSQL database",
		zap.String("database", cfg.Database),
		zap.String("host", cfg.Host))

	return &PostgresRepository{
		db:     db,
		logger: logger,
	}, nil
}

// NewRedisClient creates a new Redis client
func NewRedisClient(cfg RedisConfig, logger *zap.Logger) (*RedisClient, error) {
	client := &RedisClient{
		address:  fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		password: cfg.Password,
		db:       cfg.DB,
		poolSize: cfg.PoolSize,
		keyPrefix: cfg.KeyPrefix,
		logger:   logger,
	}

	if err := client.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Connected to Redis",
		zap.String("address", client.address))

	return client, nil
}

// Close closes the database connection
func (r *PostgresRepository) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}

// Close closes the Redis connection
func (c *RedisClient) Close() error {
	if c.pool != nil {
		return c.pool.Close()
	}
	return nil
}

// InitSchema initializes the database schema
func (r *PostgresRepository) InitSchema(ctx context.Context) error {
	queries := []string{
		// Rules table
		`CREATE TABLE IF NOT EXISTS rules (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			description TEXT,
			type VARCHAR(50) NOT NULL,
			severity VARCHAR(20) NOT NULL,
			enabled BOOLEAN DEFAULT true,
			version INTEGER DEFAULT 1,
			expression TEXT,
			parameters JSONB,
			expires_at TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// Rulesets table
		`CREATE TABLE IF NOT EXISTS rulesets (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			description TEXT,
			version INTEGER DEFAULT 1,
			active BOOLEAN DEFAULT false,
			rules UUID[],
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// Entities table
		`CREATE TABLE IF NOT EXISTS entities (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			type VARCHAR(50) NOT NULL,
			name VARCHAR(255) NOT NULL,
			country VARCHAR(3),
			risk_score DECIMAL(5,4) DEFAULT 0,
			tags TEXT[],
			metadata JSONB,
			blacklisted BOOLEAN DEFAULT false,
			watchlist BOOLEAN DEFAULT false,
			kyc_verified BOOLEAN DEFAULT false,
			aml_status VARCHAR(50) DEFAULT 'PENDING',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// Compliance results table
		`CREATE TABLE IF NOT EXISTS compliance_results (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			transaction_id VARCHAR(255) NOT NULL,
			overall_status VARCHAR(20) NOT NULL,
			risk_score DECIMAL(5,4) DEFAULT 0,
			checks JSONB NOT NULL,
			violations JSONB,
			summary JSONB NOT NULL,
			processing_time_ms BIGINT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// Violations table
		`CREATE TABLE IF NOT EXISTS violations (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			transaction_id VARCHAR(255) NOT NULL,
			rule_id UUID NOT NULL,
			rule_name VARCHAR(255) NOT NULL,
			severity VARCHAR(20) NOT NULL,
			status VARCHAR(20) DEFAULT 'OPEN',
			resolution TEXT,
			resolved_by VARCHAR(255),
			resolved_at TIMESTAMP WITH TIME ZONE,
			details JSONB,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// Watchlist table
		`CREATE TABLE IF NOT EXISTS watchlist (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			type VARCHAR(50) NOT NULL,
			name VARCHAR(255) NOT NULL,
			country VARCHAR(3),
			list_source VARCHAR(100) NOT NULL,
			match_score DECIMAL(5,4) DEFAULT 1.0,
			active BOOLEAN DEFAULT true,
			expires_at TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// Create indexes
		`CREATE INDEX IF NOT EXISTS idx_rules_type ON rules(type)`,
		`CREATE INDEX IF NOT EXISTS idx_rules_enabled ON rules(enabled)`,
		`CREATE INDEX IF NOT EXISTS idx_entities_risk_score ON entities(risk_score DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_entities_blacklisted ON entities(blacklisted)`,
		`CREATE INDEX IF NOT EXISTS idx_compliance_results_transaction_id ON compliance_results(transaction_id)`,
		`CREATE INDEX IF NOT EXISTS idx_compliance_results_created_at ON compliance_results(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_violations_status ON violations(status)`,
		`CREATE INDEX IF NOT EXISTS idx_violations_created_at ON violations(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_watchlist_active ON watchlist(active)`,
	}

	for _, query := range queries {
		if _, err := r.db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("failed to execute schema query: %w", err)
		}
	}

	r.logger.Info("Database schema initialized")
	return nil
}

// Rule operations

// CreateRule creates a new compliance rule
func (r *PostgresRepository) CreateRule(ctx context.Context, rule *domain.Rule) error {
	paramsJSON, err := json.Marshal(rule.Parameters)
	if err != nil {
		return fmt.Errorf("failed to marshal parameters: %w", err)
	}

	query := `
		INSERT INTO rules (id, name, description, type, severity, enabled, version, expression, parameters, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
	`

	_, err = r.db.ExecContext(ctx, query,
		rule.ID, rule.Name, rule.Description, rule.Type, rule.Severity,
		rule.Enabled, rule.Version, rule.Expression, paramsJSON, rule.ExpiresAt)

	if err != nil {
		return fmt.Errorf("failed to create rule: %w", err)
	}

	return nil
}

// GetRule retrieves a rule by ID
func (r *PostgresRepository) GetRule(ctx context.Context, id string) (*domain.Rule, error) {
	query := `
		SELECT id, name, description, type, severity, enabled, version, expression, parameters, expires_at, created_at, updated_at
		FROM rules WHERE id = $1
	`

	var rule domain.Rule
	var paramsJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&rule.ID, &rule.Name, &rule.Description, &rule.Type, &rule.Severity,
		&rule.Enabled, &rule.Version, &rule.Expression, &paramsJSON,
		&rule.ExpiresAt, &rule.CreatedAt, &rule.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get rule: %w", err)
	}

	if err := json.Unmarshal(paramsJSON, &rule.Parameters); err != nil {
		return nil, fmt.Errorf("failed to unmarshal parameters: %w", err)
	}

	return &rule, nil
}

// GetEnabledRules retrieves all enabled rules
func (r *PostgresRepository) GetEnabledRules(ctx context.Context) ([]domain.Rule, error) {
	query := `
		SELECT id, name, description, type, severity, enabled, version, expression, parameters, expires_at, created_at, updated_at
		FROM rules
		WHERE enabled = true
		AND (expires_at IS NULL OR expires_at > NOW())
		ORDER BY severity DESC, created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query rules: %w", err)
	}
	defer rows.Close()

	var rules []domain.Rule
	for rows.Next() {
		var rule domain.Rule
		var paramsJSON []byte

		err := rows.Scan(
			&rule.ID, &rule.Name, &rule.Description, &rule.Type, &rule.Severity,
			&rule.Enabled, &rule.Version, &rule.Expression, &paramsJSON,
			&rule.ExpiresAt, &rule.CreatedAt, &rule.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rule: %w", err)
		}

		if err := json.Unmarshal(paramsJSON, &rule.Parameters); err != nil {
			return nil, fmt.Errorf("failed to unmarshal parameters: %w", err)
		}

		rules = append(rules, rule)
	}

	return rules, nil
}

// GetRulesByType retrieves rules by type
func (r *PostgresRepository) GetRulesByType(ctx context.Context, ruleType domain.RuleType) ([]domain.Rule, error) {
	query := `
		SELECT id, name, description, type, severity, enabled, version, expression, parameters, expires_at, created_at, updated_at
		FROM rules
		WHERE type = $1 AND enabled = true
		AND (expires_at IS NULL OR expires_at > NOW())
		ORDER BY severity DESC
	`

	rows, err := r.db.QueryContext(ctx, query, ruleType)
	if err != nil {
		return nil, fmt.Errorf("failed to query rules: %w", err)
	}
	defer rows.Close()

	var rules []domain.Rule
	for rows.Next() {
		var rule domain.Rule
		var paramsJSON []byte

		err := rows.Scan(
			&rule.ID, &rule.Name, &rule.Description, &rule.Type, &rule.Severity,
			&rule.Enabled, &rule.Version, &rule.Expression, &paramsJSON,
			&rule.ExpiresAt, &rule.CreatedAt, &rule.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rule: %w", err)
		}

		if err := json.Unmarshal(paramsJSON, &rule.Parameters); err != nil {
			return nil, fmt.Errorf("failed to unmarshal parameters: %w", err)
		}

		rules = append(rules, rule)
	}

	return rules, nil
}

// UpdateRule updates an existing rule
func (r *PostgresRepository) UpdateRule(ctx context.Context, rule *domain.Rule) error {
	paramsJSON, err := json.Marshal(rule.Parameters)
	if err != nil {
		return fmt.Errorf("failed to marshal parameters: %w", err)
	}

	query := `
		UPDATE rules
		SET name = $1, description = $2, type = $3, severity = $4, enabled = $5,
			version = $6, expression = $7, parameters = $8, expires_at = $9, updated_at = NOW()
		WHERE id = $10
	`

	result, err := r.db.ExecContext(ctx, query,
		rule.Name, rule.Description, rule.Type, rule.Severity, rule.Enabled,
		rule.Version+1, rule.Expression, paramsJSON, rule.ExpiresAt, rule.ID)
	if err != nil {
		return fmt.Errorf("failed to update rule: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("rule not found: %s", rule.ID)
	}

	return nil
}

// Compliance result operations

// SaveComplianceResult saves a compliance check result
func (r *PostgresRepository) SaveComplianceResult(ctx context.Context, result *domain.ComplianceResult) error {
	checksJSON, err := json.Marshal(result.Checks)
	if err != nil {
		return fmt.Errorf("failed to marshal checks: %w", err)
	}

	violationsJSON, err := json.Marshal(result.Violations)
	if err != nil {
		return fmt.Errorf("failed to marshal violations: %w", err)
	}

	summaryJSON, err := json.Marshal(result.Summary)
	if err != nil {
		return fmt.Errorf("failed to marshal summary: %w", err)
	}

	query := `
		INSERT INTO compliance_results (transaction_id, overall_status, risk_score, checks, violations, summary, processing_time_ms, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
	`

	_, err = r.db.ExecContext(ctx, query,
		result.TransactionID, result.OverallStatus, result.RiskScore,
		checksJSON, violationsJSON, summaryJSON, result.ProcessingTime)

	if err != nil {
		return fmt.Errorf("failed to save compliance result: %w", err)
	}

	return nil
}

// GetComplianceResults retrieves compliance results with pagination
func (r *PostgresRepository) GetComplianceResults(ctx context.Context, limit, offset int) ([]domain.ComplianceResult, error) {
	query := `
		SELECT id, transaction_id, overall_status, risk_score, checks, violations, summary, processing_time_ms, created_at
		FROM compliance_results
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query compliance results: %w", err)
	}
	defer rows.Close()

	var results []domain.ComplianceResult
	for rows.Next() {
		var result domain.ComplianceResult
		var checksJSON, violationsJSON, summaryJSON []byte

		err := rows.Scan(
			&result.TransactionID, &result.OverallStatus, &result.RiskScore,
			&checksJSON, &violationsJSON, &summaryJSON,
			&result.ProcessingTime, &result.CheckedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan compliance result: %w", err)
		}

		if err := json.Unmarshal(checksJSON, &result.Checks); err != nil {
			return nil, fmt.Errorf("failed to unmarshal checks: %w", err)
		}
		if err := json.Unmarshal(violationsJSON, &result.Violations); err != nil {
			return nil, fmt.Errorf("failed to unmarshal violations: %w", err)
		}
		if err := json.Unmarshal(summaryJSON, &result.Summary); err != nil {
			return nil, fmt.Errorf("failed to unmarshal summary: %w", err)
		}

		results = append(results, result)
	}

	return results, nil
}

// Entity operations

// CreateEntity creates a new entity
func (r *PostgresRepository) CreateEntity(ctx context.Context, entity *domain.Entity) error {
	metadataJSON, err := json.Marshal(entity.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	tagsJSON, err := json.Marshal(entity.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	query := `
		INSERT INTO entities (id, type, name, country, risk_score, tags, metadata, blacklisted, watchlist, kyc_verified, aml_status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())
	`

	_, err = r.db.ExecContext(ctx, query,
		entity.ID, entity.Type, entity.Name, entity.Country, entity.RiskScore,
		tagsJSON, metadataJSON, entity.Blacklisted, entity.Watchlist,
		entity.KYCVerified, entity.AMLStatus)

	if err != nil {
		return fmt.Errorf("failed to create entity: %w", err)
	}

	return nil
}

// GetEntity retrieves an entity by ID
func (r *PostgresRepository) GetEntity(ctx context.Context, id string) (*domain.Entity, error) {
	query := `
		SELECT id, type, name, country, risk_score, tags, metadata, blacklisted, watchlist, kyc_verified, aml_status, created_at, updated_at
		FROM entities WHERE id = $1
	`

	var entity domain.Entity
	var tagsJSON, metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&entity.ID, &entity.Type, &entity.Name, &entity.Country, &entity.RiskScore,
		&tagsJSON, &metadataJSON, &entity.Blacklisted, &entity.Watchlist,
		&entity.KYCVerified, &entity.AMLStatus, &entity.CreatedAt, &entity.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get entity: %w", err)
	}

	if err := json.Unmarshal(tagsJSON, &entity.Tags); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
	}
	if err := json.Unmarshal(metadataJSON, &entity.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &entity, nil
}

// GetBlacklistedEntities retrieves all blacklisted entities
func (r *PostgresRepository) GetBlacklistedEntities(ctx context.Context) ([]domain.Entity, error) {
	query := `
		SELECT id, type, name, country, risk_score, tags, metadata, blacklisted, watchlist, kyc_verified, aml_status, created_at, updated_at
		FROM entities WHERE blacklisted = true
		ORDER BY risk_score DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query blacklisted entities: %w", err)
	}
	defer rows.Close()

	var entities []domain.Entity
	for rows.Next() {
		var entity domain.Entity
		var tagsJSON, metadataJSON []byte

		err := rows.Scan(
			&entity.ID, &entity.Type, &entity.Name, &entity.Country, &entity.RiskScore,
			&tagsJSON, &metadataJSON, &entity.Blacklisted, &entity.Watchlist,
			&entity.KYCVerified, &entity.AMLStatus, &entity.CreatedAt, &entity.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan entity: %w", err)
		}

		if err := json.Unmarshal(tagsJSON, &entity.Tags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
		}
		if err := json.Unmarshal(metadataJSON, &entity.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		entities = append(entities, entity)
	}

	return entities, nil
}

// Violation operations

// CreateViolation creates a new violation
func (r *PostgresRepository) CreateViolation(ctx context.Context, violation *domain.Violation) error {
	detailsJSON, err := json.Marshal(violation.Details)
	if err != nil {
		return fmt.Errorf("failed to marshal details: %w", err)
	}

	query := `
		INSERT INTO violations (id, transaction_id, rule_id, rule_name, severity, status, details, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
	`

	_, err = r.db.ExecContext(ctx, query,
		violation.ID, violation.TransactionID, violation.RuleID,
		violation.RuleName, violation.Severity, violation.Status, detailsJSON)

	if err != nil {
		return fmt.Errorf("failed to create violation: %w", err)
	}

	return nil
}

// GetOpenViolations retrieves all open violations
func (r *PostgresRepository) GetOpenViolations(ctx context.Context, limit, offset int) ([]domain.Violation, error) {
	query := `
		SELECT id, transaction_id, rule_id, rule_name, severity, status, resolution, resolved_by, resolved_at, details, created_at
		FROM violations
		WHERE status = 'OPEN'
		ORDER BY severity DESC, created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query violations: %w", err)
	}
	defer rows.Close()

	var violations []domain.Violation
	for rows.Next() {
		var v domain.Violation
		var detailsJSON []byte

		err := rows.Scan(
			&v.ID, &v.TransactionID, &v.RuleID, &v.RuleName,
			&v.Severity, &v.Status, &v.Resolution, &v.ResolvedBy,
			&v.ResolvedAt, &detailsJSON, &v.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan violation: %w", err)
		}

		if err := json.Unmarshal(detailsJSON, &v.Details); err != nil {
			return nil, fmt.Errorf("failed to unmarshal details: %w", err)
		}

		violations = append(violations, v)
	}

	return violations, nil
}

// ResolveViolation marks a violation as resolved
func (r *PostgresRepository) ResolveViolation(ctx context.Context, id, resolution, resolvedBy string) error {
	query := `
		UPDATE violations
		SET status = 'RESOLVED', resolution = $1, resolved_by = $2, resolved_at = NOW()
		WHERE id = $3 AND status = 'OPEN'
	`

	result, err := r.db.ExecContext(ctx, query, resolution, resolvedBy, id)
	if err != nil {
		return fmt.Errorf("failed to resolve violation: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("violation not found or already resolved: %s", id)
	}

	return nil
}

// Watchlist operations

// AddToWatchlist adds an entry to the watchlist
func (r *PostgresRepository) AddToWatchlist(ctx context.Context, entry *domain.WatchlistEntry) error {
	query := `
		INSERT INTO watchlist (id, type, name, country, list_source, match_score, active, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
	`

	_, err := r.db.ExecContext(ctx, query,
		entry.ID, entry.Type, entry.Name, entry.Country,
		entry.ListSource, entry.MatchScore, entry.Active, entry.ExpiresAt)

	if err != nil {
		return fmt.Errorf("failed to add to watchlist: %w", err)
	}

	return nil
}

// SearchWatchlist searches the watchlist by name
func (r *PostgresRepository) SearchWatchlist(ctx context.Context, name string) ([]domain.WatchlistEntry, error) {
	query := `
		SELECT id, type, name, country, list_source, match_score, active, expires_at, created_at
		FROM watchlist
		WHERE active = true
		AND (expires_at IS NULL OR expires_at > NOW())
		AND LOWER(name) LIKE LOWER($1)
		LIMIT 100
	`

	rows, err := r.db.QueryContext(ctx, query, "%"+name+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to search watchlist: %w", err)
	}
	defer rows.Close()

	var entries []domain.WatchlistEntry
	for rows.Next() {
		var entry domain.WatchlistEntry

		err := rows.Scan(
			&entry.ID, &entry.Type, &entry.Name, &entry.Country,
			&entry.ListSource, &entry.MatchScore, &entry.Active,
			&entry.ExpiresAt, &entry.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan watchlist entry: %w", err)
		}

		entries = append(entries, entry)
	}

	return entries, nil
}
