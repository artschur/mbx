package handler

import (
	"encoding/json"
	"log/slog"
	"mbx/models"
	"mbx/sender"
	"net/http"
	"time"
)

type MessageHandler struct {
	sender  sender.WhatsappSender
	fetcher sender.WhatsappFetcher
}

func NewMessageHandler(whatsapp sender.WhatsappSender, fetcher sender.WhatsappFetcher) *MessageHandler {
	return &MessageHandler{
		sender:  whatsapp,
		fetcher: fetcher,
	}
}

func (h *MessageHandler) parseScheduledTime(timeStr string) (*time.Time, error) {
	if timeStr == "" {
		return nil, nil
	}

	parsedTime, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return nil, err
	}

	return &parsedTime, nil
}

func (h *MessageHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
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

	messages, err := h.fetcher.GetMessages(r.Context(), afterTime)
	if err != nil {
		slog.Error("Failed to retrieve messages", "error", err)
		http.Error(w, "Failed to retrieve messages", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(messages)
	if err != nil {
		slog.Error("Failed to encode messages response", "error", err)
		http.Error(w, "Failed to encode messages response", http.StatusInternalServerError)
		return
	}
}

func (h *MessageHandler) NormalMessage(w http.ResponseWriter, r *http.Request) {
	var req models.WhatsappBodyDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Body == "" {
		http.Error(w, "Message body cannot be empty", http.StatusBadRequest)
		return
	}
	if req.To == "" {
		http.Error(w, "Recipient number cannot be empty", http.StatusBadRequest)
		return
	}

	// Parse optional scheduled time
	scheduledTime, err := h.parseScheduledTime(req.TimeFromNow)
	if err != nil {
		http.Error(w, "Invalid time format. Use RFC3339 format", http.StatusBadRequest)
		return
	}

	whatsappMessage := models.WhatsappBody{
		To:          req.To,
		Body:        req.Body,
		TimeFromNow: scheduledTime,
	}

	msgResponse, err := h.sender.Send(r.Context(), whatsappMessage)
	if err != nil {
		slog.Error("Failed to send message", "error", err)
		http.Error(w, "Failed to send message", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(msgResponse)
	if err != nil {
		slog.Error("Failed to encode message response", "error", err)
		http.Error(w, "Failed to encode message response", http.StatusInternalServerError)
		return
	}
}

func (h *MessageHandler) TemplateMessage(w http.ResponseWriter, r *http.Request) {
	var req models.WhatsappTemplateDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	slog.Info("Received template message request", "to", req.To, "template_id", req.TemplateId, "content", req.Content, "language", req.Language)

	// Validate required fields
	if req.TemplateId == "" {
		http.Error(w, "Template ID cannot be empty", http.StatusBadRequest)
		return
	}
	if req.To == "" {
		http.Error(w, "Recipient number cannot be empty", http.StatusBadRequest)
		return
	}

	// Parse optional scheduled time
	scheduledTime, err := h.parseScheduledTime(req.TimeFromNow)
	if err != nil {
		http.Error(w, "Invalid time format. Use RFC3339 format", http.StatusBadRequest)
		return
	}

	var contentStr string
	if len(req.Content) > 0 {
		contentJSON, err := json.Marshal(req.Content)
		if err != nil {
			slog.Error("Failed to marshal content variables", "error", err)
			http.Error(w, "Invalid content variables", http.StatusBadRequest)
			return
		}
		contentStr = string(contentJSON)
	}

	slog.Info("Marshaled content variables", "content_json", contentStr)

	whatsappTemplate := models.WhatsappTemplate{
		To:          req.To,
		TemplateId:  req.TemplateId,
		Content:     contentStr,
		Language:    req.Language,
		TimeFromNow: scheduledTime,
	}

	msgResponse, err := h.sender.SendTemplate(r.Context(), whatsappTemplate)
	if err != nil {
		slog.Error("Failed to send template message", "error", err)
		http.Error(w, "Failed to send template message", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(msgResponse)
	if err != nil {
		slog.Error("Failed to encode template message response", "error", err)
		http.Error(w, "Failed to encode template message response", http.StatusInternalServerError)
		return
	}
}

func (h *MessageHandler) GetTemplates(w http.ResponseWriter, r *http.Request) {
	templates, err := h.fetcher.GetTemplates(r.Context())
	if err != nil {
		slog.Error("Failed to retrieve templates", "error", err)
		http.Error(w, "Failed to retrieve templates", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(templates)
	if err != nil {
		slog.Error("Failed to encode templates response", "error", err)
		http.Error(w, "Failed to encode templates response", http.StatusInternalServerError)
		return
	}
}

func (h *MessageHandler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTemplateDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Failed to decode create template request", "error", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.FriendlyName == "" {
		http.Error(w, "Friendly name cannot be empty", http.StatusBadRequest)
		return
	}
	if req.Language == "" {
		http.Error(w, "Language cannot be empty", http.StatusBadRequest)
		return
	}
	if req.Body == "" {
		http.Error(w, "Body cannot be empty", http.StatusBadRequest)
		return
	}

	// Create the template
	createdTemplate, err := h.sender.CreateTemplate(r.Context(), req)
	if err != nil {
		slog.Error("Failed to create template", "error", err)
		http.Error(w, "Failed to create template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(createdTemplate); err != nil {
		slog.Error("Failed to encode template response", "error", err)
		http.Error(w, "Failed to encode template response", http.StatusInternalServerError)
		return
	}
}

func (h *MessageHandler) CancelMessage(w http.ResponseWriter, r *http.Request) {
	incReq := struct {
		TwilioMessageId string `json:"message_id"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&incReq); err != nil {
		slog.Error("Failed to decode cancel message request", "error", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if incReq.TwilioMessageId == "" {
		http.Error(w, "Message ID cannot be empty", http.StatusBadRequest)
		return
	}

	err := h.sender.CancelMessage(r.Context(), incReq.TwilioMessageId)
	if err != nil {
		slog.Error("Failed to cancel message", "error", err, "message_id", incReq.TwilioMessageId)
		http.Error(w, "Failed to cancel message: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
