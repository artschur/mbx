package templates

import "context"

type TemplateRepository interface {
	List(context.Context) ([]SavedTemplate, error)
	Create(context.Context, CreateTemplateDTO) (*SavedTemplate, error)
}
