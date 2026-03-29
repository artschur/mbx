package schedules

import (
	"context"
	"log/slog"
	"mbx/messaging"
	"mbx/models"
	"mbx/templates"
	"time"
)

type Config struct {
	PoolingRate time.Duration
}

type Worker struct {
	config  Config
	service messaging.Service
	repo    Repository
}

func NewWorker(config Config, service messaging.Service, repo Repository) *Worker {
	return &Worker{
		config:  config,
		service: service,
		repo:    repo,
	}
}

func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.config.PoolingRate)

	for {
		select {
		case <-ticker.C:
			upcoming, err := w.repo.ListUpcoming(ctx, 5*time.Minute)
			if err != nil {
				slog.Error("failed to list upcoming messages", slog.Any("error", err))
				continue
			}
			for _, msg := range upcoming {
				w.Send(ctx, msg)
			}

		case <-ctx.Done():
			return
		}
	}
}

func (w *Worker) Send(ctx context.Context, msg models.ScheduledMessage) {
	switch msg.Type {
	case models.ScheduleTypeTemplate:
		err := w.service.SendTemplate(ctx,
			&templates.WhatsappTemplate{
				To:         msg.To,
				TemplateId: msg.ProviderTemplateId,
				Content:    msg.Content,
				Language:   "pt_BR",
			})
		if err != nil {
			slog.Error("failed to send template message", slog.Any("error", err))
		}

	case models.ScheduleTypeFreeform:
		err := w.service.Send(ctx, models.WhatsappBody{
			To:   msg.To,
			Body: msg.Content,
		})

		if err != nil {
			slog.Error("failed to send freeform message", slog.Any("error", err))
		}
	}

}
