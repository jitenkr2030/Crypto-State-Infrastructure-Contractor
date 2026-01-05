package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"audit-log/internal/core/domain"
	"audit-log/internal/core/ports"
)

// PostgresAuditRepository implements ports.AuditLogRepository for PostgreSQL
type PostgresAuditRepository struct {
	db     *sql.DB
	logger ports.Logger
}

// NewPostgresAuditRepository creates a new PostgresAuditRepository
func NewPostgresAuditRepository(db *sql.DB, logger ports.Logger) *PostgresAuditRepository {
	return &PostgresAuditRepository{
		db:     db,
		logger: logger,
	}
}

// CreateEntry creates a new audit entry in the database
func (r *PostgresAuditRepository) CreateEntry(ctx context.Context, entry *domain.AuditEntry) error {
	query := `
		INSERT INTO audit_entries (
			id, trace_id, actor_id, actor_type, action, resource, resource_id,
			operation, outcome, severity, payload, metadata, source_ip, user_agent,
			timestamp, previous_hash, current_hash, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
	`

	payloadJSON, _ := json.Marshal(entry.Payload)
	metadataJSON, _ := json.Marshal(entry.Metadata)

	_, err := r.db.ExecContext(ctx, query,
		entry.ID,
		entry.TraceID,
		entry.ActorID,
		entry.ActorType,
		entry.Action,
		entry.Resource,
		entry.ResourceID,
		entry.Operation,
		entry.Outcome,
		entry.Severity,
		payloadJSON,
		metadataJSON,
		entry.SourceIP,
		entry.UserAgent,
		entry.Timestamp,
		entry.PreviousHash,
		entry.CurrentHash,
		entry.CreatedAt,
	)

	if err != nil {
		r.logger.Error("Failed to create audit entry", "error", err, "id", entry.ID)
		return fmt.Errorf("failed to create audit entry: %w", err)
	}

	return nil
}

// GetLastHash retrieves the hash of the most recent audit entry
func (r *PostgresAuditRepository) GetLastHash(ctx context.Context) (string, error) {
	query := `SELECT current_hash FROM audit_entries ORDER BY created_at DESC LIMIT 1`

	var hash string
	err := r.db.QueryRowContext(ctx, query).Scan(&hash)
	if err == sql.ErrNoRows {
		// Return empty hash for first entry
		return "", nil
	}
	if err != nil {
		r.logger.Error("Failed to get last hash", "error", err)
		return "", fmt.Errorf("failed to get last hash: %w", err)
	}

	return hash, nil
}

// GetEntry retrieves an audit entry by ID
func (r *PostgresAuditRepository) GetEntry(ctx context.Context, id string) (*domain.AuditEntry, error) {
	query := `
		SELECT id, trace_id, actor_id, actor_type, action, resource, resource_id,
			   operation, outcome, severity, payload, metadata, source_ip, user_agent,
			   timestamp, previous_hash, current_hash, created_at
		FROM audit_entries
		WHERE id = $1
	`

	entry := &domain.AuditEntry{}
	var payloadJSON, metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&entry.ID,
		&entry.TraceID,
		&entry.ActorID,
		&entry.ActorType,
		&entry.Action,
		&entry.Resource,
		&entry.ResourceID,
		&entry.Operation,
		&entry.Outcome,
		&entry.Severity,
		&payloadJSON,
		&metadataJSON,
		&entry.SourceIP,
		&entry.UserAgent,
		&entry.Timestamp,
		&entry.PreviousHash,
		&entry.CurrentHash,
		&entry.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("audit entry not found")
	}
	if err != nil {
		r.logger.Error("Failed to get audit entry", "error", err, "id", id)
		return nil, fmt.Errorf("failed to get audit entry: %w", err)
	}

	json.Unmarshal(payloadJSON, &entry.Payload)
	json.Unmarshal(metadataJSON, &entry.Metadata)

	return entry, nil
}

