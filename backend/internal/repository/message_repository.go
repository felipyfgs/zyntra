package repository

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/zyntra/backend/internal/domain"
	"github.com/zyntra/backend/internal/ports"
)

// MessageRepository repositorio de mensagens
type MessageRepository struct {
	db *sql.DB
}

// NewMessageRepository cria novo repositorio
func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

// Create cria uma mensagem
func (r *MessageRepository) Create(ctx context.Context, msg *domain.Message) error {
	attrsJSON, _ := json.Marshal(msg.ContentAttributes)
	query := `
		INSERT INTO messages (id, conversation_id, inbox_id, sender_type, sender_id, 
		                      content, content_type, content_attributes, source_id, status, private, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := r.db.ExecContext(ctx, query,
		msg.ID, msg.ConversationID, msg.InboxID, msg.SenderType, msg.SenderID,
		msg.Content, msg.ContentType, attrsJSON, msg.SourceID, msg.Status, msg.Private, msg.CreatedAt,
	)
	return err
}

// GetByID busca mensagem por ID
func (r *MessageRepository) GetByID(ctx context.Context, id string) (*domain.Message, error) {
	query := `
		SELECT id, conversation_id, inbox_id, sender_type, sender_id, 
		       COALESCE(content, ''), content_type, COALESCE(content_attributes, '{}'),
		       COALESCE(source_id, ''), status, private, created_at
		FROM messages WHERE id = $1
	`
	msg := &domain.Message{}
	var attrsJSON []byte
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&msg.ID, &msg.ConversationID, &msg.InboxID, &msg.SenderType, &msg.SenderID,
		&msg.Content, &msg.ContentType, &attrsJSON, &msg.SourceID, &msg.Status, &msg.Private, &msg.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	json.Unmarshal(attrsJSON, &msg.ContentAttributes)
	return msg, nil
}

// GetBySourceID busca mensagem por source_id
func (r *MessageRepository) GetBySourceID(ctx context.Context, inboxID, sourceID string) (*domain.Message, error) {
	query := `
		SELECT id, conversation_id, inbox_id, sender_type, sender_id, 
		       COALESCE(content, ''), content_type, COALESCE(content_attributes, '{}'),
		       COALESCE(source_id, ''), status, private, created_at
		FROM messages WHERE inbox_id = $1 AND source_id = $2
	`
	msg := &domain.Message{}
	var attrsJSON []byte
	err := r.db.QueryRowContext(ctx, query, inboxID, sourceID).Scan(
		&msg.ID, &msg.ConversationID, &msg.InboxID, &msg.SenderType, &msg.SenderID,
		&msg.Content, &msg.ContentType, &attrsJSON, &msg.SourceID, &msg.Status, &msg.Private, &msg.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	json.Unmarshal(attrsJSON, &msg.ContentAttributes)
	return msg, nil
}

// ListByConversation lista mensagens de uma conversa
func (r *MessageRepository) ListByConversation(ctx context.Context, conversationID string, limit, offset int) ([]*domain.Message, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `
		SELECT id, conversation_id, inbox_id, sender_type, sender_id, 
		       COALESCE(content, ''), content_type, COALESCE(content_attributes, '{}'),
		       COALESCE(source_id, ''), status, private, created_at
		FROM messages WHERE conversation_id = $1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, conversationID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*domain.Message
	for rows.Next() {
		msg := &domain.Message{}
		var attrsJSON []byte
		if err := rows.Scan(
			&msg.ID, &msg.ConversationID, &msg.InboxID, &msg.SenderType, &msg.SenderID,
			&msg.Content, &msg.ContentType, &attrsJSON, &msg.SourceID, &msg.Status, &msg.Private, &msg.CreatedAt,
		); err != nil {
			return nil, err
		}
		json.Unmarshal(attrsJSON, &msg.ContentAttributes)
		messages = append(messages, msg)
	}
	return messages, rows.Err()
}

