package sender

import (
	"context"
	"mbx/models"

	api "github.com/twilio/twilio-go/rest/api/v2010"
)

type WhatsappSender interface {
	Send(context.Context, models.WhatsappBody) (*api.ApiV2010Message, error)
	SendTemplate(context.Context, models.WhatsappTemplate) (*api.ApiV2010Message, error)
	GetTemplates(context.Context) ([]models.SavedTemplate, error)
	CreateTemplate(context.Context, models.CreateTemplateDTO) (*models.SavedTemplate, error)
}
