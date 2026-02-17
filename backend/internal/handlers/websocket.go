package handlers

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	wspkg "github.com/zyntra/backend/pkg/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WebSocketMessage mensagem do WebSocket
type WebSocketMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// WebSocketHandler handler de WebSocket
type WebSocketHandler struct {
	hub *wspkg.Hub
}

// NewWebSocketHandler cria novo handler
func NewWebSocketHandler(hub *wspkg.Hub) *WebSocketHandler {
	return &WebSocketHandler{hub: hub}
}

// Handle faz upgrade da conexao
func (h *WebSocketHandler) Handle(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	h.hub.Register(ws)

	defer func() {
		h.hub.Unregister(ws)
	}()

	for {
		var msg WebSocketMessage
		err := ws.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[WebSocket] Error: %v", err)
			}
			break
		}

		switch msg.Type {
		case "ping":
			ws.WriteJSON(WebSocketMessage{Type: "pong"})
		case "subscribe":
			ws.WriteJSON(WebSocketMessage{Type: "subscribed", Payload: msg.Payload})
		}
	}

	return nil
}

// Hub retorna o hub subjacente
func (h *WebSocketHandler) Hub() *wspkg.Hub {
	return h.hub
}
