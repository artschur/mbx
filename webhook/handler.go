package webhook

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
)

type WebhookHandler struct {
	service WebhookService
}

func NewWebhookHandler(svc WebhookService) *WebhookHandler {
	return &WebhookHandler{service: svc}
}

// PostTwilioStatus is the endpoint for Twilio StatusCallback
func (h *WebhookHandler) PostTwilioStatus(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		slog.Error("Failed to parse form in webhook", "error", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	sid := r.FormValue("MessageSid")
	status := r.FormValue("MessageStatus")

	// Capture all form data for raw data
	rawData, _ := json.Marshal(r.Form)

	slog.Info("Webhook received", "sid", sid, "status", status)

	err := h.service.HandleWebhook(r.Context(), sid, status, string(rawData))
	if err != nil {
		slog.Error("Failed to handle webhook", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *WebhookHandler) ListWebhooks(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}

	webhooks, err := h.service.GetWebhookRequests(r.Context(), limit)
	if err != nil {
		slog.Error("Failed to list webhooks", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(webhooks)
}
