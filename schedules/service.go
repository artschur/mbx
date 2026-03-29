package schedules

import (
	"context"
	"mbx/models"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	FindById(context.Context, uuid.UUID) (*models.ScheduledMessage, error)
	ListUpcoming(context.Context, time.Duration) ([]models.ScheduledMessage, error)
	Create(context.Context, models.ScheduledMessage) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) *service {
	return &service{repo: repo}
}

func (s *service) FindById(ctx context.Context, id uuid.UUID) (*models.ScheduledMessage, error) {
	return s.repo.FindById(ctx, id)
}

func (s *service) Create(ctx context.Context, message models.ScheduledMessage) error {
	return s.repo.Create(ctx, message)
}
