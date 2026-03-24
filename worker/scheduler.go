package worker

import (
	"context"
	"log/slog"
	"mbx/persistence/postgres"
	"time"

	api "github.com/twilio/twilio-go/rest/api/v2010"
)

type SendFunc func(ctx context.Context, to string, body string, templateId string, content string, language string) (*api.ApiV2010Message, error)

type SchedulerWorker struct {
	repo     postgres.SchedulerRepository
	sendFunc SendFunc
	interval time.Duration
}

func NewSchedulerWorker(repo postgres.SchedulerRepository, sendFunc SendFunc, interval time.Duration) *SchedulerWorker {
	return &SchedulerWorker{
		repo:     repo,
		sendFunc: sendFunc,
		interval: interval,
	}
}

func (w *SchedulerWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	slog.Info("Scheduler worker started", "interval", w.interval)

	for {
		select {
		case <-ctx.Done():
			slog.Info("Scheduler worker stopping")
			return
		case <-ticker.C:
			w.processPending(ctx)
		}
	}
}

func (w *SchedulerWorker) processPending(ctx context.Context) {
	msgs, err := w.repo.GetPending(ctx, 10)
	if err != nil {
		slog.Error("Failed to fetch pending messages", "error", err)
		return
	}

	for _, msg := range msgs {
		slog.Info("Processing scheduled message", "id", msg.ID, "to", msg.To)

		_, err := w.sendFunc(ctx, msg.To, msg.Body, msg.TemplateId, msg.Content, msg.Language)
		if err != nil {
			slog.Error("Failed to send scheduled message", "id", msg.ID, "error", err)
			_ = w.repo.MarkAsFailed(ctx, msg.ID, err.Error())
			continue
		}

		err = w.repo.MarkAsSent(ctx, msg.ID)
		if err != nil {
			slog.Error("Failed to mark message as sent", "id", msg.ID, "error", err)
		}
	}
}