// UpdateStatus atualiza status de uma mensagem
func (r *MessageRepository) UpdateStatus(ctx context.Context, id string, status ports.MessageStatus) error {
	query := `UPDATE messages SET status = $2 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, status)
	return err
}

// UpdateStatusBySourceID atualiza status por source_id
func (r *MessageRepository) UpdateStatusBySourceID(ctx context.Context, inboxID, sourceID string, status ports.MessageStatus) error {
	query := `UPDATE messages SET status = $3 WHERE inbox_id = $1 AND source_id = $2`
	_, err := r.db.ExecContext(ctx, query, inboxID, sourceID, status)
	return err
}

// Delete remove uma mensagem
func (r *MessageRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM messages WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// Count conta mensagens de uma conversa
func (r *MessageRepository) Count(ctx context.Context, conversationID string) (int, error) {
	query := `SELECT COUNT(*) FROM messages WHERE conversation_id = $1`
	var count int
	err := r.db.QueryRowContext(ctx, query, conversationID).Scan(&count)
	return count, err
}

// MessageFilter filtro para busca de mensagens (compatibilidade legado)
type MessageFilter struct {
	ChatID  string
	Page    int
	PerPage int
}

// LegacyMessage mensagem no formato legado
type LegacyMessage struct {
	ID        string `json:"id"`
	ChatID    string `json:"chat_id"`
	Direction string `json:"direction"`
	Content   string `json:"content"`
	MediaType string `json:"media_type,omitempty"`
	MediaURL  string `json:"media_url,omitempty"`
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	CreatedAt string `json:"created_at"`
}

// List lista mensagens (compatibilidade legado - busca por chat_jid na tabela antiga)
func (r *MessageRepository) List(ctx context.Context, userID string, filter MessageFilter) ([]*LegacyMessage, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage < 1 || filter.PerPage > 100 {
		filter.PerPage = 50
	}

	// Contar total
	var total int64
	countQuery := `
		SELECT COUNT(*) FROM messages m
		JOIN conversations c ON m.conversation_id = c.id
		WHERE c.id = $1
	`
	if err := r.db.QueryRowContext(ctx, countQuery, filter.ChatID).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Buscar mensagens
	query := `
		SELECT m.id, m.conversation_id, 
		       CASE WHEN m.sender_type = 'contact' THEN 'inbound' ELSE 'outbound' END as direction,
		       COALESCE(m.content, ''), COALESCE(m.content_type, 'text'),
		       m.status, m.created_at
		FROM messages m
		WHERE m.conversation_id = $1
		ORDER BY m.created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, filter.ChatID, filter.PerPage, (filter.Page-1)*filter.PerPage)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var messages []*LegacyMessage
	for rows.Next() {
		msg := &LegacyMessage{}
		var createdAt interface{}
		if err := rows.Scan(&msg.ID, &msg.ChatID, &msg.Direction, &msg.Content, &msg.MediaType, &msg.Status, &createdAt); err != nil {
			return nil, 0, err
		}
		messages = append(messages, msg)
	}

	return messages, total, rows.Err()
}

// AttachmentRepository repositorio de attachments
type AttachmentRepository struct {
	db *sql.DB
}

// NewAttachmentRepository cria novo repositorio
func NewAttachmentRepository(db *sql.DB) *AttachmentRepository {
	return &AttachmentRepository{db: db}
}

// Create cria um attachment
func (r *AttachmentRepository) Create(ctx context.Context, att *domain.Attachment) error {
	query := `
		INSERT INTO attachments (id, message_id, file_type, file_url, file_name, file_size, mime_type, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
	`
	_, err := r.db.ExecContext(ctx, query,
		att.ID, att.MessageID, att.FileType, att.FileURL, att.FileName, att.FileSize, att.MimeType,
	)
	return err
}

// GetByMessageID lista attachments de uma mensagem
func (r *AttachmentRepository) GetByMessageID(ctx context.Context, messageID string) ([]*domain.Attachment, error) {
	query := `
		SELECT id, message_id, COALESCE(file_type, ''), COALESCE(file_url, ''),
		       COALESCE(file_name, ''), COALESCE(file_size, 0), COALESCE(mime_type, ''), created_at
		FROM attachments WHERE message_id = $1
	`
	rows, err := r.db.QueryContext(ctx, query, messageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attachments []*domain.Attachment
	for rows.Next() {
		att := &domain.Attachment{}
		if err := rows.Scan(
			&att.ID, &att.MessageID, &att.FileType, &att.FileURL,
			&att.FileName, &att.FileSize, &att.MimeType, &att.CreatedAt,
		); err != nil {
			return nil, err
		}
		attachments = append(attachments, att)
	}
	return attachments, rows.Err()
}

// Delete remove um attachment
func (r *AttachmentRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM attachments WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
