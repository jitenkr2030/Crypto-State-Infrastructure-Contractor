package service

import (
	"context"
	"fmt"
	"time"

	"github.com/csic/platform/service/reporting/regulatory/internal/config"
	"github.com/csic/platform/service/reporting/regulatory/internal/domain"
	"github.com/csic/platform/service/reporting/regulatory/internal/messaging"
	"github.com/csic/platform/service/reporting/regulatory/internal/repository"
	"github.com/google/uuid"
)

// ScheduleService handles scheduled report operations
type ScheduleService struct {
	config       *config.Config
	scheduleRepo repository.ScheduleRepository
	reportService *ReportService
	producer     messaging.KafkaProducer
}

// NewScheduleService creates a new ScheduleService instance
func NewScheduleService(
	cfg *config.Config,
	scheduleRepo repository.ScheduleRepository,
	reportService *ReportService,
	producer messaging.KafkaProducer,
) *ScheduleService {
	return &ScheduleService{
		config:        cfg,
		scheduleRepo:  scheduleRepo,
		reportService: reportService,
		producer:      producer,
	}
}

// CreateSchedule creates a new schedule
func (s *ScheduleService) CreateSchedule(ctx context.Context, req *domain.CreateScheduleRequest) (*domain.Schedule, error) {
	schedule := &domain.Schedule{
		Name:       req.Name,
		ReportType: req.ReportType,
		Format:     req.Format,
		Cron:       req.Cron,
		Enabled:    req.Enabled,
		Parameters: req.Parameters,
		Filters:    req.Filters,
		Recipients: req.Recipients,
	}

	if schedule.Format == "" {
		schedule.Format = domain.ReportFormatPDF
	}

	if schedule.Enabled {
		nextRun, err := s.calculateNextRun(schedule.Cron)
		if err != nil {
			return nil, fmt.Errorf("invalid cron expression: %w", err)
		}
		schedule.NextRun = &nextRun
	}

	if err := s.scheduleRepo.Create(schedule); err != nil {
		return nil, fmt.Errorf("failed to create schedule: %w", err)
	}

	return schedule, nil
}

// GetSchedule retrieves a schedule by ID
func (s *ScheduleService) GetSchedule(ctx context.Context, id string) (*domain.Schedule, error) {
	schedule, err := s.scheduleRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get schedule: %w", err)
	}
	if schedule == nil {
		return nil, fmt.Errorf("schedule not found: %s", id)
	}
	return schedule, nil
}

// ListSchedules lists all schedules with optional filtering
func (s *ScheduleService) ListSchedules(ctx context.Context, filter repository.ScheduleListFilter) (*domain.PaginatedSchedules, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	return s.scheduleRepo.List(filter)
}

// UpdateSchedule updates an existing schedule
func (s *ScheduleService) UpdateSchedule(ctx context.Context, id string, req *domain.UpdateScheduleRequest) (*domain.Schedule, error) {
	schedule, err := s.scheduleRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get schedule: %w", err)
	}
	if schedule == nil {
		return nil, fmt.Errorf("schedule not found: %s", id)
	}

	if req.Name != nil {
		schedule.Name = *req.Name
	}
	if req.Cron != nil {
		schedule.Cron = *req.Cron
		if schedule.Enabled {
			nextRun, err := s.calculateNextRun(schedule.Cron)
			if err != nil {
				return nil, fmt.Errorf("invalid cron expression: %w", err)
			}
			schedule.NextRun = &nextRun
		}
	}
	if req.Enabled != nil {
		schedule.Enabled = *req.Enabled
		if schedule.Enabled && schedule.NextRun == nil {
			nextRun, err := s.calculateNextRun(schedule.Cron)
			if err != nil {
				return nil, fmt.Errorf("invalid cron expression: %w", err)
			}
			schedule.NextRun = &nextRun
		}
	}
	if req.Parameters != nil {
		schedule.Parameters = req.Parameters
	}
	if req.Filters != nil {
		schedule.Filters = *req.Filters
	}
	if req.Recipients != nil {
		schedule.Recipients = req.Recipients
	}

	if err := s.scheduleRepo.Update(schedule); err != nil {
		return nil, fmt.Errorf("failed to update schedule: %w", err)
	}

	return schedule, nil
}

// DeleteSchedule deletes a schedule
func (s *ScheduleService) DeleteSchedule(ctx context.Context, id string) error {
	schedule, err := s.scheduleRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get schedule: %w", err)
	}
	if schedule == nil {
		return fmt.Errorf("schedule not found: %s", id)
	}

	if err := s.scheduleRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete schedule: %w", err)
	}

	return nil
}

// TriggerSchedule triggers immediate report generation for a schedule
func (s *ScheduleService) TriggerSchedule(ctx context.Context, id string) (*domain.Report, error) {
	schedule, err := s.scheduleRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get schedule: %w", err)
	}
	if schedule == nil {
		return nil, fmt.Errorf("schedule not found: %s", id)
	}

	return s.reportService.TriggerScheduledReport(ctx, schedule)
}

// StartScheduler starts the background scheduler
func (s *ScheduleService) StartScheduler(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.runScheduledReports(ctx)
		}
	}
}

// runScheduledReports checks and executes due schedules
func (s *ScheduleService) runScheduledReports(ctx context.Context) {
	enabled := true
	filter := repository.ScheduleListFilter{Enabled: &enabled, Limit: 100}

	schedules, err := s.scheduleRepo.List(filter)
	if err != nil {
		return
	}

	now := time.Now()
	for _, schedule := range schedules.Schedules {
		if schedule.NextRun == nil || schedule.NextRun.After(now) {
			continue
		}

		// Trigger report generation
		report, err := s.reportService.TriggerScheduledReport(ctx, schedule)
		if err != nil {
			continue
		}

		// Calculate next run time
		nextRun, err := s.calculateNextRun(schedule.Cron)
		if err != nil {
			continue
		}

		// Update schedule
		s.scheduleRepo.UpdateLastRun(schedule.ID, now)
		s.scheduleRepo.UpdateNextRun(schedule.ID, nextRun)

		_ = report // Use the report variable
	}
}

// calculateNextRun calculates the next run time from a cron expression
func (s *ScheduleService) calculateNextRun(cronExpr string) (time.Time, error) {
	// Simplified cron parsing - in production, use a proper cron library
	// This is a placeholder that returns a default next run time
	// For full cron support, integrate with github.com/robfig/cron/v3

	now := time.Now()

	// Parse simple cron expressions (minute hour day month dow)
	// This is a simplified implementation
	switch {
	case len(cronExpr) == 0:
		return now.Add(24 * time.Hour), nil
	default:
		// For complex cron expressions, use a default of 1 hour
		return now.Add(1 * time.Hour), nil
	}
}

// GenerateExecutionID generates a unique execution ID
func GenerateExecutionID() string {
	return uuid.New().String()
}
