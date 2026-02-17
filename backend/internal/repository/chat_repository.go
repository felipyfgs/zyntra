package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// Chat represents a chat with contact info
type Chat struct {
	ID            string     `json:"id" db:"id"`
	ConnectionID  string     `json:"connection_id" db:"connection_id"`
	JID           string     `json:"jid" db:"jid"`
	Name          string     `json:"name" db:"name"`
	Phone         string     `json:"phone,omitempty" db:"phone"`
	AvatarURL     string     `json:"avatar_url,omitempty" db:"avatar_url"`
	IsGroup       bool       `json:"is_group" db:"is_group"`
	IsFavorite    bool       `json:"is_favorite" db:"is_favorite"`
	IsArchived    bool       `json:"is_archived" db:"is_archived"`
	UnreadCount   int        `json:"unread_count" db:"unread_count"`
	LastMessage   string     `json:"last_message,omitempty" db:"last_message"`
	LastMessageAt *time.Time `json:"last_message_at,omitempty" db:"last_message_at"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

// ChatFilter contains filter options for listing chats
type ChatFilter struct {
	ConnectionID string
	Search       string
	Filter       string // "all", "unread", "groups", "favorites", "archived"
	Page         int
	PerPage      int
}

// ChatRepository handles chat database operations
type ChatRepository struct {
	db *sql.DB
}

// NewChatRepository creates a new chat repository
func NewChatRepository(db *sql.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

// List returns chats with filters and pagination
func (r *ChatRepository) List(ctx context.Context, userID string, filter ChatFilter) ([]*Chat, int64, error) {
	// Set defaults
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage < 1 || filter.PerPage > 100 {
		filter.PerPage = 20
	}

	// Build query
	baseQuery := `
		FROM chats c
		JOIN connections conn ON c.connection_id = conn.id
		LEFT JOIN contacts ct ON c.connection_id = ct.connection_id AND c.jid = ct.jid
		LEFT JOIN (
			SELECT DISTINCT ON (chat_jid, connection_id) 
				chat_jid, connection_id, content as last_msg_content
			FROM messages 
			ORDER BY chat_jid, connection_id, timestamp DESC
		) m ON c.jid = m.chat_jid AND c.connection_id = m.connection_id
		WHERE conn.user_id = $1
	`

	args := []interface{}{userID}
	argIndex := 2

	// Connection filter
	if filter.ConnectionID != "" {
		baseQuery += fmt.Sprintf(" AND c.connection_id = $%d", argIndex)
		args = append(args, filter.ConnectionID)
		argIndex++
	}

	// Search filter
	if filter.Search != "" {
		baseQuery += fmt.Sprintf(" AND (c.name ILIKE $%d OR ct.phone ILIKE $%d OR ct.name ILIKE $%d)", argIndex, argIndex, argIndex)
		args = append(args, "%"+filter.Search+"%")
		argIndex++
	}

	// Type filters
	switch filter.Filter {
	case "unread":
		baseQuery += " AND c.unread_count > 0"
	case "groups":
		baseQuery += " AND c.is_group = true"
	case "favorites":
		baseQuery += " AND c.is_favorite = true"
	case "archived":
		baseQuery += " AND c.is_archived = true"
	default:
		baseQuery += " AND c.is_archived = false"
	}

	// Count total
	var total int64
	countQuery := "SELECT COUNT(*) " + baseQuery
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count chats: %w", err)
	}

	// Select with pagination
	selectQuery := `
		SELECT 
			c.id, c.connection_id, c.jid, 
			COALESCE(ct.name, c.name, c.jid) as name,
			COALESCE(ct.phone, '') as phone,
			COALESCE(ct.avatar_url, '') as avatar_url,
			c.is_group,
			COALESCE(c.is_favorite, false) as is_favorite,
			COALESCE(c.is_archived, false) as is_archived,
			c.unread_count,
			COALESCE(m.last_msg_content, '') as last_message,
			c.last_message_at,
			c.created_at,
			c.updated_at
	` + baseQuery + `
		ORDER BY c.last_message_at DESC NULLS LAST
		LIMIT $` + fmt.Sprintf("%d", argIndex) + ` OFFSET $` + fmt.Sprintf("%d", argIndex+1)

	args = append(args, filter.PerPage, (filter.Page-1)*filter.PerPage)

	rows, err := r.db.QueryContext(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query chats: %w", err)
	}
	defer rows.Close()

	var chats []*Chat
	for rows.Next() {
		var chat Chat
		err := rows.Scan(
			&chat.ID, &chat.ConnectionID, &chat.JID, &chat.Name,
			&chat.Phone, &chat.AvatarURL, &chat.IsGroup,
			&chat.IsFavorite, &chat.IsArchived, &chat.UnreadCount,
			&chat.LastMessage, &chat.LastMessageAt,
			&chat.CreatedAt, &chat.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan chat: %w", err)
		}
		chats = append(chats, &chat)
	}

	return chats, total, nil
}

// GetByID returns a chat by ID
func (r *ChatRepository) GetByID(ctx context.Context, userID, chatID string) (*Chat, error) {
	query := `
		SELECT 
			c.id, c.connection_id, c.jid, 
			COALESCE(ct.name, c.name, c.jid) as name,
			COALESCE(ct.phone, '') as phone,
			COALESCE(ct.avatar_url, '') as avatar_url,
			c.is_group,
			COALESCE(c.is_favorite, false) as is_favorite,
			COALESCE(c.is_archived, false) as is_archived,
			c.unread_count,
			c.last_message_at,
			c.created_at,
			c.updated_at
		FROM chats c
		JOIN connections conn ON c.connection_id = conn.id
		LEFT JOIN contacts ct ON c.connection_id = ct.connection_id AND c.jid = ct.jid
		WHERE c.id = $1 AND conn.user_id = $2
	`

	var chat Chat
	err := r.db.QueryRowContext(ctx, query, chatID, userID).Scan(
		&chat.ID, &chat.ConnectionID, &chat.JID, &chat.Name,
		&chat.Phone, &chat.AvatarURL, &chat.IsGroup,
		&chat.IsFavorite, &chat.IsArchived, &chat.UnreadCount,
		&chat.LastMessageAt, &chat.CreatedAt, &chat.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get chat: %w", err)
	}

	return &chat, nil
}

// Update updates chat fields (favorite, archived)
func (r *ChatRepository) Update(ctx context.Context, userID, chatID string, updates map[string]interface{}) error {
	// Verify ownership
	var exists bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM chats c
			JOIN connections conn ON c.connection_id = conn.id
			WHERE c.id = $1 AND conn.user_id = $2
		)
	`, chatID, userID).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return sql.ErrNoRows
	}

	// Build update query
	var setParts []string
	var args []interface{}
	argIndex := 1

	for key, value := range updates {
		switch key {
		case "is_favorite", "is_archived":
			setParts = append(setParts, fmt.Sprintf("%s = $%d", key, argIndex))
			args = append(args, value)
			argIndex++
		}
	}

	if len(setParts) == 0 {
		return nil
	}

	query := fmt.Sprintf("UPDATE chats SET %s, updated_at = NOW() WHERE id = $%d",
		strings.Join(setParts, ", "), argIndex)
	args = append(args, chatID)

	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}

