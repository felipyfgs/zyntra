package repository

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/zyntra/backend/internal/domain"
	"github.com/zyntra/backend/internal/ports"
)

// InboxRepository repositorio de inboxes
type InboxRepository struct {
	db *sql.DB
}

// NewInboxRepository cria novo repositorio
func NewInboxRepository(db *sql.DB) *InboxRepository {
	return &InboxRepository{db: db}
}

// Create cria um inbox
func (r *InboxRepository) Create(ctx context.Context, inbox *domain.Inbox) error {
	query := `
		INSERT INTO inboxes (id, name, channel_type, channel_id, status, greeting_message, auto_assignment, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
	`
	_, err := r.db.ExecContext(ctx, query,
		inbox.ID, inbox.Name, inbox.ChannelType, inbox.ChannelID,
		inbox.Status, inbox.GreetingMessage, inbox.AutoAssignment,
	)
	return err
}

// GetByID busca inbox por ID
func (r *InboxRepository) GetByID(ctx context.Context, id string) (*domain.Inbox, error) {
	query := `
		SELECT id, name, channel_type, channel_id, status, COALESCE(qrcode, ''), 
		       COALESCE(greeting_message, ''), auto_assignment, created_at, updated_at
		FROM inboxes WHERE id = $1
	`
	inbox := &domain.Inbox{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&inbox.ID, &inbox.Name, &inbox.ChannelType, &inbox.ChannelID,
		&inbox.Status, &inbox.QRCode, &inbox.GreetingMessage,
		&inbox.AutoAssignment, &inbox.CreatedAt, &inbox.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return inbox, err
}

// GetAll lista todos os inboxes
func (r *InboxRepository) GetAll(ctx context.Context) ([]*domain.Inbox, error) {
	query := `
		SELECT id, name, channel_type, channel_id, status, COALESCE(qrcode, ''),
		       COALESCE(greeting_message, ''), auto_assignment, created_at, updated_at
		FROM inboxes ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var inboxes []*domain.Inbox
	for rows.Next() {
		inbox := &domain.Inbox{}
		if err := rows.Scan(
			&inbox.ID, &inbox.Name, &inbox.ChannelType, &inbox.ChannelID,
			&inbox.Status, &inbox.QRCode, &inbox.GreetingMessage,
			&inbox.AutoAssignment, &inbox.CreatedAt, &inbox.UpdatedAt,
		); err != nil {
			return nil, err
		}
		inboxes = append(inboxes, inbox)
	}
	return inboxes, rows.Err()
}

// GetByChannelType lista inboxes por tipo de canal
func (r *InboxRepository) GetByChannelType(ctx context.Context, channelType ports.ChannelType) ([]*domain.Inbox, error) {
	query := `
		SELECT id, name, channel_type, channel_id, status, COALESCE(qrcode, ''),
		       COALESCE(greeting_message, ''), auto_assignment, created_at, updated_at
		FROM inboxes WHERE channel_type = $1 ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, channelType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var inboxes []*domain.Inbox
	for rows.Next() {
		inbox := &domain.Inbox{}
		if err := rows.Scan(
			&inbox.ID, &inbox.Name, &inbox.ChannelType, &inbox.ChannelID,
			&inbox.Status, &inbox.QRCode, &inbox.GreetingMessage,
			&inbox.AutoAssignment, &inbox.CreatedAt, &inbox.UpdatedAt,
		); err != nil {
			return nil, err
		}
		inboxes = append(inboxes, inbox)
	}
	return inboxes, rows.Err()
}

// Update atualiza um inbox
func (r *InboxRepository) Update(ctx context.Context, inbox *domain.Inbox) error {
	query := `
		UPDATE inboxes SET name = $2, status = $3, greeting_message = $4, 
		       auto_assignment = $5, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query,
		inbox.ID, inbox.Name, inbox.Status, inbox.GreetingMessage, inbox.AutoAssignment,
	)
	return err
}

// UpdateStatus atualiza status do inbox
func (r *InboxRepository) UpdateStatus(ctx context.Context, id string, status ports.ChannelStatus) error {
	query := `UPDATE inboxes SET status = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, status)
	return err
}

