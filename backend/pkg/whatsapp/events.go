package whatsapp

import (
	"context"
	"encoding/json"
	"log"

	natspkg "github.com/zyntra/backend/pkg/nats"
)

// EventType represents the type of WhatsApp event
type EventType string

const (
	EventTypeQRCode        EventType = "qr_code"
	EventTypeConnected     EventType = "connected"
	EventTypeDisconnected  EventType = "disconnected"
	EventTypeMessage       EventType = "message"
	EventTypeMessageStatus EventType = "message_status"
)

// Event represents a generic event that can be broadcast
type Event struct {
	Type         EventType   `json:"type"`
	ConnectionID string      `json:"connection_id"`
	Data         interface{} `json:"data"`
}

// EventBroadcaster defines the interface for broadcasting events
type EventBroadcaster interface {
	Broadcast(event Event)
}

// DefaultEventHandler implements EventHandler and broadcasts events
type DefaultEventHandler struct {
	broadcaster     EventBroadcaster
	messageRepo     *MessageRepository
	connectionRepo  *ConnectionRepository
	manager         *Manager
}

// NewDefaultEventHandler creates a new event handler
func NewDefaultEventHandler(broadcaster EventBroadcaster, messageRepo *MessageRepository, connectionRepo *ConnectionRepository) *DefaultEventHandler {
	return &DefaultEventHandler{
		broadcaster:    broadcaster,
		messageRepo:    messageRepo,
		connectionRepo: connectionRepo,
	}
}

// SetManager sets the manager reference for QR code caching
func (h *DefaultEventHandler) SetManager(manager *Manager) {
	h.manager = manager
}

// OnQRCode handles QR code events (wuzapi pattern: store base64 in DB)
func (h *DefaultEventHandler) OnQRCode(event QRCodeEvent) {
	log.Printf("[WhatsApp] QR Code generated for connection %s", event.ConnectionID)

	// Store QR code in manager cache (for backward compatibility)
	if h.manager != nil {
		h.manager.SetQRCode(event.ConnectionID, event.Base64Image)
		log.Printf("[WhatsApp] QR code stored in manager cache")
	} else {
		log.Printf("[WhatsApp] WARNING: manager is nil, cannot store in cache")
	}

	// Store QR code base64 in database for retrieval via API (wuzapi pattern)
	if h.connectionRepo != nil {
		ctx := context.Background()
		if err := h.connectionRepo.SetQRCode(ctx, event.ConnectionID, event.Base64Image); err != nil {
			log.Printf("[WhatsApp] Failed to save QR code to DB: %v", err)
		} else {
			log.Printf("[WhatsApp] QR code saved to DB for connection %s", event.ConnectionID)
		}
	} else {
		log.Printf("[WhatsApp] WARNING: connectionRepo is nil, cannot save to DB")
	}

	h.broadcast(Event{
		Type:         EventTypeQRCode,
		ConnectionID: event.ConnectionID,
		Data: map[string]string{
			"code":         event.Code,
			"base64_image": event.Base64Image,
		},
	})
}

// OnPairSuccess handles successful QR code pairing (wuzapi pattern: save JID to DB)
func (h *DefaultEventHandler) OnPairSuccess(event PairSuccessEvent) {
	log.Printf("[WhatsApp] Pair success for connection %s: JID=%s", event.ConnectionID, event.JID)

	// Clear QR code and save JID to database (wuzapi pattern)
	if h.connectionRepo != nil {
		ctx := context.Background()
		// Clear QR code and set status to connected
		if err := h.connectionRepo.ClearQRCode(ctx, event.ConnectionID, string(StatusConnected)); err != nil {
			log.Printf("[WhatsApp] Failed to clear QR code from DB: %v", err)
		}
		// Save JID to connection record (important for reconnection)
		if conn, err := h.connectionRepo.GetByID(ctx, event.ConnectionID); err == nil && conn != nil {
			conn.JID = event.JID
			conn.Status = StatusConnected
			if err := h.connectionRepo.Update(ctx, conn); err != nil {
				log.Printf("[WhatsApp] Failed to save JID to DB: %v", err)
			} else {
				log.Printf("[WhatsApp] JID saved to DB for connection %s", event.ConnectionID)
			}
		}
	}

	// Clear from manager cache
	if h.manager != nil {
		h.manager.ClearQRCode(event.ConnectionID)
	}

	h.broadcast(Event{
		Type:         EventTypeConnected,
		ConnectionID: event.ConnectionID,
		Data: map[string]interface{}{
			"status":        StatusConnected,
			"jid":           event.JID,
			"business_name": event.BusinessName,
			"platform":      event.Platform,
		},
	})
}

