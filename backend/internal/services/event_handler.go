package services

import (
	"context"
	"log"

	"github.com/zyntra/backend/internal/ports"
)

// ChannelEventHandler processa eventos dos canais
type ChannelEventHandler struct {
	inboxService   *InboxService
	messageService *MessageService
}

// NewChannelEventHandler cria novo handler
func NewChannelEventHandler(inboxService *InboxService, messageService *MessageService) *ChannelEventHandler {
	return &ChannelEventHandler{
		inboxService:   inboxService,
		messageService: messageService,
	}
}

// OnMessage processa mensagem recebida
func (h *ChannelEventHandler) OnMessage(event ports.IncomingEvent) {
	log.Printf("[EventHandler] Message received for inbox %s from %s", event.InboxID, event.ContactID)

	ctx := context.Background()
	if err := h.messageService.ProcessIncomingMessage(ctx, event); err != nil {
		log.Printf("[EventHandler] Failed to process message: %v", err)
	}
}

// OnStatusUpdate processa atualizacao de status
func (h *ChannelEventHandler) OnStatusUpdate(inboxID, sourceID string, status ports.MessageStatus) {
	log.Printf("[EventHandler] Status update for inbox %s: %s -> %s", inboxID, sourceID, status)

	ctx := context.Background()
	if err := h.messageService.ProcessStatusUpdate(ctx, inboxID, sourceID, status); err != nil {
		log.Printf("[EventHandler] Failed to update status: %v", err)
	}
}

// OnQRCode processa QR code gerado
func (h *ChannelEventHandler) OnQRCode(inboxID, qrCode, base64Image string) {
	log.Printf("[EventHandler] QR code for inbox %s", inboxID)

	ctx := context.Background()
	if err := h.inboxService.OnQRCode(ctx, inboxID, qrCode, base64Image); err != nil {
		log.Printf("[EventHandler] Failed to save QR code: %v", err)
	}
}

// OnConnected processa conexao estabelecida
func (h *ChannelEventHandler) OnConnected(inboxID, phone string) {
	log.Printf("[EventHandler] Connected inbox %s with phone %s", inboxID, phone)

	ctx := context.Background()
	if err := h.inboxService.OnConnected(ctx, inboxID, phone); err != nil {
		log.Printf("[EventHandler] Failed to handle connection: %v", err)
	}
}

// OnDisconnected processa desconexao
func (h *ChannelEventHandler) OnDisconnected(inboxID string) {
	log.Printf("[EventHandler] Disconnected inbox %s", inboxID)

	ctx := context.Background()
	if err := h.inboxService.OnDisconnected(ctx, inboxID); err != nil {
		log.Printf("[EventHandler] Failed to handle disconnection: %v", err)
	}
}

// Verify interface implementation
var _ ports.ChannelEventHandler = (*ChannelEventHandler)(nil)
