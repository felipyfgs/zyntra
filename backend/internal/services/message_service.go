package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/zyntra/backend/internal/domain"
	"github.com/zyntra/backend/internal/ports"
	"github.com/zyntra/backend/internal/repository"
	"github.com/zyntra/backend/internal/channels/whatsapp"
)

// MessageService servico de mensagens
type MessageService struct {
	messageRepo      *repository.MessageRepository
	conversationRepo *repository.ConversationRepository
	contactRepo      *repository.ContactRepository
	contactInboxRepo *repository.ContactInboxRepository
	inboxRepo        *repository.InboxRepository
	waManager        *whatsapp.Manager
	broadcaster      EventBroadcaster
}

// EventBroadcaster interface para broadcast de eventos
type EventBroadcaster interface {
	BroadcastMessage(inboxID string, msg *domain.Message)
	BroadcastConversationUpdate(inboxID string, conv *domain.Conversation)
}

// NewMessageService cria novo servico
func NewMessageService(
	messageRepo *repository.MessageRepository,
	conversationRepo *repository.ConversationRepository,
	contactRepo *repository.ContactRepository,
	contactInboxRepo *repository.ContactInboxRepository,
	inboxRepo *repository.InboxRepository,
	waManager *whatsapp.Manager,
) *MessageService {
	return &MessageService{
		messageRepo:      messageRepo,
		conversationRepo: conversationRepo,
		contactRepo:      contactRepo,
		contactInboxRepo: contactInboxRepo,
		inboxRepo:        inboxRepo,
		waManager:        waManager,
	}
}

// SetBroadcaster define o broadcaster de eventos
func (s *MessageService) SetBroadcaster(b EventBroadcaster) {
	s.broadcaster = b
}

// SendMessage envia uma mensagem
func (s *MessageService) SendMessage(ctx context.Context, conversationID string, req domain.SendMessageRequest, senderID string) (*domain.Message, error) {
	// Buscar conversa
	conv, err := s.conversationRepo.GetByID(ctx, conversationID)
	if err != nil || conv == nil {
		return nil, fmt.Errorf("conversation not found")
	}

	// Buscar contact_inbox para obter source_id (telefone/JID)
	contactInbox, err := s.contactInboxRepo.GetByID(ctx, conv.ContactInboxID)
	if err != nil || contactInbox == nil {
		return nil, fmt.Errorf("contact inbox not found")
	}

	// Buscar inbox
	inbox, err := s.inboxRepo.GetByID(ctx, conv.InboxID)
	if err != nil || inbox == nil {
		return nil, fmt.Errorf("inbox not found")
	}

	var sourceID string

	// Enviar pelo canal apropriado
	switch inbox.ChannelType {
	case ports.ChannelTypeWhatsApp:
		if s.waManager == nil {
			return nil, fmt.Errorf("whatsapp manager not initialized")
		}
		sourceID, err = s.waManager.SendText(ctx, inbox.ID, contactInbox.SourceID, req.Content)
		if err != nil {
			return nil, fmt.Errorf("failed to send message: %w", err)
		}

	default:
		return nil, fmt.Errorf("channel type %s not supported for sending", inbox.ChannelType)
	}

	// Criar mensagem no banco
	msg := &domain.Message{
		ID:             uuid.New().String(),
		ConversationID: conversationID,
		InboxID:        inbox.ID,
		SenderType:     domain.SenderTypeUser,
		SenderID:       &senderID,
		Content:        req.Content,
		ContentType:    req.ContentType,
		SourceID:       sourceID,
		Status:         ports.MessageStatusSent,
		Private:        req.Private,
		CreatedAt:      time.Now(),
	}

	if msg.ContentType == "" {
		msg.ContentType = domain.ContentTypeText
	}

	if err := s.messageRepo.Create(ctx, msg); err != nil {
		log.Printf("Failed to save message: %v", err)
	}

	// Atualizar conversa
	now := time.Now()
	conv.LastMessageAt = &now
	s.conversationRepo.UpdateLastMessage(ctx, conversationID, now)

	// Broadcast
	if s.broadcaster != nil {
		s.broadcaster.BroadcastMessage(inbox.ID, msg)
	}

	return msg, nil
}

