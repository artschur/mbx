package mbx

import (
	"mbx/handler"
	"net/http"
)

func SetupRouter(messageHandler *handler.MessageHandler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /templates", messageHandler.TemplateMessage)
	mux.HandleFunc("POST /templates", messageHandler.CreateTemplate)

	mux.HandleFunc("POST /send-message", messageHandler.NormalMessage)
	mux.HandleFunc("POST /send-template", messageHandler.TemplateMessage)

	return mux
}