// SetQRCode define QR code do inbox
func (r *InboxRepository) SetQRCode(ctx context.Context, id, qrcode string) error {
	query := `UPDATE inboxes SET qrcode = $2, status = 'qr_code', updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, qrcode)
	return err
}

// ClearQRCode limpa QR code do inbox
func (r *InboxRepository) ClearQRCode(ctx context.Context, id string, status ports.ChannelStatus) error {
	query := `UPDATE inboxes SET qrcode = '', status = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, status)
	return err
}

// Delete remove um inbox
func (r *InboxRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM inboxes WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// ChannelWhatsAppRepository repositorio de canais WhatsApp
type ChannelWhatsAppRepository struct {
	db *sql.DB
}

// NewChannelWhatsAppRepository cria novo repositorio
func NewChannelWhatsAppRepository(db *sql.DB) *ChannelWhatsAppRepository {
	return &ChannelWhatsAppRepository{db: db}
}

// Create cria canal WhatsApp
func (r *ChannelWhatsAppRepository) Create(ctx context.Context, channel *domain.ChannelWhatsApp) error {
	configJSON, _ := json.Marshal(channel.ProviderConfig)
	query := `
		INSERT INTO channel_whatsapp (id, phone_number, jid, provider, provider_config, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
	`
	_, err := r.db.ExecContext(ctx, query,
		channel.ID, channel.PhoneNumber, channel.JID, channel.Provider, configJSON,
	)
	return err
}

// GetByID busca canal por ID
func (r *ChannelWhatsAppRepository) GetByID(ctx context.Context, id string) (*domain.ChannelWhatsApp, error) {
	query := `
		SELECT id, COALESCE(phone_number, ''), COALESCE(jid, ''), provider, 
		       COALESCE(provider_config, '{}'), created_at, updated_at
		FROM channel_whatsapp WHERE id = $1
	`
	channel := &domain.ChannelWhatsApp{}
	var configJSON []byte
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&channel.ID, &channel.PhoneNumber, &channel.JID,
		&channel.Provider, &configJSON, &channel.CreatedAt, &channel.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	json.Unmarshal(configJSON, &channel.ProviderConfig)
	return channel, nil
}

// Update atualiza canal WhatsApp
func (r *ChannelWhatsAppRepository) Update(ctx context.Context, channel *domain.ChannelWhatsApp) error {
	configJSON, _ := json.Marshal(channel.ProviderConfig)
	query := `
		UPDATE channel_whatsapp SET phone_number = $2, jid = $3, provider_config = $4, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, channel.ID, channel.PhoneNumber, channel.JID, configJSON)
	return err
}

// UpdateJID atualiza JID do canal
func (r *ChannelWhatsAppRepository) UpdateJID(ctx context.Context, id, jid, phone string) error {
	query := `UPDATE channel_whatsapp SET jid = $2, phone_number = $3, updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, jid, phone)
	return err
}

