package repository

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// Database wraps the sql.DB connection with additional functionality
type Database struct {
	*sql.DB
}

// NewDatabase creates a new Database connection
func NewDatabase(config interface{}) (*Database, error) {
	// Type assertion to get config values
	cfg, ok := config.(struct {
		Host            string
		Port            int
		Username        string
		Password        string
		Name            string
		SSLMode         string
		MaxOpenConns    int
		MaxIdleConns    int
		ConnMaxLifetime int
	})
	if !ok {
		return nil, fmt.Errorf("invalid database config type")
	}

	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Name, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Database{db}, nil
}

// RunMigrations applies database migrations
func RunMigrations(db *Database) error {
	migrations := []string{
		// Reports table
		`CREATE TABLE IF NOT EXISTS reports (
			id VARCHAR(36) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			type VARCHAR(100) NOT NULL,
			format VARCHAR(20) NOT NULL DEFAULT 'pdf',
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			description TEXT,
			parameters JSONB DEFAULT '{}',
			filters JSONB DEFAULT '{}',
			result JSONB,
			metadata JSONB DEFAULT '{}',
			scheduled_id VARCHAR(36),
			generated_at TIMESTAMP,
			expires_at TIMESTAMP,
			file_path VARCHAR(500),
			file_size BIGINT DEFAULT 0,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,

		// Templates table
		`CREATE TABLE IF NOT EXISTS templates (
			id VARCHAR(36) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			type VARCHAR(100) NOT NULL,
			content TEXT,
			parameters JSONB DEFAULT '[]',
			variables JSONB DEFAULT '[]',
			metadata JSONB DEFAULT '{}',
			version INT DEFAULT 1,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,

		// Schedules table
		`CREATE TABLE IF NOT EXISTS schedules (
			id VARCHAR(36) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			report_type VARCHAR(100) NOT NULL,
			format VARCHAR(20) NOT NULL DEFAULT 'pdf',
			cron VARCHAR(100) NOT NULL,
			enabled BOOLEAN DEFAULT true,
			parameters JSONB DEFAULT '{}',
			filters JSONB DEFAULT '{}',
			recipients JSONB DEFAULT '[]',
			metadata JSONB DEFAULT '{}',
			last_run TIMESTAMP,
			next_run TIMESTAMP,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,

		// Report executions table (for tracking scheduled report runs)
		`CREATE TABLE IF NOT EXISTS report_executions (
			id VARCHAR(36) PRIMARY KEY,
			schedule_id VARCHAR(36) REFERENCES schedules(id) ON DELETE CASCADE,
			report_id VARCHAR(36) REFERENCES reports(id) ON DELETE SET NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			started_at TIMESTAMP NOT NULL DEFAULT NOW(),
			completed_at TIMESTAMP,
			error_message TEXT,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,

		// Indexes
		`CREATE INDEX IF NOT EXISTS idx_reports_type ON reports(type)`,
		`CREATE INDEX IF NOT EXISTS idx_reports_status ON reports(status)`,
		`CREATE INDEX IF NOT EXISTS idx_reports_created_at ON reports(created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_reports_scheduled ON reports(scheduled_id)`,
		`CREATE INDEX IF NOT EXISTS idx_templates_type ON templates(type)`,
		`CREATE INDEX IF NOT EXISTS idx_schedules_report_type ON schedules(report_type)`,
		`CREATE INDEX IF NOT EXISTS idx_schedules_enabled ON schedules(enabled)`,
		`CREATE INDEX IF NOT EXISTS idx_schedules_next_run ON schedules(next_run)`,
		`CREATE INDEX IF NOT EXISTS idx_executions_schedule ON report_executions(schedule_id)`,
		`CREATE INDEX IF NOT EXISTS idx_executions_status ON report_executions(status)`,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return nil
}
