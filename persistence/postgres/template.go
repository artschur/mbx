package postgres

import (
	"context"
	"database/sql"
	"mbx/models"
)

//go:generate mockgen -destination=mock_template_test.go -package=postgres mbx/persistence/postgres TemplateRepository

type TemplateRepository interface {
	SaveTemplate(ctx context.Context, template models.SavedTemplate) error
	GetTemplates(ctx context.Context) ([]models.SavedTemplate, error)
}

type templateRepo struct {
	db *sql.DB
}

func NewTemplateRepository(db *sql.DB) TemplateRepository {
	return &templateRepo{db: db}
}

func (r *templateRepo) SaveTemplate(ctx context.Context, t models.SavedTemplate) error {
	query := `
		INSERT INTO templates (content_id, friendly_name, language, body, variables, types, date_created, date_updated)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (content_id) DO UPDATE SET
			friendly_name = EXCLUDED.friendly_name,
			body = EXCLUDED.body,
			variables = EXCLUDED.variables,
			date_updated = EXCLUDED.date_updated`

	// Note: You'll need to handle JSON serialization for variables and types
	// This is a simplified version
	_, err := r.db.ExecContext(ctx, query,
		t.ContentId, t.FriendlyName, t.Language, t.Body, t.Variables, t.Types, t.DateCreated, t.DateUpdated)
	return err
}

func (r *templateRepo) GetTemplates(ctx context.Context) ([]models.SavedTemplate, error) {
	query := `SELECT content_id, friendly_name, language, body, date_created FROM templates`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []models.SavedTemplate
	for rows.Next() {
		var t models.SavedTemplate
		if err := rows.Scan(&t.ContentId, &t.FriendlyName, &t.Language, &t.Body, &t.DateCreated); err != nil {
			return nil, err
		}
		templates = append(templates, t)
	}
	return templates, nil
}
