package domain

import (
	"time"
)

// ConversationStatus status da conversa
type ConversationStatus string

const (
	ConversationStatusOpen     ConversationStatus = "open"
	ConversationStatusResolved ConversationStatus = "resolved"
	ConversationStatusPending  ConversationStatus = "pending"
)

// ConversationPriority prioridade da conversa
type ConversationPriority string

const (
	ConversationPriorityLow    ConversationPriority = "low"
	ConversationPriorityMedium ConversationPriority = "medium"
	ConversationPriorityHigh   ConversationPriority = "high"
	ConversationPriorityUrgent ConversationPriority = "urgent"
)

// Conversation conversa
type Conversation struct {
	ID                   string                 `json:"id" db:"id"`
	InboxID              string                 `json:"inbox_id" db:"inbox_id"`
	ContactID            string                 `json:"contact_id" db:"contact_id"`
	ContactInboxID       string                 `json:"contact_inbox_id,omitempty" db:"contact_inbox_id"`
	AssigneeID           *string                `json:"assignee_id,omitempty" db:"assignee_id"`
	Status               ConversationStatus     `json:"status" db:"status"`
	Priority             *ConversationPriority  `json:"priority,omitempty" db:"priority"`
	UnreadCount          int                    `json:"unread_count" db:"unread_count"`
	IsFavorite           bool                   `json:"is_favorite" db:"is_favorite"`
	IsArchived           bool                   `json:"is_archived" db:"is_archived"`
	LastMessageAt        *time.Time             `json:"last_message_at,omitempty" db:"last_message_at"`
	AdditionalAttributes map[string]interface{} `json:"additional_attributes,omitempty" db:"additional_attributes"`
	CreatedAt            time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time              `json:"updated_at" db:"updated_at"`
}

// ConversationWithDetails conversa com detalhes relacionados
type ConversationWithDetails struct {
	Conversation
	Contact      *Contact `json:"contact,omitempty"`
	Inbox        *Inbox   `json:"inbox,omitempty"`
	Assignee     *User    `json:"assignee,omitempty"`
	LastMessage  *Message `json:"last_message,omitempty"`
	MessagesCount int     `json:"messages_count,omitempty"`
}

// Label etiqueta para conversas
type Label struct {
	ID          string    `json:"id" db:"id"`
	Title       string    `json:"title" db:"title"`
	Color       string    `json:"color" db:"color"`
	Description string    `json:"description,omitempty" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// ConversationLabel associacao conversa-label
type ConversationLabel struct {
	ConversationID string    `json:"conversation_id" db:"conversation_id"`
	LabelID        string    `json:"label_id" db:"label_id"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// ConversationFilter filtros para busca de conversas
type ConversationFilter struct {
	InboxID    *string             `json:"inbox_id,omitempty"`
	Status     *ConversationStatus `json:"status,omitempty"`
	AssigneeID *string             `json:"assignee_id,omitempty"`
	ContactID  *string             `json:"contact_id,omitempty"`
	IsFavorite *bool               `json:"is_favorite,omitempty"`
	IsArchived *bool               `json:"is_archived,omitempty"`
	LabelIDs   []string            `json:"label_ids,omitempty"`
	Search     *string             `json:"search,omitempty"`
	Limit      int                 `json:"limit,omitempty"`
	Offset     int                 `json:"offset,omitempty"`
}

// UpdateConversationRequest request para atualizar conversa
type UpdateConversationRequest struct {
	Status     *ConversationStatus   `json:"status,omitempty"`
	Priority   *ConversationPriority `json:"priority,omitempty"`
	AssigneeID *string               `json:"assignee_id,omitempty"`
	IsFavorite *bool                 `json:"is_favorite,omitempty"`
	IsArchived *bool                 `json:"is_archived,omitempty"`
}
