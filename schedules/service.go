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

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) FindById(ctx context.Context, id uuid.UUID) (*models.ScheduledMessage, error) {
	return s.repo.FindById(ctx, id)
}

func (s *Service) Create(ctx context.Context, message models.ScheduledMessage) error {
	return s.repo.Create(ctx, message)
}

func (s *Service) ListUpcoming(ctx context.Context, duration time.Duration) ([]models.ScheduledMessage, error) {
	return s.repo.ListUpcoming(ctx, duration)
}
