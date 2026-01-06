package service

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/csic/platform/compliance/internal/config"
	"github.com/csic/platform/compliance/internal/domain"
	"github.com/csic/platform/compliance/internal/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ComplianceService handles compliance check operations
type ComplianceService struct {
	repo           *repository.PostgresRepository
	redis          *repository.RedisClient
	violationProducer repository.KafkaProducer
	config         *config.Config
	logger         *zap.Logger
}

// NewComplianceService creates a new compliance service
func NewComplianceService(
	repo *repository.PostgresRepository,
	redis *repository.RedisClient,
	violationProducer repository.KafkaProducer,
	cfg *config.Config,
	logger *zap.Logger,
) *ComplianceService {
	return &ComplianceService{
		repo:             repo,
		redis:            redis,
		violationProducer: violationProducer,
		config:           cfg,
		logger:           logger,
	}
}

// CheckCompliance performs a comprehensive compliance check on a transaction
func (s *ComplianceService) CheckCompliance(ctx context.Context, tx *domain.Transaction) (*domain.ComplianceResult, error) {
	startTime := time.Now()

	// Get enabled rules from cache or database
	rules, err := s.getEnabledRules(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get enabled rules: %w", err)
	}

	// Initialize result
	result := &domain.ComplianceResult{
		TransactionID:  tx.ID,
		OverallStatus:  domain.StatusPass,
		RiskScore:      0,
		Checks:         make([]domain.ComplianceCheck, 0),
		Violations:     make([]domain.Violation, 0),
		Summary:        domain.ResultSummary{},
		CheckedAt:      time.Now(),
	}

	// Run compliance checks
	for _, rule := range rules {
		check := s.evaluateRule(ctx, rule, tx)
		result.Checks = append(result.Checks, check)

		// Update summary
		result.Summary.TotalChecks++
		switch check.Status {
		case domain.StatusPass:
			result.Summary.PassedChecks++
		case domain.StatusFail:
			result.Summary.FailedChecks++
			result.OverallStatus = s.worstStatus(result.OverallStatus, domain.StatusFail)
		case domain.StatusWarn:
			result.Summary.WarningChecks++
			result.OverallStatus = s.worstStatus(result.OverallStatus, domain.StatusWarn)
		}

		// Update severity counts
		switch check.Severity {
		case domain.SeverityCritical:
			result.Summary.CriticalCount++
			result.RiskScore += 1.0
		case domain.SeverityHigh:
			result.Summary.HighCount++
			result.RiskScore += 0.75
		case domain.SeverityMedium:
			result.Summary.MediumCount++
			result.RiskScore += 0.5
		case domain.SeverityLow:
			result.Summary.LowCount++
			result.RiskScore += 0.25
		}

		// Create violation if check failed
		if check.Status == domain.StatusFail || check.Status == domain.StatusWarn {
			violation := s.createViolation(tx, check)
			result.Violations = append(result.Violations, violation)

			// Send violation to Kafka
			if err := s.sendViolation(ctx, &violation); err != nil {
				s.logger.Error("Failed to send violation to Kafka", zap.Error(err))
			}
		}
	}

	// Normalize risk score
	result.RiskScore = min(result.RiskScore, 1.0)

	// Set overall status based on risk score
	if result.RiskScore >= 0.8 {
		result.OverallStatus = domain.StatusFail
	} else if result.RiskScore >= 0.5 {
		result.OverallStatus = domain.StatusWarn
	}

	// Calculate processing time
	result.ProcessingTime = time.Since(startTime).Milliseconds()

	// Cache the result
	if err := s.cacheResult(ctx, tx.ID, result); err != nil {
		s.logger.Warn("Failed to cache compliance result", zap.Error(err))
	}

	// Save result to database
	if err := s.repo.SaveComplianceResult(ctx, result); err != nil {
		s.logger.Error("Failed to save compliance result", zap.Error(err))
	}

	// Update metrics
	if err := s.redis.IncrementMetrics(ctx, "total_checks"); err == nil {
		s.redis.IncrementMetrics(ctx, fmt.Sprintf("status:%s", result.OverallStatus))
	}

	s.logger.Info("Compliance check completed",
		zap.String("transaction_id", tx.ID),
		zap.String("overall_status", result.OverallStatus),
		zap.Float64("risk_score", result.RiskScore),
		zap.Int64("processing_time_ms", result.ProcessingTime))

	return result, nil
}

