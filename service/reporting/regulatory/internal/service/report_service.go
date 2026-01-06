package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/csic/platform/service/reporting/regulatory/internal/config"
	"github.com/csic/platform/service/reporting/regulatory/internal/domain"
	"github.com/csic/platform/service/reporting/regulatory/internal/generator"
	"github.com/csic/platform/service/reporting/regulatory/internal/messaging"
	"github.com/csic/platform/service/reporting/regulatory/internal/repository"
)

// ReportService handles report operations
type ReportService struct {
	config          *config.Config
	reportRepo      *repository.ReportRepository
	producer        messaging.KafkaProducer
	reportGenerator *generator.ReportGenerator
}

// NewReportService creates a new ReportService instance
func NewReportService(
	cfg *config.Config,
	reportRepo *repository.ReportRepository,
	producer messaging.KafkaProducer,
	reportGenerator *generator.ReportGenerator,
) *ReportService {
	return &ReportService{
		config:          cfg,
		reportRepo:      reportRepo,
		producer:        producer,
		reportGenerator: reportGenerator,
	}
}

// CreateReport creates a new report
func (s *ReportService) CreateReport(ctx context.Context, req *domain.CreateReportRequest) (*domain.Report, error) {
	report := &domain.Report{
		Name:        req.Name,
		Type:        req.Type,
		Format:      req.Format,
		Description: req.Description,
		Parameters:  req.Parameters,
		Filters:     req.Filters,
		Status:      domain.ReportStatusPending,
	}

	if report.Format == "" {
		report.Format = domain.ReportFormatPDF
	}

	if err := s.reportRepo.Create(report); err != nil {
		return nil, fmt.Errorf("failed to create report: %w", err)
	}

	return report, nil
}

// GetReport retrieves a report by ID
func (s *ReportService) GetReport(ctx context.Context, id string) (*domain.Report, error) {
	report, err := s.reportRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get report: %w", err)
	}
	if report == nil {
		return nil, fmt.Errorf("report not found: %s", id)
	}
	return report, nil
}

// ListReports lists all reports with optional filtering
func (s *ReportService) ListReports(ctx context.Context, filter repository.ReportListFilter) (*domain.PaginatedReports, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	return s.reportRepo.List(filter)
}

// GenerateReport generates a report synchronously
func (s *ReportService) GenerateReport(ctx context.Context, id string) (*domain.Report, error) {
	report, err := s.GetReport(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update status to generating
	if err := s.reportRepo.UpdateStatus(id, domain.ReportStatusGenerating); err != nil {
		return nil, fmt.Errorf("failed to update status: %w", err)
	}

	// Generate the report
	result, filePath, fileSize, err := s.reportGenerator.Generate(ctx, report)
	if err != nil {
		s.reportRepo.UpdateStatus(id, domain.ReportStatusFailed)
		return nil, fmt.Errorf("failed to generate report: %w", err)
	}

	// Update report with result
	expiresAt := time.Now().Add(s.config.GetRetentionPeriod())
	if err := s.reportRepo.UpdateResult(id, result, filePath, fileSize); err != nil {
		return nil, fmt.Errorf("failed to update result: %w", err)
	}

	// Get updated report
	return s.GetReport(ctx, id)
}

// DeleteReport deletes a report
func (s *ReportService) DeleteReport(ctx context.Context, id string) error {
	report, err := s.reportRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get report: %w", err)
	}
	if report == nil {
		return fmt.Errorf("report not found: %s", id)
	}

	if err := s.reportRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete report: %w", err)
	}

	return nil
}

