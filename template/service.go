package template

import (
	"context"
	"fmt"
	"mbx/models"
)

type TemplateService interface {
	GetTemplates(ctx context.Context) ([]models.SavedTemplate, error)
	CreateTemplate(ctx context.Context, dto models.CreateTemplateDTO) (*models.SavedTemplate, error)
	ListMessagingServices(ctx context.Context) ([]models.MessagingService, error)
}

//go:generate mockgen -destination=mock_service_test.go -package=template mbx/template TemplateFetcher,TemplateSender,TemplateRepo

type TemplateFetcher interface {
	GetTemplates(ctx context.Context) ([]models.SavedTemplate, error)
	ListMessagingServices(ctx context.Context) ([]models.MessagingService, error)
}

type TemplateSender interface {
	CreateTemplate(ctx context.Context, dto models.CreateTemplateDTO) (*models.SavedTemplate, error)
}

type TemplateRepo interface {
	SaveTemplate(ctx context.Context, template models.SavedTemplate) error
	GetTemplates(ctx context.Context) ([]models.SavedTemplate, error)
}

type templateService struct {
	fetcher TemplateFetcher
	sender  TemplateSender
	repo    TemplateRepo
}

func NewTemplateService(f TemplateFetcher, s TemplateSender, r TemplateRepo) TemplateService {
	return &templateService{
		fetcher: f,
		sender:  s,
		repo:    r,
	}
}

func (s *templateService) GetTemplates(ctx context.Context) ([]models.SavedTemplate, error) {
	templates, err := s.fetcher.GetTemplates(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get templates: %w", err)
	}

	return templates, nil
}

func (s *templateService) CreateTemplate(ctx context.Context, dto models.CreateTemplateDTO) (*models.SavedTemplate, error) {
	t, err := s.sender.CreateTemplate(ctx, dto)
	if err != nil {
		return nil, fmt.Errorf("failed to create template: %w", err)
	}

	err = s.repo.SaveTemplate(ctx, *t)
	if err != nil {
		return nil, fmt.Errorf("failed to save template: %w", err)
	}

	return t, nil
}

func (s *templateService) ListMessagingServices(ctx context.Context) ([]models.MessagingService, error) {
	return s.fetcher.ListMessagingServices(ctx)
}
