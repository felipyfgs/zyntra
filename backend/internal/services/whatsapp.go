package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/zyntra/backend/pkg/whatsapp"
)

// WhatsAppService handles WhatsApp business logic
type WhatsAppService struct {
	manager        *whatsapp.Manager
	connectionRepo *whatsapp.ConnectionRepository
	messageRepo    *whatsapp.MessageRepository
}

// NewWhatsAppService creates a new WhatsApp service
func NewWhatsAppService(manager *whatsapp.Manager, connectionRepo *whatsapp.ConnectionRepository, messageRepo *whatsapp.MessageRepository) *WhatsAppService {
	return &WhatsAppService{
		manager:        manager,
		connectionRepo: connectionRepo,
		messageRepo:    messageRepo,
	}
}

// CreateConnection creates a new WhatsApp connection record
func (s *WhatsAppService) CreateConnection(ctx context.Context, userID, name string) (*whatsapp.Connection, error) {
	conn := &whatsapp.Connection{
		ID:        uuid.New().String(),
		UserID:    userID,
		Name:      name,
		Status:    whatsapp.StatusDisconnected,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.connectionRepo.Create(ctx, conn); err != nil {
		return nil, fmt.Errorf("failed to create connection: %w", err)
	}

	return conn, nil
}

// GetConnection retrieves a connection by ID
func (s *WhatsAppService) GetConnection(ctx context.Context, id string) (*whatsapp.Connection, error) {
	conn, err := s.connectionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}
	if conn == nil {
		return nil, fmt.Errorf("connection not found")
	}

	// Update status from manager
	conn.Status = s.manager.GetStatus(id)
	
	return conn, nil
}

// GetUserConnections retrieves all connections for a user
func (s *WhatsAppService) GetUserConnections(ctx context.Context, userID string) ([]*whatsapp.Connection, error) {
	connections, err := s.connectionRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get connections: %w", err)
	}

	// Update status from manager for each connection
	for _, conn := range connections {
		conn.Status = s.manager.GetStatus(conn.ID)
	}

	return connections, nil
}

// Connect initiates a WhatsApp connection (starts QR code process - wuzapi pattern)
func (s *WhatsAppService) Connect(ctx context.Context, connectionID string) error {
	// Verify connection exists
	conn, err := s.connectionRepo.GetByID(ctx, connectionID)
	if err != nil {
		return fmt.Errorf("failed to get connection: %w", err)
	}
	if conn == nil {
		return fmt.Errorf("connection not found")
	}

	// Update status to connecting
	conn.Status = whatsapp.StatusConnecting
	conn.UpdatedAt = time.Now()
	if err := s.connectionRepo.Update(ctx, conn); err != nil {
		return fmt.Errorf("failed to update connection: %w", err)
	}

	// Connect via manager - pass JID if available (for reconnection)
	if err := s.manager.Connect(ctx, connectionID, conn.JID); err != nil {
		// Revert status on failure
		conn.Status = whatsapp.StatusDisconnected
		conn.UpdatedAt = time.Now()
		s.connectionRepo.Update(ctx, conn)
		return fmt.Errorf("failed to connect: %w", err)
	}

	return nil
}

// Disconnect disconnects a WhatsApp connection
func (s *WhatsAppService) Disconnect(ctx context.Context, connectionID string) error {
	if err := s.manager.Disconnect(connectionID); err != nil {
		return fmt.Errorf("failed to disconnect: %w", err)
	}

	// Update status in database
	conn, _ := s.connectionRepo.GetByID(ctx, connectionID)
	if conn != nil {
		conn.Status = whatsapp.StatusDisconnected
		conn.UpdatedAt = time.Now()
		s.connectionRepo.Update(ctx, conn)
	}

	return nil
}

// DeleteConnection removes a WhatsApp connection
func (s *WhatsAppService) DeleteConnection(ctx context.Context, connectionID string) error {
	// Remove from manager (this will logout and disconnect)
	if err := s.manager.Remove(ctx, connectionID); err != nil {
		// Log but continue with database deletion
		fmt.Printf("Warning: failed to remove from manager: %v\n", err)
	}

	// Delete from database
	if err := s.connectionRepo.Delete(ctx, connectionID); err != nil {
		return fmt.Errorf("failed to delete connection: %w", err)
	}

	return nil
}

// SendMessage sends a message through a connection
func (s *WhatsAppService) SendMessage(ctx context.Context, connectionID, to, content string) (*whatsapp.Message, error) {
	// Send via manager
	msg, err := s.manager.SendMessage(ctx, connectionID, to, content)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	// Save to database
	if err := s.messageRepo.Create(ctx, msg); err != nil {
		// Log but don't fail - message was sent
		fmt.Printf("Warning: failed to save message to database: %v\n", err)
	}

	return msg, nil
}

// GetMessages retrieves messages for a chat
func (s *WhatsAppService) GetMessages(ctx context.Context, connectionID, chatJID string, limit, offset int) ([]*whatsapp.Message, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	messages, err := s.messageRepo.GetByChatJID(ctx, connectionID, chatJID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	return messages, nil
}

// GetConnectionStatus returns the current status of a connection
func (s *WhatsAppService) GetConnectionStatus(connectionID string) whatsapp.ConnectionStatus {
	return s.manager.GetStatus(connectionID)
}

// GetQRCode returns the current QR code for a connection (from database - wuzapi pattern)
func (s *WhatsAppService) GetQRCode(ctx context.Context, connectionID string) string {
	// First try database (persistent storage - wuzapi pattern)
	if qr, err := s.connectionRepo.GetQRCode(ctx, connectionID); err == nil && qr != "" {
		return qr
	}
	// Fallback to manager cache
	return s.manager.GetQRCode(connectionID)
}

// IsConnected checks if a connection is active
func (s *WhatsAppService) IsConnected(connectionID string) bool {
	return s.manager.IsConnected(connectionID)
}

// OnConnectionUpdate handles connection status updates (called from event handler)
func (s *WhatsAppService) OnConnectionUpdate(ctx context.Context, connectionID string, status whatsapp.ConnectionStatus, phone string) error {
	conn, err := s.connectionRepo.GetByID(ctx, connectionID)
	if err != nil || conn == nil {
		return fmt.Errorf("connection not found")
	}

	conn.Status = status
	conn.UpdatedAt = time.Now()
	
	if phone != "" {
		conn.Phone = phone
	}

	return s.connectionRepo.Update(ctx, conn)
}
