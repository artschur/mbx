package postgres

import (
	"context"
	"database/sql"
	"mbx/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDB(ctx context.Context, t *testing.T) (*sql.DB, func()) {
	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(15*time.Second)),
	)
	if err != nil {
		t.Fatalf("failed to start container: %v", err)
	}

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	// Create tables
	_, err = db.Exec(`
		CREATE TABLE templates (
			content_id TEXT PRIMARY KEY,
			friendly_name TEXT NOT NULL,
			language TEXT,
			body TEXT,
			variables JSONB,
			types JSONB,
			date_created TIMESTAMP WITH TIME ZONE,
			date_updated TIMESTAMP WITH TIME ZONE
		);
	`)
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	return db, func() {
		db.Close()
		container.Terminate(ctx)
	}
}

func TestTemplateRepository_SaveAndGet(t *testing.T) {
	ctx := context.Background()
	db, cleanup := setupTestDB(ctx, t)
	defer cleanup()

	repo := NewTemplateRepository(db)

	template := models.SavedTemplate{
		ContentId:    "HX123",
		FriendlyName: "Test Template",
		Language:     "en",
		Body:         "Hello {{1}}",
		DateCreated:  time.Now().Format(time.RFC3339),
	}

	t.Run("Save Template", func(t *testing.T) {
		err := repo.SaveTemplate(ctx, template)
		assert.NoError(t, err)
	})

	t.Run("Get Templates", func(t *testing.T) {
		templates, err := repo.GetTemplates(ctx)
		assert.NoError(t, err)
		assert.Len(t, templates, 1)
		assert.Equal(t, template.ContentId, templates[0].ContentId)
	})
}
