package sender

import (
	"context"
	"mbx/models"
	"mbx/templates"

	api "github.com/twilio/twilio-go/rest/api/v2010"
)

type Whatsapp interface {
	Send(context.Context, models.WhatsappBody) (*api.ApiV2010Message, error)
	CancelMessage(ctx context.Context, twilioId string) error
}

type WhatsappTemplate interface {
	SendTemplate(context.Context, templates.WhatsappTemplate) (*api.ApiV2010Message, error)
	CreateTemplate(context.Context, templates.CreateTemplateDTO) (*templates.SavedTemplate, error)
}
