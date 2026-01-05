package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"forensic-tools/internal/core/domain"
	"forensic-tools/internal/core/ports"
)

// PostgresRepository implements ports.Repository for PostgreSQL
type PostgresRepository struct {
	db     *sql.DB
	logger ports.Logger
}

// NewPostgresRepository creates a new PostgresRepository
func NewPostgresRepository(db *sql.DB, logger ports.Logger) *PostgresRepository {
	return &PostgresRepository{
		db:     db,
		logger: logger,
	}
}

// CreateEvidence creates new evidence in the database
func (r *PostgresRepository) CreateEvidence(ctx context.Context, evidence *domain.Evidence) error {
	query := `
		INSERT INTO evidence (id, name, type, description, source, hash, hash_algorithm, size, location,
		                      tags, metadata, status, collected_at, collected_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`

	tagsJSON, _ := json.Marshal(evidence.Tags)
	metadataJSON, _ := json.Marshal(evidence.Metadata)

	_, err := r.db.ExecContext(ctx, query,
		evidence.ID,
		evidence.Name,
		evidence.Type,
		evidence.Description,
		evidence.Source,
		evidence.Hash,
		evidence.HashAlgorithm,
		evidence.Size,
		evidence.Location,
		tagsJSON,
		metadataJSON,
		evidence.Status,
		evidence.CollectedAt,
		evidence.CollectedBy,
		evidence.CreatedAt,
		evidence.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("Failed to create evidence", "error", err, "id", evidence.ID)
		return fmt.Errorf("failed to create evidence: %w", err)
	}

	return nil
}

// GetEvidence retrieves evidence by ID
func (r *PostgresRepository) GetEvidence(ctx context.Context, id string) (*domain.Evidence, error) {
	query := `
		SELECT id, name, type, description, source, hash, hash_algorithm, size, location,
		       tags, metadata, status, collected_at, collected_by, verified_at, verified_by,
		       created_at, updated_at
		FROM evidence
		WHERE id = $1 AND status != 'deleted'
	`

	evidence := &domain.Evidence{}
	var tagsJSON, metadataJSON []byte
	var verifiedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&evidence.ID,
		&evidence.Name,
		&evidence.Type,
		&evidence.Description,
		&evidence.Source,
		&evidence.Hash,
		&evidence.HashAlgorithm,
		&evidence.Size,
		&evidence.Location,
		&tagsJSON,
		&metadataJSON,
		&evidence.Status,
		&evidence.CollectedAt,
		&evidence.CollectedBy,
		&verifiedAt,
		&evidence.VerifiedBy,
		&evidence.CreatedAt,
		&evidence.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("evidence not found: %s", id)
	}
	if err != nil {
		r.logger.Error("Failed to get evidence", "error", err, "id", id)
		return nil, fmt.Errorf("failed to get evidence: %w", err)
	}

	if verifiedAt.Valid {
		evidence.VerifiedAt = &verifiedAt.Time
	}

	json.Unmarshal(tagsJSON, &evidence.Tags)
	json.Unmarshal(metadataJSON, &evidence.Metadata)

	return evidence, nil
}

