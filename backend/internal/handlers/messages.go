package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/zyntra/backend/internal/services"
)

// MessageHandler handles message endpoints
type MessageHandler struct {
	service *services.WhatsAppService
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(service *services.WhatsAppService) *MessageHandler {
	return &MessageHandler{service: service}
}

// WhatsAppSendMessageRequest represents the request body for sending a WhatsApp message
type WhatsAppSendMessageRequest struct {
	ConnectionID string `json:"connection_id" validate:"required"`
	To           string `json:"to" validate:"required"`
	Content      string `json:"content" validate:"required"`
	MediaType    string `json:"media_type,omitempty"`
	MediaURL     string `json:"media_url,omitempty"`
}

// SendMessage sends a message through a WhatsApp connection
func (h *MessageHandler) SendMessage(c echo.Context) error {
	var req WhatsAppSendMessageRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	msg, err := h.service.SendMessage(c.Request().Context(), req.ConnectionID, req.To, req.Content)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, msg)
}

// GetMessages retrieves messages for a chat
func (h *MessageHandler) GetMessages(c echo.Context) error {
	connectionID := c.QueryParam("connection_id")
	chatJID := c.QueryParam("chat_jid")
	
	if connectionID == "" || chatJID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "connection_id and chat_jid are required"})
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	if limit == 0 {
		limit = 50
	}

	messages, err := h.service.GetMessages(c.Request().Context(), connectionID, chatJID, limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, messages)
}
