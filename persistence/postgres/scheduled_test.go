package postgres

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"mbx/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	testDB        *pgxpool.Pool
	testContainer testcontainers.Container
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Start container using basic testcontainers (not the postgres module)
	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start postgres container: %v\n", err)
		os.Exit(1)
	}
	testContainer = container

	// Get host and port
	host, err := container.Host(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get host: %v\n", err)
		os.Exit(1)
	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get port: %v\n", err)
		os.Exit(1)
	}

	// Build connection string with explicit sslmode=disable
	connStr := fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb?sslmode=disable", host, port.Port())

	// Create pool with retry logic
	var pool *pgxpool.Pool
	maxAttempts := 10
	for i := 0; i < maxAttempts; i++ {
		pool, err = pgxpool.New(ctx, connStr)
		if err == nil {
			if err = pool.Ping(ctx); err == nil {
				break
			}
			pool.Close()
		}
		if i < maxAttempts-1 {
			time.Sleep(500 * time.Millisecond)
		}
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create pool: %v\n", err)
		os.Exit(1)
	}
	testDB = pool

	// Run migrations
	if err := runMigrations(ctx, pool); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to run migrations: %v\n", err)
		os.Exit(1)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	pool.Close()
	if err := container.Terminate(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to terminate container: %v\n", err)
	}

	os.Exit(code)
}

func runMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	migrationSQL := `
		DROP TYPE IF EXISTS message_status CASCADE;
		CREATE TYPE message_status AS ENUM('pending', 'sent', 'failed');

		DROP TABLE IF EXISTS scheduled_messages;
		CREATE TABLE scheduled_messages (
			id UUID PRIMARY KEY,
			to_number VARCHAR(255) NOT NULL,
			send_at TIMESTAMP NOT NULL,
			content TEXT NOT NULL,
			provider_template_id VARCHAR(255) NOT NULL,
			message_type VARCHAR(255) NOT NULL,
			status message_status NOT NULL DEFAULT 'pending',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`

	_, err := pool.Exec(ctx, migrationSQL)
	return err
}

func TestScheduledMessages_Create(t *testing.T) {
	ctx := context.Background()
	messageRepo := NewMessageRepository(testDB)

	id := uuid.New()
	scheduledMessage := models.ScheduledMessage{
		Id:                 id,
		To:                 "1234567890",
		SendAt:             time.Now().Add(10 * time.Minute),
		Content:            "Test message",
		ProviderTemplateId: "template-123",
		Type:               models.ScheduleTypeFreeform,
		Status:             models.StatusPending,
		CreatedAt:          time.Now(),
	}

	err := messageRepo.Create(ctx, scheduledMessage)
	require.NoError(t, err)

	gotten, err := messageRepo.FindById(ctx, id)
	require.NoError(t, err)
	require.NotNil(t, gotten)
	require.Equal(t, scheduledMessage.Id, gotten.Id)
	require.Equal(t, scheduledMessage.To, gotten.To)
	require.Equal(t, scheduledMessage.Content, gotten.Content)
	require.Equal(t, scheduledMessage.Type, gotten.Type)
	require.Equal(t, models.StatusPending, gotten.Status)
}

func TestScheduledMessages_FindById(t *testing.T) {
	ctx := context.Background()
	messageRepo := NewMessageRepository(testDB)

	id := uuid.New()
	scheduledMessage := models.ScheduledMessage{
		Id:                 id,
		To:                 "9876543210",
		SendAt:             time.Now().Add(5 * time.Minute),
		Content:            "Another message",
		ProviderTemplateId: "",
		Type:               models.ScheduleTypeFreeform,
		Status:             models.StatusPending,
		CreatedAt:          time.Now(),
	}

	err := messageRepo.Create(ctx, scheduledMessage)
	require.NoError(t, err)

	gotten, err := messageRepo.FindById(ctx, id)
	require.NoError(t, err)
	require.NotNil(t, gotten)
	require.Equal(t, scheduledMessage.Id, gotten.Id)
	require.Equal(t, scheduledMessage.To, gotten.To)
}

func TestScheduledMessages_FindById_NotFound(t *testing.T) {
	ctx := context.Background()
	messageRepo := NewMessageRepository(testDB)

	fakeId := uuid.New()
	gotten, err := messageRepo.FindById(ctx, fakeId)
	require.NoError(t, err)
	require.Nil(t, gotten)
}

func TestScheduledMessages_Template(t *testing.T) {
	ctx := context.Background()
	messageRepo := NewMessageRepository(testDB)

	id := uuid.New()
	scheduledMessage := models.ScheduledMessage{
		Id:                 id,
		To:                 "5555555555",
		SendAt:             time.Now().Add(20 * time.Minute),
		Content:            "Welcome {{name}}",
		ProviderTemplateId: "template-456",
		Type:               models.ScheduleTypeTemplate,
		Status:             models.StatusPending,
		CreatedAt:          time.Now(),
	}

	err := messageRepo.Create(ctx, scheduledMessage)
	require.NoError(t, err)

	gotten, err := messageRepo.FindById(ctx, id)
	require.NoError(t, err)
	require.NotNil(t, gotten)
	require.Equal(t, models.ScheduleTypeTemplate, gotten.Type)
	require.Equal(t, "template-456", gotten.ProviderTemplateId)
}

func TestScheduledMessages_ListUpcoming(t *testing.T) {
	ctx := context.Background()
	messageRepo := NewMessageRepository(testDB)

	id := uuid.New()
	futureTime := time.Now().UTC().Add(10 * time.Minute)
	scheduledMessage := models.ScheduledMessage{
		Id:                 id,
		To:                 "7777777777",
		SendAt:             futureTime,
		Content:            "Future message",
		ProviderTemplateId: "",
		Type:               models.ScheduleTypeFreeform,
		Status:             models.StatusPending,
		CreatedAt:          time.Now(),
	}

	err := messageRepo.Create(ctx, scheduledMessage)
	require.NoError(t, err)

	// Test ListUpcoming with duration that doesn't include the message
	falseUpcoming, err := messageRepo.ListUpcoming(ctx, 5*time.Minute)
	require.NoError(t, err)
	require.Empty(t, falseUpcoming)

	// Test ListUpcoming with duration that includes the message
	upcoming, err := messageRepo.ListUpcoming(ctx, 15*time.Minute)
	require.NoError(t, err)
	require.Len(t, upcoming, 1)

	msg := upcoming[0]
	require.Equal(t, scheduledMessage.Id, msg.Id)
	require.Equal(t, scheduledMessage.To, msg.To)
	require.Equal(t, scheduledMessage.Content, msg.Content)

}
