package postgres

import (
	"context"
	"errors"
	"mbx/models"
	"mbx/schedules"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MessageRepository struct {
	db *pgxpool.Pool
}

func NewMessageRepository(db *pgxpool.Pool) *MessageRepository {
	return &MessageRepository{db: db}
}

var _ schedules.Repository = &MessageRepository{}

func (r *MessageRepository) Create(ctx context.Context, message models.ScheduledMessage) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO scheduled_messages
		(id, to_number, send_at, content, provider_template_id, message_type, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`,
		message.Id,
		message.To,
		message.SendAt,
		message.Content,
		message.ProviderId,
		message.Type,
		message.Status,
		message.CreatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *MessageRepository) FindById(ctx context.Context, id uuid.UUID) (*models.ScheduledMessage, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, to_number, send_at, content, provider_template_id, message_type, status, created_at
		FROM scheduled_messages
		WHERE id = $1
		`, id)
	var message models.ScheduledMessage
	err := row.Scan(&message.Id, &message.To, &message.SendAt, &message.Content, &message.ProviderId, &message.Type, &message.Status, &message.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &message, nil
}

func (r *MessageRepository) ListUpcoming(ctx context.Context, duration time.Duration) ([]models.ScheduledMessage, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, to_number, send_at, content, provider_template_id, message_type, status, created_at
		FROM scheduled_messages
		WHERE send_at >= NOW() AND send_at < NOW() + $1
		`, duration)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.ScheduledMessage
	for rows.Next() {
		var message models.ScheduledMessage
		if err := rows.Scan(&message.Id, &message.To, &message.SendAt, &message.Content, &message.ProviderId, &message.Type, &message.Status, &message.CreatedAt); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, nil
			}
			return nil, err
		}
		messages = append(messages, message)
	}
	return messages, nil
}
