package messaging

import (
	"encoding/json"
	"log/slog"
	"mbx/models"
	"net/http"
	"slices"
	"time"
)

type MessagingHandler struct {
	service MessagingService
}

func NewMessagingHandler(svc MessagingService) *MessagingHandler {
	return &MessagingHandler{
		service: svc,
	}
}

func parseScheduledTime(timeStr string) (*time.Time, error) {
	if timeStr == "" {
		return nil, nil
	}
	parsedTime, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return nil, err
	}
	return &parsedTime, nil
}

func (h *MessagingHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	afterStr := r.URL.Query().Get("after")
	if afterStr == "" {
		http.Error(w, "Missing 'after' query parameter", http.StatusBadRequest)
		return
	}
	afterTime, err := time.Parse("2006-01-02", afterStr)
	if err != nil {
		http.Error(w, "Invalid 'after' time format. Use YYYY-MM-DD format", http.StatusBadRequest)
		return
	}
	status := r.URL.Query().Get("status")
	validStatuses := []string{"", "sent", "read", "delivered", "failed", "scheduled", "queued", "sending"}
	if status != "" && !slices.Contains(validStatuses, status) {
		http.Error(w, "Invalid 'status' value", http.StatusBadRequest)
		return
	}
	messages, err := h.service.GetMessages(r.Context(), afterTime)
	if err != nil {
		slog.Error("Failed to retrieve messages", "error", err)
		http.Error(w, "Failed to retrieve messages", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

func (h *MessagingHandler) NormalMessage(w http.ResponseWriter, r *http.Request) {
	var req models.WhatsappBodyDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	if req.Body == "" || req.To == "" {
		http.Error(w, "Body and To are required", http.StatusBadRequest)
		return
	}
	scheduledTime, err := parseScheduledTime(req.TimeFromNow)
	if err != nil {
		http.Error(w, "Invalid time format", http.StatusBadRequest)
		return
	}
	whatsappMessage := models.WhatsappBody{
		To:          req.To,
		Body:        req.Body,
		TimeFromNow: scheduledTime,
	}
	msgResponse, err := h.service.Send(r.Context(), whatsappMessage)
	if err != nil {
		slog.Error("Failed to send message", "error", err)
		http.Error(w, "Failed to send message", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(msgResponse)
}

func (h *MessagingHandler) CancelMessage(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MessageId string `json:"message_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	if req.MessageId == "" {
		http.Error(w, "Message ID is required", http.StatusBadRequest)
		return
	}
	err := h.service.CancelMessage(r.Context(), req.MessageId)
	if err != nil {
		slog.Error("Failed to cancel message", "error", err)
		http.Error(w, "Failed to cancel message", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *MessagingHandler) TemplateMessage(w http.ResponseWriter, r *http.Request) {
	var req models.WhatsappTemplateDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.TemplateId == "" || req.To == "" {
		http.Error(w, "Template ID and To are required", http.StatusBadRequest)
		return
	}

	scheduledTime, err := parseScheduledTime(req.TimeFromNow)
	if err != nil {
		http.Error(w, "Invalid time format", http.StatusBadRequest)
		return
	}

	var contentStr string
	if len(req.Content) > 0 {
		contentJSON, _ := json.Marshal(req.Content)
		contentStr = string(contentJSON)
	}

	whatsappTemplate := models.WhatsappTemplate{
		To:          req.To,
		TemplateId:  req.TemplateId,
		Content:     contentStr,
		Language:    req.Language,
		TimeFromNow: scheduledTime,
	}

	msgResponse, err := h.service.SendTemplate(r.Context(), whatsappTemplate)
	if err != nil {
		slog.Error("Failed to send template message", "error", err)
		http.Error(w, "Failed to send template message", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(msgResponse)
}

func (h *MessagingHandler) GetScheduledMessages(w http.ResponseWriter, r *http.Request) {
	messages, err := h.service.GetScheduledMessages(r.Context(), time.Now())
	if err != nil {
		slog.Error("Failed to retrieve scheduled messages", "error", err)
		http.Error(w, "Failed to retrieve scheduled messages", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}
