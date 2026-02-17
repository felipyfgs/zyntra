package services

import (
	"github.com/zyntra/backend/internal/domain"
)

// BroadcastHub interface para o hub de broadcast (implementado por handlers.WebSocketHub)
type BroadcastHub interface {
	BroadcastMessage(inboxID string, message interface{})
	BroadcastConversationUpdate(inboxID string, conversation interface{})
}

// WebSocketBroadcaster implementa EventBroadcaster usando BroadcastHub
type WebSocketBroadcaster struct {
	hub BroadcastHub
}

// NewWebSocketBroadcaster cria novo broadcaster
func NewWebSocketBroadcaster(hub BroadcastHub) *WebSocketBroadcaster {
	return &WebSocketBroadcaster{hub: hub}
}

// BroadcastMessage envia mensagem via WebSocket
func (b *WebSocketBroadcaster) BroadcastMessage(inboxID string, msg *domain.Message) {
	if b.hub == nil {
		return
	}
	b.hub.BroadcastMessage(inboxID, msg)
}

// BroadcastConversationUpdate envia atualizacao de conversa via WebSocket
func (b *WebSocketBroadcaster) BroadcastConversationUpdate(inboxID string, conv *domain.Conversation) {
	if b.hub == nil {
		return
	}
	b.hub.BroadcastConversationUpdate(inboxID, conv)
}

// Verify interface implementation
var _ EventBroadcaster = (*WebSocketBroadcaster)(nil)
