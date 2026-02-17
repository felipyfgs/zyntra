package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Message represents a chat message
type Message struct {
	ID           string    `json:"id" db:"id"`
	ConnectionID string    `json:"connection_id" db:"connection_id"`
	ChatJID      string    `json:"chat_jid" db:"chat_jid"`
	SenderJID    string    `json:"sender_jid" db:"sender_jid"`
	SenderName   string    `json:"sender_name,omitempty" db:"sender_name"`
	Direction    string    `json:"direction" db:"direction"` // "inbound" or "outbound"
	Content      string    `json:"content" db:"content"`
	MediaType    string    `json:"media_type,omitempty" db:"media_type"`
	MediaURL     string    `json:"media_url,omitempty" db:"media_url"`
	Status       string    `json:"status" db:"status"` // "pending", "sent", "delivered", "read"
	Timestamp    time.Time `json:"timestamp" db:"timestamp"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// MessageFilter contains filter options for listing messages
type MessageFilter struct {
	ChatID   string
	Page     int
	PerPage  int
	Before   *time.Time // For cursor-based pagination
	After    *time.Time
}

// MessageRepository handles message database operations
type MessageRepository struct {
	db *sql.DB
}

// NewMessageRepository creates a new message repository
func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

// List returns messages for a chat with pagination
func (r *MessageRepository) List(ctx context.Context, userID string, filter MessageFilter) ([]*Message, int64, error) {
	// Set defaults
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage < 1 || filter.PerPage > 100 {
		filter.PerPage = 50
	}

	// First verify user has access to this chat
	var chatJID, connectionID string
	err := r.db.QueryRowContext(ctx, `
		SELECT c.jid, c.connection_id 
		FROM chats c
		JOIN connections conn ON c.connection_id = conn.id
		WHERE c.id = $1 AND conn.user_id = $2
	`, filter.ChatID, userID).Scan(&chatJID, &connectionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, 0, fmt.Errorf("chat not found")
		}
		return nil, 0, err
	}

	// Count total messages
	var total int64
	err = r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM messages 
		WHERE connection_id = $1 AND chat_jid = $2
	`, connectionID, chatJID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Build query with cursor or offset pagination
	var query string
	var args []interface{}

	if filter.Before != nil {
		query = `
			SELECT m.id, m.connection_id, m.chat_jid, m.sender_jid, 
				COALESCE(ct.name, m.sender_jid) as sender_name,
				m.direction, m.content, 
				COALESCE(m.media_type, '') as media_type,
				COALESCE(m.media_url, '') as media_url,
				m.status, m.timestamp, m.created_at
			FROM messages m
			LEFT JOIN contacts ct ON m.connection_id = ct.connection_id AND m.sender_jid = ct.jid
			WHERE m.connection_id = $1 AND m.chat_jid = $2 AND m.timestamp < $3
			ORDER BY m.timestamp DESC
			LIMIT $4
		`
		args = []interface{}{connectionID, chatJID, filter.Before, filter.PerPage}
	} else if filter.After != nil {
		query = `
			SELECT * FROM (
				SELECT m.id, m.connection_id, m.chat_jid, m.sender_jid, 
					COALESCE(ct.name, m.sender_jid) as sender_name,
					m.direction, m.content, 
					COALESCE(m.media_type, '') as media_type,
					COALESCE(m.media_url, '') as media_url,
					m.status, m.timestamp, m.created_at
				FROM messages m
				LEFT JOIN contacts ct ON m.connection_id = ct.connection_id AND m.sender_jid = ct.jid
				WHERE m.connection_id = $1 AND m.chat_jid = $2 AND m.timestamp > $3
				ORDER BY m.timestamp ASC
				LIMIT $4
			) sub ORDER BY timestamp DESC
		`
		args = []interface{}{connectionID, chatJID, filter.After, filter.PerPage}
	} else {
		// Offset pagination (get latest messages)
		query = `
			SELECT m.id, m.connection_id, m.chat_jid, m.sender_jid, 
				COALESCE(ct.name, m.sender_jid) as sender_name,
				m.direction, m.content, 
				COALESCE(m.media_type, '') as media_type,
				COALESCE(m.media_url, '') as media_url,
				m.status, m.timestamp, m.created_at
			FROM messages m
			LEFT JOIN contacts ct ON m.connection_id = ct.connection_id AND m.sender_jid = ct.jid
			WHERE m.connection_id = $1 AND m.chat_jid = $2
			ORDER BY m.timestamp DESC
			LIMIT $3 OFFSET $4
		`
		args = []interface{}{connectionID, chatJID, filter.PerPage, (filter.Page - 1) * filter.PerPage}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		var msg Message
		err := rows.Scan(
			&msg.ID, &msg.ConnectionID, &msg.ChatJID, &msg.SenderJID,
			&msg.SenderName, &msg.Direction, &msg.Content,
			&msg.MediaType, &msg.MediaURL, &msg.Status,
			&msg.Timestamp, &msg.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		messages = append(messages, &msg)
	}

	return messages, total, nil
}

// GetByID returns a message by ID
func (r *MessageRepository) GetByID(ctx context.Context, userID, messageID string) (*Message, error) {
	query := `
		SELECT m.id, m.connection_id, m.chat_jid, m.sender_jid, 
			COALESCE(ct.name, m.sender_jid) as sender_name,
			m.direction, m.content, 
			COALESCE(m.media_type, '') as media_type,
			COALESCE(m.media_url, '') as media_url,
			m.status, m.timestamp, m.created_at
		FROM messages m
		JOIN connections conn ON m.connection_id = conn.id
		LEFT JOIN contacts ct ON m.connection_id = ct.connection_id AND m.sender_jid = ct.jid
		WHERE m.id = $1 AND conn.user_id = $2
	`

	var msg Message
	err := r.db.QueryRowContext(ctx, query, messageID, userID).Scan(
		&msg.ID, &msg.ConnectionID, &msg.ChatJID, &msg.SenderJID,
		&msg.SenderName, &msg.Direction, &msg.Content,
		&msg.MediaType, &msg.MediaURL, &msg.Status,
		&msg.Timestamp, &msg.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &msg, nil
}

// Create creates a new message
func (r *MessageRepository) Create(ctx context.Context, msg *Message) error {
	query := `
		INSERT INTO messages (id, connection_id, chat_jid, sender_jid, direction, content, media_type, media_url, status, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.ExecContext(ctx, query,
		msg.ID, msg.ConnectionID, msg.ChatJID, msg.SenderJID,
		msg.Direction, msg.Content, msg.MediaType, msg.MediaURL,
		msg.Status, msg.Timestamp,
	)
	return err
}

// UpdateStatus updates message status
func (r *MessageRepository) UpdateStatus(ctx context.Context, messageID, status string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE messages SET status = $1 WHERE id = $2
	`, status, messageID)
	return err
}

// UpdateStatusBatch updates status for multiple messages
func (r *MessageRepository) UpdateStatusBatch(ctx context.Context, messageIDs []string, status string) error {
	if len(messageIDs) == 0 {
		return nil
	}

	// Build query with IN clause
	placeholders := make([]string, len(messageIDs))
	args := make([]interface{}, len(messageIDs)+1)
	args[0] = status

	for i, id := range messageIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args[i+1] = id
	}

	query := fmt.Sprintf(`
		UPDATE messages SET status = $1 WHERE id IN (%s)
	`, joinStrings(placeholders, ", "))

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

// GetLatestForChat returns the latest message for a chat
func (r *MessageRepository) GetLatestForChat(ctx context.Context, connectionID, chatJID string) (*Message, error) {
	query := `
		SELECT id, connection_id, chat_jid, sender_jid, direction, content, 
			COALESCE(media_type, '') as media_type,
			COALESCE(media_url, '') as media_url,
			status, timestamp, created_at
		FROM messages
		WHERE connection_id = $1 AND chat_jid = $2
		ORDER BY timestamp DESC
		LIMIT 1
	`

	var msg Message
	err := r.db.QueryRowContext(ctx, query, connectionID, chatJID).Scan(
		&msg.ID, &msg.ConnectionID, &msg.ChatJID, &msg.SenderJID,
		&msg.Direction, &msg.Content, &msg.MediaType, &msg.MediaURL,
		&msg.Status, &msg.Timestamp, &msg.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &msg, nil
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
