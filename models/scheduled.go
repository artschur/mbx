package models

import (
	"time"
)

type ScheduledMessage struct {
	ID           int64      `json:"id"`
	To           string     `json:"to"`
	Body         string     `json:"body"`
	TemplateId   string     `json:"template_id,omitempty"`
	Content      string     `json:"content,omitempty"` // JSON variables
	Language     string     `json:"language,omitempty"`
	ScheduledFor time.Time  `json:"scheduled_for"`
	SentAt       *time.Time `json:"sent_at,omitempty"`
	Status       string     `json:"status"` // pending, sent, failed
	ErrorMessage string     `json:"error_message,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}
