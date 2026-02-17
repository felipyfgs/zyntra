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

// InboxService servico de inboxes
type InboxService struct {
	inboxRepo      *repository.InboxRepository
	waChannelRepo  *repository.ChannelWhatsAppRepository
	memberRepo     *repository.InboxMemberRepository
	waManager      *whatsapp.Manager
}

// NewInboxService cria novo servico
func NewInboxService(
	inboxRepo *repository.InboxRepository,
	waChannelRepo *repository.ChannelWhatsAppRepository,
	memberRepo *repository.InboxMemberRepository,
	waManager *whatsapp.Manager,
) *InboxService {
	return &InboxService{
		inboxRepo:     inboxRepo,
		waChannelRepo: waChannelRepo,
		memberRepo:    memberRepo,
		waManager:     waManager,
	}
}

// Create cria um inbox
func (s *InboxService) Create(ctx context.Context, req domain.CreateInboxRequest) (*domain.Inbox, error) {
	channelID := uuid.New().String()

	// Criar canal especifico
	switch req.ChannelType {
	case ports.ChannelTypeWhatsApp:
		channel := &domain.ChannelWhatsApp{
			ID:       channelID,
			Provider: "whatsmeow",
		}
		if err := s.waChannelRepo.Create(ctx, channel); err != nil {
			return nil, fmt.Errorf("failed to create whatsapp channel: %w", err)
		}

	case ports.ChannelTypeTelegram:
		// TODO: implementar canal telegram
		return nil, fmt.Errorf("telegram channel not implemented yet")

	case ports.ChannelTypeAPI:
		// TODO: implementar canal API
		return nil, fmt.Errorf("api channel not implemented yet")

	default:
		return nil, fmt.Errorf("unknown channel type: %s", req.ChannelType)
	}

	// Criar inbox
	inbox := &domain.Inbox{
		ID:              uuid.New().String(),
		Name:            req.Name,
		ChannelType:     req.ChannelType,
		ChannelID:       channelID,
		Status:          ports.ChannelStatusDisconnected,
		GreetingMessage: req.GreetingMessage,
		AutoAssignment:  req.AutoAssignment,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.inboxRepo.Create(ctx, inbox); err != nil {
		return nil, fmt.Errorf("failed to create inbox: %w", err)
	}

	return inbox, nil
}

// GetByID busca inbox por ID
func (s *InboxService) GetByID(ctx context.Context, id string) (*domain.Inbox, error) {
	inbox, err := s.inboxRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get inbox: %w", err)
	}
	if inbox == nil {
		return nil, fmt.Errorf("inbox not found")
	}

	// Atualizar status do manager
	if inbox.ChannelType == ports.ChannelTypeWhatsApp && s.waManager != nil {
		inbox.Status = s.waManager.Status(id)
	}

	return inbox, nil
}

// GetAll lista todos os inboxes
func (s *InboxService) GetAll(ctx context.Context) ([]*domain.Inbox, error) {
	inboxes, err := s.inboxRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list inboxes: %w", err)
	}

	// Atualizar status do manager
	for _, inbox := range inboxes {
		if inbox.ChannelType == ports.ChannelTypeWhatsApp && s.waManager != nil {
			inbox.Status = s.waManager.Status(inbox.ID)
		}
	}

	return inboxes, nil
}

// Connect conecta um inbox
func (s *InboxService) Connect(ctx context.Context, inboxID string) error {
	inbox, err := s.inboxRepo.GetByID(ctx, inboxID)
	if err != nil {
		return fmt.Errorf("failed to get inbox: %w", err)
	}
	if inbox == nil {
		return fmt.Errorf("inbox not found")
	}

	switch inbox.ChannelType {
	case ports.ChannelTypeWhatsApp:
		if s.waManager == nil {
			return fmt.Errorf("whatsapp manager not initialized")
		}

		// Buscar JID existente
		channel, _ := s.waChannelRepo.GetByID(ctx, inbox.ChannelID)
		jid := ""
		if channel != nil {
			jid = channel.JID
		}

		// Atualizar status
		if err := s.inboxRepo.UpdateStatus(ctx, inboxID, ports.ChannelStatusConnecting); err != nil {
			log.Printf("Failed to update inbox status: %v", err)
		}

		// Conectar
		if err := s.waManager.Connect(ctx, inboxID, jid); err != nil {
			s.inboxRepo.UpdateStatus(ctx, inboxID, ports.ChannelStatusDisconnected)
			return fmt.Errorf("failed to connect: %w", err)
		}

	default:
		return fmt.Errorf("channel type %s not supported", inbox.ChannelType)
	}

	return nil
}

