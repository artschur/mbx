package messaging

import (
	"context"
	"mbx/models"
	"time"

	api "github.com/twilio/twilio-go/rest/api/v2010"
)

type MessagingService interface {
	Send(ctx context.Context, body models.WhatsappBody) (*api.ApiV2010Message, error)
	SendTemplate(ctx context.Context, template models.WhatsappTemplate) (*api.ApiV2010Message, error)
	CancelMessage(ctx context.Context, twilioId string) error
	GetMessages(ctx context.Context, after time.Time) ([]models.SentMessage, error)
	GetScheduledMessages(ctx context.Context, after time.Time) ([]models.SentMessage, error)
}

//go:generate mockgen -destination=mock_messaging_test.go -package=messaging mbx/messaging MessageSender,MessageFetcher,SchedulerRepo

type MessageSender interface {
	Send(context.Context, models.WhatsappBody) (*api.ApiV2010Message, error)
	SendTemplate(context.Context, models.WhatsappTemplate) (*api.ApiV2010Message, error)
	CancelMessage(ctx context.Context, twilioId string) error
}

type MessageFetcher interface {
	GetMessages(ctx context.Context, after time.Time) ([]models.SentMessage, error)
	GetScheduledMessages(ctx context.Context, after time.Time) ([]models.SentMessage, error)
}

type SchedulerRepo interface {
	Save(ctx context.Context, msg *models.ScheduledMessage) error
	GetPending(ctx context.Context, limit int) ([]models.ScheduledMessage, error)
	MarkAsSent(ctx context.Context, id int64) error
	MarkAsFailed(ctx context.Context, id int64, errStr string) error
}

type messagingService struct {
	sender  MessageSender
	fetcher MessageFetcher
	repo    SchedulerRepo
}

func NewMessagingService(s MessageSender, f MessageFetcher, r SchedulerRepo) MessagingService {
	return &messagingService{
		sender:  s,
		fetcher: f,
		repo:    r,
	}
}

func (s *messagingService) Send(ctx context.Context, body models.WhatsappBody) (*api.ApiV2010Message, error) {
	// If scheduled for more than 35 days (or any threshold), store locally
	if body.TimeFromNow != nil && body.TimeFromNow.After(time.Now().Add(24*time.Hour)) {
		msg := &models.ScheduledMessage{
			To:           body.To,
			Body:         body.Body,
			ScheduledFor: *body.TimeFromNow,
			Status:       "pending",
		}
		err := s.repo.Save(ctx, msg)
		return nil, err // Return nil for Twilio message since it's not sent yet
	}
	return s.sender.Send(ctx, body)
}

func (s *messagingService) SendTemplate(ctx context.Context, template models.WhatsappTemplate) (*api.ApiV2010Message, error) {
	if template.TimeFromNow != nil && template.TimeFromNow.After(time.Now().Add(24*time.Hour)) {
		msg := &models.ScheduledMessage{
			To:           template.To,
			TemplateId:   template.TemplateId,
			Content:      template.Content,
			Language:     template.Language,
			ScheduledFor: *template.TimeFromNow,
			Status:       "pending",
		}
		err := s.repo.Save(ctx, msg)
		return nil, err
	}
	return s.sender.SendTemplate(ctx, template)
}

func (s *messagingService) CancelMessage(ctx context.Context, twilioId string) error {
	return s.sender.CancelMessage(ctx, twilioId)
}

func (s *messagingService) GetMessages(ctx context.Context, after time.Time) ([]models.SentMessage, error) {
	return s.fetcher.GetMessages(ctx, after)
}

func (s *messagingService) GetScheduledMessages(ctx context.Context, after time.Time) ([]models.SentMessage, error) {
	return s.fetcher.GetScheduledMessages(ctx, after)
}
