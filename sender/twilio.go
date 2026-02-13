package sender

import (
	"context"
	"fmt"
	"log/slog"
	"mbx/models"

	"github.com/twilio/twilio-go"
	api "github.com/twilio/twilio-go/rest/api/v2010"
	content "github.com/twilio/twilio-go/rest/content/v1"
)

type TwilioSender struct {
	cfg    *Config
	client *twilio.RestClient
}

var _ WhatsappSender = (*TwilioSender)(nil)

func NewTwilioClient(cfg *Config) *twilio.RestClient {

	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username:   cfg.TwilioAccountSID,
		Password:   cfg.TwilioAuthToken,
		AccountSid: cfg.TwilioAccountSID,
	})

	return client
}

func NewTwilioSender(client *twilio.RestClient, cfg *Config) *TwilioSender {
	return &TwilioSender{
		client: client,
		cfg:    cfg,
	}
}

func (s *TwilioSender) Send(ctx context.Context, message models.WhatsappBody) (*api.ApiV2010Message, error) {
	messageParams := &api.CreateMessageParams{
		To:   &message.To,
		From: &s.cfg.TwilioFromNumber,
		Body: &message.Body,
	}

	if message.TimeFromNow != nil {
		messageParams.SendAt = message.TimeFromNow
	}

	resp, err := s.client.Api.CreateMessage(messageParams)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %v", err)
	}

	return resp, nil
}

func (s *TwilioSender) SendTemplate(ctx context.Context, template models.WhatsappTemplate) (*api.ApiV2010Message, error) {
	messageParams := &api.CreateMessageParams{}

	messageParams.SetTo(fmt.Sprintf("whatsapp:%s", template.To))
	messageParams.SetFrom(fmt.Sprintf("whatsapp:%s", s.cfg.TwilioFromNumber))

	messageParams.SetContentSid(template.TemplateId)

	if template.Content != "" && template.Content != "{}" && template.Content != "null" {
		messageParams.SetContentVariables(template.Content)
	}

	slog.Info("Sending template message with parameters",
		"template_id", template.TemplateId,
		"content_variables", template.Content)

	if template.TimeFromNow != nil {
		messageParams.SendAt = template.TimeFromNow
	}

	resp, err := s.client.Api.CreateMessage(messageParams)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %v", err)
	}
	slog.Info("Sent template message", "sid", *resp.Sid)

	return resp, nil
}

func (s *TwilioSender) CreateTemplate(ctx context.Context, dto models.CreateTemplateDTO) (*models.SavedTemplate, error) {
	contentService := content.NewApiServiceWithClient(s.client.Client)

	// Convert DTO to Twilio's Types struct
	types := dto.ToTwilioTypes()

	createParams := &content.CreateContentParams{
		ContentCreateRequest: &content.ContentCreateRequest{
			FriendlyName: dto.FriendlyName,
			Language:     dto.Language,
			Types:        types,
			Variables:    dto.Variables,
		},
	}

	slog.Info("Creating WhatsApp template", "friendly_name", dto.FriendlyName, "language", dto.Language)
	createdContent, err := contentService.CreateContent(createParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create template: %v", err)
	}

	slog.Info("Created WhatsApp template", "sid", *createdContent.Sid)

	// Build response with safe nil checks
	friendlyName := ""
	if createdContent.FriendlyName != nil {
		friendlyName = *createdContent.FriendlyName
	}
	language := ""
	if createdContent.Language != nil {
		language = *createdContent.Language
	}
	vars := make(map[string]any)
	if createdContent.Variables != nil {
		vars = *createdContent.Variables
	}
	dateCreated := ""
	if createdContent.DateCreated != nil {
		dateCreated = createdContent.DateCreated.String()
	}
	dateUpdated := ""
	if createdContent.DateUpdated != nil {
		dateUpdated = createdContent.DateUpdated.String()
	}

	// Extract body from the Types response
	var body string
	if createdContent.Types != nil {
		typesMap := *createdContent.Types
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
	}

	createdTemplate := &models.SavedTemplate{
		ContentId:    *createdContent.Sid,
		FriendlyName: friendlyName,
		Language:     language,
		Variables:    vars,
		Types:        createdContent.Types,
		Body:         body,
		DateCreated:  dateCreated,
		DateUpdated:  dateUpdated,
	}

	return createdTemplate, nil
}
