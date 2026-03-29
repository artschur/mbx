package models

import (
	"time"

	"github.com/google/uuid"
)

type ScheduledMessageType string

const (
	ScheduleTypeTemplate ScheduledMessageType = "template"
	ScheduleTypeFreeform ScheduledMessageType = "freeform"
)

type ScheduledMessage struct {
	Id                 uuid.UUID
	To                 string
	SendAt             time.Time
	Content            string
	ProviderTemplateId string
	Type               ScheduledMessageType
	Status             Status
	CreatedAt          time.Time
}
