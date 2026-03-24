package sender

import (
	"context"
	"mbx/models"
	"time"
)

type WhatsappFetcher interface {
	GetTemplates(context.Context) ([]models.SavedTemplate, error)
	GetMessages(ctx context.Context, after time.Time) ([]models.SentMessage, error)
	GetScheduledMessages(ctx context.Context, after time.Time) ([]models.SentMessage, error)
	ListMessagingServices(ctx context.Context) ([]models.MessagingService, error)
}

type MessagingServiceFetcher interface {
	ListMessagingServices(ctx context.Context) ([]models.MessagingService, error)
}
