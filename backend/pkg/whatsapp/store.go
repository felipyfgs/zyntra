package whatsapp

import (
	"context"
	"database/sql"
	"fmt"

	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"
)

// Store wraps the whatsmeow sqlstore.Container for session persistence
// Uses whatsmeow's built-in SQL store - no custom session logic needed
type Store struct {
	*sqlstore.Container
	db *sql.DB
}

// NewStore creates a new session store using PostgreSQL
// Uses whatsmeow's sqlstore which handles all session/crypto tables automatically
func NewStore(databaseURL string) (*Store, error) {
	ctx := context.Background()

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Use whatsmeow's sqlstore directly - it manages all WhatsApp session data
	container := sqlstore.NewWithDB(db, "postgres", waLog.Noop)

	if err := container.Upgrade(ctx); err != nil {
		return nil, fmt.Errorf("failed to upgrade database schema: %w", err)
	}

	return &Store{
		Container: container,
		db:        db,
	}, nil
}

// GetDeviceByJID retrieves a device by JID string (convenience wrapper)
func (s *Store) GetDeviceByJID(ctx context.Context, jidStr string) (*store.Device, error) {
	if jidStr == "" {
		return nil, nil
	}

	jid, err := types.ParseJID(jidStr)
	if err != nil {
		return nil, nil
	}

	return s.Container.GetDevice(ctx, jid)
}

// DB returns the underlying database connection for app-level queries
func (s *Store) DB() *sql.DB {
	return s.db
}

// Close closes the database connection
func (s *Store) Close() error {
	return s.db.Close()
}

// ConnectionRepository handles connection records in the database
type ConnectionRepository struct {
	db *sql.DB
}

// NewConnectionRepository creates a new connection repository
func NewConnectionRepository(db *sql.DB) *ConnectionRepository {
	return &ConnectionRepository{db: db}
}

// Create inserts a new connection
func (r *ConnectionRepository) Create(ctx context.Context, conn *Connection) error {
	query := `
		INSERT INTO connections (id, user_id, name, phone, jid, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(ctx, query,
		conn.ID, conn.UserID, conn.Name, conn.Phone, conn.JID, conn.Status,
		conn.CreatedAt, conn.UpdatedAt,
	)
	return err
}

// Update updates an existing connection
func (r *ConnectionRepository) Update(ctx context.Context, conn *Connection) error {
	query := `
		UPDATE connections 
		SET name = $2, phone = $3, jid = $4, status = $5, updated_at = $6
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query,
		conn.ID, conn.Name, conn.Phone, conn.JID, conn.Status, conn.UpdatedAt,
	)
	return err
}

// GetByID retrieves a connection by ID
func (r *ConnectionRepository) GetByID(ctx context.Context, id string) (*Connection, error) {
	query := `
		SELECT id, user_id, name, phone, jid, status, created_at, updated_at
		FROM connections WHERE id = $1
	`
	conn := &Connection{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&conn.ID, &conn.UserID, &conn.Name, &conn.Phone, &conn.JID, &conn.Status,
		&conn.CreatedAt, &conn.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// GetByUserID retrieves all connections for a user
func (r *ConnectionRepository) GetByUserID(ctx context.Context, userID string) ([]*Connection, error) {
	query := `
		SELECT id, user_id, name, phone, jid, status, created_at, updated_at
		FROM connections WHERE user_id = $1 ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var connections []*Connection
	for rows.Next() {
		conn := &Connection{}
		if err := rows.Scan(
			&conn.ID, &conn.UserID, &conn.Name, &conn.Phone, &conn.JID, &conn.Status,
			&conn.CreatedAt, &conn.UpdatedAt,
		); err != nil {
			return nil, err
		}
		connections = append(connections, conn)
	}
	return connections, rows.Err()
}

// Delete removes a connection
func (r *ConnectionRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM connections WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// SetQRCode stores the QR code base64 for a connection (wuzapi pattern)
func (r *ConnectionRepository) SetQRCode(ctx context.Context, id, qrcode string) error {
	query := `UPDATE connections SET qrcode = $2, status = 'qr_code' WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, qrcode)
	return err
}

// GetQRCode retrieves the QR code for a connection
func (r *ConnectionRepository) GetQRCode(ctx context.Context, id string) (string, error) {
	query := `SELECT COALESCE(qrcode, '') FROM connections WHERE id = $1`
	var qrcode string
	err := r.db.QueryRowContext(ctx, query, id).Scan(&qrcode)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return qrcode, err
}

// ClearQRCode removes the QR code and updates status
func (r *ConnectionRepository) ClearQRCode(ctx context.Context, id string, status string) error {
	query := `UPDATE connections SET qrcode = '', status = $2 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, status)
	return err
}

// MessageRepository handles message records in the database
type MessageRepository struct {
	db *sql.DB
}

// NewMessageRepository creates a new message repository
func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

// Create inserts a new message
func (r *MessageRepository) Create(ctx context.Context, msg *Message) error {
	query := `
		INSERT INTO messages (id, connection_id, chat_jid, sender_jid, direction, content, media_type, media_url, status, timestamp, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.db.ExecContext(ctx, query,
		msg.ID, msg.ConnectionID, msg.ChatJID, msg.SenderJID, msg.Direction,
		msg.Content, msg.MediaType, msg.MediaURL, msg.Status, msg.Timestamp, msg.CreatedAt,
	)
	return err
}

// UpdateStatus updates the status of a message
func (r *MessageRepository) UpdateStatus(ctx context.Context, id string, status MessageStatus) error {
	query := `UPDATE messages SET status = $2 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, status)
	return err
}

// GetByChatJID retrieves messages for a chat
func (r *MessageRepository) GetByChatJID(ctx context.Context, connectionID, chatJID string, limit, offset int) ([]*Message, error) {
	query := `
		SELECT id, connection_id, chat_jid, sender_jid, direction, content, media_type, media_url, status, timestamp, created_at
		FROM messages 
		WHERE connection_id = $1 AND chat_jid = $2 
		ORDER BY timestamp DESC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.db.QueryContext(ctx, query, connectionID, chatJID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		msg := &Message{}
		if err := rows.Scan(
			&msg.ID, &msg.ConnectionID, &msg.ChatJID, &msg.SenderJID, &msg.Direction,
			&msg.Content, &msg.MediaType, &msg.MediaURL, &msg.Status, &msg.Timestamp, &msg.CreatedAt,
		); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	return messages, rows.Err()
}