// getEnabledRules retrieves enabled rules from cache or database
func (s *ComplianceService) getEnabledRules(ctx context.Context) ([]domain.Rule, error) {
	// Try to get from cache first
	cachedData, err := s.redis.GetCachedRuleset(ctx, "active")
	if err == nil && cachedData != nil {
		var rules []domain.Rule
		if err := json.Unmarshal(cachedData, &rules); err == nil {
			return rules, nil
		}
	}

	// Fallback to database
	rules, err := s.repo.GetEnabledRules(ctx)
	if err != nil {
		return nil, err
	}

	// Cache the rules
	if data, err := json.Marshal(rules); err == nil {
		cacheTTL := time.Duration(s.config.RulesEngine.CacheTTL) * time.Second
		s.redis.CacheRuleset(ctx, "active", data, cacheTTL)
	}

	return rules, nil
}

// evaluateRule evaluates a single rule against a transaction
func (s *ComplianceService) evaluateRule(ctx context.Context, rule domain.Rule, tx *domain.Transaction) domain.ComplianceCheck {
	startTime := time.Now()

	check := domain.ComplianceCheck{
		ID:        uuid.New().String(),
		RuleID:    rule.ID,
		RuleName:  rule.Name,
		RuleType:  rule.Type,
		Severity:  rule.Severity,
		CheckedAt: time.Now(),
	}

	// Evaluate based on rule type
	switch rule.Type {
	case domain.RuleTypeAML:
		check = s.evaluateAMLRule(check, rule, tx)
	case domain.RuleTypeKYC:
		check = s.evaluateKYCRule(check, rule, tx)
	case domain.RuleTypeSanctions:
		check = s.evaluateSanctionsRule(check, rule, tx)
	case domain.RuleTypeTransaction:
		check = s.evaluateTransactionRule(check, rule, tx)
	case domain.RuleTypeGeographic:
		check = s.evaluateGeographicRule(check, rule, tx)
	case domain.RuleTypeAmount:
		check = s.evaluateAmountRule(check, rule, tx)
	case domain.RuleTypeFrequency:
		check = s.evaluateFrequencyRule(ctx, check, rule, tx)
	default:
		check = s.evaluateCustomRule(check, rule, tx)
	}

	check.Duration = time.Since(startTime).Milliseconds()

	return check
}

// evaluateAMLRule evaluates AML-related rules
func (s *ComplianceService) evaluateAMLRule(check domain.ComplianceCheck, rule domain.Rule, tx *domain.Transaction) domain.ComplianceCheck {
	// Check if source or target is blacklisted
	entities, err := s.repo.GetBlacklistedEntities(context.Background())
	if err != nil {
		check.Status = domain.StatusError
		check.Message = "Failed to check blacklisted entities"
		return check
	}

	for _, entity := range entities {
		if entity.ID == tx.SourceID || entity.ID == tx.TargetID {
			check.Status = domain.StatusFail
			check.Severity = domain.SeverityCritical
			check.Message = fmt.Sprintf("Transaction involves blacklisted entity: %s", entity.Name)
			check.Details = map[string]interface{}{
				"entity_id":   entity.ID,
				"entity_name": entity.Name,
				"entity_type": entity.Type,
			}
			return check
		}
	}

	// Check watchlist
	watchlistEntries, err := s.repo.SearchWatchlist(context.Background(), tx.SourceName)
	if err == nil && len(watchlistEntries) > 0 {
		for _, entry := range watchlistEntries {
			if entry.MatchScore > 0.9 {
				check.Status = domain.StatusFail
				check.Severity = domain.SeverityHigh
				check.Message = fmt.Sprintf("Source matches watchlist entry: %s", entry.Name)
				check.Details = map[string]interface{}{
					"watchlist_name": entry.Name,
					"list_source":    entry.ListSource,
					"match_score":    entry.MatchScore,
				}
				return check
			}
		}
	}

	check.Status = domain.StatusPass
	check.Message = "AML check passed"
	return check
}

