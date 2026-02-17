package handlers

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
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

// WebSocketEvent evento para broadcast
type WebSocketEvent struct {
	Type    string      `json:"type"`
	InboxID string      `json:"inbox_id,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// WebSocketHub gerencia conexoes WebSocket
type WebSocketHub struct {
	clients    map[*websocket.Conn]string
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	broadcast  chan WebSocketEvent
	mu         sync.RWMutex
}

// NewWebSocketHub cria novo hub
func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		clients:    make(map[*websocket.Conn]string),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
		broadcast:  make(chan WebSocketEvent, 256),
	}
}

// Run inicia o hub
func (h *WebSocketHub) Run() {
	for {
		select {
		case conn := <-h.register:
			h.mu.Lock()
			h.clients[conn] = "default-user"
			h.mu.Unlock()
			log.Printf("[WebSocket] Client connected, total: %d", len(h.clients))

		case conn := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[conn]; ok {
				delete(h.clients, conn)
				conn.Close()
			}
			h.mu.Unlock()
			log.Printf("[WebSocket] Client disconnected, total: %d", len(h.clients))

		case event := <-h.broadcast:
			h.mu.RLock()
			for conn := range h.clients {
				msg := WebSocketMessage{
					Type:    event.Type,
					Payload: event,
				}
				if err := conn.WriteJSON(msg); err != nil {
					log.Printf("[WebSocket] Error sending: %v", err)
					conn.Close()
					delete(h.clients, conn)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast envia evento para todos os clientes
func (h *WebSocketHub) Broadcast(event WebSocketEvent) {
	select {
	case h.broadcast <- event:
	default:
		log.Printf("[WebSocket] Broadcast channel full")
	}
}

// BroadcastQRCode envia QR code
func (h *WebSocketHub) BroadcastQRCode(inboxID, qrCode string) {
	h.Broadcast(WebSocketEvent{
		Type:    "qr_code",
		InboxID: inboxID,
		Data:    map[string]string{"qr_code": qrCode},
	})
}

// BroadcastConnectionStatus envia status de conexao
func (h *WebSocketHub) BroadcastConnectionStatus(inboxID, status, phone string) {
	h.Broadcast(WebSocketEvent{
		Type:    "connection_status",
		InboxID: inboxID,
		Data:    map[string]string{"status": status, "phone": phone},
	})
}

// BroadcastMessage envia nova mensagem
func (h *WebSocketHub) BroadcastMessage(inboxID string, message interface{}) {
	h.Broadcast(WebSocketEvent{
		Type:    "message",
		InboxID: inboxID,
		Data:    message,
	})
}

// BroadcastConversationUpdate envia atualizacao de conversa
func (h *WebSocketHub) BroadcastConversationUpdate(inboxID string, conversation interface{}) {
	h.Broadcast(WebSocketEvent{
		Type:    "conversation_update",
		InboxID: inboxID,
		Data:    conversation,
	})
}

// WebSocketHandler handler de WebSocket
type WebSocketHandler struct {
	hub *WebSocketHub
}

// NewWebSocketHandler cria novo handler
func NewWebSocketHandler(hub *WebSocketHub) *WebSocketHandler {
	return &WebSocketHandler{hub: hub}
}

// Handle faz upgrade da conexao
func (h *WebSocketHandler) Handle(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	h.hub.register <- ws

	defer func() {
		h.hub.unregister <- ws
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
