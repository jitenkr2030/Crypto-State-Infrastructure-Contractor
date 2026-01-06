package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/csic/platform/service/reporting/regulatory/internal/domain"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// ReportRepository handles database operations for reports
type ReportRepository struct {
	db        *sql.DB
	keyPrefix string
}

// NewReportRepository creates a new ReportRepository instance
func NewReportRepository(db *sql.DB, keyPrefix string) *ReportRepository {
	return &ReportRepository{
		db:        db,
		keyPrefix: keyPrefix,
	}
}

// Create creates a new report record
func (r *ReportRepository) Create(report *domain.Report) error {
	query := `
		INSERT INTO reports (
			id, name, type, format, status, description, parameters,
			filters, result, metadata, scheduled_id, generated_at,
			expires_at, file_path, file_size, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`

	report.ID = uuid.New().String()
	report.CreatedAt = time.Now()
	report.UpdatedAt = time.Now()
	if report.Status == "" {
		report.Status = domain.ReportStatusPending
	}

	_, err := r.db.Exec(query,
		report.ID, report.Name, report.Type, report.Format, report.Status,
		report.Description, report.Parameters, report.Filters, report.Result,
		report.Metadata, report.ScheduledID, report.GeneratedAt, report.ExpiresAt,
		report.FilePath, report.FileSize, report.CreatedAt, report.UpdatedAt,
	)

	return err
}

// GetByID retrieves a report by its ID
func (r *ReportRepository) GetByID(id string) (*domain.Report, error) {
	query := `
		SELECT id, name, type, format, status, description, parameters,
			filters, result, metadata, scheduled_id, generated_at,
			expires_at, file_path, file_size, created_at, updated_at
		FROM reports WHERE id = $1
	`

	var report domain.Report
	err := r.db.QueryRow(query, id).Scan(
		&report.ID, &report.Name, &report.Type, &report.Format, &report.Status,
		&report.Description, &report.Parameters, &report.Filters, &report.Result,
		&report.Metadata, &report.ScheduledID, &report.GeneratedAt, &report.ExpiresAt,
		&report.FilePath, &report.FileSize, &report.CreatedAt, &report.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &report, nil
}

// Update updates an existing report record
func (r *ReportRepository) Update(report *domain.Report) error {
	query := `
		UPDATE reports SET
			name = $1, description = $2, parameters = $3, filters = $4,
			result = $5, metadata = $6, updated_at = $7
		WHERE id = $8
	`

	report.UpdatedAt = time.Now()

	_, err := r.db.Exec(query,
		report.Name, report.Description, report.Parameters, report.Filters,
		report.Result, report.Metadata, report.UpdatedAt, report.ID,
	)

	return err
}

// UpdateStatus updates the status of a report
func (r *ReportRepository) UpdateStatus(id string, status domain.ReportStatus) error {
	query := `
		UPDATE reports SET status = $1, updated_at = $2 WHERE id = $3
	`

	_, err := r.db.Exec(query, status, time.Now(), id)
	return err
}

// UpdateResult updates the result of a report
func (r *ReportRepository) UpdateResult(id string, result *domain.ReportResult, filePath string, fileSize int64) error {
	query := `
		UPDATE reports SET
			result = $1, file_path = $2, file_size = $3,
			generated_at = $4, updated_at = $5
		WHERE id = $6
	`

	now := time.Now()
	_, err := r.db.Exec(query, result, filePath, fileSize, now, now, id)
	return err
}

// Delete deletes a report by its ID
func (r *ReportRepository) Delete(id string) error {
	query := "DELETE FROM reports WHERE id = $1"
	_, err := r.db.Exec(query, id)
	return err
}

// List retrieves all reports with optional filtering
func (r *ReportRepository) List(filter ReportListFilter) (*domain.PaginatedReports, error) {
	baseQuery := `SELECT id, name, type, format, status, description, parameters,
			filters, result, metadata, scheduled_id, generated_at,
			expires_at, file_path, file_size, created_at, updated_at FROM reports`
	countQuery := "SELECT COUNT(*) FROM reports"

	whereClause := ""
	args := []interface{}{}
	argIndex := 1

	if filter.Type != "" {
		if whereClause != "" {
			whereClause += " AND "
		}
		whereClause += fmt.Sprintf("type = $%d", argIndex)
		args = append(args, filter.Type)
		argIndex++
	}

	if filter.Status != "" {
		if whereClause != "" {
			whereClause += " AND "
		}
		whereClause += fmt.Sprintf("status = $%d", argIndex)
		args = append(args, filter.Status)
		argIndex++
	}

	if filter.StartDate != nil {
		if whereClause != "" {
			whereClause += " AND "
		}
		whereClause += fmt.Sprintf("created_at >= $%d", argIndex)
		args = append(args, *filter.StartDate)
		argIndex++
	}

	if whereClause != "" {
		baseQuery += " WHERE " + whereClause
		countQuery += " WHERE " + whereClause
	}

	// Add pagination
	baseQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, filter.Limit, filter.Offset)

	// Get total count
	var total int
	countArgs := args[:len(args)-2]
	err := r.db.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, err
	}

	// Execute query
	rows, err := r.db.Query(baseQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reports := make([]*domain.Report, 0)
	for rows.Next() {
		var report domain.Report
		err := rows.Scan(
			&report.ID, &report.Name, &report.Type, &report.Format, &report.Status,
			&report.Description, &report.Parameters, &report.Filters, &report.Result,
			&report.Metadata, &report.ScheduledID, &report.GeneratedAt, &report.ExpiresAt,
			&report.FilePath, &report.FileSize, &report.CreatedAt, &report.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		reports = append(reports, &report)
	}

	return &domain.PaginatedReports{
		Reports: reports,
		Total:   total,
		Offset:  filter.Offset,
		Limit:   filter.Limit,
	}, nil
}

// GetExpiredReports retrieves reports that have expired
func (r *ReportRepository) GetExpiredReports(before time.Time) ([]*domain.Report, error) {
	query := `
		SELECT id, name, type, format, status, description, parameters,
			filters, result, metadata, scheduled_id, generated_at,
			expires_at, file_path, file_size, created_at, updated_at
		FROM reports WHERE expires_at IS NOT NULL AND expires_at < $1
	`

	rows, err := r.db.Query(query, before)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reports := make([]*domain.Report, 0)
	for rows.Next() {
		var report domain.Report
		err := rows.Scan(
			&report.ID, &report.Name, &report.Type, &report.Format, &report.Status,
			&report.Description, &report.Parameters, &report.Filters, &report.Result,
			&report.Metadata, &report.ScheduledID, &report.GeneratedAt, &report.ExpiresAt,
			&report.FilePath, &report.FileSize, &report.CreatedAt, &report.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		reports = append(reports, &report)
	}

	return reports, nil
}

// ReportListFilter represents filter options for listing reports
type ReportListFilter struct {
	Type      domain.ReportType
	Status    domain.ReportStatus
	StartDate *time.Time
	Offset    int
	Limit     int
}
