package domain

import (
	"time"
)

// Contact contato global unificado
type Contact struct {
	ID               string                 `json:"id" db:"id"`
	Name             string                 `json:"name" db:"name"`
	Email            string                 `json:"email,omitempty" db:"email"`
	PhoneNumber      string                 `json:"phone_number,omitempty" db:"phone_number"`
	AvatarURL        string                 `json:"avatar_url,omitempty" db:"avatar_url"`
	CustomAttributes map[string]interface{} `json:"custom_attributes,omitempty" db:"custom_attributes"`
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at" db:"updated_at"`
}

// ContactInbox identidade do contato em um canal especifico
type ContactInbox struct {
	ID        string    `json:"id" db:"id"`
	ContactID string    `json:"contact_id" db:"contact_id"`
	InboxID   string    `json:"inbox_id" db:"inbox_id"`
	SourceID  string    `json:"source_id" db:"source_id"` // JID, chat_id, etc
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// ContactWithInboxes contato com suas identidades por canal
type ContactWithInboxes struct {
	Contact
	ContactInboxes []ContactInbox `json:"contact_inboxes,omitempty"`
}

// Tag tag para contatos
type Tag struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Color     string    `json:"color" db:"color"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// ContactTag associacao contato-tag
type ContactTag struct {
	ContactID string `json:"contact_id" db:"contact_id"`
	TagID     string `json:"tag_id" db:"tag_id"`
}

// CreateContactRequest request para criar contato
type CreateContactRequest struct {
	Name             string                 `json:"name"`
	Email            string                 `json:"email,omitempty"`
	PhoneNumber      string                 `json:"phone_number,omitempty"`
	CustomAttributes map[string]interface{} `json:"custom_attributes,omitempty"`
}

// UpdateContactRequest request para atualizar contato
type UpdateContactRequest struct {
	Name             *string                `json:"name,omitempty"`
	Email            *string                `json:"email,omitempty"`
	PhoneNumber      *string                `json:"phone_number,omitempty"`
	AvatarURL        *string                `json:"avatar_url,omitempty"`
	CustomAttributes map[string]interface{} `json:"custom_attributes,omitempty"`
}