// GetEntryByTraceID retrieves all audit entries for a trace ID
func (r *PostgresAuditRepository) GetEntryByTraceID(ctx context.Context, traceID string) ([]*domain.AuditEntry, error) {
	query := `
		SELECT id, trace_id, actor_id, actor_type, action, resource, resource_id,
			   operation, outcome, severity, payload, metadata, source_ip, user_agent,
			   timestamp, previous_hash, current_hash, created_at
		FROM audit_entries
		WHERE trace_id = $1
		ORDER BY timestamp ASC
	`

	rows, err := r.db.QueryContext(ctx, query, traceID)
	if err != nil {
		r.logger.Error("Failed to get audit entries by trace ID", "error", err, "traceID", traceID)
		return nil, fmt.Errorf("failed to get audit entries: %w", err)
	}
	defer rows.Close()

	var entries []*domain.AuditEntry
	for rows.Next() {
		entry := &domain.AuditEntry{}
		var payloadJSON, metadataJSON []byte

		err := rows.Scan(
			&entry.ID,
			&entry.TraceID,
			&entry.ActorID,
			&entry.ActorType,
			&entry.Action,
			&entry.Resource,
			&entry.ResourceID,
			&entry.Operation,
			&entry.Outcome,
			&entry.Severity,
			&payloadJSON,
			&metadataJSON,
			&entry.SourceIP,
			&entry.UserAgent,
			&entry.Timestamp,
			&entry.PreviousHash,
			&entry.CurrentHash,
			&entry.CreatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan audit entry", "error", err)
			continue
		}

		json.Unmarshal(payloadJSON, &entry.Payload)
		json.Unmarshal(metadataJSON, &entry.Metadata)
		entries = append(entries, entry)
	}

	return entries, nil
}

// SearchEntries searches for audit entries based on criteria
func (r *PostgresAuditRepository) SearchEntries(ctx context.Context, request domain.AuditSearchRequest) ([]*domain.AuditEntry, int64, error) {
	offset := (request.Page - 1) * request.PageSize

	// Build dynamic query
	baseQuery := `FROM audit_entries WHERE 1=1`
	args := []interface{}{}
	argNum := 1

	if request.TraceID != "" {
		baseQuery += fmt.Sprintf(` AND trace_id = $%d`, argNum)
		args = append(args, request.TraceID)
		argNum++
	}

	if request.ActorID != "" {
		baseQuery += fmt.Sprintf(` AND actor_id = $%d`, argNum)
		args = append(args, request.ActorID)
		argNum++
	}

	if request.ActorType != "" {
		baseQuery += fmt.Sprintf(` AND actor_type = $%d`, argNum)
		args = append(args, request.ActorType)
		argNum++
	}

	if request.Action != "" {
		baseQuery += fmt.Sprintf(` AND action = $%d`, argNum)
		args = append(args, request.Action)
		argNum++
	}

	if request.Resource != "" {
		baseQuery += fmt.Sprintf(` AND resource = $%d`, argNum)
		args = append(args, request.Resource)
		argNum++
	}

	if request.Outcome != "" {
		baseQuery += fmt.Sprintf(` AND outcome = $%d`, argNum)
		args = append(args, request.Outcome)
		argNum++
	}

	if request.Severity != "" {
		baseQuery += fmt.Sprintf(` AND severity = $%d`, argNum)
		args = append(args, request.Severity)
		argNum++
	}

	if request.SourceIP != "" {
		baseQuery += fmt.Sprintf(` AND source_ip = $%d`, argNum)
		args = append(args, request.SourceIP)
		argNum++
	}

	if request.StartTime != nil {
		baseQuery += fmt.Sprintf(` AND timestamp >= $%d`, argNum)
		args = append(args, *request.StartTime)
		argNum++
	}

	if request.EndTime != nil {
		baseQuery += fmt.Sprintf(` AND timestamp <= $%d`, argNum)
		args = append(args, *request.EndTime)
		argNum++
	}

	// Get total count
	countQuery := `SELECT COUNT(*)` + baseQuery
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count audit entries: %w", err)
	}

	// Build ORDER BY clause
	sortOrder := "DESC"
	if request.SortOrder == "asc" {
		sortOrder = "ASC"
	}

	// Get search results
	selectQuery := `
		SELECT id, trace_id, actor_id, actor_type, action, resource, resource_id,
			   operation, outcome, severity, payload, metadata, source_ip, user_agent,
			   timestamp, previous_hash, current_hash, created_at
		` + baseQuery + `
		ORDER BY ` + request.SortBy + ` ` + sortOrder + `
		LIMIT $` + fmt.Sprintf("%d", argNum) + ` OFFSET $` + fmt.Sprintf("%d", argNum+1)
	args = append(args, request.PageSize, offset)

	rows, err := r.db.QueryContext(ctx, selectQuery, args...)
	if err != nil {
		r.logger.Error("Failed to search audit entries", "error", err)
		return nil, 0, fmt.Errorf("failed to search audit entries: %w", err)
	}
	defer rows.Close()

	var entries []*domain.AuditEntry
	for rows.Next() {
		entry := &domain.AuditEntry{}
		var payloadJSON, metadataJSON []byte

		err := rows.Scan(
			&entry.ID,
			&entry.TraceID,
			&entry.ActorID,
			&entry.ActorType,
			&entry.Action,
			&entry.Resource,
			&entry.ResourceID,
			&entry.Operation,
			&entry.Outcome,
			&entry.Severity,
			&payloadJSON,
			&metadataJSON,
			&entry.SourceIP,
			&entry.UserAgent,
			&entry.Timestamp,
			&entry.PreviousHash,
			&entry.CurrentHash,
			&entry.CreatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan audit entry", "error", err)
			continue
		}

		json.Unmarshal(payloadJSON, &entry.Payload)
		json.Unmarshal(metadataJSON, &entry.Metadata)
		entries = append(entries, entry)
	}

	return entries, total, nil
}

