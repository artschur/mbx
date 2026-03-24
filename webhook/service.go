package webhook

import (
	"context"
	"mbx/models"
	"mbx/webhook/persistence/postgres"
	"time"
)

type WebhookService interface {
	HandleWebhook(ctx context.Context, sid string, status string, raw string) error
	GetWebhookRequests(ctx context.Context, limit int) ([]models.WebhookRequest, error)
}

type webhookService struct {
	repo postgres.WebhookRepository
}

func NewWebhookService(repo postgres.WebhookRepository) WebhookService {
	return &webhookService{repo: repo}
}

func (s *webhookService) HandleWebhook(ctx context.Context, sid string, status string, raw string) error {
	req := &models.WebhookRequest{
		MessageSid: sid,
		Status:     status,
		RawData:    raw,
		ReceivedAt: time.Now(),
	}
	return s.repo.SaveWebhookRequest(ctx, req)
}

func (s *webhookService) GetWebhookRequests(ctx context.Context, limit int) ([]models.WebhookRequest, error) {
	return s.repo.ListWebhooks(ctx, limit)
}
