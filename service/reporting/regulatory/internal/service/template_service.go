package service

import (
	"fmt"

	"github.com/csic/platform/service/reporting/regulatory/internal/config"
	"github.com/csic/platform/service/reporting/regulatory/internal/domain"
	"github.com/csic/platform/service/reporting/regulatory/internal/repository"
)

// TemplateService handles template operations
type TemplateService struct {
	config       *config.Config
	templateRepo *repository.TemplateRepository
}

// NewTemplateService creates a new TemplateService instance
func NewTemplateService(cfg *config.Config, templateRepo *repository.TemplateRepository) *TemplateService {
	return &TemplateService{
		config:       cfg,
		templateRepo: templateRepo,
	}
}

// CreateTemplate creates a new template
func (s *TemplateService) CreateTemplate(ctx interface{}, req *domain.CreateTemplateRequest) (*domain.Template, error) {
	template := &domain.Template{
		Name:       req.Name,
		Type:       req.Type,
		Content:    req.Content,
		Parameters: req.Parameters,
		Variables:  req.Variables,
	}

	if err := s.templateRepo.Create(template); err != nil {
		return nil, fmt.Errorf("failed to create template: %w", err)
	}

	return template, nil
}

// GetTemplate retrieves a template by ID
func (s *TemplateService) GetTemplate(ctx interface{}, id string) (*domain.Template, error) {
	template, err := s.templateRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}
	if template == nil {
		return nil, fmt.Errorf("template not found: %s", id)
	}
	return template, nil
}

// ListTemplates lists all templates with optional filtering
func (s *TemplateService) ListTemplates(ctx interface{}, filter repository.TemplateListFilter) (*domain.PaginatedTemplates, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	return s.templateRepo.List(filter)
}

// UpdateTemplate updates an existing template
func (s *TemplateService) UpdateTemplate(ctx interface{}, id string, req *domain.UpdateTemplateRequest) (*domain.Template, error) {
	template, err := s.templateRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}
	if template == nil {
		return nil, fmt.Errorf("template not found: %s", id)
	}

	if req.Name != nil {
		template.Name = *req.Name
	}
	if req.Content != nil {
		template.Content = *req.Content
	}
	if req.Parameters != nil {
		template.Parameters = req.Parameters
	}
	if req.Variables != nil {
		template.Variables = req.Variables
	}

	if err := s.templateRepo.Update(template); err != nil {
		return nil, fmt.Errorf("failed to update template: %w", err)
	}

	return template, nil
}

// DeleteTemplate deletes a template
func (s *TemplateService) DeleteTemplate(ctx interface{}, id string) error {
	template, err := s.templateRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get template: %w", err)
	}
	if template == nil {
		return fmt.Errorf("template not found: %s", id)
	}

	if err := s.templateRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}

	return nil
}