// evaluateKYCRule evaluates KYC-related rules
func (s *ComplianceService) evaluateKYCRule(check domain.ComplianceCheck, rule domain.Rule, tx *domain.Transaction) domain.ComplianceCheck {
	// Check if source entity is KYC verified
	sourceEntity, err := s.repo.GetEntity(context.Background(), tx.SourceID)
	if err != nil {
		check.Status = domain.StatusError
		check.Message = "Failed to check entity KYC status"
		return check
	}

	if sourceEntity != nil && !sourceEntity.KYCVerified {
		check.Status = domain.StatusFail
		check.Severity = domain.SeverityHigh
		check.Message = "Source entity is not KYC verified"
		check.Details = map[string]interface{}{
			"entity_id":   sourceEntity.ID,
			"entity_name": sourceEntity.Name,
		}
		return check
	}

	check.Status = domain.StatusPass
	check.Message = "KYC check passed"
	return check
}

// evaluateSanctionsRule evaluates sanctions-related rules
func (s *ComplianceService) evaluateSanctionsRule(check domain.ComplianceCheck, rule domain.Rule, tx *domain.Transaction) domain.ComplianceCheck {
	// Check against blocked countries
	for _, country := range rule.Parameters.BlockedCountries {
		if strings.EqualFold(tx.SourceCountry, country) || strings.EqualFold(tx.TargetCountry, country) {
			check.Status = domain.StatusFail
			check.Severity = domain.SeverityCritical
			check.Message = fmt.Sprintf("Transaction involves sanctioned country: %s", country)
			check.Details = map[string]interface{}{
				"source_country": tx.SourceCountry,
				"target_country": tx.TargetCountry,
				"blocked_country": country,
			}
			return check
		}
	}

	check.Status = domain.StatusPass
	check.Message = "Sanctions check passed"
	return check
}

// evaluateTransactionRule evaluates general transaction rules
func (s *ComplianceService) evaluateTransactionRule(check domain.ComplianceCheck, rule domain.Rule, tx *domain.Transaction) domain.ComplianceCheck {
	// Validate required fields
	for _, field := range rule.Parameters.RequiredFields {
		switch field {
		case "source_account":
			if tx.SourceAccount == "" {
				check.Status = domain.StatusFail
				check.Message = "Missing required field: source_account"
				return check
			}
		case "target_account":
			if tx.TargetAccount == "" {
				check.Status = domain.StatusFail
				check.Message = "Missing required field: target_account"
				return check
			}
		case "amount":
			if tx.Amount <= 0 {
				check.Status = domain.StatusFail
				check.Message = "Invalid amount: must be positive"
				return check
			}
		}
	}

	check.Status = domain.StatusPass
	check.Message = "Transaction validation passed"
	return check
}

// evaluateGeographicRule evaluates geographic restrictions
func (s *ComplianceService) evaluateGeographicRule(check domain.ComplianceCheck, rule domain.Rule, tx *domain.Transaction) domain.ComplianceCheck {
	// Check allowed countries if specified
	if len(rule.Parameters.AllowedCountries) > 0 {
		allowed := false
		for _, country := range rule.Parameters.AllowedCountries {
			if strings.EqualFold(tx.SourceCountry, country) || strings.EqualFold(tx.TargetCountry, country) {
				allowed = true
				break
			}
		}
		if !allowed {
			check.Status = domain.StatusFail
			check.Severity = domain.SeverityMedium
			check.Message = "Transaction involves non-allowed countries"
			check.Details = map[string]interface{}{
				"source_country":   tx.SourceCountry,
				"target_country":   tx.TargetCountry,
				"allowed_countries": rule.Parameters.AllowedCountries,
			}
			return check
		}
	}

	check.Status = domain.StatusPass
	check.Message = "Geographic check passed"
	return check
}

