package domain

import (
	"time"

	"github.com/zyntra/backend/internal/ports"
)

// SenderType tipo do remetente
type SenderType string

const (
	SenderTypeContact SenderType = "contact"
	SenderTypeUser    SenderType = "user"
	SenderTypeSystem  SenderType = "system"
	SenderTypeBot     SenderType = "bot"
)

// ContentType tipo de conteudo
type ContentType string

const (
	ContentTypeText     ContentType = "text"
	ContentTypeImage    ContentType = "image"
	ContentTypeVideo    ContentType = "video"
	ContentTypeAudio    ContentType = "audio"
	ContentTypeDocument ContentType = "document"
	ContentTypeSticker  ContentType = "sticker"
	ContentTypeLocation ContentType = "location"
)

// Message mensagem
type Message struct {
	ID                string                 `json:"id" db:"id"`
	ConversationID    string                 `json:"conversation_id" db:"conversation_id"`
	InboxID           string                 `json:"inbox_id" db:"inbox_id"`
	SenderType        SenderType             `json:"sender_type" db:"sender_type"`
	SenderID          *string                `json:"sender_id,omitempty" db:"sender_id"`
	Content           string                 `json:"content" db:"content"`
	ContentType       ContentType            `json:"content_type" db:"content_type"`
	ContentAttributes map[string]interface{} `json:"content_attributes,omitempty" db:"content_attributes"`
	SourceID          string                 `json:"source_id,omitempty" db:"source_id"`
	Status            ports.MessageStatus    `json:"status" db:"status"`
	Private           bool                   `json:"private" db:"private"`
	CreatedAt         time.Time              `json:"created_at" db:"created_at"`
}

// MessageWithSender mensagem com dados do remetente
type MessageWithSender struct {
	Message
	Sender      interface{}  `json:"sender,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

// Attachment anexo de mensagem
type Attachment struct {
	ID        string    `json:"id" db:"id"`
	MessageID string    `json:"message_id" db:"message_id"`
	FileType  string    `json:"file_type" db:"file_type"`
	FileURL   string    `json:"file_url" db:"file_url"`
	FileName  string    `json:"file_name,omitempty" db:"file_name"`
	FileSize  int64     `json:"file_size,omitempty" db:"file_size"`
	MimeType  string    `json:"mime_type,omitempty" db:"mime_type"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// SendMessageRequest request para enviar mensagem
type SendMessageRequest struct {
	Content     string                 `json:"content" validate:"required"`
	ContentType ContentType            `json:"content_type,omitempty"`
	Private     bool                   `json:"private,omitempty"`
	Attachments []AttachmentRequest    `json:"attachments,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// AttachmentRequest request de anexo
type AttachmentRequest struct {
	FileType string `json:"file_type"`
	FileURL  string `json:"file_url,omitempty"`
	FileName string `json:"file_name,omitempty"`
	Data     []byte `json:"data,omitempty"`
	MimeType string `json:"mime_type,omitempty"`
}

// MessageFilter filtros para busca de mensagens
type MessageFilter struct {
	ConversationID string    `json:"conversation_id"`
	Before         *time.Time `json:"before,omitempty"`
	After          *time.Time `json:"after,omitempty"`
	SenderType     *SenderType `json:"sender_type,omitempty"`
	Limit          int        `json:"limit,omitempty"`
	Offset         int        `json:"offset,omitempty"`
}

// IncomingMessageEvent evento de mensagem recebida (do canal)
type IncomingMessageEvent struct {
	InboxID         string
	SourceID        string
	ContactSourceID string
	ContactName     string
	IsFromMe        bool
	Content         string
	ContentType     ContentType
	MediaURL        string
	Timestamp       time.Time
}
