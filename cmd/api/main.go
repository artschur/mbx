package main

import (
	"context"
	"database/sql"
	_ "github.com/lib/pq"
	"log"
	"log/slog"
	"mbx"
	"mbx/messaging"
	"mbx/models"
	"mbx/persistence/postgres"
	"mbx/sender"
	"mbx/template"
	"mbx/webhook"
	webhookPostgres "mbx/webhook/persistence/postgres"
	"mbx/worker"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	api "github.com/twilio/twilio-go/rest/api/v2010"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/mbx?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

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

	// Persistence
	templateRepo := postgres.NewTemplateRepository(db)
	webhookRepo := webhookPostgres.NewWebhookRepository(db)
	schedulerRepo := postgres.NewSchedulerRepository(db)

	// Services
	messagingService := messaging.NewMessagingService(twilioSender, twilioFetcher, schedulerRepo)
	templateService := template.NewTemplateService(twilioFetcher, twilioSender, templateRepo)
	webhookService := webhook.NewWebhookService(webhookRepo)

	// Worker
	schedulerWorker := worker.NewSchedulerWorker(schedulerRepo, func(ctx context.Context, to string, body string, templateId string, content string, language string) (*api.ApiV2010Message, error) {
		if templateId != "" {
			return twilioSender.SendTemplate(ctx, models.WhatsappTemplate{
				To:         to,
				TemplateId: templateId,
				Content:    content,
				Language:   language,
			})
		}
		return twilioSender.Send(ctx, models.WhatsappBody{
			To:   to,
			Body: body,
		})
	}, 1*time.Minute)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go schedulerWorker.Start(ctx)

	// Handlers
	messagingHandler := messaging.NewMessagingHandler(messagingService)
	templateHandler := template.NewTemplateHandler(templateService)
	webhookHandler := webhook.NewWebhookHandler(webhookService)

	router := mbx.SetupRouter(messagingHandler, templateHandler, webhookHandler)

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
	<-ctx.Done()
	log.Println("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