// UpdateEvidence updates existing evidence
func (r *PostgresRepository) UpdateEvidence(ctx context.Context, evidence *domain.Evidence) error {
	query := `
		UPDATE evidence
		SET name = $1, type = $2, description = $3, source = $4, hash = $5,
		    size = $6, location = $7, tags = $8, metadata = $9, status = $10,
		    verified_at = $11, verified_by = $12, updated_at = $13
		WHERE id = $14
	`

	tagsJSON, _ := json.Marshal(evidence.Tags)
	metadataJSON, _ := json.Marshal(evidence.Metadata)

	result, err := r.db.ExecContext(ctx, query,
		evidence.Name,
		evidence.Type,
		evidence.Description,
		evidence.Source,
		evidence.Hash,
		evidence.Size,
		evidence.Location,
		tagsJSON,
		metadataJSON,
		evidence.Status,
		evidence.VerifiedAt,
		evidence.VerifiedBy,
		time.Now().UTC(),
		evidence.ID,
	)

	if err != nil {
		r.logger.Error("Failed to update evidence", "error", err, "id", evidence.ID)
		return fmt.Errorf("failed to update evidence: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("evidence not found: %s", evidence.ID)
	}

	return nil
}

// DeleteEvidence marks evidence as deleted
func (r *PostgresRepository) DeleteEvidence(ctx context.Context, id string) error {
	query := `UPDATE evidence SET status = 'deleted', updated_at = $1 WHERE id = $2`

	_, err := r.db.ExecContext(ctx, query, time.Now().UTC(), id)
	if err != nil {
		r.logger.Error("Failed to delete evidence", "error", err, "id", id)
		return fmt.Errorf("failed to delete evidence: %w", err)
	}

	return nil
}

// ListEvidence lists evidence with pagination
func (r *PostgresRepository) ListEvidence(ctx context.Context, page, pageSize int) ([]*domain.Evidence, int64, error) {
	offset := (page - 1) * pageSize

	// Get total count
	var total int64
	countQuery := `SELECT COUNT(*) FROM evidence WHERE status != 'deleted'`
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count evidence: %w", err)
	}

	// Get evidence list
	query := `
		SELECT id, name, type, description, source, hash, hash_algorithm, size, location,
		       tags, metadata, status, collected_at, collected_by, created_at, updated_at
		FROM evidence
		WHERE status != 'deleted'
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, pageSize, offset)
	if err != nil {
		r.logger.Error("Failed to list evidence", "error", err)
		return nil, 0, fmt.Errorf("failed to list evidence: %w", err)
	}
	defer rows.Close()

	var evidenceList []*domain.Evidence
	for rows.Next() {
		evidence := &domain.Evidence{}
		var tagsJSON, metadataJSON []byte

		err := rows.Scan(
			&evidence.ID,
			&evidence.Name,
			&evidence.Type,
			&evidence.Description,
			&evidence.Source,
			&evidence.Hash,
			&evidence.HashAlgorithm,
			&evidence.Size,
			&evidence.Location,
			&tagsJSON,
			&metadataJSON,
			&evidence.Status,
			&evidence.CollectedAt,
			&evidence.CollectedBy,
			&evidence.CreatedAt,
			&evidence.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan evidence row", "error", err)
			continue
		}

		json.Unmarshal(tagsJSON, &evidence.Tags)
		json.Unmarshal(metadataJSON, &evidence.Metadata)
		evidenceList = append(evidenceList, evidence)
	}

	return evidenceList, total, nil
}

// SearchEvidence searches for evidence
func (r *PostgresRepository) SearchEvidence(ctx context.Context, query string, evidenceTypes []string, dateFrom, dateTo *time.Time, tags []string, page, pageSize int) ([]domain.EvidenceSummary, int64, error) {
	offset := (page - 1) * pageSize

	// Build dynamic query
	baseQuery := `FROM evidence WHERE status != 'deleted'`
	args := []interface{}{}
	argNum := 1

	if query != "" {
		baseQuery += fmt.Sprintf(` AND (name ILIKE $%d OR description ILIKE $%d OR source ILIKE $%d)`, argNum, argNum+1, argNum+2)
		args = append(args, "%"+query+"%", "%"+query+"%", "%"+query+"%")
		argNum += 3
	}

	if len(evidenceTypes) > 0 {
		baseQuery += fmt.Sprintf(` AND type = ANY($%d)`, argNum)
		args = append(args, evidenceTypes)
		argNum++
	}

	if dateFrom != nil {
		baseQuery += fmt.Sprintf(` AND collected_at >= $%d`, argNum)
		args = append(args, *dateFrom)
		argNum++
	}

	if dateTo != nil {
		baseQuery += fmt.Sprintf(` AND collected_at <= $%d`, argNum)
		args = append(args, *dateTo)
		argNum++
	}

	if len(tags) > 0 {
		baseQuery += fmt.Sprintf(` AND tags @> $%d`, argNum)
		args = append(args, tags)
		argNum++
	}

	// Get total count
	countQuery := `SELECT COUNT(*) ` + baseQuery
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count evidence: %w", err)
	}

	// Get search results
	selectQuery := `
		SELECT id, name, type, hash, description, source, tags, collected_at, status
		` + baseQuery + `
		ORDER BY collected_at DESC
		LIMIT $` + fmt.Sprintf("%d", argNum) + ` OFFSET $` + fmt.Sprintf("%d", argNum+1)
	args = append(args, pageSize, offset)

	rows, err := r.db.QueryContext(ctx, selectQuery, args...)
	if err != nil {
		r.logger.Error("Failed to search evidence", "error", err)
		return nil, 0, fmt.Errorf("failed to search evidence: %w", err)
	}
	defer rows.Close()

	var results []domain.EvidenceSummary
	for rows.Next() {
		var summary domain.EvidenceSummary
		var tagsJSON []byte

		err := rows.Scan(
			&summary.ID,
			&summary.Name,
			&summary.Type,
			&summary.Hash,
			&summary.Description,
			&summary.Source,
			&tagsJSON,
			&summary.CollectedAt,
			&summary.Status,
		)
		if err != nil {
			r.logger.Error("Failed to scan evidence row", "error", err)
			continue
		}

		json.Unmarshal(tagsJSON, &summary.Tags)
		results = append(results, summary)
	}

	return results, total, nil
}

// AddCustodyRecord adds a custody record
func (r *PostgresRepository) AddCustodyRecord(ctx context.Context, record *domain.CustodyRecord) error {
	query := `
		INSERT INTO custody_records (id, evidence_id, handler, action, location, notes,
		                            digital_signature, previous_hash, record_hash, timestamp, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.ExecContext(ctx, query,
		record.ID,
		record.EvidenceID,
		record.Handler,
		record.Action,
		record.Location,
		record.Notes,
		record.DigitalSig,
		record.PrevHash,
		record.RecordHash,
		record.Timestamp,
		record.CreatedAt,
	)

	if err != nil {
		r.logger.Error("Failed to add custody record", "error", err, "evidenceId", record.EvidenceID)
		return fmt.Errorf("failed to add custody record: %w", err)
	}

	return nil
}

// GetChainOfCustody retrieves chain of custody records
func (r *PostgresRepository) GetChainOfCustody(ctx context.Context, evidenceID string) ([]*domain.CustodyRecord, error) {
	query := `
		SELECT id, evidence_id, handler, action, location, notes, digital_signature,
		       previous_hash, record_hash, timestamp, created_at
		FROM custody_records
		WHERE evidence_id = $1
		ORDER BY timestamp ASC
	`

	rows, err := r.db.QueryContext(ctx, query, evidenceID)
	if err != nil {
		r.logger.Error("Failed to get chain of custody", "error", err, "evidenceId", evidenceID)
		return nil, fmt.Errorf("failed to get chain of custody: %w", err)
	}
	defer rows.Close()

	var records []*domain.CustodyRecord
	for rows.Next() {
		record := &domain.CustodyRecord{}
		err := rows.Scan(
			&record.ID,
			&record.EvidenceID,
			&record.Handler,
			&record.Action,
			&record.Location,
			&record.Notes,
			&record.DigitalSig,
			&record.PrevHash,
			&record.RecordHash,
			&record.Timestamp,
			&record.CreatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan custody record", "error", err)
			continue
		}
		records = append(records, record)
	}

	return records, nil
}

// CreateAnalysis creates a new analysis
func (r *PostgresRepository) CreateAnalysis(ctx context.Context, analysis *domain.Analysis) error {
	query := `
		INSERT INTO analysis (id, evidence_id, evidence_name, analysis_type, status, parameters,
		                     results, findings, processed_by, started_at, completed_at,
		                     error_message, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	paramsJSON, _ := json.Marshal(analysis.Parameters)
	resultsJSON, _ := json.Marshal(analysis.Results)
	findingsJSON, _ := json.Marshal(analysis.Findings)

	var completedAt interface{}
	if analysis.CompletedAt != nil {
		completedAt = *analysis.CompletedAt
	}

	_, err := r.db.ExecContext(ctx, query,
		analysis.ID,
		analysis.EvidenceID,
		analysis.EvidenceName,
		analysis.AnalysisType,
		analysis.Status,
		paramsJSON,
		resultsJSON,
		findingsJSON,
		analysis.ProcessedBy,
		analysis.StartedAt,
		completedAt,
		analysis.ErrorMessage,
		analysis.CreatedAt,
		analysis.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("Failed to create analysis", "error", err, "id", analysis.ID)
		return fmt.Errorf("failed to create analysis: %w", err)
	}

	return nil
}

// GetAnalysis retrieves analysis by ID
func (r *PostgresRepository) GetAnalysis(ctx context.Context, id string) (*domain.Analysis, error) {
	query := `
		SELECT id, evidence_id, evidence_name, analysis_type, status, parameters, results,
		       findings, processed_by, started_at, completed_at, error_message, created_at, updated_at
		FROM analysis
		WHERE id = $1
	`

	analysis := &domain.Analysis{}
	var paramsJSON, resultsJSON, findingsJSON []byte
	var completedAt sql.NullTime
	var errorMsg sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&analysis.ID,
		&analysis.EvidenceID,
		&analysis.EvidenceName,
		&analysis.AnalysisType,
		&analysis.Status,
		&paramsJSON,
		&resultsJSON,
		&findingsJSON,
		&analysis.ProcessedBy,
		&analysis.StartedAt,
		&completedAt,
		&errorMsg,
		&analysis.CreatedAt,
		&analysis.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("analysis not found: %s", id)
	}
	if err != nil {
		r.logger.Error("Failed to get analysis", "error", err, "id", id)
		return nil, fmt.Errorf("failed to get analysis: %w", err)
	}

	if completedAt.Valid {
		analysis.CompletedAt = &completedAt.Time
	}

	if errorMsg.Valid {
		analysis.ErrorMessage = errorMsg.String
	}

	json.Unmarshal(paramsJSON, &analysis.Parameters)
	json.Unmarshal(resultsJSON, &analysis.Results)
	json.Unmarshal(findingsJSON, &analysis.Findings)

	return analysis, nil
}

// UpdateAnalysis updates an analysis
func (r *PostgresRepository) UpdateAnalysis(ctx context.Context, analysis *domain.Analysis) error {
	query := `
		UPDATE analysis
		SET status = $1, results = $2, findings = $3, completed_at = $4,
		    error_message = $5, updated_at = $6
		WHERE id = $7
	`

	resultsJSON, _ := json.Marshal(analysis.Results)
	findingsJSON, _ := json.Marshal(analysis.Findings)

	var completedAt interface{}
	if analysis.CompletedAt != nil {
		completedAt = *analysis.CompletedAt
	}

	var errorMsg interface{}
	if analysis.ErrorMessage != "" {
		errorMsg = analysis.ErrorMessage
	}

	result, err := r.db.ExecContext(ctx, query,
		analysis.Status,
		resultsJSON,
		findingsJSON,
		completedAt,
		errorMsg,
		time.Now().UTC(),
		analysis.ID,
	)

	if err != nil {
		r.logger.Error("Failed to update analysis", "error", err, "id", analysis.ID)
		return fmt.Errorf("failed to update analysis: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("analysis not found: %s", analysis.ID)
	}

	return nil
}

// ListAnalyses lists analyses with filters
func (r *PostgresRepository) ListAnalyses(ctx context.Context, evidenceID, status string, page, pageSize int) ([]*domain.Analysis, int64, error) {
	offset := (page - 1) * pageSize

	// Build dynamic query
	baseQuery := `FROM analysis WHERE 1=1`
	args := []interface{}{}
	argNum := 1

	if evidenceID != "" {
		baseQuery += fmt.Sprintf(` AND evidence_id = $%d`, argNum)
		args = append(args, evidenceID)
		argNum++
	}

	if status != "" {
		baseQuery += fmt.Sprintf(` AND status = $%d`, argNum)
		args = append(args, status)
		argNum++
	}

	// Get total count
	countQuery := `SELECT COUNT(*) ` + baseQuery
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count analyses: %w", err)
	}

	// Get analyses
	selectQuery := `
		SELECT id, evidence_id, evidence_name, analysis_type, status, parameters, results,
		       findings, processed_by, started_at, completed_at, error_message, created_at, updated_at
		` + baseQuery + `
		ORDER BY created_at DESC
		LIMIT $` + fmt.Sprintf("%d", argNum) + ` OFFSET $` + fmt.Sprintf("%d", argNum+1)
	args = append(args, pageSize, offset)

	rows, err := r.db.QueryContext(ctx, selectQuery, args...)
	if err != nil {
		r.logger.Error("Failed to list analyses", "error", err)
		return nil, 0, fmt.Errorf("failed to list analyses: %w", err)
	}
	defer rows.Close()

	var analyses []*domain.Analysis
	for rows.Next() {
		analysis := &domain.Analysis{}
		var paramsJSON, resultsJSON, findingsJSON []byte
		var completedAt sql.NullTime
		var errorMsg sql.NullString

		err := rows.Scan(
			&analysis.ID,
			&analysis.EvidenceID,
			&analysis.EvidenceName,
			&analysis.AnalysisType,
			&analysis.Status,
			&paramsJSON,
			&resultsJSON,
			&findingsJSON,
			&analysis.ProcessedBy,
			&analysis.StartedAt,
			&completedAt,
			&errorMsg,
			&analysis.CreatedAt,
			&analysis.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan analysis row", "error", err)
			continue
		}

		if completedAt.Valid {
			analysis.CompletedAt = &completedAt.Time
		}

		if errorMsg.Valid {
			analysis.ErrorMessage = errorMsg.String
		}

		json.Unmarshal(paramsJSON, &analysis.Parameters)
		json.Unmarshal(resultsJSON, &analysis.Results)
		json.Unmarshal(findingsJSON, &analysis.Findings)

		analyses = append(analyses, analysis)
	}

	return analyses, total, nil
}

// HealthCheck checks database connectivity
func (r *PostgresRepository) HealthCheck(ctx context.Context) error {
	return r.db.PingContext(ctx)
}
