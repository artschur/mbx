package handler

import (
	"encoding/json"
	"log/slog"
	"mbx/models"
	"mbx/schedules"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type ScheduledMessageHandler struct {
	scheduleService *schedules.Service
}

func NewScheduledMessageHandler(scheduleService *schedules.Service) *ScheduledMessageHandler {
	return &ScheduledMessageHandler{
		scheduleService: scheduleService,
	}
}

// CreateScheduledMessageRequest represents the request payload for scheduling a message
type CreateScheduledMessageRequest struct {
	To                 string                      `json:"to"`
	Content            string                      `json:"content"`
	SendAt             time.Time                   `json:"send_at"`
	ProviderTemplateId string                      `json:"provider_template_id,omitempty"`
	Type               models.ScheduledMessageType `json:"type"` // "template" or "freeform"
}

// CreateScheduledMessage handles POST /scheduled-messages
func (h *ScheduledMessageHandler) CreateScheduledMessage(w http.ResponseWriter, r *http.Request) {
	var req CreateScheduledMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Failed to decode create scheduled message request", "error", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.To == "" {
		http.Error(w, "Recipient number cannot be empty", http.StatusBadRequest)
		return
	}
	if req.Content == "" {
		http.Error(w, "Message content cannot be empty", http.StatusBadRequest)
		return
	}
	if req.Type != models.ScheduleTypeTemplate && req.Type != models.ScheduleTypeFreeform {
		http.Error(w, "Invalid message type. Must be 'template' or 'freeform'", http.StatusBadRequest)
		return
	}
	if req.Type == models.ScheduleTypeTemplate && req.ProviderTemplateId == "" {
		http.Error(w, "Provider template ID required for template messages", http.StatusBadRequest)
		return
	}
	if req.SendAt.Before(time.Now()) {
		http.Error(w, "Send time cannot be in the past", http.StatusBadRequest)
		return
	}

	message := models.ScheduledMessage{
		Id:                 uuid.New(),
		To:                 req.To,
		Content:            req.Content,
		SendAt:             req.SendAt,
		ProviderTemplateId: req.ProviderTemplateId,
		Type:               req.Type,
		Status:             models.StatusPending,
		CreatedAt:          time.Now(),
	}

	err := h.scheduleService.Create(r.Context(), message)
	if err != nil {
		slog.Error("Failed to create scheduled message", "error", err)
		http.Error(w, "Failed to create scheduled message", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(message)
}

// GetScheduledMessage handles GET /scheduled-messages/:id
func (h *ScheduledMessageHandler) GetScheduledMessage(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		http.Error(w, "Message ID is required", http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid message ID format", http.StatusBadRequest)
		return
	}

	message, err := h.scheduleService.FindById(r.Context(), id)
	if err != nil {
		slog.Error("Failed to fetch scheduled message", "error", err, "id", id)
		http.Error(w, "Failed to fetch scheduled message", http.StatusInternalServerError)
		return
	}
	if message == nil {
		http.Error(w, "Scheduled message not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(message)
}
