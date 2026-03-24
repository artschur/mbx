package postgres

import (
	"context"
	"database/sql"
	"mbx/models"
	"time"
)

//go:generate mockgen -destination=mock_scheduler_test.go -package=postgres mbx/persistence/postgres SchedulerRepository

type SchedulerRepository interface {
	Save(ctx context.Context, msg *models.ScheduledMessage) error
	GetPending(ctx context.Context, limit int) ([]models.ScheduledMessage, error)
	MarkAsSent(ctx context.Context, id int64) error
	MarkAsFailed(ctx context.Context, id int64, errStr string) error
}

type schedulerRepo struct {
	db *sql.DB
}

func NewSchedulerRepository(db *sql.DB) SchedulerRepository {
	return &schedulerRepo{db: db}
}

func (r *schedulerRepo) Save(ctx context.Context, m *models.ScheduledMessage) error {
	query := `
		INSERT INTO scheduled_messages (to_phone, body, template_id, content, language, scheduled_for, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`
	return r.db.QueryRowContext(ctx, query, m.To, m.Body, m.TemplateId, m.Content, m.Language, m.ScheduledFor, "pending", time.Now()).Scan(&m.ID)
}

func (r *schedulerRepo) GetPending(ctx context.Context, limit int) ([]models.ScheduledMessage, error) {
	query := `
		SELECT id, to_phone, body, template_id, content, language, scheduled_for, created_at
		FROM scheduled_messages
		WHERE status = 'pending' AND scheduled_for <= $1
		ORDER BY scheduled_for ASC
		LIMIT $2`
	rows, err := r.db.QueryContext(ctx, query, time.Now(), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []models.ScheduledMessage
	for rows.Next() {
		var m models.ScheduledMessage
		err := rows.Scan(&m.ID, &m.To, &m.Body, &m.TemplateId, &m.Content, &m.Language, &m.ScheduledFor, &m.CreatedAt)
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}
	return msgs, nil
}

func (r *schedulerRepo) MarkAsSent(ctx context.Context, id int64) error {
	query := `UPDATE scheduled_messages SET status = 'sent', sent_at = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	return err
}

func (r *schedulerRepo) MarkAsFailed(ctx context.Context, id int64, errStr string) error {
	query := `UPDATE scheduled_messages SET status = 'failed', error_message = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, errStr, id)
	return err
}
