package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/zyntra/backend/internal/domain"
)

// ConversationRepository repositorio de conversas
type ConversationRepository struct {
	db *sql.DB
}

// NewConversationRepository cria novo repositorio
func NewConversationRepository(db *sql.DB) *ConversationRepository {
	return &ConversationRepository{db: db}
}

// Create cria uma conversa
func (r *ConversationRepository) Create(ctx context.Context, conv *domain.Conversation) error {
	attrsJSON, _ := json.Marshal(conv.AdditionalAttributes)
	query := `
		INSERT INTO conversations (id, inbox_id, contact_id, contact_inbox_id, assignee_id, 
		                           status, priority, unread_count, is_favorite, is_archived,
		                           last_message_at, additional_attributes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW(), NOW())
	`
	_, err := r.db.ExecContext(ctx, query,
		conv.ID, conv.InboxID, conv.ContactID, nullString(conv.ContactInboxID),
		conv.AssigneeID, conv.Status, conv.Priority, conv.UnreadCount,
		conv.IsFavorite, conv.IsArchived, conv.LastMessageAt, attrsJSON,
	)
	return err
}

// GetByID busca conversa por ID
func (r *ConversationRepository) GetByID(ctx context.Context, id string) (*domain.Conversation, error) {
	query := `
		SELECT id, inbox_id, contact_id, COALESCE(contact_inbox_id::text, ''), assignee_id,
		       status, priority, unread_count, is_favorite, is_archived,
		       last_message_at, COALESCE(additional_attributes, '{}'), created_at, updated_at
		FROM conversations WHERE id = $1
	`
	conv := &domain.Conversation{}
	var contactInboxID, priority sql.NullString
	var attrsJSON []byte
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&conv.ID, &conv.InboxID, &conv.ContactID, &contactInboxID, &conv.AssigneeID,
		&conv.Status, &priority, &conv.UnreadCount, &conv.IsFavorite, &conv.IsArchived,
		&conv.LastMessageAt, &attrsJSON, &conv.CreatedAt, &conv.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	conv.ContactInboxID = contactInboxID.String
	if priority.Valid {
		p := domain.ConversationPriority(priority.String)
		conv.Priority = &p
	}
	json.Unmarshal(attrsJSON, &conv.AdditionalAttributes)
	return conv, nil
}

// GetByContactInboxID busca conversa por contact_inbox_id
func (r *ConversationRepository) GetByContactInboxID(ctx context.Context, contactInboxID string) (*domain.Conversation, error) {
	query := `
		SELECT id, inbox_id, contact_id, COALESCE(contact_inbox_id::text, ''), assignee_id,
		       status, priority, unread_count, is_favorite, is_archived,
		       last_message_at, COALESCE(additional_attributes, '{}'), created_at, updated_at
		FROM conversations WHERE contact_inbox_id = $1
		ORDER BY created_at DESC LIMIT 1
	`
	conv := &domain.Conversation{}
	var ciID, priority sql.NullString
	var attrsJSON []byte
	err := r.db.QueryRowContext(ctx, query, contactInboxID).Scan(
		&conv.ID, &conv.InboxID, &conv.ContactID, &ciID, &conv.AssigneeID,
		&conv.Status, &priority, &conv.UnreadCount, &conv.IsFavorite, &conv.IsArchived,
		&conv.LastMessageAt, &attrsJSON, &conv.CreatedAt, &conv.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	conv.ContactInboxID = ciID.String
	if priority.Valid {
		p := domain.ConversationPriority(priority.String)
		conv.Priority = &p
	}
	json.Unmarshal(attrsJSON, &conv.AdditionalAttributes)
	return conv, nil
}

// List lista conversas com filtros
func (r *ConversationRepository) List(ctx context.Context, filter domain.ConversationFilter) ([]*domain.Conversation, error) {
	query := `
		SELECT id, inbox_id, contact_id, COALESCE(contact_inbox_id::text, ''), assignee_id,
		       status, priority, unread_count, is_favorite, is_archived,
		       last_message_at, COALESCE(additional_attributes, '{}'), created_at, updated_at
		FROM conversations WHERE 1=1
	`
	var args []interface{}
	argNum := 1

	if filter.InboxID != nil {
		query += fmt.Sprintf(" AND inbox_id = $%d", argNum)
		args = append(args, *filter.InboxID)
		argNum++
	}
	if filter.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argNum)
		args = append(args, *filter.Status)
		argNum++
	}
	if filter.AssigneeID != nil {
		query += fmt.Sprintf(" AND assignee_id = $%d", argNum)
		args = append(args, *filter.AssigneeID)
		argNum++
	}
	if filter.ContactID != nil {
		query += fmt.Sprintf(" AND contact_id = $%d", argNum)
		args = append(args, *filter.ContactID)
		argNum++
	}
	if filter.IsFavorite != nil {
		query += fmt.Sprintf(" AND is_favorite = $%d", argNum)
		args = append(args, *filter.IsFavorite)
		argNum++
	}
	if filter.IsArchived != nil {
		query += fmt.Sprintf(" AND is_archived = $%d", argNum)
		args = append(args, *filter.IsArchived)
		argNum++
	}

	query += " ORDER BY last_message_at DESC NULLS LAST"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argNum)
		args = append(args, filter.Limit)
		argNum++
	} else {
		query += " LIMIT 50"
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argNum)
		args = append(args, filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []*domain.Conversation
	for rows.Next() {
		conv := &domain.Conversation{}
		var ciID, priority sql.NullString
		var attrsJSON []byte
		if err := rows.Scan(
			&conv.ID, &conv.InboxID, &conv.ContactID, &ciID, &conv.AssigneeID,
			&conv.Status, &priority, &conv.UnreadCount, &conv.IsFavorite, &conv.IsArchived,
			&conv.LastMessageAt, &attrsJSON, &conv.CreatedAt, &conv.UpdatedAt,
		); err != nil {
			return nil, err
		}
		conv.ContactInboxID = ciID.String
		if priority.Valid {
			p := domain.ConversationPriority(priority.String)
			conv.Priority = &p
		}
		json.Unmarshal(attrsJSON, &conv.AdditionalAttributes)
		conversations = append(conversations, conv)
	}
	return conversations, rows.Err()
}

