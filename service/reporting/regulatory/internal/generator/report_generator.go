package generator

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/csic/platform/service/reporting/regulatory/internal/config"
	"github.com/csic/platform/service/reporting/regulatory/internal/domain"
	"github.com/csic/platform/service/reporting/regulatory/internal/repository"
)

// ReportGenerator generates reports in various formats
type ReportGenerator struct {
	config       *config.Config
	templateRepo *repository.TemplateRepository
}

// NewReportGenerator creates a new ReportGenerator instance
func NewReportGenerator(cfg *config.Config, templateRepo *repository.TemplateRepository) *ReportGenerator {
	return &ReportGenerator{
		config:       cfg,
		templateRepo: templateRepo,
	}
}

// Generate generates a report based on the report configuration
func (g *ReportGenerator) Generate(ctx context.Context, report *domain.Report) (*domain.ReportResult, string, int64, error) {
	// Create reports directory if it doesn't exist
	reportsDir := g.config.Reporting.Storage.BasePath
	if err := os.MkdirAll(reportsDir, 0755); err != nil {
		return nil, "", 0, fmt.Errorf("failed to create reports directory: %w", err)
	}

	// Generate filename
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s_%s.%s", report.Type, report.ID[:8], timestamp, report.Format)
	filePath := filepath.Join(reportsDir, filename)

	// Generate report data
	result, err := g.generateReportData(ctx, report)
	if err != nil {
		return nil, "", 0, fmt.Errorf("failed to generate report data: %w", err)
	}

	// Write report file
	switch report.Format {
	case domain.ReportFormatCSV:
		if err := g.writeCSV(filePath, result); err != nil {
			return nil, "", 0, fmt.Errorf("failed to write CSV: %w", err)
		}
	case domain.ReportFormatJSON:
		if err := g.writeJSON(filePath, result); err != nil {
			return nil, "", 0, fmt.Errorf("failed to write JSON: %w", err)
		}
	case domain.ReportFormatPDF:
		// For PDF, we'll create a simple text file as placeholder
		// In production, integrate with a PDF library like gofpdf
		if err := g.writeText(filePath, result); err != nil {
			return nil, "", 0, fmt.Errorf("failed to write PDF: %w", err)
		}
	case domain.ReportFormatXLSX:
		// For XLSX, we'll create a CSV with .xlsx extension as placeholder
		// In production, integrate with excelize library
		if err := g.writeCSV(filePath, result); err != nil {
			return nil, "", 0, fmt.Errorf("failed to write XLSX: %w", err)
		}
	}

	// Get file size
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, "", 0, fmt.Errorf("failed to get file info: %w", err)
	}

	return result, filePath, fileInfo.Size(), nil
}

// generateReportData generates the data for a report
func (g *ReportGenerator) generateReportData(ctx context.Context, report *domain.Report) (*domain.ReportResult, error) {
	// This is a placeholder implementation
	// In production, this would fetch data from the compliance and blockchain services

	result := &domain.ReportResult{
		TotalRecords:    100,
		FilteredRecords: 50,
		Summary: map[string]interface{}{
			"report_type":    report.Type,
			"generated_at":   time.Now().Format(time.RFC3339),
			"total_amount":   1500000.00,
			"risk_score_avg": 0.35,
			"high_risk_count": 5,
			"medium_risk_count": 15,
			"low_risk_count": 30,
		},
		Data: []map[string]interface{}{
			{
				"id":              "txn_001",
				"timestamp":       time.Now().Format(time.RFC3339),
				"type":            "transfer",
				"amount":          5000.00,
				"currency":        "ETH",
				"from_address":    "0x1234567890abcdef",
				"to_address":      "0xfedcba0987654321",
				"risk_level":      "low",
				"risk_score":      0.15,
				"status":          "completed",
			},
			{
				"id":              "txn_002",
				"timestamp":       time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
				"type":            "swap",
				"amount":          25000.00,
				"currency":        "USDT",
				"from_address":    "0xabcdef1234567890",
				"to_address":      "0x0987654321abcdef",
				"risk_level":      "medium",
				"risk_score":      0.45,
				"status":          "pending",
			},
		},
		Errors: []domain.ReportError{},
	}

	return result, nil
}

// writeCSV writes the report data to a CSV file
func (g *ReportGenerator) writeCSV(filePath string, result *domain.ReportResult) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	if len(result.Data) > 0 {
		headers := getHeaders(result.Data[0])
		if err := writer.Write(headers); err != nil {
			return fmt.Errorf("failed to write header: %w", err)
		}

		// Write data rows
		for _, row := range result.Data {
			values := make([]string, len(headers))
			for i, header := range headers {
				values[i] = fmt.Sprintf("%v", row[header])
			}
			if err := writer.Write(values); err != nil {
				return fmt.Errorf("failed to write row: %w", err)
			}
		}
	}

	return nil
}

// writeJSON writes the report data to a JSON file
func (g *ReportGenerator) writeJSON(filePath string, result *domain.ReportResult) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

// writeText writes the report data to a text file
func (g *ReportGenerator) writeText(filePath string, result *domain.ReportResult) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Write report summary
	fmt.Fprintf(file, "Regulatory Report\n")
	fmt.Fprintf(file, "=================\n\n")
	fmt.Fprintf(file, "Report Type: %s\n", result.Summary["report_type"])
	fmt.Fprintf(file, "Generated At: %s\n", result.Summary["generated_at"])
	fmt.Fprintf(file, "Total Records: %d\n", result.TotalRecords)
	fmt.Fprintf(file, "Filtered Records: %d\n\n", result.FilteredRecords)

	// Write summary statistics
	fmt.Fprintf(file, "Summary Statistics:\n")
	for key, value := range result.Summary {
		if key != "report_type" && key != "generated_at" {
			fmt.Fprintf(file, "  %s: %v\n", key, value)
		}
	}

	// Write sample data
	fmt.Fprintf(file, "\nSample Records:\n")
	for i, row := range result.Data {
		if i >= 10 { // Limit to first 10 records
			break
		}
		fmt.Fprintf(file, "  - %s: %s %s from %s\n",
			row["id"], row["amount"], row["currency"], row["type"])
	}

	return nil
}

// getHeaders returns the column headers from a data row
func getHeaders(row map[string]interface{}) []string {
	headers := make([]string, 0, len(row))
	for key := range row {
		headers = append(headers, key)
	}
	return headers
}
