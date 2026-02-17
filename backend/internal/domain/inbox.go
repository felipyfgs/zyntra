package domain

import (
	"time"

	"github.com/zyntra/backend/internal/ports"
)

// Inbox representa um inbox (ponte entre sistema e canal)
type Inbox struct {
	ID              string              `json:"id" db:"id"`
	Name            string              `json:"name" db:"name"`
	ChannelType     ports.ChannelType   `json:"channel_type" db:"channel_type"`
	ChannelID       string              `json:"channel_id" db:"channel_id"`
	Status          ports.ChannelStatus `json:"status" db:"status"`
	QRCode          string              `json:"qrcode,omitempty" db:"qrcode"`
	GreetingMessage string              `json:"greeting_message,omitempty" db:"greeting_message"`
	AutoAssignment  bool                `json:"auto_assignment" db:"auto_assignment"`
	CreatedAt       time.Time           `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at" db:"updated_at"`
}

// ChannelWhatsApp configuracao do canal WhatsApp
type ChannelWhatsApp struct {
	ID             string                 `json:"id" db:"id"`
	PhoneNumber    string                 `json:"phone_number" db:"phone_number"`
	JID            string                 `json:"jid" db:"jid"`
	Provider       string                 `json:"provider" db:"provider"`
	ProviderConfig map[string]interface{} `json:"provider_config" db:"provider_config"`
	CreatedAt      time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at" db:"updated_at"`
}

// ChannelTelegram configuracao do canal Telegram
type ChannelTelegram struct {
	ID          string    `json:"id" db:"id"`
	BotToken    string    `json:"bot_token" db:"bot_token"`
	BotUsername string    `json:"bot_username" db:"bot_username"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// ChannelAPI configuracao do canal API
type ChannelAPI struct {
	ID         string    `json:"id" db:"id"`
	WebhookURL string    `json:"webhook_url" db:"webhook_url"`
	APIKey     string    `json:"api_key" db:"api_key"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// InboxMember associacao inbox-usuario
type InboxMember struct {
	InboxID   string    `json:"inbox_id" db:"inbox_id"`
	UserID    string    `json:"user_id" db:"user_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// InboxWithChannel inbox com dados do canal
type InboxWithChannel struct {
	Inbox
	Channel interface{} `json:"channel,omitempty"`
}

// CreateInboxRequest request para criar inbox
type CreateInboxRequest struct {
	Name            string            `json:"name" validate:"required"`
	ChannelType     ports.ChannelType `json:"channel_type" validate:"required"`
	GreetingMessage string            `json:"greeting_message,omitempty"`
	AutoAssignment  bool              `json:"auto_assignment"`
	ChannelConfig   map[string]string `json:"channel_config,omitempty"`
}
