package service

import (
	"context"
	"mbx/models"
	"mbx/persistence/postgres"
	"mbx/sender"
)

type MessagingService struct {
	w       sender.Whatsapp
	wt      sender.WhatsappTemplate
	repo    postgres.MessageRepository
	msgChan chan models.ScheduledMessage
}

func (m *MessagingService) SendMessage(ctx context.Context, msg models.ScheduledMessage) error {
	msg.Status = models.StatusPending

	err := m.repo.Create(ctx, msg)
	if err != nil {
		return err
	}

	_, err = m.w.Send(ctx, models.WhatsappBody{
		To:   msg.To,
		Body: msg.Content,
	})
	if err != nil {
		return err
	}

	return nil
}
