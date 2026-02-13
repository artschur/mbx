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
	scheduledTime, err := parseScheduledTime(req.TimeFromNow)
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
