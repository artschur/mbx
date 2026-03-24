package postgres

import (
	"context"
	"database/sql"
	"mbx/models"
)

type WebhookRepository interface {
	SaveWebhookRequest(ctx context.Context, req *models.WebhookRequest) error
	ListWebhooks(ctx context.Context, limit int) ([]models.WebhookRequest, error)
}

type webhookRepo struct {
	db *sql.DB
}

func NewWebhookRepository(db *sql.DB) WebhookRepository {
	return &webhookRepo{db: db}
}

func (r *webhookRepo) SaveWebhookRequest(ctx context.Context, req *models.WebhookRequest) error {
	query := `
		INSERT INTO webhook_requests (message_sid, status, raw_data, received_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id`
	return r.db.QueryRowContext(ctx, query, req.MessageSid, req.Status, req.RawData, req.ReceivedAt).Scan(&req.ID)
}

func (r *webhookRepo) ListWebhooks(ctx context.Context, limit int) ([]models.WebhookRequest, error) {
	query := `SELECT id, message_sid, status, raw_data, received_at FROM webhook_requests ORDER BY received_at DESC LIMIT $1`
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []models.WebhookRequest
	for rows.Next() {
		var req models.WebhookRequest
		if err := rows.Scan(&req.ID, &req.MessageSid, &req.Status, &req.RawData, &req.ReceivedAt); err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}
	return requests, nil
}