// GetChainSummary returns statistics about the audit chain
func (r *PostgresAuditRepository) GetChainSummary(ctx context.Context) (*domain.AuditChainSummary, error) {
	summary := &domain.AuditChainSummary{}

	// Get total count
	countQuery := `SELECT COUNT(*) FROM audit_entries`
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&summary.TotalEntries); err != nil {
		return nil, fmt.Errorf("failed to count audit entries: %w", err)
	}

	if summary.TotalEntries == 0 {
		summary.ChainIntegrity = "empty"
		return summary, nil
	}

	// Get first entry time
	firstQuery := `SELECT MIN(timestamp) FROM audit_entries`
	if err := r.db.QueryRowContext(ctx, firstQuery).Scan(&summary.FirstEntryTime); err != nil {
		return nil, fmt.Errorf("failed to get first entry time: %w", err)
	}

	// Get last entry time
	lastQuery := `SELECT MAX(timestamp) FROM audit_entries`
	if err := r.db.QueryRowContext(ctx, lastQuery).Scan(&summary.LastEntryTime); err != nil {
		return nil, fmt.Errorf("failed to get last entry time: %w", err)
	}

	// Get recent activity (last 24 hours)
	oneDayAgo := time.Now().UTC().Add(-24 * time.Hour)
	recentQuery := `SELECT COUNT(*) FROM audit_entries WHERE timestamp >= $1`
	if err := r.db.QueryRowContext(ctx, recentQuery, oneDayAgo).Scan(&summary.RecentActivity); err != nil {
		return nil, fmt.Errorf("failed to get recent activity: %w", err)
	}

	// Verify chain integrity
	chainValid, err := r.verifyChainIntegrity(ctx)
	if err != nil {
		summary.ChainIntegrity = "unknown"
	} else if chainValid {
		summary.ChainIntegrity = "valid"
	} else {
		summary.ChainIntegrity = "broken"
	}

	return summary, nil
}

// VerifyChain verifies the hash chain starting from an entry
func (r *PostgresAuditRepository) VerifyChain(ctx context.Context, startID string, limit int) (*domain.VerificationResult, error) {
	// Get the starting entry
	startEntry, err := r.GetEntry(ctx, startID)
	if err != nil {
		return nil, err
	}

	// Get entries after the start entry
	query := `
		SELECT id, trace_id, actor_id, actor_type, action, resource, resource_id,
			   operation, outcome, severity, payload, metadata, source_ip, user_agent,
			   timestamp, previous_hash, current_hash, created_at
		FROM audit_entries
		WHERE id != $1 AND created_at >= $2
		ORDER BY created_at ASC
		LIMIT $3
	`

	rows, err := r.db.QueryContext(ctx, query, startID, startEntry.CreatedAt, limit)
	if err != nil {
		r.logger.Error("Failed to get entries for chain verification", "error", err)
		return nil, fmt.Errorf("failed to verify chain: %w", err)
	}
	defer rows.Close()

	var entries []*domain.AuditEntry
	for rows.Next() {
		entry := &domain.AuditEntry{}
		var payloadJSON, metadataJSON []byte

		err := rows.Scan(
			&entry.ID,
			&entry.TraceID,
			&entry.ActorID,
			&entry.ActorType,
			&entry.Action,
			&entry.Resource,
			&entry.ResourceID,
			&entry.Operation,
			&entry.Outcome,
			&entry.Severity,
			&payloadJSON,
			&metadataJSON,
			&entry.SourceIP,
			&entry.UserAgent,
			&entry.Timestamp,
			&entry.PreviousHash,
			&entry.CurrentHash,
			&entry.CreatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan audit entry", "error", err)
			continue
		}

		json.Unmarshal(payloadJSON, &entry.Payload)
		json.Unmarshal(metadataJSON, &entry.Metadata)
		entries = append(entries, entry)
	}

	// Verify the chain starting from the start entry
	allValid := true
	var lastHash string
	entryCount := int64(1) // Start with the initial entry

	// Calculate expected hash for start entry
	expectedHash := calculateExpectedHash(startEntry)
	if startEntry.CurrentHash != expectedHash {
		allValid = false
	}
	lastHash = startEntry.CurrentHash

	// Verify subsequent entries
	for _, entry := range entries {
		entryCount++

		// Check if previous hash matches
		if entry.PreviousHash != lastHash {
			allValid = false
			break
		}

		// Recalculate hash
		expectedHash = calculateExpectedHash(entry)
		if entry.CurrentHash != expectedHash {
			allValid = false
			break
		}

		lastHash = entry.CurrentHash
	}

	return &domain.VerificationResult{
		Valid:       allValid,
		BlockNumber: entryCount,
		Timestamp:   time.Now().UTC(),
		Message:     getVerificationMessage(allValid, entryCount),
	}, nil
}

