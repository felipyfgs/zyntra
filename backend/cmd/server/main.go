package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/zyntra/backend/internal/auth"
	"github.com/zyntra/backend/internal/channels/whatsapp"
	"github.com/zyntra/backend/internal/database"
	"github.com/zyntra/backend/internal/handlers"
	"github.com/zyntra/backend/internal/middleware"
	"github.com/zyntra/backend/internal/repository"
	"github.com/zyntra/backend/internal/router"
	"github.com/zyntra/backend/internal/services"
	natspkg "github.com/zyntra/backend/pkg/nats"
	wapkg "github.com/zyntra/backend/pkg/whatsapp"
	wspkg "github.com/zyntra/backend/pkg/websocket"
)

func main() {
	log.Println("=== Zyntra API Server v2.0 ===")

	// Database
	db, err := database.New(database.DefaultConfig())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Migrations
	migrator := database.NewMigrator(db.DB)
	if err := migrator.Run(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// NATS
	natsClient, err := natspkg.NewClient(nil)
	if err != nil {
		log.Printf("Warning: NATS not available: %v", err)
	} else {
		defer natsClient.Close()
		natsClient.SetupStreams(context.Background())
	}

	// WhatsApp Store (pkg/whatsapp)
	waStore, err := wapkg.NewStore(database.DefaultConfig().DSN())
	if err != nil {
		log.Fatalf("Failed to initialize WhatsApp store: %v", err)
	}

	// WhatsApp Manager (internal/channels/whatsapp)
	waManager := whatsapp.NewManager(waStore)

	// Repositories
	inboxRepo := repository.NewInboxRepository(db.DB)
	waChannelRepo := repository.NewChannelWhatsAppRepository(db.DB)
	memberRepo := repository.NewInboxMemberRepository(db.DB)
	contactRepo := repository.NewContactRepository(db.DB)
	contactInboxRepo := repository.NewContactInboxRepository(db.DB)
	conversationRepo := repository.NewConversationRepository(db.DB)
	messageRepo := repository.NewMessageRepository(db.DB)
	labelRepo := repository.NewLabelRepository(db.DB)

	// Services
	inboxService := services.NewInboxService(inboxRepo, waChannelRepo, memberRepo, waManager)
	contactService := services.NewContactService(contactRepo, contactInboxRepo)
	conversationService := services.NewConversationService(conversationRepo, contactRepo, labelRepo, inboxRepo, messageRepo)
	messageService := services.NewMessageService(messageRepo, conversationRepo, contactRepo, contactInboxRepo, inboxRepo, waManager)

	// Event Handler (conecta canal aos services)
	eventHandler := services.NewChannelEventHandler(inboxService, messageService)
	waManager.SetEventHandler(eventHandler)

	// Auth
	jwtService := auth.NewJWTService(nil)
	authMiddleware := middleware.NewAuthMiddleware(jwtService, db.DB)

	// Rate Limiters
	defaultRateLimiter := middleware.DefaultRateLimiter()
	strictRateLimiter := middleware.StrictRateLimiter()

	// Handlers
	authHandler := handlers.NewAuthHandler(db.DB, jwtService)
	apiKeyHandler := handlers.NewAPIKeyHandler(db.DB)
	inboxHandler := handlers.NewInboxHandler(inboxService)
	conversationHandler := handlers.NewConversationHandler(conversationService)
	messageHandler := handlers.NewMessageHandler(messageService)
	contactHandler := handlers.NewContactHandler(contactService, conversationService)
	labelHandler := handlers.NewLabelHandler(labelRepo)

	// WebSocket Hub (pkg/websocket)
	wsHub := wspkg.NewHub()
	go wsHub.Run()
	wsHandler := handlers.NewWebSocketHandler(wsHub)

	// Broadcaster (conecta WebSocket ao MessageService para real-time)
	broadcaster := services.NewWebSocketBroadcaster(wsHub)
	messageService.SetBroadcaster(broadcaster)

	// Echo
	e := echo.New()
	e.HideBanner = true

	// Setup Router
	router.Setup(e, router.Config{
		AuthMiddleware:    authMiddleware,
		RateLimiter:       defaultRateLimiter,
		StrictRateLimiter: strictRateLimiter,
	}, router.Handlers{
		Auth:         authHandler,
		APIKey:       apiKeyHandler,
		Inbox:        inboxHandler,
		Conversation: conversationHandler,
		Message:      messageHandler,
		Contact:      contactHandler,
		Label:        labelHandler,
		WebSocket:    wsHandler,
	})

	// Restore WhatsApp connections
	go func() {
		ctx := context.Background()
		if err := inboxService.RestoreConnections(ctx); err != nil {
			log.Printf("Warning: Failed to restore connections: %v", err)
		}
	}()

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	go func() {
		log.Printf("Server starting on port %s", port)
		if err := e.Start(":" + port); err != nil {
			log.Printf("Server stopped: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Println("Shutting down...")

	waManager.Shutdown()

	if natsClient != nil {
		natsClient.Close()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	e.Shutdown(ctx)

	log.Println("Server stopped")
}
