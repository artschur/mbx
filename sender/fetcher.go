package sender

import (
	"context"
	"fmt"
	"log/slog"
	"mbx/models"
	"time"

	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
	content "github.com/twilio/twilio-go/rest/content/v1"
)

type WhatsappFetcher interface {
	GetTemplates(context.Context) ([]models.SavedTemplate, error)
	GetMessages(ctx context.Context, after time.Time) ([]models.SentMessage, error)
	// GetTemplateMessages(ctx context.Context, after time.Time) ([]models.SentMessage, error)
}

var _ WhatsappFetcher = (*TwilioFetcher)(nil)

type TwilioFetcher struct {
	cfg    *Config
	client *twilio.RestClient
}

func sp[T any](ptr *T) T {
	if ptr == nil {
		var zero T
		return zero
	}
	return *ptr
}

func NewTwilioFetcher(client *twilio.RestClient, cfg *Config) *TwilioFetcher {
	return &TwilioFetcher{
		client: client,
		cfg:    cfg,
	}
}

func (s *TwilioFetcher) GetMessages(ctx context.Context, after time.Time) ([]models.SentMessage, error) {
	params := &openapi.ListMessageParams{}
	params.SetPageSize(1000)
	params.SetLimit(1000)

	params.SetDateSentAfter(after)

	messages, err := s.client.Api.ListMessage(params)
	if err != nil {
		slog.Error("Error fetching messages from Twilio", "error", err, "after", after)
		return nil, fmt.Errorf("error getting messages from twilio: %w", err)
	}
	slog.Info("Fetched messages from Twilio", "count", len(messages), "after", after)

	sentMessages := make([]models.SentMessage, len(messages))
	for i, msg := range messages {
		sentMessages[i] = models.SentMessage{
			ID:           sp(msg.Sid),
			To:           sp(msg.To),
			Body:         sp(msg.Body),
			DateSent:     sp(msg.DateSent),
			ErrorCode:    sp(msg.ErrorCode),
			ErrorMessage: sp(msg.ErrorMessage),
			Status:       sp(msg.Status),
			Price:        sp(msg.Price),
			PriceUnit:    sp(msg.PriceUnit),
		}
	}

	return sentMessages, nil
}

func (s *TwilioFetcher) GetTemplates(ctx context.Context) ([]models.SavedTemplate, error) {
	contentService := content.NewApiServiceWithClient(s.client.Client)

	contentParams := &content.ListContentParams{}
	contentParams.SetLimit(100)
	contentParams.SetPageSize(1000)

	slog.Info("Fetching WhatsApp templates")
	contents, err := contentService.ListContent(contentParams)
	if err != nil {
		return nil, err
	}

	slog.Info("Fetched WhatsApp templates", "count", len(contents))

	var templatesOut []models.SavedTemplate = make([]models.SavedTemplate, len(contents))
	for i, c := range contents {
		var body string
		if c.Types != nil {
			typesMap := *c.Types

			// Check for twilio/text type
			if textType, ok := typesMap["twilio/text"]; ok {
				if textMap, ok := textType.(map[string]interface{}); ok {
					if bodyVal, ok := textMap["body"].(string); ok {
						body = bodyVal
					}
				}
			}
			// Check for twilio/call-to-action type
			if ctaType, ok := typesMap["twilio/call-to-action"]; ok {
				if ctaMap, ok := ctaType.(map[string]interface{}); ok {
					if bodyVal, ok := ctaMap["body"].(string); ok {
						body = bodyVal
					}
				}
			}

			// Add more types as needed (twilio/card, twilio/quick-reply, etc.)
		}

		friendlyName := ""
		if c.FriendlyName != nil {
			friendlyName = *c.FriendlyName
		}
		language := ""
		if c.Language != nil {
			language = *c.Language
		}
		variables := make(map[string]any)
		if c.Variables != nil {
			variables = *c.Variables
		}
		dateCreated := ""
		if c.DateCreated != nil {
			dateCreated = c.DateCreated.String()
		}
		dateUpdated := ""
		if c.DateUpdated != nil {
			dateUpdated = c.DateUpdated.String()
		}

		contentId := ""
		if c.Sid != nil {
			contentId = *c.Sid
		}

		template := models.SavedTemplate{
			ContentId:    contentId,
			FriendlyName: friendlyName,
			Language:     language,
			Variables:    variables,
			Types:        c.Types,
			Body:         body,
			DateCreated:  dateCreated,
			DateUpdated:  dateUpdated,
		}

		templatesOut[i] = template
	}

	return templatesOut, nil
}