// OnQRTimeout handles QR code timeout events (wuzapi pattern)
func (h *DefaultEventHandler) OnQRTimeout(connectionID string) {
	log.Printf("[WhatsApp] QR Code timeout for connection %s", connectionID)

	// Clear QR code from manager cache
	if h.manager != nil {
		h.manager.ClearQRCode(connectionID)
	}

	// Clear QR code from database and set status to disconnected
	if h.connectionRepo != nil {
		ctx := context.Background()
		if err := h.connectionRepo.ClearQRCode(ctx, connectionID, string(StatusDisconnected)); err != nil {
			log.Printf("[WhatsApp] Failed to clear QR code from DB: %v", err)
		}
	}

	h.broadcast(Event{
		Type:         EventTypeDisconnected,
		ConnectionID: connectionID,
		Data: map[string]interface{}{
			"status": StatusDisconnected,
			"reason": "qr_timeout",
		},
	})
}

// OnConnected handles connection success events (wuzapi pattern: clear QR from DB)
func (h *DefaultEventHandler) OnConnected(event ConnectionEvent) {
	log.Printf("[WhatsApp] Connected: %s (phone: %s)", event.ConnectionID, event.Phone)

	// Clear QR code from cache since we're now connected
	if h.manager != nil {
		h.manager.ClearQRCode(event.ConnectionID)
	}

	// Clear QR code from database and update status to connected (wuzapi pattern)
	if h.connectionRepo != nil {
		ctx := context.Background()
		// First clear the QR code
		if err := h.connectionRepo.ClearQRCode(ctx, event.ConnectionID, string(StatusConnected)); err != nil {
			log.Printf("[WhatsApp] Failed to clear QR code from DB: %v", err)
		}
		// Then update phone number
		if conn, err := h.connectionRepo.GetByID(ctx, event.ConnectionID); err == nil && conn != nil {
			conn.Phone = event.Phone
			h.connectionRepo.Update(ctx, conn)
		}
	}

	h.broadcast(Event{
		Type:         EventTypeConnected,
		ConnectionID: event.ConnectionID,
		Data: map[string]interface{}{
			"status": event.Status,
			"phone":  event.Phone,
		},
	})
}

// OnDisconnected handles disconnection events
func (h *DefaultEventHandler) OnDisconnected(event ConnectionEvent) {
	log.Printf("[WhatsApp] Disconnected: %s", event.ConnectionID)

	// Update connection status in database
	// TODO: Implement context properly
	// h.connectionRepo.UpdateStatus(context.Background(), event.ConnectionID, StatusDisconnected, "")

	h.broadcast(Event{
		Type:         EventTypeDisconnected,
		ConnectionID: event.ConnectionID,
		Data: map[string]interface{}{
			"status": event.Status,
		},
	})
}

// OnMessage handles incoming message events
func (h *DefaultEventHandler) OnMessage(event MessageEvent) {
	if event.Message == nil {
		return
	}

	log.Printf("[WhatsApp] Message received: %s -> %s", event.Message.SenderJID, event.Message.Content[:min(50, len(event.Message.Content))])

	// Save message to database
	// TODO: Implement context properly
	// h.messageRepo.Create(context.Background(), event.Message)

	h.broadcast(Event{
		Type:         EventTypeMessage,
		ConnectionID: event.ConnectionID,
		Data:         event.Message,
	})
}

// OnMessageStatus handles message status update events
func (h *DefaultEventHandler) OnMessageStatus(event MessageEvent) {
	log.Printf("[WhatsApp] Message status update: %s -> %s", event.MessageID, event.Status)

	// Update message status in database
	// TODO: Implement context properly
	// h.messageRepo.UpdateStatus(context.Background(), event.MessageID, event.Status)

	h.broadcast(Event{
		Type:         EventTypeMessageStatus,
		ConnectionID: event.ConnectionID,
		Data: map[string]interface{}{
			"message_id": event.MessageID,
			"status":     event.Status,
		},
	})
}