// ListReportTypes returns available report types
func (s *ReportService) ListReportTypes(ctx context.Context) []domain.ReportTypeInfo {
	return []domain.ReportTypeInfo{
		{
			Type:               domain.ReportTypeAMLMonitoring,
			Name:               "AML Monitoring Report",
			Description:        "Anti-Money Laundering monitoring and suspicious activity analysis",
			Category:           "Compliance",
			Formats:            []domain.ReportFormat{domain.ReportFormatPDF, domain.ReportFormatXLSX, domain.ReportFormatCSV},
			RegulatoryFramework: "FATF, BSA",
			RetentionDays:       2555,
		},
		{
			Type:               domain.ReportTypeCTFReport,
			Name:               "Counter Terrorist Financing Report",
			Description:        "CTF compliance and transaction monitoring report",
			Category:           "Compliance",
			Formats:            []domain.ReportFormat{domain.ReportFormatPDF, domain.ReportFormatXLSX},
			RegulatoryFramework: "FATF, USA PATRIOT Act",
			RetentionDays:       3650,
		},
		{
			Type:               domain.ReportTypeKYCReport,
			Name:               "Know Your Customer Report",
			Description:        "Customer due diligence and identity verification report",
			Category:           "KYC",
			Formats:            []domain.ReportFormat{domain.ReportFormatPDF, domain.ReportFormatXLSX},
			RegulatoryFramework: "FATF, GDPR, MiCA",
			RetentionDays:       2555,
		},
		{
			Type:               domain.ReportTypeSAR,
			Name:               "Suspicious Activity Report",
			Description:        "SAR filing for suspicious transactions and activities",
			Category:           "Regulatory",
			Formats:            []domain.ReportFormat{domain.ReportFormatPDF},
			RegulatoryFramework: "BSA, FinCEN",
			RetentionDays:       3650,
		},
		{
			Type:               domain.ReportTypeCTR,
			Name:               "Currency Transaction Report",
			Description:        "CTR for cash transactions exceeding reporting thresholds",
			Category:           "Regulatory",
			Formats:            []domain.ReportFormat{domain.ReportFormatPDF},
			RegulatoryFramework: "BSA, FinCEN",
			RetentionDays:       3650,
		},
		{
			Type:               domain.ReportTypeComplianceSummary,
			Name:               "Compliance Summary Report",
			Description:        "Executive summary of compliance activities and metrics",
			Category:           "Management",
			Formats:            []domain.ReportFormat{domain.ReportFormatPDF, domain.ReportFormatXLSX},
			RegulatoryFramework: "Multiple",
			RetentionDays:       1095,
		},
		{
			Type:               domain.ReportTypeRiskAssessment,
			Name:               "Risk Assessment Report",
			Description:        "Comprehensive risk analysis and assessment report",
			Category:           "Risk",
			Formats:            []domain.ReportFormat{domain.ReportFormatPDF, domain.ReportFormatXLSX},
			RegulatoryFramework: "Basel III, MiCA",
			RetentionDays:      1825,
		},
		{
			Type:               domain.ReportTypeTransactionHistory,
			Name:               "Transaction History Report",
			Description:        "Detailed transaction history and audit trail",
			Category:           "Operations",
			Formats:            []domain.ReportFormat{domain.ReportFormatPDF, domain.ReportFormatXLSX, domain.ReportFormatCSV, domain.ReportFormatJSON},
			RegulatoryFramework: "GDPR, Multiple",
			RetentionDays:       1825,
		},
		{
			Type:               domain.ReportTypeAuditTrail,
			Name:               "Audit Trail Report",
			Description:        "System and data access audit trail",
			Category:           "Audit",
			Formats:            []domain.ReportFormat{domain.ReportFormatPDF, domain.ReportFormatXLSX, domain.ReportFormatCSV},
			RegulatoryFramework: "SOX, GDPR",
			RetentionDays:       2555,
		},
		{
			Type:               domain.ReportTypeRegulatoryFiling,
			Name:               "Regulatory Filing Report",
			Description:        "Periodic regulatory filing and compliance documentation",
			Category:           "Regulatory",
			Formats:            []domain.ReportFormat{domain.ReportFormatPDF},
			RegulatoryFramework: "MiCA, FATF",
			RetentionDays:       3650,
		},
	}
}

// GetReportType returns information about a specific report type
func (s *ReportService) GetReportType(ctx context.Context, reportType domain.ReportType) (*domain.ReportTypeInfo, error) {
	types := s.ListReportTypes(ctx)
	for _, t := range types {
		if t.Type == reportType {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("unknown report type: %s", reportType)
}

// TriggerScheduledReport triggers report generation for a scheduled report
func (s *ReportService) TriggerScheduledReport(ctx context.Context, schedule *domain.Schedule) (*domain.Report, error) {
	req := &domain.CreateReportRequest{
		Name:        fmt.Sprintf("Scheduled: %s", schedule.Name),
		Type:        schedule.ReportType,
		Format:      schedule.Format,
		Parameters:  schedule.Parameters,
		Filters:     schedule.Filters,
		Description: fmt.Sprintf("Auto-generated report from schedule: %s", schedule.Name),
	}

	report, err := s.CreateReport(ctx, req)
	if err != nil {
		return nil, err
	}

	report.ScheduledID = schedule.ID

	// Update schedule last run time
	now := time.Now()

	// Publish event
	if s.producer != nil {
		event := map[string]interface{}{
			"event_type":  "report.triggered",
			"schedule_id": schedule.ID,
			"report_id":   report.ID,
			"report_type": schedule.ReportType,
			"timestamp":   now,
		}
		data, _ := json.Marshal(event)
		s.producer.Publish(ctx, "csic.reports.events", data)
	}

	// Update schedule last run
	if err := s.reportRepo.UpdateResult(report.ID, &domain.ReportResult{}, "", 0); err != nil {
		// Handle error but don't fail the operation
	}

	// Start async generation
	go func() {
		ctx := context.Background()
		s.GenerateReport(ctx, report.ID)
	}()

	return report, nil
}
