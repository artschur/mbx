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

func NewTwilioSender(cfg *Config) *TwilioSender {

	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username:   cfg.TwilioAccountSID,
		Password:   cfg.TwilioAuthToken,
		AccountSid: cfg.TwilioAccountSID,
	})

	return &TwilioSender{
		client: client,
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
	messageParams := &api.CreateMessageParams{
		To:               &template.To,
		From:             &s.cfg.TwilioFromNumber,
		ContentSid:       &template.TemplateId,
		ContentVariables: &template.Content,
	}

	if template.TimeFromNow != nil {
		messageParams.SendAt = template.TimeFromNow
	}

	resp, err := s.client.Api.CreateMessage(messageParams)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %v", err)
	}

	return resp, nil
}

func (s *TwilioSender) GetTemplates(ctx context.Context) ([]models.SavedTemplate, error) {
	contentService := content.NewApiServiceWithClient(s.client.Client)

	contentParams := &content.ListContentParams{}
	contentParams.SetLimit(100)
	contentParams.SetPageSize(1000)

	slog.Info("Fetching WhatsApp templates")
	content, err := contentService.ListContent(contentParams)
	if err != nil {
		return nil, err
	}

	slog.Info("Fetched WhatsApp templates", "count", len(content))

	var templatesOut []models.SavedTemplate = make([]models.SavedTemplate, len(content))
	for i, c := range content {
		var body string
		if c.Types != nil {
			for _, t := range *c.Types {
				if typ, ok := t.(map[string]interface{}); ok {
					if typeStr, ok := typ["type"].(string); ok {
						contentType := models.ContentType(typeStr)
						if contentType == models.ContentTypeTwilioText {
							if bodyVal, ok := typ["body"].(string); ok {
								body = bodyVal
								break
							}
						}
					}
				}
			}
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

		template := models.SavedTemplate{
			FriendlyName: friendlyName,
			Language:     language,
			Variables:    variables,
			Body:         body,
			DateCreated:  dateCreated,
			DateUpdated:  dateUpdated,
		}

		templatesOut[i] = template

	}

	return templatesOut, nil
}

func (s *TwilioSender) CreateTemplate(ctx context.Context, dto models.CreateTemplateDTO) (*models.SavedTemplate, error) {
	contentService := content.NewApiServiceWithClient(s.client.Client)

	// Convert DTO to Twilio's types format
	types := dto.ToTwilioTypes()

	// Prepare variables - convert to map[string]interface{}
	var variables map[string]string
	if dto.Variables != nil {
		variables = make(map[string]string)
		for k, v := range dto.Variables {
			variables[k] = v
		}
	}

	createParams := &content.CreateContentParams{
		ContentCreateRequest: &content.ContentCreateRequest{
			FriendlyName: dto.FriendlyName,
			Language:     dto.Language,
			Types:        types,
			Variables:    variables,
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

	createdTemplate := &models.SavedTemplate{
		FriendlyName: friendlyName,
		Language:     language,
		Variables:    vars,
		Types:        createdContent.Types,
		Body:         dto.Body,
		DateCreated:  dateCreated,
		DateUpdated:  dateUpdated,
	}

	return createdTemplate, nil
}
