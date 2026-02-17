package handlers

import (
	"database/sql"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/zyntra/backend/internal/api"
	"github.com/zyntra/backend/internal/middleware"
	"github.com/zyntra/backend/internal/repository"
	"github.com/zyntra/backend/internal/services"
)

// ChatHandler handles chat-related endpoints
type ChatHandler struct {
	chatRepo *repository.ChatRepository
	msgRepo  *repository.MessageRepository
	waService *services.WhatsAppService
}

// NewChatHandler creates a new chat handler
func NewChatHandler(db *sql.DB, waService *services.WhatsAppService) *ChatHandler {
	return &ChatHandler{
		chatRepo:  repository.NewChatRepository(db),
		msgRepo:   repository.NewMessageRepository(db),
		waService: waService,
	}
}

// GetChats returns a list of chats with filters and pagination
func (h *ChatHandler) GetChats(c echo.Context) error {
	user := middleware.GetUser(c)
	if user == nil {
		return api.Unauthorized(c, "Not authenticated")
	}

	// Parse query parameters
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))

	filter := repository.ChatFilter{
		ConnectionID: c.QueryParam("connection_id"),
		Search:       c.QueryParam("search"),
		Filter:       c.QueryParam("filter"),
		Page:         page,
		PerPage:      perPage,
	}

	chats, total, err := h.chatRepo.List(c.Request().Context(), user.UserID, filter)
	if err != nil {
		return api.InternalError(c, "Failed to get chats")
	}

	// Calculate pagination meta
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage < 1 {
		filter.PerPage = 20
	}

	return api.SuccessWithMeta(c, chats, api.NewMeta(filter.Page, filter.PerPage, total))
}

// GetChat returns a single chat by ID
func (h *ChatHandler) GetChat(c echo.Context) error {
	user := middleware.GetUser(c)
	if user == nil {
		return api.Unauthorized(c, "Not authenticated")
	}

	chatID := c.Param("id")
	if chatID == "" {
		return api.BadRequest(c, "Chat ID is required")
	}

	chat, err := h.chatRepo.GetByID(c.Request().Context(), user.UserID, chatID)
	if err != nil {
		return api.InternalError(c, "Failed to get chat")
	}
	if chat == nil {
		return api.NotFound(c, "Chat not found")
	}

	return api.Success(c, chat)
}

// UpdateChatRequest represents the update chat request body
type UpdateChatRequest struct {
	IsFavorite *bool `json:"is_favorite,omitempty"`
	IsArchived *bool `json:"is_archived,omitempty"`
}

// UpdateChat updates chat fields (favorite, archived)
func (h *ChatHandler) UpdateChat(c echo.Context) error {
	user := middleware.GetUser(c)
	if user == nil {
		return api.Unauthorized(c, "Not authenticated")
	}

	chatID := c.Param("id")
	if chatID == "" {
		return api.BadRequest(c, "Chat ID is required")
	}

	var req UpdateChatRequest
	if err := c.Bind(&req); err != nil {
		return api.BadRequest(c, "Invalid request body")
	}

	updates := make(map[string]interface{})
	if req.IsFavorite != nil {
		updates["is_favorite"] = *req.IsFavorite
	}
	if req.IsArchived != nil {
		updates["is_archived"] = *req.IsArchived
	}

	if len(updates) == 0 {
		return api.BadRequest(c, "No fields to update")
	}

	err := h.chatRepo.Update(c.Request().Context(), user.UserID, chatID, updates)
	if err != nil {
		if err == sql.ErrNoRows {
			return api.NotFound(c, "Chat not found")
		}
		return api.InternalError(c, "Failed to update chat")
	}

	// Return updated chat
	chat, _ := h.chatRepo.GetByID(c.Request().Context(), user.UserID, chatID)
	return api.Success(c, chat)
}

