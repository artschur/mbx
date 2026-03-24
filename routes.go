package mbx

import (
	"mbx/messaging"
	"mbx/template"
	"mbx/webhook"
	"net/http"
)

// CORSMiddleware adds CORS headers to all responses
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*") // Change "*" to specific domain in production
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "3600")

		// Handle preflight OPTIONS request
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

func SetupRouter(
	messagingHandler *messaging.MessagingHandler,
	templateHandler *template.TemplateHandler,
	webhookHandler *webhook.WebhookHandler,
) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /messages", messagingHandler.GetMessages)
	mux.HandleFunc("GET /messages/templates", messagingHandler.GetScheduledMessages)
	mux.HandleFunc("POST /messages/cancel", messagingHandler.CancelMessage)

	mux.HandleFunc("GET /templates", templateHandler.GetTemplates)
	mux.HandleFunc("POST /templates", templateHandler.CreateTemplate)
	mux.HandleFunc("GET /templates/services", templateHandler.ListMessagingServices)

	mux.HandleFunc("POST /send-message", messagingHandler.NormalMessage)
	mux.HandleFunc("POST /send-template", messagingHandler.TemplateMessage)

	// Webhook and Status
	mux.HandleFunc("POST /webhooks/twilio", webhookHandler.PostTwilioStatus)
	mux.HandleFunc("GET /webhooks", webhookHandler.ListWebhooks)

	return CORSMiddleware(mux)
}