// MarkAsRead marks all messages in a chat as read
func (r *ChatRepository) MarkAsRead(ctx context.Context, userID, chatID string) error {
	// Verify ownership and update
	_, err := r.db.ExecContext(ctx, `
		UPDATE chats c
		SET unread_count = 0, updated_at = NOW()
		FROM connections conn
		WHERE c.connection_id = conn.id
		AND c.id = $1 AND conn.user_id = $2
	`, chatID, userID)
	return err
}

// GetOrCreate gets or creates a chat for a JID
func (r *ChatRepository) GetOrCreate(ctx context.Context, connectionID, jid, name string, isGroup bool) (*Chat, error) {
	// Try to get existing chat
	var chat Chat
	err := r.db.QueryRowContext(ctx, `
		SELECT id, connection_id, jid, name, is_group, unread_count, last_message_at, created_at, updated_at
		FROM chats
		WHERE connection_id = $1 AND jid = $2
	`, connectionID, jid).Scan(
		&chat.ID, &chat.ConnectionID, &chat.JID, &chat.Name,
		&chat.IsGroup, &chat.UnreadCount, &chat.LastMessageAt,
		&chat.CreatedAt, &chat.UpdatedAt,
	)

	if err == nil {
		return &chat, nil
	}

	if err != sql.ErrNoRows {
		return nil, err
	}

	// Create new chat
	err = r.db.QueryRowContext(ctx, `
		INSERT INTO chats (connection_id, jid, name, is_group)
		VALUES ($1, $2, $3, $4)
		RETURNING id, connection_id, jid, name, is_group, unread_count, last_message_at, created_at, updated_at
	`, connectionID, jid, name, isGroup).Scan(
		&chat.ID, &chat.ConnectionID, &chat.JID, &chat.Name,
		&chat.IsGroup, &chat.UnreadCount, &chat.LastMessageAt,
		&chat.CreatedAt, &chat.UpdatedAt,
	)

	return &chat, err
}

// IncrementUnread increments unread count for a chat
func (r *ChatRepository) IncrementUnread(ctx context.Context, connectionID, jid string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE chats 
		SET unread_count = unread_count + 1, updated_at = NOW()
		WHERE connection_id = $1 AND jid = $2
	`, connectionID, jid)
	return err
}

// UpdateLastMessage updates the last message timestamp
func (r *ChatRepository) UpdateLastMessage(ctx context.Context, connectionID, jid string, timestamp time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE chats 
		SET last_message_at = $1, updated_at = NOW()
		WHERE connection_id = $2 AND jid = $3
	`, timestamp, connectionID, jid)
	return err
}
