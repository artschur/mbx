package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"mbx/models"
	"mbx/schedules"
	"mbx/schedules/mocks"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
)

// Test: Create scheduled message successfully
func TestCreateScheduledMessage_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)

	// Expect Create to be called with any message
	mockRepo.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(nil).
		Times(1)

	service := schedules.NewService(mockRepo)
	handler := NewScheduledMessageHandler(service)

	futureTime := time.Now().Add(1 * time.Hour)
	req := CreateScheduledMessageRequest{
		To:      "1234567890",
		Content: "Test message",
		SendAt:  futureTime,
		Type:    models.ScheduleTypeFreeform,
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/scheduled-messages", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.CreateScheduledMessage(w, httpReq)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var response models.ScheduledMessage
	json.NewDecoder(w.Body).Decode(&response)
	if response.To != req.To {
		t.Errorf("Expected To %s, got %s", req.To, response.To)
	}
	if response.Status != models.StatusPending {
		t.Errorf("Expected status %s, got %s", models.StatusPending, response.Status)
	}
}

// Test: Create scheduled message with template
func TestCreateScheduledMessage_Template(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockRepo.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(nil).
		Times(1)

	service := schedules.NewService(mockRepo)
	handler := NewScheduledMessageHandler(service)

	futureTime := time.Now().Add(2 * time.Hour)
	req := CreateScheduledMessageRequest{
		To:                 "1234567890",
		Content:            "Welcome {{1}}",
		SendAt:             futureTime,
		Type:               models.ScheduleTypeTemplate,
		ProviderTemplateId: "template-123",
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/scheduled-messages", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.CreateScheduledMessage(w, httpReq)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var response models.ScheduledMessage
	json.NewDecoder(w.Body).Decode(&response)
	if response.Type != models.ScheduleTypeTemplate {
		t.Errorf("Expected type %s, got %s", models.ScheduleTypeTemplate, response.Type)
	}
	if response.ProviderTemplateId != req.ProviderTemplateId {
		t.Errorf("Expected ProviderTemplateId %s, got %s", req.ProviderTemplateId, response.ProviderTemplateId)
	}
}

// Test: Create scheduled message with missing recipient
func TestCreateScheduledMessage_MissingTo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	// Create should NOT be called for invalid requests
	mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Times(0)

	service := schedules.NewService(mockRepo)
	handler := NewScheduledMessageHandler(service)

	req := CreateScheduledMessageRequest{
		To:      "",
		Content: "Test message",
		SendAt:  time.Now().Add(1 * time.Hour),
		Type:    models.ScheduleTypeFreeform,
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/scheduled-messages", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.CreateScheduledMessage(w, httpReq)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// Test: Create scheduled message with missing content
func TestCreateScheduledMessage_MissingContent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Times(0)

	service := schedules.NewService(mockRepo)
	handler := NewScheduledMessageHandler(service)

	req := CreateScheduledMessageRequest{
		To:      "1234567890",
		Content: "",
		SendAt:  time.Now().Add(1 * time.Hour),
		Type:    models.ScheduleTypeFreeform,
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/scheduled-messages", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.CreateScheduledMessage(w, httpReq)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// Test: Create scheduled message with past send time
func TestCreateScheduledMessage_PastSendTime(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Times(0)

	service := schedules.NewService(mockRepo)
	handler := NewScheduledMessageHandler(service)

	pastTime := time.Now().Add(-1 * time.Hour)
	req := CreateScheduledMessageRequest{
		To:      "1234567890",
		Content: "Test message",
		SendAt:  pastTime,
		Type:    models.ScheduleTypeFreeform,
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/scheduled-messages", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.CreateScheduledMessage(w, httpReq)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// Test: Create template message without provider template ID
func TestCreateScheduledMessage_TemplateMissingId(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Times(0)

	service := schedules.NewService(mockRepo)
	handler := NewScheduledMessageHandler(service)

	req := CreateScheduledMessageRequest{
		To:      "1234567890",
		Content: "Test message",
		SendAt:  time.Now().Add(1 * time.Hour),
		Type:    models.ScheduleTypeTemplate,
		// ProviderTemplateId is missing
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/scheduled-messages", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.CreateScheduledMessage(w, httpReq)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// Test: Create scheduled message with invalid message type
func TestCreateScheduledMessage_InvalidType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Times(0)

	service := schedules.NewService(mockRepo)
	handler := NewScheduledMessageHandler(service)

	body := []byte(`{
		"to": "1234567890",
		"content": "Test message",
		"send_at": "2025-01-20T15:30:00Z",
		"type": "invalid"
	}`)

	httpReq := httptest.NewRequest("POST", "/scheduled-messages", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.CreateScheduledMessage(w, httpReq)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// Test: Create scheduled message with invalid JSON
func TestCreateScheduledMessage_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Times(0)

	service := schedules.NewService(mockRepo)
	handler := NewScheduledMessageHandler(service)

	body := []byte(`{invalid json}`)
	httpReq := httptest.NewRequest("POST", "/scheduled-messages", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.CreateScheduledMessage(w, httpReq)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// Test: Create scheduled message with repository error
func TestCreateScheduledMessage_RepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockRepo.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(fmt.Errorf("database error")).
		Times(1)

	service := schedules.NewService(mockRepo)
	handler := NewScheduledMessageHandler(service)

	futureTime := time.Now().Add(1 * time.Hour)
	req := CreateScheduledMessageRequest{
		To:      "1234567890",
		Content: "Test message",
		SendAt:  futureTime,
		Type:    models.ScheduleTypeFreeform,
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/scheduled-messages", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.CreateScheduledMessage(w, httpReq)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// Test: Get scheduled message success
func TestGetScheduledMessage_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)

	msgId := uuid.New()
	expectedMessage := &models.ScheduledMessage{
		Id:                 msgId,
		To:                 "1234567890",
		Content:            "Test message",
		SendAt:             time.Now().Add(1 * time.Hour),
		ProviderTemplateId: "",
		Type:               models.ScheduleTypeFreeform,
		Status:             models.StatusPending,
		CreatedAt:          time.Now(),
	}

	mockRepo.EXPECT().
		FindById(gomock.Any(), msgId).
		Return(expectedMessage, nil).
		Times(1)

	service := schedules.NewService(mockRepo)
	handler := NewScheduledMessageHandler(service)

	httpReq := httptest.NewRequest("GET", fmt.Sprintf("/scheduled-messages/%s", msgId), nil)
	httpReq.SetPathValue("id", msgId.String())
	w := httptest.NewRecorder()

	handler.GetScheduledMessage(w, httpReq)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response models.ScheduledMessage
	json.NewDecoder(w.Body).Decode(&response)
	if response.Id != msgId {
		t.Errorf("Expected ID %s, got %s", msgId, response.Id)
	}
	if response.To != expectedMessage.To {
		t.Errorf("Expected To %s, got %s", expectedMessage.To, response.To)
	}
}

// Test: Get non-existent scheduled message
func TestGetScheduledMessage_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	fakeId := uuid.New()

	mockRepo.EXPECT().
		FindById(gomock.Any(), fakeId).
		Return(nil, nil).
		Times(1)

	service := schedules.NewService(mockRepo)
	handler := NewScheduledMessageHandler(service)

	httpReq := httptest.NewRequest("GET", fmt.Sprintf("/scheduled-messages/%s", fakeId), nil)
	httpReq.SetPathValue("id", fakeId.String())
	w := httptest.NewRecorder()

	handler.GetScheduledMessage(w, httpReq)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

// Test: Get scheduled message with invalid ID
func TestGetScheduledMessage_InvalidId(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	// Should not call FindById for invalid input
	mockRepo.EXPECT().FindById(gomock.Any(), gomock.Any()).Times(0)

	service := schedules.NewService(mockRepo)
	handler := NewScheduledMessageHandler(service)

	httpReq := httptest.NewRequest("GET", "/scheduled-messages/invalid-id", nil)
	httpReq.SetPathValue("id", "invalid-id")
	w := httptest.NewRecorder()

	handler.GetScheduledMessage(w, httpReq)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// Test: Get scheduled message with empty ID
func TestGetScheduledMessage_EmptyId(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockRepo.EXPECT().FindById(gomock.Any(), gomock.Any()).Times(0)

	service := schedules.NewService(mockRepo)
	handler := NewScheduledMessageHandler(service)

	httpReq := httptest.NewRequest("GET", "/scheduled-messages/", nil)
	httpReq.SetPathValue("id", "")
	w := httptest.NewRecorder()

	handler.GetScheduledMessage(w, httpReq)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// Test: Get scheduled message with repository error
func TestGetScheduledMessage_RepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	msgId := uuid.New()

	mockRepo.EXPECT().
		FindById(gomock.Any(), msgId).
		Return(nil, fmt.Errorf("database error")).
		Times(1)

	service := schedules.NewService(mockRepo)
	handler := NewScheduledMessageHandler(service)

	httpReq := httptest.NewRequest("GET", fmt.Sprintf("/scheduled-messages/%s", msgId), nil)
	httpReq.SetPathValue("id", msgId.String())
	w := httptest.NewRecorder()

	handler.GetScheduledMessage(w, httpReq)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// Test: Get scheduled message with template type
func TestGetScheduledMessage_TemplateMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)

	msgId := uuid.New()
	expectedMessage := &models.ScheduledMessage{
		Id:                 msgId,
		To:                 "1234567890",
		Content:            "Welcome {{1}}",
		SendAt:             time.Now().Add(2 * time.Hour),
		ProviderTemplateId: "template-123",
		Type:               models.ScheduleTypeTemplate,
		Status:             models.StatusPending,
		CreatedAt:          time.Now(),
	}

	mockRepo.EXPECT().
		FindById(gomock.Any(), msgId).
		Return(expectedMessage, nil).
		Times(1)

	service := schedules.NewService(mockRepo)
	handler := NewScheduledMessageHandler(service)

	httpReq := httptest.NewRequest("GET", fmt.Sprintf("/scheduled-messages/%s", msgId), nil)
	httpReq.SetPathValue("id", msgId.String())
	w := httptest.NewRecorder()

	handler.GetScheduledMessage(w, httpReq)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response models.ScheduledMessage
	json.NewDecoder(w.Body).Decode(&response)
	if response.Type != models.ScheduleTypeTemplate {
		t.Errorf("Expected type %s, got %s", models.ScheduleTypeTemplate, response.Type)
	}
	if response.ProviderTemplateId != "template-123" {
		t.Errorf("Expected ProviderTemplateId template-123, got %s", response.ProviderTemplateId)
	}
}