// evaluateAmountRule evaluates amount-based rules
func (s *ComplianceService) evaluateAmountRule(check domain.ComplianceCheck, rule domain.Rule, tx *domain.Transaction) domain.ComplianceCheck {
	// Check minimum amount
	if rule.Parameters.MinAmount > 0 && tx.Amount < rule.Parameters.MinAmount {
		check.Status = domain.StatusWarn
		check.Severity = domain.SeverityLow
		check.Message = fmt.Sprintf("Transaction amount below minimum: %f", rule.Parameters.MinAmount)
		check.Details = map[string]interface{}{
			"transaction_amount": tx.Amount,
			"min_amount":         rule.Parameters.MinAmount,
		}
		return check
	}

	// Check maximum amount
	if rule.Parameters.MaxAmount > 0 && tx.Amount > rule.Parameters.MaxAmount {
		check.Status = domain.StatusFail
		check.Severity = domain.SeverityHigh
		check.Message = fmt.Sprintf("Transaction amount exceeds maximum: %f", rule.Parameters.MaxAmount)
		check.Details = map[string]interface{}{
			"transaction_amount": tx.Amount,
			"max_amount":         rule.Parameters.MaxAmount,
		}
		return check
	}

	check.Status = domain.StatusPass
	check.Message = "Amount check passed"
	return check
}

// evaluateFrequencyRule evaluates frequency-based rules
func (s *ComplianceService) evaluateFrequencyRule(ctx context.Context, check domain.ComplianceCheck, rule domain.Rule, tx *domain.Transaction) domain.ComplianceCheck {
	window := time.Duration(rule.Parameters.WindowSeconds) * time.Second

	// Get transaction rate for source
	txRate, err := s.redis.TrackTransactionRate(ctx, tx.SourceID, window)
	if err != nil {
		check.Status = domain.StatusError
		check.Message = "Failed to check transaction frequency"
		return check
	}

	maxAttempts := rule.Parameters.MaxAttempts
	if maxAttempts == 0 {
		maxAttempts = 10
	}

	if int(txRate) >= maxAttempts {
		check.Status = domain.StatusFail
		check.Severity = domain.SeverityMedium
		check.Message = fmt.Sprintf("Transaction frequency exceeded: %d transactions in window", txRate)
		check.Details = map[string]interface{}{
			"current_count":     txRate,
			"max_attempts":      maxAttempts,
			"window_seconds":    rule.Parameters.WindowSeconds,
		}
		return check
	}

	check.Status = domain.StatusPass
	check.Message = "Frequency check passed"
	return check
}

// evaluateCustomRule evaluates custom rules using expression evaluation
func (s *ComplianceService) evaluateCustomRule(check domain.ComplianceCheck, rule domain.Rule, tx *domain.Transaction) domain.ComplianceCheck {
	// Simple pattern matching for custom rules
	if rule.Expression != "" {
		// Check if expression matches any transaction metadata
		for key, value := range tx.Metadata {
			pattern := strings.ReplaceAll(rule.Expression, "{value}", fmt.Sprintf("%v", value))
			if matched, _ := regexp.MatchString(pattern, key); matched {
				check.Status = domain.StatusPass
				check.Message = "Custom rule check passed"
				return check
			}
		}
	}

	check.Status = domain.StatusPass
	check.Message = "Custom rule check passed"
	return check
}

// createViolation creates a violation from a failed check
func (s *ComplianceService) createViolation(tx *domain.Transaction, check domain.ComplianceCheck) domain.Violation {
	return domain.Violation{
		ID:            uuid.New().String(),
		TransactionID: tx.ID,
		RuleID:        check.RuleID,
		RuleName:      check.RuleName,
		Severity:      check.Severity,
		Status:        "OPEN",
		CreatedAt:     time.Now(),
		Details:       check.Details,
	}
}

// sendViolation sends a violation to Kafka
func (s *ComplianceService) sendViolation(ctx context.Context, violation *domain.Violation) error {
	if s.violationProducer == nil {
		return nil
	}

	data, err := json.Marshal(violation)
	if err != nil {
		return fmt.Errorf("failed to marshal violation: %w", err)
	}

	return s.violationProducer.Send(ctx, s.config.Kafka.Topics.Violations, string(data))
}

// cacheResult caches the compliance result
func (s *ComplianceService) cacheResult(ctx context.Context, txID string, result *domain.ComplianceResult) error {
	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	cacheTTL := 24 * time.Hour
	return s.redis.CacheCheckResult(ctx, txID, data, cacheTTL)
}

// GetCachedResult retrieves a cached compliance result
func (s *ComplianceService) GetCachedResult(ctx context.Context, txID string) (*domain.ComplianceResult, error) {
	data, err := s.redis.GetCachedCheckResult(ctx, txID)
	if err != nil {
		return nil, err
	}

	var result domain.ComplianceResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return &result, nil
}

