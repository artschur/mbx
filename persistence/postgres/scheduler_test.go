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

func setupSchedulerTestDB(ctx context.Context, t *testing.T) (*sql.DB, func()) {
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
		CREATE TABLE scheduled_messages (
			id SERIAL PRIMARY KEY,
			to_phone TEXT NOT NULL,
			body TEXT,
			template_id TEXT,
			content JSONB,
			language TEXT,
			scheduled_for TIMESTAMP WITH TIME ZONE NOT NULL,
			sent_at TIMESTAMP WITH TIME ZONE,
			status TEXT DEFAULT 'pending',
			error_message TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
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

func TestSchedulerRepository_GetPending(t *testing.T) {
	ctx := context.Background()
	db, cleanup := setupSchedulerTestDB(ctx, t)
	defer cleanup()

	repo := NewSchedulerRepository(db)

	now := time.Now()

	// 1. Message scheduled for now (should be picked up)
	msg1 := &models.ScheduledMessage{
		To:           "123",
		Body:         "Due now",
		ScheduledFor: now.Add(-1 * time.Minute),
		Status:       "pending",
	}
	// 2. Message scheduled for future (should NOT be picked up)
	msg2 := &models.ScheduledMessage{
		To:           "456",
		Body:         "Future",
		ScheduledFor: now.Add(1 * time.Hour),
		Status:       "pending",
	}
	// 3. Message already sent (should NOT be picked up)
	msg3 := &models.ScheduledMessage{
		To:           "789",
		Body:         "Sent",
		ScheduledFor: now.Add(-1 * time.Hour),
		Status:       "sent",
	}

	assert.NoError(t, repo.Save(ctx, msg1))
	assert.NoError(t, repo.Save(ctx, msg2))
	// Manually insert sent message as Save defaults to pending
	_, _ = db.Exec("INSERT INTO scheduled_messages (to_phone, body, scheduled_for, status) VALUES ($1, $2, $3, $4)",
		msg3.To, msg3.Body, msg3.ScheduledFor, "sent")

	t.Run("Pick up only pending and due messages", func(t *testing.T) {
		msgs, err := repo.GetPending(ctx, 10)
		assert.NoError(t, err)
		assert.Len(t, msgs, 1)
		assert.Equal(t, "123", msgs[0].To)
	})
}
