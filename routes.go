package mbx

import (
	"mbx/handler"
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

func SetupRouter(messageHandler *handler.MessageHandler, templateHandler *handler.TemplateHandler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /messages", messageHandler.GetMessages)
	mux.HandleFunc("GET /messages/templates", templateHandler.GetScheduledMessages)
	mux.HandleFunc("POST /messages/cancel", messageHandler.CancelMessage)

	mux.HandleFunc("GET /templates", templateHandler.GetTemplates)
	mux.HandleFunc("GET /templates/services", templateHandler.GetTemplates)
	mux.HandleFunc("POST /templates", templateHandler.CreateTemplate)

	mux.HandleFunc("POST /send-message", messageHandler.NormalMessage)
	mux.HandleFunc("POST /send-template", templateHandler.TemplateMessage)

	// mux.HandleFunc("POST /callbacks/twilio", messageHandler.GetMessages)

	return CORSMiddleware(mux)
}