// worstStatus returns the worst status between two statuses
func (s *ComplianceService) worstStatus(a, b domain.ComplianceStatus) domain.ComplianceStatus {
	order := map[domain.ComplianceStatus]int{
		domain.StatusPass:    0,
		domain.StatusPending: 1,
		domain.StatusWarn:    2,
		domain.StatusError:   3,
		domain.StatusFail:    4,
	}

	if order[b] > order[a] {
		return b
	}
	return a
}

// Rule management

// CreateRule creates a new compliance rule
func (s *ComplianceService) CreateRule(ctx context.Context, rule *domain.Rule) error {
	if rule.ID == "" {
		rule.ID = uuid.New().String()
	}

	if err := s.repo.CreateRule(ctx, rule); err != nil {
		return err
	}

	// Invalidate rules cache
	s.redis.InvalidateRuleCache(ctx, "active")

	s.logger.Info("Created compliance rule",
		zap.String("rule_id", rule.ID),
		zap.String("rule_name", rule.Name),
		zap.String("rule_type", rule.Type))

	return nil
}

// GetRule retrieves a rule by ID
func (s *ComplianceService) GetRule(ctx context.Context, id string) (*domain.Rule, error) {
	return s.repo.GetRule(ctx, id)
}

// UpdateRule updates an existing rule
func (s *ComplianceService) UpdateRule(ctx context.Context, rule *domain.Rule) error {
	if err := s.repo.UpdateRule(ctx, rule); err != nil {
		return err
	}

	// Invalidate rules cache
	s.redis.InvalidateRuleCache(ctx, "active")

	s.logger.Info("Updated compliance rule",
		zap.String("rule_id", rule.ID),
		zap.String("rule_name", rule.Name))

	return nil
}

// GetRulesByType retrieves rules by type
func (s *ComplianceService) GetRulesByType(ctx context.Context, ruleType domain.RuleType) ([]domain.Rule, error) {
	return s.repo.GetRulesByType(ctx, ruleType)
}

// Entity management

// CreateEntity creates a new entity
func (s *ComplianceService) CreateEntity(ctx context.Context, entity *domain.Entity) error {
	if entity.ID == "" {
		entity.ID = uuid.New().String()
	}

	return s.repo.CreateEntity(ctx, entity)
}

// GetEntity retrieves an entity by ID
func (s *ComplianceService) GetEntity(ctx context.Context, id string) (*domain.Entity, error) {
	return s.repo.GetEntity(ctx, id)
}

// Violation management

// GetOpenViolations retrieves open violations
func (s *ComplianceService) GetOpenViolations(ctx context.Context, limit, offset int) ([]domain.Violation, error) {
	return s.repo.GetOpenViolations(ctx, limit, offset)
}

// ResolveViolation resolves a violation
func (s *ComplianceService) ResolveViolation(ctx context.Context, id, resolution, resolvedBy string) error {
	return s.repo.ResolveViolation(ctx, id, resolution, resolvedBy)
}

// Watchlist management

// AddToWatchlist adds an entry to the watchlist
func (s *ComplianceService) AddToWatchlist(ctx context.Context, entry *domain.WatchlistEntry) error {
	if entry.ID == "" {
		entry.ID = uuid.New().String()
	}

	return s.repo.AddToWatchlist(ctx, entry)
}

// SearchWatchlist searches the watchlist
func (s *ComplianceService) SearchWatchlist(ctx context.Context, name string) ([]domain.WatchlistEntry, error) {
	return s.repo.SearchWatchlist(ctx, name)
}

// Compliance results

// GetComplianceResults retrieves compliance results
func (s *ComplianceService) GetComplianceResults(ctx context.Context, limit, offset int) ([]domain.ComplianceResult, error) {
	return s.repo.GetComplianceResults(ctx, limit, offset)
}

// Health check

// HealthCheck checks the service health
func (s *ComplianceService) HealthCheck(ctx context.Context) error {
	if err := s.repo.Close(); err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}

	if err := s.redis.Ping(); err != nil {
		return fmt.Errorf("redis connection failed: %w", err)
	}

	return nil
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