// Delete remove canal
func (r *ChannelWhatsAppRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM channel_whatsapp WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// GetByJID busca canal por JID
func (r *ChannelWhatsAppRepository) GetByJID(ctx context.Context, jid string) (*domain.ChannelWhatsApp, error) {
	query := `
		SELECT id, COALESCE(phone_number, ''), COALESCE(jid, ''), provider,
		       COALESCE(provider_config, '{}'), created_at, updated_at
		FROM channel_whatsapp WHERE jid = $1
	`
	channel := &domain.ChannelWhatsApp{}
	var configJSON []byte
	err := r.db.QueryRowContext(ctx, query, jid).Scan(
		&channel.ID, &channel.PhoneNumber, &channel.JID,
		&channel.Provider, &configJSON, &channel.CreatedAt, &channel.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	json.Unmarshal(configJSON, &channel.ProviderConfig)
	return channel, nil
}

// GetInboxByJID busca inbox por JID do canal WhatsApp
func (r *InboxRepository) GetByChannelJID(ctx context.Context, jid string) (*domain.Inbox, error) {
	query := `
		SELECT i.id, i.name, i.channel_type, i.channel_id, i.status, COALESCE(i.qrcode, ''),
		       COALESCE(i.greeting_message, ''), i.auto_assignment, i.created_at, i.updated_at
		FROM inboxes i
		JOIN channel_whatsapp cw ON i.channel_id = cw.id
		WHERE cw.jid = $1 AND i.channel_type = 'whatsapp'
	`
	inbox := &domain.Inbox{}
	err := r.db.QueryRowContext(ctx, query, jid).Scan(
		&inbox.ID, &inbox.Name, &inbox.ChannelType, &inbox.ChannelID,
		&inbox.Status, &inbox.QRCode, &inbox.GreetingMessage,
		&inbox.AutoAssignment, &inbox.CreatedAt, &inbox.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return inbox, err
}

// InboxMemberRepository repositorio de membros do inbox
type InboxMemberRepository struct {
	db *sql.DB
}

// NewInboxMemberRepository cria novo repositorio
func NewInboxMemberRepository(db *sql.DB) *InboxMemberRepository {
	return &InboxMemberRepository{db: db}
}

// Add adiciona membro ao inbox
func (r *InboxMemberRepository) Add(ctx context.Context, inboxID, userID string) error {
	query := `INSERT INTO inbox_members (inbox_id, user_id, created_at) VALUES ($1, $2, NOW()) ON CONFLICT DO NOTHING`
	_, err := r.db.ExecContext(ctx, query, inboxID, userID)
	return err
}

// Remove remove membro do inbox
func (r *InboxMemberRepository) Remove(ctx context.Context, inboxID, userID string) error {
	query := `DELETE FROM inbox_members WHERE inbox_id = $1 AND user_id = $2`
	_, err := r.db.ExecContext(ctx, query, inboxID, userID)
	return err
}

// GetByInboxID lista membros de um inbox
func (r *InboxMemberRepository) GetByInboxID(ctx context.Context, inboxID string) ([]string, error) {
	query := `SELECT user_id FROM inbox_members WHERE inbox_id = $1`
	rows, err := r.db.QueryContext(ctx, query, inboxID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userIDs []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}
	return userIDs, rows.Err()
}

// GetInboxesByUserID lista inboxes de um usuario
func (r *InboxMemberRepository) GetInboxesByUserID(ctx context.Context, userID string) ([]string, error) {
	query := `SELECT inbox_id FROM inbox_members WHERE user_id = $1`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var inboxIDs []string
	for rows.Next() {
		var inboxID string
		if err := rows.Scan(&inboxID); err != nil {
			return nil, err
		}
		inboxIDs = append(inboxIDs, inboxID)
	}
	return inboxIDs, rows.Err()
}

// GetAllWhatsAppForRestore lista todos inboxes WhatsApp com JID para restaurar
func (r *InboxRepository) GetAllWhatsAppForRestore(ctx context.Context) ([]struct{ ID, JID string }, error) {
	query := `
		SELECT i.id, COALESCE(cw.jid, '')
		FROM inboxes i
		JOIN channel_whatsapp cw ON i.channel_id = cw.id
		WHERE i.channel_type = 'whatsapp' AND cw.jid IS NOT NULL AND cw.jid != ''
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []struct{ ID, JID string }
	for rows.Next() {
		var item struct{ ID, JID string }
		if err := rows.Scan(&item.ID, &item.JID); err != nil {
			return nil, err
		}
		results = append(results, item)
	}
	return results, rows.Err()
}
