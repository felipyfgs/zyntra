package websocket

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// Event evento para broadcast
type Event struct {
	Type    string      `json:"type"`
	InboxID string      `json:"inbox_id,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// Message mensagem do WebSocket
type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// Hub gerencia conexoes WebSocket
type Hub struct {
	clients    map[*websocket.Conn]string
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	broadcast  chan Event
	mu         sync.RWMutex
}

// NewHub cria novo hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*websocket.Conn]string),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
		broadcast:  make(chan Event, 256),
	}
}

// Run inicia o loop do hub
func (h *Hub) Run() {
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
				msg := Message{
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

// Register registra nova conexao
func (h *Hub) Register(conn *websocket.Conn) {
	h.register <- conn
}

// Unregister remove conexao
func (h *Hub) Unregister(conn *websocket.Conn) {
	h.unregister <- conn
}

// Broadcast envia evento para todos os clientes
func (h *Hub) Broadcast(event Event) {
	select {
	case h.broadcast <- event:
	default:
		log.Printf("[WebSocket] Broadcast channel full")
	}
}

// BroadcastQRCode envia QR code
func (h *Hub) BroadcastQRCode(inboxID, qrCode string) {
	h.Broadcast(Event{
		Type:    "qr_code",
		InboxID: inboxID,
		Data:    map[string]string{"qr_code": qrCode},
	})
}

// BroadcastConnectionStatus envia status de conexao
func (h *Hub) BroadcastConnectionStatus(inboxID, status, phone string) {
	h.Broadcast(Event{
		Type:    "connection_status",
		InboxID: inboxID,
		Data:    map[string]string{"status": status, "phone": phone},
	})
}

// BroadcastMessage envia nova mensagem
func (h *Hub) BroadcastMessage(inboxID string, message interface{}) {
	h.Broadcast(Event{
		Type:    "message",
		InboxID: inboxID,
		Data:    message,
	})
}

// BroadcastConversationUpdate envia atualizacao de conversa
func (h *Hub) BroadcastConversationUpdate(inboxID string, conversation interface{}) {
	h.Broadcast(Event{
		Type:    "conversation_update",
		InboxID: inboxID,
		Data:    conversation,
	})
}

// ClientCount retorna numero de clientes conectados
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
