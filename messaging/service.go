package messaging

import (
	"context"
	"mbx/models"
	"mbx/templates"
)

type templateService interface {
	List(context.Context) ([]templates.SavedTemplate, error)
	Create(context.Context, templates.CreateTemplateDTO) (*templates.SavedTemplate, error)
}

type whatsappService interface {
	Send(context.Context, models.WhatsappBody)
}

type Service struct {
	templateService templateService
}

func NewService(templateService templateService) *Service {
	return &Service{templateService: templateService}
}
