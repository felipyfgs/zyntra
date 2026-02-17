package models

import (
	"time"
)

type User struct {
	ID           string    `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Role         string    `json:"role" db:"role"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type Connection struct {
	ID          string    `json:"id" db:"id"`
	UserID      string    `json:"user_id" db:"user_id"`
	Phone       string    `json:"phone" db:"phone"`
	Status      string    `json:"status" db:"status"`
	QRCode      string    `json:"qr_code,omitempty" db:"qr_code"`
	SessionData []byte    `json:"-" db:"session_data"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type Contact struct {
	ID           string    `json:"id" db:"id"`
	ConnectionID string    `json:"connection_id" db:"connection_id"`
	Phone        string    `json:"phone" db:"phone"`
	Name         string    `json:"name" db:"name"`
	AvatarURL    string    `json:"avatar_url,omitempty" db:"avatar_url"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type Chat struct {
	ID            string    `json:"id" db:"id"`
	ConnectionID  string    `json:"connection_id" db:"connection_id"`
	ContactID     string    `json:"contact_id" db:"contact_id"`
	LastMessageAt time.Time `json:"last_message_at" db:"last_message_at"`
	UnreadCount   int       `json:"unread_count" db:"unread_count"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

type Message struct {
	ID        string    `json:"id" db:"id"`
	ChatID    string    `json:"chat_id" db:"chat_id"`
	Direction string    `json:"direction" db:"direction"` // "inbound" or "outbound"
	Content   string    `json:"content" db:"content"`
	MediaURL  string    `json:"media_url,omitempty" db:"media_url"`
	MediaType string    `json:"media_type,omitempty" db:"media_type"`
	Status    string    `json:"status" db:"status"` // "pending", "sent", "delivered", "read"
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Connection statuses
const (
	ConnectionStatusDisconnected = "disconnected"
	ConnectionStatusConnecting   = "connecting"
	ConnectionStatusConnected    = "connected"
	ConnectionStatusQRCode       = "qr_code"
)

// Message directions
const (
	MessageDirectionInbound  = "inbound"
	MessageDirectionOutbound = "outbound"
)

// Message statuses
const (
	MessageStatusPending   = "pending"
	MessageStatusSent      = "sent"
	MessageStatusDelivered = "delivered"
	MessageStatusRead      = "read"
)

// User roles
const (
	UserRoleAdmin    = "admin"
	UserRoleOperator = "operator"
)
