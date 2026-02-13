package handler

import (
	"encoding/json"
	"log/slog"
	"mbx/models"
	"mbx/sender"
	"net/http"
)

type TemplateHandler struct {
	sender  sender.WhatsappSender
	fetcher sender.WhatsappFetcher
}

func NewTemplateHandler(whatsapp sender.WhatsappSender, fetcher sender.WhatsappFetcher) *TemplateHandler {
	return &TemplateHandler{
		sender:  whatsapp,
		fetcher: fetcher,
	}
}

func (h *TemplateHandler) TemplateMessage(w http.ResponseWriter, r *http.Request) {
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
	scheduledTime, err := parseScheduledTime(req.TimeFromNow)
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

func (h *TemplateHandler) GetTemplates(w http.ResponseWriter, r *http.Request) {
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

func (h *TemplateHandler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
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
