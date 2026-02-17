package whatsapp

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/types"
)

// Manager handles multiple WhatsApp client connections (wuzapi ClientManager pattern)
type Manager struct {
	store            *Store
	clients          map[string]*Client           // Our wrapped clients
	whatsmeowClients map[string]*whatsmeow.Client // Raw whatsmeow clients
	killChannels     map[string]chan bool         // Kill channels per connection
	handler          EventHandler
	broadcaster      EventBroadcaster
	qrCodes          map[string]string // Cache of QR codes by connection ID
	mu               sync.RWMutex
}

// NewManager creates a new WhatsApp connection manager
func NewManager(store *Store, broadcaster EventBroadcaster) *Manager {
	return &Manager{
		store:            store,
		clients:          make(map[string]*Client),
		whatsmeowClients: make(map[string]*whatsmeow.Client),
		killChannels:     make(map[string]chan bool),
		broadcaster:      broadcaster,
		qrCodes:          make(map[string]string),
	}
}

// SetQRCode stores a QR code for a connection
func (m *Manager) SetQRCode(connectionID, code string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.qrCodes[connectionID] = code
}

// GetQRCode retrieves the QR code for a connection
func (m *Manager) GetQRCode(connectionID string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.qrCodes[connectionID]
}

// ClearQRCode removes the QR code for a connection
func (m *Manager) ClearQRCode(connectionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.qrCodes, connectionID)
}

// SetEventHandler sets the event handler for all clients
func (m *Manager) SetEventHandler(handler EventHandler) {
	m.handler = handler
}

// GetWhatsmeowClient returns the raw whatsmeow client (wuzapi pattern)
func (m *Manager) GetWhatsmeowClient(connectionID string) *whatsmeow.Client {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.whatsmeowClients[connectionID]
}

// SetWhatsmeowClient stores the raw whatsmeow client (wuzapi pattern)
func (m *Manager) SetWhatsmeowClient(connectionID string, client *whatsmeow.Client) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.whatsmeowClients[connectionID] = client
}

// DeleteWhatsmeowClient removes the raw whatsmeow client (wuzapi pattern)
func (m *Manager) DeleteWhatsmeowClient(connectionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.whatsmeowClients, connectionID)
}

// Connect creates and connects a WhatsApp client (wuzapi pattern with startClient goroutine)
func (m *Manager) Connect(ctx context.Context, connectionID string, jid string) error {
	m.mu.Lock()

	// Check if already connected
	if client, exists := m.clients[connectionID]; exists {
		m.mu.Unlock()
		if client.WAClient.IsConnected() {
			return fmt.Errorf("connection %s already connected", connectionID)
		}
	}

	// Get or create device
	var device *Device
	var err error

	if jid != "" {
		// Try to get existing device by JID (for reconnection)
		parsedJID, parseErr := types.ParseJID(jid)
		if parseErr == nil {
			device, err = m.store.Container.GetDevice(ctx, parsedJID)
			if err != nil {
				log.Printf("[Manager] Failed to get device by JID %s: %v", jid, err)
				device = nil
			}
		}
	}

	// Create new device if not found
	if device == nil {
		log.Printf("[Manager] Creating new device for connection %s", connectionID)
		device = m.store.NewDevice()
	}

	// Create kill channel
	killChan := make(chan bool, 1)
	m.killChannels[connectionID] = killChan

	// Create our client wrapper
	client := NewClient(device, connectionID, m.handler)
	m.clients[connectionID] = client
	m.whatsmeowClients[connectionID] = client.WAClient

	m.mu.Unlock()

	// Start client in goroutine (wuzapi pattern)
	go client.StartClient(ctx, killChan)

	log.Printf("[Manager] Client %s starting...", connectionID)
	return nil
}

// ConnectSimple creates and connects without JID (for new connections)
func (m *Manager) ConnectSimple(ctx context.Context, connectionID string) error {
	return m.Connect(ctx, connectionID, "")
}

// Disconnect disconnects a WhatsApp client (wuzapi pattern with kill signal)
func (m *Manager) Disconnect(connectionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Send kill signal
	if killChan, exists := m.killChannels[connectionID]; exists {
		select {
		case killChan <- true:
			log.Printf("[Manager] Kill signal sent to %s", connectionID)
		default:
			log.Printf("[Manager] Kill channel full for %s", connectionID)
		}
		delete(m.killChannels, connectionID)
	}

	// Clean up
	delete(m.clients, connectionID)
	delete(m.whatsmeowClients, connectionID)
	delete(m.qrCodes, connectionID)

	log.Printf("[Manager] Client %s disconnected", connectionID)
	return nil
}

// Remove disconnects and removes a client completely
func (m *Manager) Remove(ctx context.Context, connectionID string) error {
	m.mu.Lock()

	client, exists := m.clients[connectionID]
	if !exists {
		m.mu.Unlock()
		return nil
	}

	// Logout from WhatsApp (removes device registration)
	if client.WAClient.IsLoggedIn() {
		if err := client.Logout(ctx); err != nil {
			log.Printf("[Manager] Failed to logout client %s: %v", connectionID, err)
		}
	}

	m.mu.Unlock()

	// Disconnect (sends kill signal)
	return m.Disconnect(connectionID)
}

