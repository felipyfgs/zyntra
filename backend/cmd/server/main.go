package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/zyntra/backend/internal/api"
	"github.com/zyntra/backend/internal/auth"
	"github.com/zyntra/backend/internal/database"
	"github.com/zyntra/backend/internal/handlers"
	"github.com/zyntra/backend/internal/middleware"
	"github.com/zyntra/backend/internal/services"
	natspkg "github.com/zyntra/backend/pkg/nats"
	"github.com/zyntra/backend/pkg/whatsapp"
)

func main() {
	log.Println("=== Zyntra API Server ===")

	// Initialize database
	db, err := database.New(database.DefaultConfig())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	migrator := database.NewMigrator(db.DB)
	if err := migrator.Run(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize NATS client
	natsClient, err := natspkg.NewClient(nil)
	if err != nil {
		log.Printf("Warning: Failed to connect to NATS: %v (real-time features disabled)", err)
	} else {
		defer natsClient.Close()
		// Setup JetStream streams
		if err := natsClient.SetupStreams(context.Background()); err != nil {
			log.Printf("Warning: Failed to setup NATS streams: %v", err)
		}
	}

	// Initialize WhatsApp store
	waStore, err := whatsapp.NewStore(database.DefaultConfig().DSN())
	if err != nil {
		log.Fatalf("Failed to initialize WhatsApp store: %v", err)
	}

	// Initialize WebSocket hub (legacy, kept for backward compatibility)
	wsHub := handlers.NewWebSocketHub()
	go wsHub.Run()

	// Initialize NATS broadcaster for real-time events
	var natsBroadcaster whatsapp.EventBroadcaster
	if natsClient != nil {
		natsBroadcaster = whatsapp.NewNATSBroadcaster(natsClient)
	} else {
		natsBroadcaster = wsHub // Fallback to WebSocket hub
	}

	// Initialize WhatsApp manager with NATS broadcaster
	waManager := whatsapp.NewManager(waStore, natsBroadcaster)

	// Initialize repositories
	connectionRepo := whatsapp.NewConnectionRepository(db.DB)
	messageRepo := whatsapp.NewMessageRepository(db.DB)

	// Initialize event handler with NATS broadcaster
	eventHandler := whatsapp.NewDefaultEventHandler(natsBroadcaster, messageRepo, connectionRepo)
	eventHandler.SetManager(waManager)
	waManager.SetEventHandler(eventHandler)

	// Initialize services
	waService := services.NewWhatsAppService(waManager, connectionRepo, messageRepo)

	// Initialize JWT service
	jwtService := auth.NewJWTService(nil)

	// Initialize auth middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtService, db.DB)

	// Initialize rate limiters
	defaultRateLimiter := middleware.DefaultRateLimiter()
	strictRateLimiter := middleware.StrictRateLimiter()

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db.DB, jwtService)
	apiKeyHandler := handlers.NewAPIKeyHandler(db.DB)
	chatHandler := handlers.NewChatHandler(db.DB, waService)
	filterHandler := handlers.NewFilterHandler(db.DB)
	connHandler := handlers.NewConnectionHandler(waService)
	msgHandler := handlers.NewMessageHandler(waService)
	wsHandler := handlers.NewWebSocketHandler(wsHub)

	// Restore existing connections
	go func() {
		ctx := context.Background()
		if err := waManager.RestoreConnections(ctx); err != nil {
			log.Printf("Warning: Failed to restore connections: %v", err)
		}
	}()

	// Initialize Echo
	e := echo.New()
	e.HideBanner = true

	// Global middleware
	e.Use(echoMiddleware.Logger())
	e.Use(echoMiddleware.Recover())
	e.Use(echoMiddleware.CORSWithConfig(echoMiddleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:3000", "http://localhost:3001", "http://127.0.0.1:3000"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions, http.MethodPatch},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, "X-API-Key"},
	}))

	// Health check (public)
	e.GET("/health", func(c echo.Context) error {
		natsStatus := "disconnected"
		if natsClient != nil {
			natsStatus = "connected"
		}
		return api.Success(c, map[string]string{
			"status":   "ok",
			"database": "connected",
			"nats":     natsStatus,
			"version":  "1.0.0",
		})
	})

	// API v1 routes
	v1 := e.Group("/api/v1")

	// Public auth routes (with strict rate limiting)
	authRoutes := v1.Group("/auth")
	authRoutes.Use(strictRateLimiter.Middleware())
	authRoutes.POST("/login", authHandler.Login)
	authRoutes.POST("/register", authHandler.Register)
	authRoutes.POST("/refresh", authHandler.RefreshToken)

	// Protected routes (require authentication)
	protected := v1.Group("")
	protected.Use(authMiddleware.Authenticate)
	protected.Use(defaultRateLimiter.Middleware())

	// User profile
	protected.GET("/auth/profile", authHandler.GetProfile)
	protected.PUT("/auth/password", authHandler.ChangePassword)

	// API Keys management
	protected.GET("/api-keys", apiKeyHandler.ListAPIKeys)
	protected.POST("/api-keys", apiKeyHandler.CreateAPIKey)
	protected.DELETE("/api-keys/:id", apiKeyHandler.RevokeAPIKey)

	// Chat routes
	protected.GET("/chats", chatHandler.GetChats)
	protected.GET("/chats/:id", chatHandler.GetChat)
	protected.PUT("/chats/:id", chatHandler.UpdateChat)
	protected.POST("/chats/:id/read", chatHandler.MarkChatAsRead)
	protected.GET("/chats/:id/messages", chatHandler.GetMessages)
	protected.POST("/chats/:id/messages", chatHandler.SendMessage)

	// Saved filters routes
	protected.GET("/filters", filterHandler.GetFilters)
	protected.POST("/filters", filterHandler.CreateFilter)
	protected.PUT("/filters/:id", filterHandler.UpdateFilter)
	protected.DELETE("/filters/:id", filterHandler.DeleteFilter)
	protected.PUT("/filters/reorder", filterHandler.ReorderFilters)

	// Contact routes
	protected.GET("/contacts", handlers.GetContacts)
	protected.POST("/contacts", handlers.CreateContact)
	protected.PUT("/contacts/:id", handlers.UpdateContact)
	protected.DELETE("/contacts/:id", handlers.DeleteContact)

	// Connection routes (with WhatsApp integration)
	protected.GET("/connections", connHandler.GetConnections)
	protected.POST("/connections", connHandler.CreateConnection)
	protected.GET("/connections/:id", connHandler.GetConnectionStatus)
	protected.DELETE("/connections/:id", connHandler.DeleteConnection)
	protected.POST("/connections/:id/connect", connHandler.ConnectWhatsApp)
	protected.POST("/connections/:id/disconnect", connHandler.DisconnectWhatsApp)
	protected.GET("/connections/:id/qr", connHandler.GetQRCode)

	// Message routes (with WhatsApp integration)
	protected.POST("/messages/send", msgHandler.SendMessage)
	protected.GET("/messages", msgHandler.GetMessages)

	// WebSocket for real-time updates (authenticated)
	protected.GET("/ws", wsHandler.Handle)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	go func() {
		log.Printf("Server starting on port %s", port)
		if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
			log.Fatal("shutting down the server")
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Println("Shutting down server...")

	// Shutdown WhatsApp manager
	waManager.Shutdown()

	// Shutdown NATS client
	if natsClient != nil {
		natsClient.Close()
	}

	// Shutdown HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}

	log.Println("Server stopped")
}
