package models

import (
	"time"

	content "github.com/twilio/twilio-go/rest/content/v1"
)

// ContentType represents a Twilio content template type
type ContentType string

const (
	// Twilio content types
	ContentTypeTwilioText         ContentType = "twilio/text"
	ContentTypeTwilioMedia        ContentType = "twilio/media"
	ContentTypeTwilioLocation     ContentType = "twilio/location"
	ContentTypeTwilioListPicker   ContentType = "twilio/list-picker"
	ContentTypeTwilioCallToAction ContentType = "twilio/call-to-action"
	ContentTypeTwilioQuickReply   ContentType = "twilio/quick-reply"
	ContentTypeTwilioCard         ContentType = "twilio/card"
	ContentTypeTwilioCarousel     ContentType = "twilio/carousel"
	ContentTypeTwilioCatalog      ContentType = "twilio/catalog"
	ContentTypeTwilioPay          ContentType = "twilio/pay"
	ContentTypeTwilioFlows        ContentType = "twilio/flows"

	// WhatsApp specific content types
	ContentTypeWhatsAppAuthentication ContentType = "whatsapp/authentication"
	ContentTypeWhatsAppCard           ContentType = "whatsapp/card"
)

// String returns the string representation of the ContentType
func (ct ContentType) String() string {
	return string(ct)
}

// IsValid checks if the ContentType is a valid Twilio content type
func (ct ContentType) IsValid() bool {
	switch ct {
	case ContentTypeTwilioText, ContentTypeTwilioMedia, ContentTypeTwilioLocation,
		ContentTypeTwilioListPicker, ContentTypeTwilioCallToAction, ContentTypeTwilioQuickReply,
		ContentTypeTwilioCard, ContentTypeTwilioCarousel, ContentTypeTwilioCatalog,
		ContentTypeTwilioPay, ContentTypeTwilioFlows, ContentTypeWhatsAppAuthentication,
		ContentTypeWhatsAppCard:
		return true
	default:
		return false
	}
}

// AllContentTypes returns all valid content types
func AllContentTypes() []ContentType {
	return []ContentType{
		ContentTypeTwilioText, ContentTypeTwilioMedia, ContentTypeTwilioLocation,
		ContentTypeTwilioListPicker, ContentTypeTwilioCallToAction, ContentTypeTwilioQuickReply,
		ContentTypeTwilioCard, ContentTypeTwilioCarousel, ContentTypeTwilioCatalog,
		ContentTypeTwilioPay, ContentTypeTwilioFlows, ContentTypeWhatsAppAuthentication,
		ContentTypeWhatsAppCard,
	}
}

// ActionType represents the type of action button
type ActionType string

const (
	ActionTypeURL         ActionType = "URL"
	ActionTypePhoneNumber ActionType = "PHONE_NUMBER"
	ActionTypeQuickReply  ActionType = "QUICK_REPLY"
	ActionTypeCopyCode    ActionType = "COPY_CODE"
	ActionTypeVoiceCall   ActionType = "VOICE_CALL"
)

// CallToActionButton represents a button in a call-to-action template
type CallToActionButton struct {
	Type  ActionType `json:"type"`
	Title string     `json:"title"`           // Max 20 chars for WhatsApp
	URL   string     `json:"url,omitempty"`   // Required for URL type
	Phone string     `json:"phone,omitempty"` // Required for PHONE_NUMBER type (E.164 format)
}

// CreateTemplateDTO represents the request to create a new template with call-to-action
type CreateTemplateDTO struct {
	FriendlyName string               `json:"friendly_name"`
	Language     string               `json:"language"`
	Body         string               `json:"body"`                // Use {{1}}, {{2}}, etc. for variables
	Variables    map[string]string    `json:"variables,omitempty"` // e.g., {"1": "name", "2": "date"}
	Actions      []CallToActionButton `json:"actions,omitempty"`
}

// ToTwilioTypesFormat converts the CreateTemplateDTO to Twilio's types format
func (dto *CreateTemplateDTO) ToTwilioTypes() content.Types {
	// If no actions, use simple text template
	if len(dto.Actions) == 0 {
		return content.Types{
			TwilioText: &content.TwilioText{
				Body: dto.Body,
			},
		}
	}

	actions := make([]content.CallToActionAction, len(dto.Actions))
	for i, action := range dto.Actions {
		actions[i] = content.CallToActionAction{
			Type:  content.CallToActionActionType(action.Type),
			Title: action.Title,
			Url:   action.URL,
			Phone: action.Phone,
		}
	}

	return content.Types{
		TwilioCallToAction: &content.TwilioCallToAction{
			Body:    dto.Body,
			Actions: actions,
		},
	}
}

type SavedTemplate struct {
	ContentId    string         `json:"content_id"`
	FriendlyName string         `json:"friendly_name"`
	Language     string         `json:"language"`
	Body         string         `json:"body"`
	Variables    map[string]any `json:"variables"`
	Types        interface{}    `json:"types"`
	DateCreated  string         `json:"date_created"`
	DateUpdated  string         `json:"date_updated"`
}

type WhatsappTemplateDTO struct {
	To          string `json:"to"`
	TimeFromNow string `json:"time_from_now"`
	TemplateId  string `json:"template"`
	Content     string `json:"content"`
	Language    string `json:"language"`
}

type WhatsappTemplate struct {
	To          string     `json:"to"`
	TimeFromNow *time.Time `json:"time_from_now"`
	TemplateId  string     `json:"template"`
	Content     string     `json:"content"`
	Language    string     `json:"language"`
}
