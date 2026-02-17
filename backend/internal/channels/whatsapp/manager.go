package whatsapp

import (
	"context"
	"fmt"
	"log"
	"sync"

	wapkg "github.com/zyntra/backend/pkg/whatsapp"
	"github.com/zyntra/backend/internal/ports"
)

// Manager gerencia multiplas conexoes WhatsApp
type Manager struct {
	store    *wapkg.Store
	adapters map[string]*Adapter
	handler  ports.ChannelEventHandler
	mu       sync.RWMutex
}

// NewManager cria novo manager
func NewManager(store *wapkg.Store) *Manager {
	return &Manager{
		store:    store,
		adapters: make(map[string]*Adapter),
	}
}

// SetEventHandler define handler global de eventos
func (m *Manager) SetEventHandler(handler ports.ChannelEventHandler) {
	m.handler = handler
}

// Connect conecta um inbox
func (m *Manager) Connect(ctx context.Context, inboxID, jid string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if adapter, exists := m.adapters[inboxID]; exists {
		if adapter.Status() == ports.ChannelStatusConnected {
			return fmt.Errorf("inbox %s already connected", inboxID)
		}
	}

	// Obter device do store
	device, err := m.store.GetDevice(ctx, jid)
	if err != nil {
		return fmt.Errorf("failed to get device: %w", err)
	}

	// Criar cliente e adapter
	client := wapkg.NewClient(device)
	adapter := NewAdapter(client, inboxID)
	adapter.SetEventHandler(m.handler)
	m.adapters[inboxID] = adapter

	// Conectar
	return adapter.Connect(ctx, inboxID)
}

// Disconnect desconecta um inbox
func (m *Manager) Disconnect(ctx context.Context, inboxID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	adapter, exists := m.adapters[inboxID]
	if !exists {
		return nil
	}

	if err := adapter.Disconnect(ctx); err != nil {
		return err
	}

	delete(m.adapters, inboxID)
	return nil
}

// Remove remove completamente um inbox (com logout)
func (m *Manager) Remove(ctx context.Context, inboxID string) error {
	m.mu.Lock()
	adapter, exists := m.adapters[inboxID]
	if !exists {
		m.mu.Unlock()
		return nil
	}
	m.mu.Unlock()

	// Fazer logout se conectado
	if adapter.client.IsConnected() {
		if err := adapter.Logout(ctx); err != nil {
			log.Printf("[WhatsApp] Failed to logout %s: %v", inboxID, err)
		}
	}

	return m.Disconnect(ctx, inboxID)
}

// Get retorna adapter de um inbox
func (m *Manager) Get(inboxID string) (*Adapter, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	adapter, exists := m.adapters[inboxID]
	if !exists {
		return nil, fmt.Errorf("inbox %s not found", inboxID)
	}
	return adapter, nil
}

// GetConnected retorna adapter apenas se conectado
func (m *Manager) GetConnected(inboxID string) (*Adapter, error) {
	adapter, err := m.Get(inboxID)
	if err != nil {
		return nil, err
	}
	if adapter.Status() != ports.ChannelStatusConnected {
		return nil, fmt.Errorf("inbox %s not connected", inboxID)
	}
	return adapter, nil
}

// Status retorna status de um inbox
func (m *Manager) Status(inboxID string) ports.ChannelStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	adapter, exists := m.adapters[inboxID]
	if !exists {
		return ports.ChannelStatusDisconnected
	}
	return adapter.Status()
}

// GetQRCode retorna QR code de um inbox
func (m *Manager) GetQRCode(inboxID string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	adapter, exists := m.adapters[inboxID]
	if !exists {
		return ""
	}
	return adapter.GetQRCode()
}

// GetInfo retorna informacoes de um inbox
func (m *Manager) GetInfo(inboxID string) map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	adapter, exists := m.adapters[inboxID]
	if !exists {
		return map[string]interface{}{
			"inbox_id": inboxID,
			"status":   ports.ChannelStatusDisconnected,
		}
	}

	return map[string]interface{}{
		"inbox_id": inboxID,
		"status":   adapter.Status(),
		"jid":      adapter.GetJID(),
		"phone":    adapter.GetPhone(),
	}
}

// ListConnected lista todos os inboxes conectados
func (m *Manager) ListConnected() []map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var list []map[string]interface{}
	for id := range m.adapters {
		list = append(list, m.GetInfo(id))
	}
	return list
}

// SendText envia mensagem de texto
func (m *Manager) SendText(ctx context.Context, inboxID, to, content string) (string, error) {
	adapter, err := m.GetConnected(inboxID)
	if err != nil {
		return "", err
	}
	return adapter.SendText(ctx, to, content)
}

// SendMedia envia mensagem com midia
func (m *Manager) SendMedia(ctx context.Context, inboxID, to string, media ports.Media) (string, error) {
	adapter, err := m.GetConnected(inboxID)
	if err != nil {
		return "", err
	}
	return adapter.SendMedia(ctx, to, media)
}

// RestoreConnections restaura conexoes salvas
func (m *Manager) RestoreConnections(ctx context.Context, inboxes []InboxInfo) error {
	for _, inbox := range inboxes {
		if inbox.JID != "" {
			log.Printf("[WhatsApp] Restoring connection for inbox %s (JID: %s)", inbox.ID, inbox.JID)
			if err := m.Connect(ctx, inbox.ID, inbox.JID); err != nil {
				log.Printf("[WhatsApp] Failed to restore %s: %v", inbox.ID, err)
			}
		}
	}
	return nil
}

// Shutdown desconecta todos os adapters
func (m *Manager) Shutdown() {
	m.mu.Lock()
	defer m.mu.Unlock()

	log.Printf("[WhatsApp] Shutting down %d adapters", len(m.adapters))

	ctx := context.Background()
	for id, adapter := range m.adapters {
		if err := adapter.Disconnect(ctx); err != nil {
			log.Printf("[WhatsApp] Failed to disconnect %s: %v", id, err)
		}
	}

	m.adapters = make(map[string]*Adapter)
}

// InboxInfo info para restaurar conexao
type InboxInfo struct {
	ID  string
	JID string
}
