package messaging

import (
	"context"
	"mbx/models"
	"mbx/templates"
)

type whatsappService interface {
	Send(context.Context, models.WhatsappBody) error
	SendTemplate(context.Context, *templates.WhatsappTemplate) error
}

type Service struct {
	whatsappService whatsappService
}

func NewService(whatsappService whatsappService) *Service {
	return &Service{whatsappService: whatsappService}
}

func (s *Service) SendTemplate(ctx context.Context, template *templates.WhatsappTemplate) error {
	return s.whatsappService.SendTemplate(ctx, template)
}

func (s *Service) Send(ctx context.Context, message models.WhatsappBody) error {
	return s.whatsappService.Send(ctx, message)
}