// Update atualiza uma conversa
func (r *ConversationRepository) Update(ctx context.Context, conv *domain.Conversation) error {
	attrsJSON, _ := json.Marshal(conv.AdditionalAttributes)
	query := `
		UPDATE conversations SET assignee_id = $2, status = $3, priority = $4,
		       unread_count = $5, is_favorite = $6, is_archived = $7,
		       last_message_at = $8, additional_attributes = $9, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query,
		conv.ID, conv.AssigneeID, conv.Status, conv.Priority, conv.UnreadCount,
		conv.IsFavorite, conv.IsArchived, conv.LastMessageAt, attrsJSON,
	)
	return err
}

// UpdateLastMessage atualiza timestamp da ultima mensagem
func (r *ConversationRepository) UpdateLastMessage(ctx context.Context, id string, lastMessageAt interface{}) error {
	query := `UPDATE conversations SET last_message_at = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, lastMessageAt)
	return err
}

// IncrementUnread incrementa contador de nao lidas
func (r *ConversationRepository) IncrementUnread(ctx context.Context, id string) error {
	query := `UPDATE conversations SET unread_count = unread_count + 1, updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// ResetUnread zera contador de nao lidas
func (r *ConversationRepository) ResetUnread(ctx context.Context, id string) error {
	query := `UPDATE conversations SET unread_count = 0, updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// Delete remove uma conversa
func (r *ConversationRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM conversations WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// FindOrCreateByContactInbox busca ou cria conversa
func (r *ConversationRepository) FindOrCreateByContactInbox(ctx context.Context, conv *domain.Conversation) (*domain.Conversation, error) {
	existing, err := r.GetByContactInboxID(ctx, conv.ContactInboxID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}
	if err := r.Create(ctx, conv); err != nil {
		return nil, err
	}
	return conv, nil
}

// LabelRepository repositorio de labels
type LabelRepository struct {
	db *sql.DB
}

// NewLabelRepository cria novo repositorio
func NewLabelRepository(db *sql.DB) *LabelRepository {
	return &LabelRepository{db: db}
}

// Create cria um label
func (r *LabelRepository) Create(ctx context.Context, label *domain.Label) error {
	query := `INSERT INTO labels (id, title, color, description, created_at) VALUES ($1, $2, $3, $4, NOW())`
	_, err := r.db.ExecContext(ctx, query, label.ID, label.Title, label.Color, label.Description)
	return err
}

// GetByID busca label por ID
func (r *LabelRepository) GetByID(ctx context.Context, id string) (*domain.Label, error) {
	query := `SELECT id, title, color, COALESCE(description, ''), created_at FROM labels WHERE id = $1`
	label := &domain.Label{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(&label.ID, &label.Title, &label.Color, &label.Description, &label.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return label, err
}

// GetAll lista todos os labels
func (r *LabelRepository) GetAll(ctx context.Context) ([]*domain.Label, error) {
	query := `SELECT id, title, color, COALESCE(description, ''), created_at FROM labels ORDER BY title`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var labels []*domain.Label
	for rows.Next() {
		label := &domain.Label{}
		if err := rows.Scan(&label.ID, &label.Title, &label.Color, &label.Description, &label.CreatedAt); err != nil {
			return nil, err
		}
		labels = append(labels, label)
	}
	return labels, rows.Err()
}

// Delete remove um label
func (r *LabelRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM labels WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// AddLabelToConversation adiciona label a conversa
func (r *LabelRepository) AddToConversation(ctx context.Context, conversationID, labelID string) error {
	query := `INSERT INTO conversation_labels (conversation_id, label_id, created_at) VALUES ($1, $2, NOW()) ON CONFLICT DO NOTHING`
	_, err := r.db.ExecContext(ctx, query, conversationID, labelID)
	return err
}

// RemoveLabelFromConversation remove label da conversa
func (r *LabelRepository) RemoveFromConversation(ctx context.Context, conversationID, labelID string) error {
	query := `DELETE FROM conversation_labels WHERE conversation_id = $1 AND label_id = $2`
	_, err := r.db.ExecContext(ctx, query, conversationID, labelID)
	return err
}

// GetConversationLabels lista labels de uma conversa
func (r *LabelRepository) GetConversationLabels(ctx context.Context, conversationID string) ([]*domain.Label, error) {
	query := `
		SELECT l.id, l.title, l.color, COALESCE(l.description, ''), l.created_at
		FROM labels l
		JOIN conversation_labels cl ON l.id = cl.label_id
		WHERE cl.conversation_id = $1
	`
	rows, err := r.db.QueryContext(ctx, query, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var labels []*domain.Label
	for rows.Next() {
		label := &domain.Label{}
		if err := rows.Scan(&label.ID, &label.Title, &label.Color, &label.Description, &label.CreatedAt); err != nil {
			return nil, err
		}
		labels = append(labels, label)
	}
	return labels, rows.Err()
}

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
