package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/csic/platform/service/reporting/regulatory/internal/domain"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// TemplateRepository handles database operations for templates
type TemplateRepository struct {
	db *sql.DB
}

// NewTemplateRepository creates a new TemplateRepository instance
func NewTemplateRepository(db *sql.DB) *TemplateRepository {
	return &TemplateRepository{db: db}
}

// Create creates a new template record
func (r *TemplateRepository) Create(template *domain.Template) error {
	query := `
		INSERT INTO templates (
			id, name, type, content, parameters, variables,
			metadata, version, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	template.ID = uuid.New().String()
	template.Version = 1
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()

	_, err := r.db.Exec(query,
		template.ID, template.Name, template.Type, template.Content,
		template.Parameters, template.Variables, template.Metadata,
		template.Version, template.CreatedAt, template.UpdatedAt,
	)

	return err
}

// GetByID retrieves a template by its ID
func (r *TemplateRepository) GetByID(id string) (*domain.Template, error) {
	query := `
		SELECT id, name, type, content, parameters, variables,
			metadata, version, created_at, updated_at
		FROM templates WHERE id = $1
	`

	var template domain.Template
	err := r.db.QueryRow(query, id).Scan(
		&template.ID, &template.Name, &template.Type, &template.Content,
		&template.Parameters, &template.Variables, &template.Metadata,
		&template.Version, &template.CreatedAt, &template.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &template, nil
}

// GetByType retrieves the latest template by type
func (r *TemplateRepository) GetByType(reportType domain.ReportType) (*domain.Template, error) {
	query := `
		SELECT id, name, type, content, parameters, variables,
			metadata, version, created_at, updated_at
		FROM templates WHERE type = $1
		ORDER BY version DESC LIMIT 1
	`

	var template domain.Template
	err := r.db.QueryRow(query, reportType).Scan(
		&template.ID, &template.Name, &template.Type, &template.Content,
		&template.Parameters, &template.Variables, &template.Metadata,
		&template.Version, &template.CreatedAt, &template.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &template, nil
}

// Update updates an existing template record
func (r *TemplateRepository) Update(template *domain.Template) error {
	query := `
		UPDATE templates SET
			name = $1, content = $2, parameters = $3, variables = $4,
			metadata = $5, version = version + 1, updated_at = $6
		WHERE id = $7
	`

	template.UpdatedAt = time.Now()

	_, err := r.db.Exec(query,
		template.Name, template.Content, template.Parameters, template.Variables,
		template.Metadata, template.UpdatedAt, template.ID,
	)

	return err
}

// Delete deletes a template by its ID
func (r *TemplateRepository) Delete(id string) error {
	query := "DELETE FROM templates WHERE id = $1"
	_, err := r.db.Exec(query, id)
	return err
}

// List retrieves all templates with optional filtering
func (r *TemplateRepository) List(filter TemplateListFilter) (*domain.PaginatedTemplates, error) {
	baseQuery := `SELECT id, name, type, content, parameters, variables,
			metadata, version, created_at, updated_at FROM templates`
	countQuery := "SELECT COUNT(*) FROM templates"

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

	templates := make([]*domain.Template, 0)
	for rows.Next() {
		var template domain.Template
		err := rows.Scan(
			&template.ID, &template.Name, &template.Type, &template.Content,
			&template.Parameters, &template.Variables, &template.Metadata,
			&template.Version, &template.CreatedAt, &template.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		templates = append(templates, &template)
	}

	return &domain.PaginatedTemplates{
		Templates: templates,
		Total:     total,
		Offset:    filter.Offset,
		Limit:     filter.Limit,
	}, nil
}

// TemplateListFilter represents filter options for listing templates
type TemplateListFilter struct {
	Type   domain.ReportType
	Offset int
	Limit  int
}

// ScheduleRepository handles database operations for schedules
type ScheduleRepository struct {
	db *sql.DB
}

// NewScheduleRepository creates a new ScheduleRepository instance
func NewScheduleRepository(db *sql.DB) *ScheduleRepository {
	return &ScheduleRepository{db: db}
}

// Create creates a new schedule record
func (r *ScheduleRepository) Create(schedule *domain.Schedule) error {
	query := `
		INSERT INTO schedules (
			id, name, report_type, format, cron, enabled, parameters,
			filters, recipients, metadata, last_run, next_run,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	schedule.ID = uuid.New().String()
	schedule.CreatedAt = time.Now()
	schedule.UpdatedAt = time.Now()

	_, err := r.db.Exec(query,
		schedule.ID, schedule.Name, schedule.ReportType, schedule.Format,
		schedule.Cron, schedule.Enabled, schedule.Parameters, schedule.Filters,
		schedule.Recipients, schedule.Metadata, schedule.LastRun, schedule.NextRun,
		schedule.CreatedAt, schedule.UpdatedAt,
	)

	return err
}

// GetByID retrieves a schedule by its ID
func (r *ScheduleRepository) GetByID(id string) (*domain.Schedule, error) {
	query := `
		SELECT id, name, report_type, format, cron, enabled, parameters,
			filters, recipients, metadata, last_run, next_run,
			created_at, updated_at
		FROM schedules WHERE id = $1
	`

	var schedule domain.Schedule
	err := r.db.QueryRow(query, id).Scan(
		&schedule.ID, &schedule.Name, &schedule.ReportType, &schedule.Format,
		&schedule.Cron, &schedule.Enabled, &schedule.Parameters, &schedule.Filters,
		&schedule.Recipients, &schedule.Metadata, &schedule.LastRun, &schedule.NextRun,
		&schedule.CreatedAt, &schedule.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &schedule, nil
}

// Update updates an existing schedule record
func (r *ScheduleRepository) Update(schedule *domain.Schedule) error {
	query := `
		UPDATE schedules SET
			name = $1, cron = $2, enabled = $3, parameters = $4,
			filters = $5, recipients = $6, metadata = $7,
			next_run = $8, updated_at = $9
		WHERE id = $10
	`

	schedule.UpdatedAt = time.Now()

	_, err := r.db.Exec(query,
		schedule.Name, schedule.Cron, schedule.Enabled, schedule.Parameters,
		schedule.Filters, schedule.Recipients, schedule.Metadata,
		schedule.NextRun, schedule.UpdatedAt, schedule.ID,
	)

	return err
}

// UpdateNextRun updates the next run time for a schedule
func (r *ScheduleRepository) UpdateNextRun(id string, nextRun time.Time) error {
	query := `UPDATE schedules SET next_run = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.Exec(query, nextRun, time.Now(), id)
	return err
}

// UpdateLastRun updates the last run time for a schedule
func (r *ScheduleRepository) UpdateLastRun(id string, lastRun time.Time) error {
	query := `UPDATE schedules SET last_run = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.Exec(query, lastRun, time.Now(), id)
	return err
}

// Delete deletes a schedule by its ID
func (r *ScheduleRepository) Delete(id string) error {
	query := "DELETE FROM schedules WHERE id = $1"
	_, err := r.db.Exec(query, id)
	return err
}

// List retrieves all schedules with optional filtering
func (r *ScheduleRepository) List(filter ScheduleListFilter) (*domain.PaginatedSchedules, error) {
	baseQuery := `SELECT id, name, report_type, format, cron, enabled, parameters,
			filters, recipients, metadata, last_run, next_run,
			created_at, updated_at FROM schedules`
	countQuery := "SELECT COUNT(*) FROM schedules"

	whereClause := ""
	args := []interface{}{}
	argIndex := 1

	if filter.ReportType != "" {
		if whereClause != "" {
			whereClause += " AND "
		}
		whereClause += fmt.Sprintf("report_type = $%d", argIndex)
		args = append(args, filter.ReportType)
		argIndex++
	}

	if filter.Enabled != nil {
		if whereClause != "" {
			whereClause += " AND "
		}
		whereClause += fmt.Sprintf("enabled = $%d", argIndex)
		args = append(args, *filter.Enabled)
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

	schedules := make([]*domain.Schedule, 0)
	for rows.Next() {
		var schedule domain.Schedule
		err := rows.Scan(
			&schedule.ID, &schedule.Name, &schedule.ReportType, &schedule.Format,
			&schedule.Cron, &schedule.Enabled, &schedule.Parameters, &schedule.Filters,
			&schedule.Recipients, &schedule.Metadata, &schedule.LastRun, &schedule.NextRun,
			&schedule.CreatedAt, &schedule.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, &schedule)
	}

	return &domain.PaginatedSchedules{
		Schedules: schedules,
		Total:     total,
		Offset:    filter.Offset,
		Limit:     filter.Limit,
	}, nil
}

// ScheduleListFilter represents filter options for listing schedules
type ScheduleListFilter struct {
	ReportType domain.ReportType
	Enabled    *bool
	Offset     int
	Limit      int
}
