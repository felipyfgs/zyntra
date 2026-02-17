package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/zyntra/backend/pkg/whatsapp"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// TODO: Implement proper origin check
		return true
	},
}

// WebSocketMessage represents a message sent/received via WebSocket
type WebSocketMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// WebSocketHub manages all WebSocket connections
type WebSocketHub struct {
	clients    map[*websocket.Conn]string // conn -> userID
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	broadcast  chan whatsapp.Event
	mu         sync.RWMutex
}

// NewWebSocketHub creates a new WebSocket hub
func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		clients:    make(map[*websocket.Conn]string),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
		broadcast:  make(chan whatsapp.Event, 256),
	}
}

// Run starts the hub's event loop
func (h *WebSocketHub) Run() {
	for {
		select {
		case conn := <-h.register:
			h.mu.Lock()
			h.clients[conn] = "default-user" // TODO: Get from auth
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
					Type:    string(event.Type),
					Payload: event,
				}
				if err := conn.WriteJSON(msg); err != nil {
					log.Printf("[WebSocket] Error sending message: %v", err)
					conn.Close()
					delete(h.clients, conn)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast implements whatsapp.EventBroadcaster
func (h *WebSocketHub) Broadcast(event whatsapp.Event) {
	select {
	case h.broadcast <- event:
	default:
		log.Printf("[WebSocket] Broadcast channel full, dropping event")
	}
}

// WebSocketHandler handles WebSocket connections
type WebSocketHandler struct {
	hub *WebSocketHub
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(hub *WebSocketHub) *WebSocketHandler {
	return &WebSocketHandler{hub: hub}
}

// Handle upgrades HTTP connection to WebSocket
func (h *WebSocketHandler) Handle(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	// Register client
	h.hub.register <- ws

	// Cleanup on exit
	defer func() {
		h.hub.unregister <- ws
	}()

	// Handle incoming messages
	for {
		var msg WebSocketMessage
		err := ws.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[WebSocket] Error: %v", err)
			}
			break
		}

		// Handle different message types
		switch msg.Type {
		case "ping":
			ws.WriteJSON(WebSocketMessage{Type: "pong", Payload: nil})
		
		case "subscribe":
			// Subscribe to specific connection events
			ws.WriteJSON(WebSocketMessage{Type: "subscribed", Payload: msg.Payload})
		
		default:
			log.Printf("[WebSocket] Unknown message type: %s", msg.Type)
		}
	}

	return nil
}

// LegacyWebSocketHandler is kept for backward compatibility
func LegacyWebSocketHandler(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	for {
		var msg WebSocketMessage
		err := ws.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		switch msg.Type {
		case "ping":
			ws.WriteJSON(WebSocketMessage{Type: "pong", Payload: nil})
		case "subscribe":
			ws.WriteJSON(WebSocketMessage{Type: "subscribed", Payload: msg.Payload})
		default:
			log.Printf("Unknown message type: %s", msg.Type)
		}
	}

	return nil
}

// SendQRCodeToClients is a helper to broadcast QR code events
func SendQRCodeToClients(hub *WebSocketHub, connectionID, qrCode string) {
	event := whatsapp.Event{
		Type:         whatsapp.EventTypeQRCode,
		ConnectionID: connectionID,
		Data: map[string]string{
			"code": qrCode,
		},
	}
	hub.Broadcast(event)
}

// Helper to marshal event data
func marshalEvent(event whatsapp.Event) []byte {
	data, _ := json.Marshal(event)
	return data
}