// GetClient returns a client by connection ID
func (m *Manager) GetClient(connectionID string) (*Client, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	client, exists := m.clients[connectionID]
	if !exists {
		return nil, fmt.Errorf("connection %s not found", connectionID)
	}

	return client, nil
}

// GetConnectedClient returns a client only if it's connected
func (m *Manager) GetConnectedClient(connectionID string) (*Client, error) {
	client, err := m.GetClient(connectionID)
	if err != nil {
		return nil, err
	}

	if !client.WAClient.IsConnected() {
		return nil, fmt.Errorf("connection %s not connected", connectionID)
	}

	return client, nil
}

// IsConnected checks if a connection is active
func (m *Manager) IsConnected(connectionID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	client, exists := m.clients[connectionID]
	if !exists {
		return false
	}

	return client.WAClient.IsConnected()
}

// IsLoggedIn checks if a connection is logged in
func (m *Manager) IsLoggedIn(connectionID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	client, exists := m.clients[connectionID]
	if !exists {
		return false
	}

	return client.WAClient.IsLoggedIn()
}

// GetStatus returns the status of a connection
func (m *Manager) GetStatus(connectionID string) ConnectionStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	client, exists := m.clients[connectionID]
	if !exists {
		return StatusDisconnected
	}

	if client.WAClient.IsConnected() {
		return StatusConnected
	}

	return StatusDisconnected
}

// GetConnectionInfo returns info about a connection
func (m *Manager) GetConnectionInfo(connectionID string) map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	client, exists := m.clients[connectionID]
	if !exists {
		return map[string]interface{}{
			"id":     connectionID,
			"status": StatusDisconnected,
		}
	}

	return map[string]interface{}{
		"id":        connectionID,
		"status":    m.GetStatus(connectionID),
		"connected": client.WAClient.IsConnected(),
		"loggedIn":  client.WAClient.IsLoggedIn(),
		"jid":       client.GetJID(),
		"phone":     client.GetPhoneNumber(),
	}
}

// ListConnections returns info about all connections
func (m *Manager) ListConnections() []map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var connections []map[string]interface{}
	for id := range m.clients {
		connections = append(connections, m.GetConnectionInfo(id))
	}
	return connections
}

// SendMessage sends a text message through a connection
func (m *Manager) SendMessage(ctx context.Context, connectionID, to, text string) (*Message, error) {
	client, err := m.GetConnectedClient(connectionID)
	if err != nil {
		return nil, err
	}

	return client.SendTextMessage(ctx, to, text)
}

// SendImage sends an image message through a connection
func (m *Manager) SendImage(ctx context.Context, connectionID, to string, imageData []byte, caption, mimeType string) (*Message, error) {
	client, err := m.GetConnectedClient(connectionID)
	if err != nil {
		return nil, err
	}

	return client.SendImageMessage(ctx, to, imageData, caption, mimeType)
}

// RestoreConnections attempts to restore all saved connections from the store
func (m *Manager) RestoreConnections(ctx context.Context) error {
	devices, err := m.store.Container.GetAllDevices(ctx)
	if err != nil {
		return fmt.Errorf("failed to get devices: %w", err)
	}

	log.Printf("[Manager] Found %d devices to restore", len(devices))

	for _, device := range devices {
		if device.ID == nil {
			continue
		}

		connectionID := device.ID.String()

		m.mu.Lock()
		killChan := make(chan bool, 1)
		m.killChannels[connectionID] = killChan

		client := NewClient(device, connectionID, m.handler)
		m.clients[connectionID] = client
		m.whatsmeowClients[connectionID] = client.WAClient
		m.mu.Unlock()

		// Start in background
		go func(c *Client, id string, kc <-chan bool) {
			c.StartClient(ctx, kc)
		}(client, connectionID, killChan)

		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

// Shutdown disconnects all clients and closes the store
func (m *Manager) Shutdown() {
	m.mu.Lock()
	defer m.mu.Unlock()

	log.Printf("[Manager] Shutting down %d clients...", len(m.clients))

	// Send kill signals to all
	for id, killChan := range m.killChannels {
		select {
		case killChan <- true:
			log.Printf("[Manager] Kill signal sent to %s", id)
		default:
		}
	}

	// Wait a bit for graceful shutdown
	time.Sleep(500 * time.Millisecond)

	// Clear everything
	m.clients = make(map[string]*Client)
	m.whatsmeowClients = make(map[string]*whatsmeow.Client)
	m.killChannels = make(map[string]chan bool)
	m.qrCodes = make(map[string]string)

	if m.store != nil {
		m.store.Close()
	}

	log.Printf("[Manager] Shutdown complete")
}

// Device is an alias for store.Device
type Device = store.Device
