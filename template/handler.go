package template

import (
	"encoding/json"
	"log/slog"
	"mbx/models"
	"net/http"
	"time"
)

type TemplateHandler struct {
	service TemplateService
}

func NewTemplateHandler(svc TemplateService) *TemplateHandler {
	return &TemplateHandler{
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

func (h *TemplateHandler) ListMessagingServices(w http.ResponseWriter, r *http.Request) {
	services, err := h.service.ListMessagingServices(r.Context())
	if err != nil {
		slog.Error("Failed to retrieve messaging services", "error", err)
		http.Error(w, "Failed to retrieve messaging services", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(services)
}

func (h *TemplateHandler) GetTemplates(w http.ResponseWriter, r *http.Request) {
	templates, err := h.service.GetTemplates(r.Context())
	if err != nil {
		slog.Error("Failed to retrieve templates", "error", err)
		http.Error(w, "Failed to retrieve templates", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(templates)
}

func (h *TemplateHandler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTemplateDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Failed to decode create template request", "error", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	if req.FriendlyName == "" || req.Language == "" || req.Body == "" {
		http.Error(w, "Friendly name, language, and body are required", http.StatusBadRequest)
		return
	}
	createdTemplate, err := h.service.CreateTemplate(r.Context(), req)
	if err != nil {
		slog.Error("Failed to create template", "error", err)
		http.Error(w, "Failed to create template", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdTemplate)
}

// Note: TemplateMessage is moved to messaging domain in service but if the user wants it here,
// we would need messaging service dependency. However, the user asked for "messaging services
// (send normal messages, and templates.)", so TemplateMessage should be in messaging handler.