// Disconnect desconecta um inbox
func (s *InboxService) Disconnect(ctx context.Context, inboxID string) error {
	inbox, err := s.inboxRepo.GetByID(ctx, inboxID)
	if err != nil {
		return fmt.Errorf("failed to get inbox: %w", err)
	}
	if inbox == nil {
		return fmt.Errorf("inbox not found")
	}

	switch inbox.ChannelType {
	case ports.ChannelTypeWhatsApp:
		if s.waManager != nil {
			if err := s.waManager.Disconnect(ctx, inboxID); err != nil {
				return fmt.Errorf("failed to disconnect: %w", err)
			}
		}
	}

	if err := s.inboxRepo.UpdateStatus(ctx, inboxID, ports.ChannelStatusDisconnected); err != nil {
		log.Printf("Failed to update inbox status: %v", err)
	}

	return nil
}

// Delete remove um inbox
func (s *InboxService) Delete(ctx context.Context, inboxID string) error {
	inbox, err := s.inboxRepo.GetByID(ctx, inboxID)
	if err != nil {
		return fmt.Errorf("failed to get inbox: %w", err)
	}
	if inbox == nil {
		return fmt.Errorf("inbox not found")
	}

	// Desconectar e remover do manager
	switch inbox.ChannelType {
	case ports.ChannelTypeWhatsApp:
		if s.waManager != nil {
			s.waManager.Remove(ctx, inboxID)
		}
		s.waChannelRepo.Delete(ctx, inbox.ChannelID)
	}

	// Remover inbox
	if err := s.inboxRepo.Delete(ctx, inboxID); err != nil {
		return fmt.Errorf("failed to delete inbox: %w", err)
	}

	return nil
}

// GetStatus retorna status do inbox
func (s *InboxService) GetStatus(inboxID string) ports.ChannelStatus {
	if s.waManager != nil {
		return s.waManager.Status(inboxID)
	}
	return ports.ChannelStatusDisconnected
}

// GetQRCode retorna QR code do inbox
func (s *InboxService) GetQRCode(ctx context.Context, inboxID string) string {
	// Primeiro verifica no banco
	inbox, _ := s.inboxRepo.GetByID(ctx, inboxID)
	if inbox != nil && inbox.QRCode != "" {
		return inbox.QRCode
	}

	// Fallback para o manager
	if s.waManager != nil {
		return s.waManager.GetQRCode(inboxID)
	}
	return ""
}

// AddMember adiciona membro ao inbox
func (s *InboxService) AddMember(ctx context.Context, inboxID, userID string) error {
	return s.memberRepo.Add(ctx, inboxID, userID)
}

// RemoveMember remove membro do inbox
func (s *InboxService) RemoveMember(ctx context.Context, inboxID, userID string) error {
	return s.memberRepo.Remove(ctx, inboxID, userID)
}

// GetMembers lista membros do inbox
func (s *InboxService) GetMembers(ctx context.Context, inboxID string) ([]string, error) {
	return s.memberRepo.GetByInboxID(ctx, inboxID)
}

// OnConnected chamado quando canal conecta
func (s *InboxService) OnConnected(ctx context.Context, inboxID, phone string) error {
	if err := s.inboxRepo.ClearQRCode(ctx, inboxID, ports.ChannelStatusConnected); err != nil {
		return err
	}

	// Atualizar canal com telefone
	inbox, _ := s.inboxRepo.GetByID(ctx, inboxID)
	if inbox != nil && inbox.ChannelType == ports.ChannelTypeWhatsApp {
		adapter, _ := s.waManager.Get(inboxID)
		if adapter != nil {
			jid := adapter.GetJID()
			s.waChannelRepo.UpdateJID(ctx, inbox.ChannelID, jid, phone)
		}
	}

	return nil
}

// OnDisconnected chamado quando canal desconecta
func (s *InboxService) OnDisconnected(ctx context.Context, inboxID string) error {
	return s.inboxRepo.UpdateStatus(ctx, inboxID, ports.ChannelStatusDisconnected)
}

// OnQRCode chamado quando QR code e gerado
func (s *InboxService) OnQRCode(ctx context.Context, inboxID, qrCode, base64Image string) error {
	return s.inboxRepo.SetQRCode(ctx, inboxID, base64Image)
}

// RestoreConnections restaura conexoes WhatsApp
func (s *InboxService) RestoreConnections(ctx context.Context) error {
	if s.waManager == nil {
		return nil
	}

	items, err := s.inboxRepo.GetAllWhatsAppForRestore(ctx)
	if err != nil {
		return fmt.Errorf("failed to get inboxes for restore: %w", err)
	}

	var inboxes []whatsapp.InboxInfo
	for _, item := range items {
		inboxes = append(inboxes, whatsapp.InboxInfo{
			ID:  item.ID,
			JID: item.JID,
		})
	}

	return s.waManager.RestoreConnections(ctx, inboxes)
}
