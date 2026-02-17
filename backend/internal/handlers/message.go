package handlers

import (
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/zyntra/backend/internal/api"
	"github.com/zyntra/backend/internal/domain"
	"github.com/zyntra/backend/internal/middleware"
	"github.com/zyntra/backend/internal/services"
)

// MessageHandler handler de mensagens
type MessageHandler struct {
	service *services.MessageService
}

// NewMessageHandler cria novo handler
func NewMessageHandler(service *services.MessageService) *MessageHandler {
	return &MessageHandler{service: service}
}

// List lista mensagens de uma conversa
func (h *MessageHandler) List(c echo.Context) error {
	conversationID := c.Param("id")
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	if limit <= 0 {
		limit = 50
	}

	messages, err := h.service.GetMessages(c.Request().Context(), conversationID, limit, offset)
	if err != nil {
		return api.InternalError(c, err.Error())
	}

	return api.Success(c, messages)
}

// SendMessageRequest request para enviar mensagem
type SendMessageRequest struct {
	Content     string `json:"content" validate:"required"`
	ContentType string `json:"content_type,omitempty"`
	Private     bool   `json:"private,omitempty"`
}

// Send envia uma mensagem
func (h *MessageHandler) Send(c echo.Context) error {
	conversationID := c.Param("id")

	var req SendMessageRequest
	if err := c.Bind(&req); err != nil {
		return api.BadRequest(c, "Invalid request body")
	}

	if req.Content == "" {
		return api.ValidationError(c, "Content is required")
	}

	// Obter usuario autenticado
	user := middleware.GetUser(c)
	senderID := ""
	if user != nil {
		senderID = user.UserID
	}

	contentType := domain.ContentTypeText
	if req.ContentType != "" {
		contentType = domain.ContentType(req.ContentType)
	}

	msg, err := h.service.SendMessage(c.Request().Context(), conversationID, domain.SendMessageRequest{
		Content:     req.Content,
		ContentType: contentType,
		Private:     req.Private,
	}, senderID)
	if err != nil {
		return api.InternalError(c, err.Error())
	}

	return api.Created(c, msg)
}