// broadcast sends an event through the broadcaster if available
func (h *DefaultEventHandler) broadcast(event Event) {
	if h.broadcaster != nil {
		h.broadcaster.Broadcast(event)
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// NATSBroadcaster broadcasts events through NATS JetStream
type NATSBroadcaster struct {
	client *natspkg.Client
}

// NewNATSBroadcaster creates a new NATS broadcaster
func NewNATSBroadcaster(client *natspkg.Client) *NATSBroadcaster {
	return &NATSBroadcaster{
		client: client,
	}
}

// Broadcast sends an event through NATS JetStream
func (b *NATSBroadcaster) Broadcast(event Event) {
	if b.client == nil {
		log.Printf("[NATS] Client not initialized, skipping broadcast")
		return
	}

	ctx := context.Background()

	switch event.Type {
	case EventTypeQRCode:
		if data, ok := event.Data.(map[string]string); ok {
			if code, exists := data["code"]; exists {
				if err := b.client.PublishQRCode(ctx, event.ConnectionID, code); err != nil {
					log.Printf("[NATS] Failed to publish QR code: %v", err)
				}
			}
		}

	case EventTypeConnected, EventTypeDisconnected:
		statusData := &natspkg.ConnectionStatusData{
			Status: string(event.Type),
		}
		if data, ok := event.Data.(map[string]interface{}); ok {
			if phone, exists := data["phone"]; exists {
				statusData.Phone = phone.(string)
			}
		}
		if err := b.client.PublishConnectionStatus(ctx, event.ConnectionID, statusData); err != nil {
			log.Printf("[NATS] Failed to publish connection status: %v", err)
		}

	case EventTypeMessage:
		if msg, ok := event.Data.(*Message); ok {
			msgData := &natspkg.MessageData{
				ID:        msg.ID,
				ChatJID:   msg.ChatJID,
				SenderJID: msg.SenderJID,
				Content:   msg.Content,
				MediaType: msg.MediaType,
				Direction: string(msg.Direction),
				Timestamp: msg.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
			}
			if err := b.client.PublishMessage(ctx, event.ConnectionID, msgData); err != nil {
				log.Printf("[NATS] Failed to publish message: %v", err)
			}
		}

	case EventTypeMessageStatus:
		if data, ok := event.Data.(map[string]interface{}); ok {
			statusData := &natspkg.MessageStatusData{}
			if msgID, exists := data["message_id"]; exists {
				statusData.MessageID = msgID.(string)
			}
			if status, exists := data["status"]; exists {
				statusData.Status = status.(string)
			}
			if err := b.client.PublishMessageStatus(ctx, event.ConnectionID, statusData); err != nil {
				log.Printf("[NATS] Failed to publish message status: %v", err)
			}
		}

	default:
		// Generic publish for unknown event types
		data, err := json.Marshal(event)
		if err != nil {
			log.Printf("[NATS] Failed to marshal event: %v", err)
			return
		}
		subject := "zyntra.events." + event.ConnectionID
		if err := b.client.Conn().Publish(subject, data); err != nil {
			log.Printf("[NATS] Failed to publish event: %v", err)
		}
	}
}

// ChannelBroadcaster broadcasts events through a Go channel (for WebSocket)
type ChannelBroadcaster struct {
	eventChan chan Event
}

// NewChannelBroadcaster creates a new channel broadcaster
func NewChannelBroadcaster(bufferSize int) *ChannelBroadcaster {
	return &ChannelBroadcaster{
		eventChan: make(chan Event, bufferSize),
	}
}

// Broadcast sends an event through the channel
func (b *ChannelBroadcaster) Broadcast(event Event) {
	select {
	case b.eventChan <- event:
	default:
		log.Printf("[Channel] Event channel full, dropping event")
	}
}

// Events returns the event channel for reading
func (b *ChannelBroadcaster) Events() <-chan Event {
	return b.eventChan
}

// Close closes the event channel
func (b *ChannelBroadcaster) Close() {
	close(b.eventChan)
}