// MarkChatAsRead marks all messages in a chat as read
func (h *ChatHandler) MarkChatAsRead(c echo.Context) error {
	user := middleware.GetUser(c)
	if user == nil {
		return api.Unauthorized(c, "Not authenticated")
	}

	chatID := c.Param("id")
	if chatID == "" {
		return api.BadRequest(c, "Chat ID is required")
	}

	err := h.chatRepo.MarkAsRead(c.Request().Context(), user.UserID, chatID)
	if err != nil {
		return api.InternalError(c, "Failed to mark chat as read")
	}

	return api.Success(c, map[string]string{"message": "Chat marked as read"})
}

// GetMessages returns messages for a chat with pagination
func (h *ChatHandler) GetMessages(c echo.Context) error {
	user := middleware.GetUser(c)
	if user == nil {
		return api.Unauthorized(c, "Not authenticated")
	}

	chatID := c.Param("id")
	if chatID == "" {
		return api.BadRequest(c, "Chat ID is required")
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))

	filter := repository.MessageFilter{
		ChatID:  chatID,
		Page:    page,
		PerPage: perPage,
	}

	messages, total, err := h.msgRepo.List(c.Request().Context(), user.UserID, filter)
	if err != nil {
		if err.Error() == "chat not found" {
			return api.NotFound(c, "Chat not found")
		}
		return api.InternalError(c, "Failed to get messages")
	}

	// Calculate pagination meta
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage < 1 {
		filter.PerPage = 50
	}

	return api.SuccessWithMeta(c, messages, api.NewMeta(filter.Page, filter.PerPage, total))
}

// SendMessageRequest represents the send message request body
type SendMessageRequest struct {
	Content   string `json:"content" validate:"required"`
	MediaURL  string `json:"media_url,omitempty"`
	MediaType string `json:"media_type,omitempty"`
}

// SendMessage sends a message to a chat
func (h *ChatHandler) SendMessage(c echo.Context) error {
	user := middleware.GetUser(c)
	if user == nil {
		return api.Unauthorized(c, "Not authenticated")
	}

	chatID := c.Param("id")
	if chatID == "" {
		return api.BadRequest(c, "Chat ID is required")
	}

	var req SendMessageRequest
	if err := c.Bind(&req); err != nil {
		return api.BadRequest(c, "Invalid request body")
	}

	if req.Content == "" && req.MediaURL == "" {
		return api.ValidationError(c, "Content or media is required")
	}

	// Get chat to find connection and JID
	chat, err := h.chatRepo.GetByID(c.Request().Context(), user.UserID, chatID)
	if err != nil || chat == nil {
		return api.NotFound(c, "Chat not found")
	}

	// Send via WhatsApp service
	msg, err := h.waService.SendMessage(c.Request().Context(), chat.ConnectionID, chat.JID, req.Content)
	if err != nil {
		return api.InternalError(c, "Failed to send message: "+err.Error())
	}

	return api.Created(c, msg)
}

// Legacy handlers for backward compatibility (deprecated)
func GetChats(c echo.Context) error {
	chats := []map[string]interface{}{
		{
			"id":              "1",
			"contact_name":    "John Doe",
			"contact_phone":   "+5511999999999",
			"last_message":    "Hello!",
			"last_message_at": "2024-02-16T10:30:00Z",
			"unread_count":    2,
		},
	}
	return c.JSON(200, chats)
}

func GetMessages(c echo.Context) error {
	chatID := c.Param("id")
	messages := []map[string]interface{}{
		{"id": "1", "chat_id": chatID, "direction": "inbound", "content": "Hello!", "status": "read"},
		{"id": "2", "chat_id": chatID, "direction": "outbound", "content": "Hi there!", "status": "delivered"},
	}
	return c.JSON(200, messages)
}

func SendMessage(c echo.Context) error {
	chatID := c.Param("id")
	var req SendMessageRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(400, map[string]string{"error": "Invalid request"})
	}
	message := map[string]interface{}{
		"id": "3", "chat_id": chatID, "direction": "outbound",
		"content": req.Content, "status": "pending",
	}
	return c.JSON(201, message)
}
