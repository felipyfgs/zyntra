package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/zyntra/backend/internal/api"
	"github.com/zyntra/backend/internal/domain"
	"github.com/zyntra/backend/internal/ports"
	"github.com/zyntra/backend/internal/services"
)

// InboxHandler handler de inboxes
type InboxHandler struct {
	service *services.InboxService
}

// NewInboxHandler cria novo handler
func NewInboxHandler(service *services.InboxService) *InboxHandler {
	return &InboxHandler{service: service}
}

// List lista todos os inboxes
func (h *InboxHandler) List(c echo.Context) error {
	inboxes, err := h.service.GetAll(c.Request().Context())
	if err != nil {
		return api.InternalError(c, err.Error())
	}
	return api.Success(c, inboxes)
}

// Get retorna um inbox por ID
func (h *InboxHandler) Get(c echo.Context) error {
	id := c.Param("id")
	inbox, err := h.service.GetByID(c.Request().Context(), id)
	if err != nil {
		return api.NotFound(c, err.Error())
	}
	return api.Success(c, inbox)
}

// CreateInboxRequest request para criar inbox
type CreateInboxRequest struct {
	Name            string `json:"name" validate:"required"`
	ChannelType     string `json:"channel_type" validate:"required"`
	GreetingMessage string `json:"greeting_message,omitempty"`
	AutoAssignment  bool   `json:"auto_assignment"`
}

// Create cria um novo inbox
func (h *InboxHandler) Create(c echo.Context) error {
	var req CreateInboxRequest
	if err := c.Bind(&req); err != nil {
		return api.BadRequest(c, "Invalid request body")
	}

	if req.Name == "" {
		return api.ValidationError(c, "Name is required")
	}
	if req.ChannelType == "" {
		return api.ValidationError(c, "Channel type is required")
	}

	inbox, err := h.service.Create(c.Request().Context(), domain.CreateInboxRequest{
		Name:            req.Name,
		ChannelType:     ports.ChannelType(req.ChannelType),
		GreetingMessage: req.GreetingMessage,
		AutoAssignment:  req.AutoAssignment,
	})
	if err != nil {
		return api.InternalError(c, err.Error())
	}

	return api.Created(c, inbox)
}

// UpdateInboxRequest request para atualizar inbox
type UpdateInboxRequest struct {
	Name            *string `json:"name,omitempty"`
	GreetingMessage *string `json:"greeting_message,omitempty"`
	AutoAssignment  *bool   `json:"auto_assignment,omitempty"`
}

// Update atualiza um inbox
func (h *InboxHandler) Update(c echo.Context) error {
	id := c.Param("id")
	var req UpdateInboxRequest
	if err := c.Bind(&req); err != nil {
		return api.BadRequest(c, "Invalid request body")
	}

	inbox, err := h.service.GetByID(c.Request().Context(), id)
	if err != nil {
		return api.NotFound(c, err.Error())
	}

	if req.Name != nil {
		inbox.Name = *req.Name
	}
	if req.GreetingMessage != nil {
		inbox.GreetingMessage = *req.GreetingMessage
	}
	if req.AutoAssignment != nil {
		inbox.AutoAssignment = *req.AutoAssignment
	}

	return api.Success(c, inbox)
}

// Delete remove um inbox
func (h *InboxHandler) Delete(c echo.Context) error {
	id := c.Param("id")
	if err := h.service.Delete(c.Request().Context(), id); err != nil {
		return api.InternalError(c, err.Error())
	}
	return api.NoContent(c)
}

// Connect conecta o canal do inbox
func (h *InboxHandler) Connect(c echo.Context) error {
	id := c.Param("id")
	if err := h.service.Connect(c.Request().Context(), id); err != nil {
		return api.InternalError(c, err.Error())
	}
	return api.Success(c, map[string]interface{}{
		"id":      id,
		"status":  "connecting",
		"message": "Connection initiated. QR code will be available via /qrcode endpoint or WebSocket.",
	})
}

// Disconnect desconecta o canal do inbox
func (h *InboxHandler) Disconnect(c echo.Context) error {
	id := c.Param("id")
	if err := h.service.Disconnect(c.Request().Context(), id); err != nil {
		return api.InternalError(c, err.Error())
	}
	return api.Success(c, map[string]interface{}{
		"id":     id,
		"status": "disconnected",
	})
}

// GetQRCode retorna o QR code do inbox
func (h *InboxHandler) GetQRCode(c echo.Context) error {
	id := c.Param("id")
	qrCode := h.service.GetQRCode(c.Request().Context(), id)
	status := h.service.GetStatus(id)

	return api.Success(c, map[string]interface{}{
		"id":      id,
		"qrcode":  qrCode,
		"status":  status,
	})
}
