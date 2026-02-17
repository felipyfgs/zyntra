package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// Client wraps NATS connection with JetStream support
type Client struct {
	conn *nats.Conn
	js   jetstream.JetStream
}

// Config holds NATS configuration
type Config struct {
	URL string
}

// DefaultConfig returns config from environment
func DefaultConfig() *Config {
	url := os.Getenv("NATS_URL")
	if url == "" {
		url = "nats://localhost:4222"
	}
	return &Config{URL: url}
}

// NewClient creates a new NATS client with JetStream
func NewClient(config *Config) (*Client, error) {
	if config == nil {
		config = DefaultConfig()
	}

	log.Printf("[NATS] Connecting to %s...", config.URL)

	// Connect to NATS
	nc, err := nats.Connect(config.URL,
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(10),
		nats.ReconnectWait(time.Second),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			log.Printf("[NATS] Disconnected: %v", err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Printf("[NATS] Reconnected to %s", nc.ConnectedUrl())
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	// Create JetStream context
	js, err := jetstream.New(nc)
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("failed to create JetStream context: %w", err)
	}

	log.Printf("[NATS] Connected successfully")

	return &Client{
		conn: nc,
		js:   js,
	}, nil
}

// Close closes the NATS connection
func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Drain()
		c.conn.Close()
		log.Printf("[NATS] Connection closed")
	}
}

// JetStream returns the JetStream context
func (c *Client) JetStream() jetstream.JetStream {
	return c.js
}

// Conn returns the underlying NATS connection
func (c *Client) Conn() *nats.Conn {
	return c.conn
}

// Publish publishes a message to a subject
func (c *Client) Publish(subject string, data interface{}) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}
	return c.conn.Publish(subject, payload)
}

// PublishJSON publishes JSON data to a subject
func (c *Client) PublishJSON(subject string, data interface{}) error {
	return c.Publish(subject, data)
}

// PublishToStream publishes a message to a JetStream stream
func (c *Client) PublishToStream(ctx context.Context, subject string, data interface{}) (*jetstream.PubAck, error) {
	payload, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}
	return c.js.Publish(ctx, subject, payload)
}

// Event types for real-time updates
type EventType string

const (
	EventTypeMessage          EventType = "message"
	EventTypeMessageStatus    EventType = "message_status"
	EventTypeConnectionStatus EventType = "connection_status"
	EventTypeQRCode           EventType = "qr_code"
)

// Event represents a real-time event
type Event struct {
	Type         EventType   `json:"type"`
	ConnectionID string      `json:"connection_id"`
	Timestamp    time.Time   `json:"timestamp"`
	Data         interface{} `json:"data"`
}

// NewEvent creates a new event
func NewEvent(eventType EventType, connectionID string, data interface{}) *Event {
	return &Event{
		Type:         eventType,
		ConnectionID: connectionID,
		Timestamp:    time.Now(),
		Data:         data,
	}
}

// Message event data
type MessageData struct {
	ID        string `json:"id"`
	ChatJID   string `json:"chat_jid"`
	SenderJID string `json:"sender_jid"`
	Content   string `json:"content"`
	MediaType string `json:"media_type,omitempty"`
	Direction string `json:"direction"`
	Timestamp string `json:"timestamp"`
}

// Message status event data
type MessageStatusData struct {
	MessageID string `json:"message_id"`
	ChatJID   string `json:"chat_jid"`
	Status    string `json:"status"` // sent, delivered, read
}

// Connection status event data
type ConnectionStatusData struct {
	Status string `json:"status"` // disconnected, connecting, qr_code, connected
	Phone  string `json:"phone,omitempty"`
	JID    string `json:"jid,omitempty"`
}

// QR code event data
type QRCodeData struct {
	Code string `json:"code"`
}

// Subject helpers
func SubjectMessages(connectionID string) string {
	return fmt.Sprintf("zyntra.messages.%s", connectionID)
}

func SubjectMessageStatus(connectionID string) string {
	return fmt.Sprintf("zyntra.messages.%s.status", connectionID)
}

func SubjectConnectionStatus(connectionID string) string {
	return fmt.Sprintf("zyntra.connections.%s", connectionID)
}

func SubjectQRCode(connectionID string) string {
	return fmt.Sprintf("zyntra.qr.%s", connectionID)
}

// PublishMessage publishes a new message event
func (c *Client) PublishMessage(ctx context.Context, connectionID string, data *MessageData) error {
	event := NewEvent(EventTypeMessage, connectionID, data)
	_, err := c.PublishToStream(ctx, SubjectMessages(connectionID), event)
	return err
}

// PublishMessageStatus publishes a message status update
func (c *Client) PublishMessageStatus(ctx context.Context, connectionID string, data *MessageStatusData) error {
	event := NewEvent(EventTypeMessageStatus, connectionID, data)
	_, err := c.PublishToStream(ctx, SubjectMessageStatus(connectionID), event)
	return err
}

// PublishConnectionStatus publishes a connection status update
func (c *Client) PublishConnectionStatus(ctx context.Context, connectionID string, data *ConnectionStatusData) error {
	event := NewEvent(EventTypeConnectionStatus, connectionID, data)
	_, err := c.PublishToStream(ctx, SubjectConnectionStatus(connectionID), event)
	return err
}

// PublishQRCode publishes a QR code event
func (c *Client) PublishQRCode(ctx context.Context, connectionID string, code string) error {
	event := NewEvent(EventTypeQRCode, connectionID, &QRCodeData{Code: code})
	_, err := c.PublishToStream(ctx, SubjectQRCode(connectionID), event)
	return err
}
