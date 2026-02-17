package whatsapp

import (
	"time"

	"go.mau.fi/whatsmeow/types"
)

// ConnectionStatus represents the current state of a WhatsApp connection
type ConnectionStatus string

const (
	StatusDisconnected ConnectionStatus = "disconnected"
	StatusConnecting   ConnectionStatus = "connecting"
	StatusConnected    ConnectionStatus = "connected"
	StatusQRCode       ConnectionStatus = "qr_code"
)

// MessageDirection indicates if message is incoming or outgoing
type MessageDirection string

const (
	DirectionInbound  MessageDirection = "inbound"
	DirectionOutbound MessageDirection = "outbound"
)

// MessageStatus represents the delivery status of a message
type MessageStatus string

const (
	MessagePending   MessageStatus = "pending"
	MessageSent      MessageStatus = "sent"
	MessageDelivered MessageStatus = "delivered"
	MessageRead      MessageStatus = "read"
	MessageFailed    MessageStatus = "failed"
)

// Connection represents a WhatsApp connection/session
type Connection struct {
	ID        string           `json:"id"`
	UserID    string           `json:"user_id"`
	Name      string           `json:"name"`
	Phone     string           `json:"phone"`
	JID       string           `json:"jid"`
	Status    ConnectionStatus `json:"status"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

// Message represents a WhatsApp message
type Message struct {
	ID           string           `json:"id"`
	ConnectionID string           `json:"connection_id"`
	ChatJID      string           `json:"chat_jid"`
	SenderJID    string           `json:"sender_jid"`
	Direction    MessageDirection `json:"direction"`
	Content      string           `json:"content"`
	MediaType    string           `json:"media_type,omitempty"`
	MediaURL     string           `json:"media_url,omitempty"`
	Status       MessageStatus    `json:"status"`
	Timestamp    time.Time        `json:"timestamp"`
	CreatedAt    time.Time        `json:"created_at"`
}

// Contact represents a WhatsApp contact
type Contact struct {
	ID           string    `json:"id"`
	ConnectionID string    `json:"connection_id"`
	JID          string    `json:"jid"`
	Phone        string    `json:"phone"`
	Name         string    `json:"name"`
	PushName     string    `json:"push_name"`
	AvatarURL    string    `json:"avatar_url,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Chat represents a WhatsApp conversation
type Chat struct {
	ID            string    `json:"id"`
	ConnectionID  string    `json:"connection_id"`
	JID           string    `json:"jid"`
	Name          string    `json:"name"`
	IsGroup       bool      `json:"is_group"`
	LastMessageAt time.Time `json:"last_message_at"`
	UnreadCount   int       `json:"unread_count"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// QRCodeEvent is sent when a QR code is generated for pairing
type QRCodeEvent struct {
	ConnectionID string `json:"connection_id"`
	Code         string `json:"code"`
	Base64Image  string `json:"base64_image"` // QR code as data:image/png;base64,...
}

// PairSuccessEvent is sent when QR code pairing is successful (wuzapi pattern)
type PairSuccessEvent struct {
	ConnectionID string `json:"connection_id"`
	JID          string `json:"jid"`
	BusinessName string `json:"business_name,omitempty"`
	Platform     string `json:"platform,omitempty"`
}

// ConnectionEvent is sent when connection status changes
type ConnectionEvent struct {
	ConnectionID string           `json:"connection_id"`
	Status       ConnectionStatus `json:"status"`
	Phone        string           `json:"phone,omitempty"`
}

// MessageEvent is sent when a new message is received or status changes
type MessageEvent struct {
	ConnectionID string        `json:"connection_id"`
	Message      *Message      `json:"message,omitempty"`
	Status       MessageStatus `json:"status,omitempty"`
	MessageID    string        `json:"message_id,omitempty"`
}

// SendMessageRequest represents a request to send a message
type SendMessageRequest struct {
	ConnectionID string `json:"connection_id"`
	To           string `json:"to"`
	Content      string `json:"content"`
	MediaType    string `json:"media_type,omitempty"`
	MediaURL     string `json:"media_url,omitempty"`
}

// EventHandler is the interface for handling WhatsApp events
type EventHandler interface {
	OnQRCode(event QRCodeEvent)
	OnQRTimeout(connectionID string)
	OnPairSuccess(event PairSuccessEvent)
	OnConnected(event ConnectionEvent)
	OnDisconnected(event ConnectionEvent)
	OnMessage(event MessageEvent)
	OnMessageStatus(event MessageEvent)
}

// JIDToPhone extracts phone number from JID
func JIDToPhone(jid types.JID) string {
	return "+" + jid.User
}

// PhoneToJID creates a JID from phone number
func PhoneToJID(phone string) types.JID {
	// Remove + and any formatting
	user := phone
	if len(user) > 0 && user[0] == '+' {
		user = user[1:]
	}
	return types.NewJID(user, types.DefaultUserServer)
}