// ProcessIncomingMessage processa mensagem recebida do canal
func (s *MessageService) ProcessIncomingMessage(ctx context.Context, event ports.IncomingEvent) error {
	log.Printf("[MessageService] Processing incoming message for inbox %s from %s", event.InboxID, event.ContactID)

	// 1. Buscar ou criar contato
	contact, err := s.findOrCreateContact(ctx, event)
	if err != nil {
		return fmt.Errorf("failed to find/create contact: %w", err)
	}

	// 2. Buscar ou criar contact_inbox
	contactInbox, err := s.findOrCreateContactInbox(ctx, event.InboxID, contact.ID, event.ContactID)
	if err != nil {
		return fmt.Errorf("failed to find/create contact_inbox: %w", err)
	}

	// 3. Buscar ou criar conversa
	conv, err := s.findOrCreateConversation(ctx, event.InboxID, contact.ID, contactInbox.ID)
	if err != nil {
		return fmt.Errorf("failed to find/create conversation: %w", err)
	}

	// 4. Criar mensagem
	senderType := domain.SenderTypeContact
	if event.IsFromMe {
		senderType = domain.SenderTypeUser
	}

	contentType := domain.ContentTypeText
	if event.MediaType != "" {
		contentType = domain.ContentType(event.MediaType)
	}

	msg := &domain.Message{
		ID:             uuid.New().String(),
		ConversationID: conv.ID,
		InboxID:        event.InboxID,
		SenderType:     senderType,
		SenderID:       &contact.ID,
		Content:        event.Content,
		ContentType:    contentType,
		SourceID:       event.SourceID,
		Status:         ports.MessageStatusDelivered,
		CreatedAt:      event.Timestamp,
	}

	if err := s.messageRepo.Create(ctx, msg); err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	// 5. Atualizar conversa
	conv.LastMessageAt = &event.Timestamp
	if !event.IsFromMe {
		conv.UnreadCount++
	}
	s.conversationRepo.Update(ctx, conv)

	// 6. Broadcast
	if s.broadcaster != nil {
		s.broadcaster.BroadcastMessage(event.InboxID, msg)
		s.broadcaster.BroadcastConversationUpdate(event.InboxID, conv)
	}

	log.Printf("[MessageService] Message saved: %s", msg.ID)
	return nil
}

// ProcessStatusUpdate processa atualizacao de status
func (s *MessageService) ProcessStatusUpdate(ctx context.Context, inboxID, sourceID string, status ports.MessageStatus) error {
	return s.messageRepo.UpdateStatusBySourceID(ctx, inboxID, sourceID, status)
}

// GetMessages lista mensagens de uma conversa
func (s *MessageService) GetMessages(ctx context.Context, conversationID string, limit, offset int) ([]*domain.Message, error) {
	return s.messageRepo.ListByConversation(ctx, conversationID, limit, offset)
}

// MarkAsRead marca conversa como lida
func (s *MessageService) MarkAsRead(ctx context.Context, conversationID string) error {
	return s.conversationRepo.ResetUnread(ctx, conversationID)
}

func (s *MessageService) findOrCreateContact(ctx context.Context, event ports.IncomingEvent) (*domain.Contact, error) {
	// Extrair telefone do source_id (JID)
	phone := extractPhoneFromSourceID(event.ContactID)

	// Buscar por telefone
	if phone != "" {
		contact, err := s.contactRepo.GetByPhone(ctx, phone)
		if err == nil && contact != nil {
			// Atualizar nome se mudou
			if event.ContactName != "" && contact.Name != event.ContactName {
				contact.Name = event.ContactName
				s.contactRepo.Update(ctx, contact)
			}
			return contact, nil
		}
	}

	// Criar novo contato
	name := event.ContactName
	if name == "" {
		name = phone
	}

	contact := &domain.Contact{
		ID:          uuid.New().String(),
		Name:        name,
		PhoneNumber: phone,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.contactRepo.Create(ctx, contact); err != nil {
		return nil, err
	}

	return contact, nil
}

func (s *MessageService) findOrCreateContactInbox(ctx context.Context, inboxID, contactID, sourceID string) (*domain.ContactInbox, error) {
	ci, err := s.contactInboxRepo.GetBySourceID(ctx, inboxID, sourceID)
	if err == nil && ci != nil {
		return ci, nil
	}

	ci = &domain.ContactInbox{
		ID:        uuid.New().String(),
		ContactID: contactID,
		InboxID:   inboxID,
		SourceID:  sourceID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.contactInboxRepo.Create(ctx, ci); err != nil {
		return nil, err
	}

	return ci, nil
}

func (s *MessageService) findOrCreateConversation(ctx context.Context, inboxID, contactID, contactInboxID string) (*domain.Conversation, error) {
	conv, err := s.conversationRepo.GetByContactInboxID(ctx, contactInboxID)
	if err == nil && conv != nil {
		// Reabrir se estava resolvida
		if conv.Status == domain.ConversationStatusResolved {
			conv.Status = domain.ConversationStatusOpen
			s.conversationRepo.Update(ctx, conv)
		}
		return conv, nil
	}

	conv = &domain.Conversation{
		ID:             uuid.New().String(),
		InboxID:        inboxID,
		ContactID:      contactID,
		ContactInboxID: contactInboxID,
		Status:         domain.ConversationStatusOpen,
		UnreadCount:    0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.conversationRepo.Create(ctx, conv); err != nil {
		return nil, err
	}

	return conv, nil
}

func extractPhoneFromSourceID(sourceID string) string {
	// sourceID pode ser JID (5511999999999@s.whatsapp.net)
	// ou telefone direto
	if len(sourceID) > 0 {
		// Remover sufixo @s.whatsapp.net ou similar
		phone := sourceID
		if idx := len(phone) - 1; idx > 0 {
			for i := 0; i < len(phone); i++ {
				if phone[i] == '@' {
					phone = phone[:i]
					break
				}
			}
		}
		if phone[0] != '+' && len(phone) > 5 {
			phone = "+" + phone
		}
		return phone
	}
	return ""
}
