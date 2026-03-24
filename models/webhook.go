package models

import "time"

type MessageStatus struct {
	MessageSid   string    `json:"SmsSid"`
	Status       string    `json:"MessageStatus"`
	To           string    `json:"To"`
	From         string    `json:"From"`
	ErrorCode    string    `json:"ErrorCode,omitempty"`
	ErrorMessage string    `json:"ErrorMessage,omitempty"`
	DateUpdated  time.Time `json:"DateUpdated"`
}

type WebhookRequest struct {
	ID         int64     `json:"id"`
	MessageSid string    `json:"message_sid"`
	Status     string    `json:"status"`
	RawData    string    `json:"raw_data"`
	ReceivedAt time.Time `json:"received_at"`
}