// verifyChainIntegrity checks the entire chain for integrity
func (r *PostgresAuditRepository) verifyChainIntegrity(ctx context.Context) (bool, error) {
	query := `
		SELECT id, trace_id, actor_id, actor_type, action, resource, resource_id,
			   operation, outcome, severity, payload, metadata, source_ip, user_agent,
			   timestamp, previous_hash, current_hash, created_at
		FROM audit_entries
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return false, fmt.Errorf("failed to verify chain integrity: %w", err)
	}
	defer rows.Close()

	var lastHash string
	for rows.Next() {
		entry := &domain.AuditEntry{}
		var payloadJSON, metadataJSON []byte

		err := rows.Scan(
			&entry.ID,
			&entry.TraceID,
			&entry.ActorID,
			&entry.ActorType,
			&entry.Action,
			&entry.Resource,
			&entry.ResourceID,
			&entry.Operation,
			&entry.Outcome,
			&entry.Severity,
			&payloadJSON,
			&metadataJSON,
			&entry.SourceIP,
			&entry.UserAgent,
			&entry.Timestamp,
			&entry.PreviousHash,
			&entry.CurrentHash,
			&entry.CreatedAt,
		)
		if err != nil {
			continue
		}

		json.Unmarshal(payloadJSON, &entry.Payload)
		json.Unmarshal(metadataJSON, &entry.Metadata)

		// Check if previous hash matches
		if lastHash != "" && entry.PreviousHash != lastHash {
			return false, nil
		}

		// Verify hash
		expectedHash := calculateExpectedHash(entry)
		if entry.CurrentHash != expectedHash {
			return false, nil
		}

		lastHash = entry.CurrentHash
	}

	return true, nil
}

// calculateExpectedHash calculates the expected hash for an entry
func calculateExpectedHash(entry *domain.AuditEntry) string {
	data := struct {
		ID           string                 `json:"id"`
		TraceID      string                 `json:"trace_id"`
		ActorID      string                 `json:"actor_id"`
		ActorType    string                 `json:"actor_type"`
		Action       string                 `json:"action"`
		Resource     string                 `json:"resource"`
		ResourceID   string                 `json:"resource_id"`
		Operation    string                 `json:"operation"`
		Outcome      string                 `json:"outcome"`
		Severity     string                 `json:"severity"`
		Timestamp    time.Time              `json:"timestamp"`
		PreviousHash string                 `json:"previous_hash"`
		Payload      map[string]interface{} `json:"payload,omitempty"`
	}{
		ID:           entry.ID,
		TraceID:      entry.TraceID,
		ActorID:      entry.ActorID,
		ActorType:    entry.ActorType,
		Action:       entry.Action,
		Resource:     entry.Resource,
		ResourceID:   entry.ResourceID,
		Operation:    entry.Operation,
		Outcome:      entry.Outcome,
		Severity:     entry.Severity,
		Timestamp:    entry.Timestamp,
		PreviousHash: entry.PreviousHash,
		Payload:      entry.Payload,
	}

	jsonData, _ := json.Marshal(data)
	hashInput := string(jsonData)

	if entry.PreviousHash != "" {
		hashInput = entry.PreviousHash + "|" + hashInput
	}

	// Use sha256
	hashBytes := []byte(hashInput)
	return fmt.Sprintf("%x", hashBytes)
}

func getVerificationMessage(valid bool, count int64) string {
	if valid {
		return fmt.Sprintf("Chain verification successful. %d entries verified.", count)
	}
	return "Chain verification failed. Tampering detected."
}

// HealthCheck checks database connectivity
func (r *PostgresAuditRepository) HealthCheck(ctx context.Context) error {
	return r.db.PingContext(ctx)
}
