package router

import (
	"net/http"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/zyntra/backend/internal/api"
	"github.com/zyntra/backend/internal/handlers"
	"github.com/zyntra/backend/internal/middleware"
)

// Config configuracao do router
type Config struct {
	AuthMiddleware    *middleware.AuthMiddleware
	RateLimiter       *middleware.RateLimiter
	StrictRateLimiter *middleware.RateLimiter
}

// Handlers handlers da aplicacao
type Handlers struct {
	Auth         *handlers.AuthHandler
	APIKey       *handlers.APIKeyHandler
	Inbox        *handlers.InboxHandler
	Conversation *handlers.ConversationHandler
	Message      *handlers.MessageHandler
	Contact      *handlers.ContactHandler
	Label        *handlers.LabelHandler
	WebSocket    *handlers.WebSocketHandler
}

// Setup configura todas as rotas
func Setup(e *echo.Echo, cfg Config, h Handlers) {
	// Global middleware
	e.Use(echoMiddleware.Logger())
	e.Use(echoMiddleware.Recover())
	e.Use(echoMiddleware.CORSWithConfig(echoMiddleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:3000", "http://localhost:3001", "http://127.0.0.1:3000", "*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions, http.MethodPatch},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, "X-API-Key"},
	}))

	// Health check
	e.GET("/health", healthCheck)

	// API v1
	v1 := e.Group("/api/v1")

	// Auth routes (public)
	setupAuthRoutes(v1, h.Auth, cfg.StrictRateLimiter)

	// Protected routes
	protected := v1.Group("")
	protected.Use(cfg.AuthMiddleware.Authenticate)
	if cfg.RateLimiter != nil {
		protected.Use(cfg.RateLimiter.Middleware())
	}

	// Setup protected routes
	setupInboxRoutes(protected, h.Inbox)
	setupConversationRoutes(protected, h.Conversation, h.Message)
	setupContactRoutes(protected, h.Contact)
	setupLabelRoutes(protected, h.Label)
	setupAPIKeyRoutes(protected, h.APIKey)
	
	// WebSocket
	if h.WebSocket != nil {
		protected.GET("/ws", h.WebSocket.Handle)
	}
}

func healthCheck(c echo.Context) error {
	return api.Success(c, map[string]string{
		"status":  "ok",
		"version": "2.0.0",
	})
}

func setupAuthRoutes(g *echo.Group, h *handlers.AuthHandler, limiter *middleware.RateLimiter) {
	auth := g.Group("/auth")
	if limiter != nil {
		auth.Use(limiter.Middleware())
	}
	auth.POST("/login", h.Login)
	auth.POST("/register", h.Register)
	auth.POST("/refresh", h.RefreshToken)
}

func setupInboxRoutes(g *echo.Group, h *handlers.InboxHandler) {
	inboxes := g.Group("/inboxes")
	inboxes.GET("", h.List)
	inboxes.POST("", h.Create)
	inboxes.GET("/:id", h.Get)
	inboxes.PUT("/:id", h.Update)
	inboxes.DELETE("/:id", h.Delete)
	inboxes.POST("/:id/connect", h.Connect)
	inboxes.POST("/:id/disconnect", h.Disconnect)
	inboxes.GET("/:id/qrcode", h.GetQRCode)
}

func setupConversationRoutes(g *echo.Group, convH *handlers.ConversationHandler, msgH *handlers.MessageHandler) {
	conversations := g.Group("/conversations")
	conversations.GET("", convH.List)
	conversations.GET("/:id", convH.Get)
	conversations.PUT("/:id", convH.Update)
	conversations.DELETE("/:id", convH.Delete)
	conversations.POST("/:id/read", convH.MarkAsRead)
	conversations.POST("/:id/assign", convH.Assign)
	conversations.POST("/:id/favorite", convH.ToggleFavorite)
	conversations.POST("/:id/archive", convH.ToggleArchive)

	// Messages nested under conversations
	conversations.GET("/:id/messages", msgH.List)
	conversations.POST("/:id/messages", msgH.Send)
}

func setupContactRoutes(g *echo.Group, h *handlers.ContactHandler) {
	contacts := g.Group("/contacts")
	contacts.GET("", h.List)
	contacts.POST("", h.Create)
	contacts.GET("/:id", h.Get)
	contacts.PUT("/:id", h.Update)
	contacts.DELETE("/:id", h.Delete)
	contacts.GET("/:id/conversations", h.GetConversations)
}

func setupLabelRoutes(g *echo.Group, h *handlers.LabelHandler) {
	labels := g.Group("/labels")
	labels.GET("", h.List)
	labels.POST("", h.Create)
	labels.DELETE("/:id", h.Delete)
}

func setupAPIKeyRoutes(g *echo.Group, h *handlers.APIKeyHandler) {
	apikeys := g.Group("/api-keys")
	apikeys.GET("", h.ListAPIKeys)
	apikeys.POST("", h.CreateAPIKey)
	apikeys.DELETE("/:id", h.RevokeAPIKey)
}
