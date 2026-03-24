package worker

import (
	"context"
	"mbx/models"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	api "github.com/twilio/twilio-go/rest/api/v2010"
)

type mockRepo struct {
	ctrl *gomock.Controller
}

func (m *mockRepo) Save(ctx context.Context, msg *models.ScheduledMessage) error {
	return nil
}

func (m *mockRepo) GetPending(ctx context.Context, limit int) ([]models.ScheduledMessage, error) {
	now := time.Now()
	return []models.ScheduledMessage{
		{ID: 1, To: "123", Body: "Hello", Status: "pending", ScheduledFor: now},
	}, nil
}

func (m *mockRepo) MarkAsSent(ctx context.Context, id int64) error {
	return nil
}

func (m *mockRepo) MarkAsFailed(ctx context.Context, id int64, errStr string) error {
	return nil
}

func TestSchedulerWorker_ProcessPending(t *testing.T) {
	ctx := context.Background()
	repo := &mockRepo{}

	sendCalled := false
	mockSendFunc := func(ctx context.Context, to string, body string, templateId string, content string, language string) (*api.ApiV2010Message, error) {
		sendCalled = true
		return &api.ApiV2010Message{}, nil
	}

	worker := NewSchedulerWorker(repo, mockSendFunc, 1*time.Minute)
	worker.processPending(ctx)

	assert.True(t, sendCalled)
}
