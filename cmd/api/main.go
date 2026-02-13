package main

import (
	"context"
	"log"
	"log/slog"
	"mbx"
	"mbx/handler"
	"mbx/sender"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	accountSid := os.Getenv("TWILIO_ACCOUNT_SID")
	authToken := os.Getenv("TWILIO_AUTH_TOKEN")
	fromNumber := os.Getenv("TWILIO_FROM_NUMBER")
	if fromNumber == "" {
		slog.Error("TWILIO_FROM_NUMBER environment variable is required")
		return
	}
	if accountSid == "" || authToken == "" {
		slog.Error("TWILIO_ACCOUNT_SID and TWILIO_AUTH_TOKEN environment variables are required")
		return
	}

	cfg := &sender.Config{
		TwilioAccountSID: accountSid,
		TwilioAuthToken:  authToken,
		TwilioFromNumber: fromNumber,
	}

	twilioClient := sender.NewTwilioClient(cfg)

	twilioSender := sender.NewTwilioSender(twilioClient, cfg)
	twilioFetcher := sender.NewTwilioFetcher(twilioClient, cfg)

	messageHandler := handler.NewMessageHandler(twilioSender, twilioFetcher)
	router := mbx.SetupRouter(messageHandler)

	server := &http.Server{
		Addr:    ":8765",
		Handler: router,
	}

	go func() {
		log.Println("Starting server on :8765")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
