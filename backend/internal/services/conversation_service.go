package services

import (
	"context"
	"fmt"

	"github.com/zyntra/backend/internal/domain"
	"github.com/zyntra/backend/internal/repository"
)

// ConversationService servico de conversas
type ConversationService struct {
	conversationRepo *repository.ConversationRepository
	contactRepo      *repository.ContactRepository
	labelRepo        *repository.LabelRepository
	inboxRepo        *repository.InboxRepository
	messageRepo      *repository.MessageRepository
}

// NewConversationService cria novo servico
func NewConversationService(
	conversationRepo *repository.ConversationRepository,
	contactRepo *repository.ContactRepository,
	labelRepo *repository.LabelRepository,
	inboxRepo *repository.InboxRepository,
	messageRepo *repository.MessageRepository,
) *ConversationService {
	return &ConversationService{
		conversationRepo: conversationRepo,
		contactRepo:      contactRepo,
		labelRepo:        labelRepo,
		inboxRepo:        inboxRepo,
		messageRepo:      messageRepo,
	}
}

// GetByID busca conversa por ID
func (s *ConversationService) GetByID(ctx context.Context, id string) (*domain.Conversation, error) {
	conv, err := s.conversationRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}
	if conv == nil {
		return nil, fmt.Errorf("conversation not found")
	}
	return conv, nil
}

// GetWithDetails busca conversa com detalhes
func (s *ConversationService) GetWithDetails(ctx context.Context, id string) (*domain.ConversationWithDetails, error) {
	conv, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	result := &domain.ConversationWithDetails{Conversation: *conv}

	// Buscar contato
	if contact, _ := s.contactRepo.GetByID(ctx, conv.ContactID); contact != nil {
		result.Contact = contact
	}

	// Buscar inbox
	if inbox, _ := s.inboxRepo.GetByID(ctx, conv.InboxID); inbox != nil {
		result.Inbox = inbox
	}

	return result, nil
}

// List lista conversas com filtros
func (s *ConversationService) List(ctx context.Context, filter domain.ConversationFilter) ([]*domain.Conversation, error) {
	return s.conversationRepo.List(ctx, filter)
}

// ListWithDetails lista conversas com detalhes
func (s *ConversationService) ListWithDetails(ctx context.Context, filter domain.ConversationFilter) ([]*domain.ConversationWithDetails, error) {
	conversations, err := s.conversationRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	var result []*domain.ConversationWithDetails
	for _, conv := range conversations {
		cwd := &domain.ConversationWithDetails{Conversation: *conv}

		// Buscar contato
		if contact, _ := s.contactRepo.GetByID(ctx, conv.ContactID); contact != nil {
			cwd.Contact = contact
		}

		// Buscar ultima mensagem
		if s.messageRepo != nil {
			if msg, _ := s.messageRepo.GetLastByConversation(ctx, conv.ID); msg != nil {
				cwd.LastMessage = msg
			}
		}

		result = append(result, cwd)
	}

	return result, nil
}

// Update atualiza uma conversa
func (s *ConversationService) Update(ctx context.Context, id string, req domain.UpdateConversationRequest) (*domain.Conversation, error) {
	conv, err := s.conversationRepo.GetByID(ctx, id)
	if err != nil || conv == nil {
		return nil, fmt.Errorf("conversation not found")
	}

	if req.Status != nil {
		conv.Status = *req.Status
	}
	if req.Priority != nil {
		conv.Priority = req.Priority
	}
	if req.AssigneeID != nil {
		conv.AssigneeID = req.AssigneeID
	}
	if req.IsFavorite != nil {
		conv.IsFavorite = *req.IsFavorite
	}
	if req.IsArchived != nil {
		conv.IsArchived = *req.IsArchived
	}

	if err := s.conversationRepo.Update(ctx, conv); err != nil {
		return nil, fmt.Errorf("failed to update conversation: %w", err)
	}

	return conv, nil
}

// ToggleStatus alterna status da conversa
func (s *ConversationService) ToggleStatus(ctx context.Context, id string) (*domain.Conversation, error) {
	conv, err := s.conversationRepo.GetByID(ctx, id)
	if err != nil || conv == nil {
		return nil, fmt.Errorf("conversation not found")
	}

	if conv.Status == domain.ConversationStatusOpen {
		conv.Status = domain.ConversationStatusResolved
	} else {
		conv.Status = domain.ConversationStatusOpen
	}

	if err := s.conversationRepo.Update(ctx, conv); err != nil {
		return nil, fmt.Errorf("failed to update conversation: %w", err)
	}

	return conv, nil
}

// SetFavorite define favorito
func (s *ConversationService) SetFavorite(ctx context.Context, id string, favorite bool) error {
	conv, err := s.conversationRepo.GetByID(ctx, id)
	if err != nil || conv == nil {
		return fmt.Errorf("conversation not found")
	}

	conv.IsFavorite = favorite
	return s.conversationRepo.Update(ctx, conv)
}

// SetArchived define arquivado
func (s *ConversationService) SetArchived(ctx context.Context, id string, archived bool) error {
	conv, err := s.conversationRepo.GetByID(ctx, id)
	if err != nil || conv == nil {
		return fmt.Errorf("conversation not found")
	}

	conv.IsArchived = archived
	return s.conversationRepo.Update(ctx, conv)
}

// Assign atribui conversa a um agente
func (s *ConversationService) Assign(ctx context.Context, id, assigneeID string) error {
	conv, err := s.conversationRepo.GetByID(ctx, id)
	if err != nil || conv == nil {
		return fmt.Errorf("conversation not found")
	}

	conv.AssigneeID = &assigneeID
	return s.conversationRepo.Update(ctx, conv)
}

// Unassign remove atribuicao
func (s *ConversationService) Unassign(ctx context.Context, id string) error {
	conv, err := s.conversationRepo.GetByID(ctx, id)
	if err != nil || conv == nil {
		return fmt.Errorf("conversation not found")
	}

	conv.AssigneeID = nil
	return s.conversationRepo.Update(ctx, conv)
}

// MarkAsRead marca como lida
func (s *ConversationService) MarkAsRead(ctx context.Context, id string) error {
	return s.conversationRepo.ResetUnread(ctx, id)
}

// AddLabel adiciona label
func (s *ConversationService) AddLabel(ctx context.Context, conversationID, labelID string) error {
	return s.labelRepo.AddToConversation(ctx, conversationID, labelID)
}

// RemoveLabel remove label
func (s *ConversationService) RemoveLabel(ctx context.Context, conversationID, labelID string) error {
	return s.labelRepo.RemoveFromConversation(ctx, conversationID, labelID)
}

// GetLabels lista labels da conversa
func (s *ConversationService) GetLabels(ctx context.Context, conversationID string) ([]*domain.Label, error) {
	return s.labelRepo.GetConversationLabels(ctx, conversationID)
}

// Delete remove conversa
func (s *ConversationService) Delete(ctx context.Context, id string) error {
	return s.conversationRepo.Delete(ctx, id)
}
