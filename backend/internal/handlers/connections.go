package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/zyntra/backend/internal/api"
	"github.com/zyntra/backend/internal/middleware"
	"github.com/zyntra/backend/internal/services"
)

// ConnectionHandler handles WhatsApp connection endpoints
type ConnectionHandler struct {
	service *services.WhatsAppService
}

// NewConnectionHandler creates a new connection handler
func NewConnectionHandler(service *services.WhatsAppService) *ConnectionHandler {
	return &ConnectionHandler{service: service}
}

// GetConnections returns all connections for the current user
func (h *ConnectionHandler) GetConnections(c echo.Context) error {
	user := middleware.GetUser(c)
	if user == nil {
		return api.Unauthorized(c, "Authentication required")
	}

	connections, err := h.service.GetUserConnections(c.Request().Context(), user.UserID)
	if err != nil {
		return api.InternalError(c, err.Error())
	}

	return api.Success(c, connections)
}

// CreateConnectionRequest represents the request body for creating a connection
type CreateConnectionRequest struct {
	Name string `json:"name" validate:"required"`
}

// CreateConnection creates a new WhatsApp connection
func (h *ConnectionHandler) CreateConnection(c echo.Context) error {
	var req CreateConnectionRequest
	if err := c.Bind(&req); err != nil {
		return api.BadRequest(c, "Invalid request")
	}

	user := middleware.GetUser(c)
	if user == nil {
		return api.Unauthorized(c, "Authentication required")
	}

	conn, err := h.service.CreateConnection(c.Request().Context(), user.UserID, req.Name)
	if err != nil {
		return api.InternalError(c, err.Error())
	}

	return api.Created(c, conn)
}

// DeleteConnection removes a WhatsApp connection
func (h *ConnectionHandler) DeleteConnection(c echo.Context) error {
	id := c.Param("id")

	if err := h.service.DeleteConnection(c.Request().Context(), id); err != nil {
		return api.InternalError(c, err.Error())
	}

	return api.NoContent(c)
}

// ConnectWhatsApp initiates WhatsApp connection (QR code flow)
func (h *ConnectionHandler) ConnectWhatsApp(c echo.Context) error {
	id := c.Param("id")

	if err := h.service.Connect(c.Request().Context(), id); err != nil {
		return api.InternalError(c, err.Error())
	}

	// QR code will be sent via NATS/WebSocket
	return api.Success(c, map[string]interface{}{
		"id":      id,
		"status":  "connecting",
		"message": "QR code will be sent via NATS",
	})
}

// DisconnectWhatsApp disconnects a WhatsApp connection
func (h *ConnectionHandler) DisconnectWhatsApp(c echo.Context) error {
	id := c.Param("id")

	if err := h.service.Disconnect(c.Request().Context(), id); err != nil {
		return api.InternalError(c, err.Error())
	}

	return api.Success(c, map[string]interface{}{
		"id":     id,
		"status": "disconnected",
	})
}

// GetConnectionStatus returns the current status of a connection
func (h *ConnectionHandler) GetConnectionStatus(c echo.Context) error {
	id := c.Param("id")

	conn, err := h.service.GetConnection(c.Request().Context(), id)
	if err != nil {
		return api.NotFound(c, err.Error())
	}

	return api.Success(c, conn)
}

// GetQRCode returns the current QR code for a connection (wuzapi pattern: from DB)
func (h *ConnectionHandler) GetQRCode(c echo.Context) error {
	id := c.Param("id")

	qrCode := h.service.GetQRCode(c.Request().Context(), id)
	if qrCode == "" {
		return api.Success(c, map[string]interface{}{
			"id":      id,
			"qr_code": nil,
			"status":  "waiting",
		})
	}

	return api.Success(c, map[string]interface{}{
		"id":      id,
		"qr_code": qrCode,
		"status":  "ready",
	})
}

// Legacy handlers for backward compatibility (will be removed)
func GetConnections(c echo.Context) error {
	connections := []map[string]interface{}{
		{
			"id":         "1",
			"phone":      "+5511988888888",
			"status":     "connected",
			"created_at": "2024-02-16T09:00:00Z",
		},
	}
	return c.JSON(http.StatusOK, connections)
}

func CreateConnection(c echo.Context) error {
	var req CreateConnectionRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}
	connection := map[string]interface{}{
		"id":         "2",
		"name":       req.Name,
		"status":     "disconnected",
		"created_at": "2024-02-16T10:00:00Z",
	}
	return c.JSON(http.StatusCreated, connection)
}

func DeleteConnection(c echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}

func ConnectWhatsApp(c echo.Context) error {
	id := c.Param("id")
	response := map[string]interface{}{
		"id":      id,
		"status":  "qr_code",
		"qr_code": "data:image/png;base64,placeholder_qr_code",
	}
	return c.JSON(http.StatusOK, response)
}

func DisconnectWhatsApp(c echo.Context) error {
	id := c.Param("id")
	response := map[string]interface{}{
		"id":     id,
		"status": "disconnected",
	}
	return c.JSON(http.StatusOK, response)
}
