package templates

import (
	"context"
)

type ITemplateService interface {
	List(context.Context) ([]SavedTemplate, error)
	Create(context.Context, CreateTemplateDTO) (*SavedTemplate, error)
}

type Service struct {
	repo TemplateRepository
}

func (s *Service) List(ctx context.Context) ([]SavedTemplate, error) {
	return s.repo.List(ctx)
}

func (s *Service) Create(ctx context.Context, dto CreateTemplateDTO) (*SavedTemplate, error) {
	return s.repo.Create(ctx, dto)
}
