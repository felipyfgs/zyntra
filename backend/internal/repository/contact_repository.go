package repository

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/zyntra/backend/internal/domain"
)

// ContactRepository repositorio de contatos
type ContactRepository struct {
	db *sql.DB
}

// NewContactRepository cria novo repositorio
func NewContactRepository(db *sql.DB) *ContactRepository {
	return &ContactRepository{db: db}
}

// Create cria um contato
func (r *ContactRepository) Create(ctx context.Context, contact *domain.Contact) error {
	attrsJSON, _ := json.Marshal(contact.CustomAttributes)
	query := `
		INSERT INTO contacts (id, name, email, phone_number, avatar_url, custom_attributes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
	`
	_, err := r.db.ExecContext(ctx, query,
		contact.ID, contact.Name, contact.Email, contact.PhoneNumber,
		contact.AvatarURL, attrsJSON,
	)
	return err
}

// GetByID busca contato por ID
func (r *ContactRepository) GetByID(ctx context.Context, id string) (*domain.Contact, error) {
	query := `
		SELECT id, COALESCE(name, ''), COALESCE(email, ''), COALESCE(phone_number, ''),
		       COALESCE(avatar_url, ''), COALESCE(custom_attributes, '{}'), created_at, updated_at
		FROM contacts WHERE id = $1
	`
	contact := &domain.Contact{}
	var attrsJSON []byte
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&contact.ID, &contact.Name, &contact.Email, &contact.PhoneNumber,
		&contact.AvatarURL, &attrsJSON, &contact.CreatedAt, &contact.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	json.Unmarshal(attrsJSON, &contact.CustomAttributes)
	return contact, nil
}

// GetByPhone busca contato por telefone
func (r *ContactRepository) GetByPhone(ctx context.Context, phone string) (*domain.Contact, error) {
	query := `
		SELECT id, COALESCE(name, ''), COALESCE(email, ''), COALESCE(phone_number, ''),
		       COALESCE(avatar_url, ''), COALESCE(custom_attributes, '{}'), created_at, updated_at
		FROM contacts WHERE phone_number = $1
	`
	contact := &domain.Contact{}
	var attrsJSON []byte
	err := r.db.QueryRowContext(ctx, query, phone).Scan(
		&contact.ID, &contact.Name, &contact.Email, &contact.PhoneNumber,
		&contact.AvatarURL, &attrsJSON, &contact.CreatedAt, &contact.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	json.Unmarshal(attrsJSON, &contact.CustomAttributes)
	return contact, nil
}

// GetAll lista todos os contatos
func (r *ContactRepository) GetAll(ctx context.Context, limit, offset int) ([]*domain.Contact, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `
		SELECT id, COALESCE(name, ''), COALESCE(email, ''), COALESCE(phone_number, ''),
		       COALESCE(avatar_url, ''), COALESCE(custom_attributes, '{}'), created_at, updated_at
		FROM contacts ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contacts []*domain.Contact
	for rows.Next() {
		contact := &domain.Contact{}
		var attrsJSON []byte
		if err := rows.Scan(
			&contact.ID, &contact.Name, &contact.Email, &contact.PhoneNumber,
			&contact.AvatarURL, &attrsJSON, &contact.CreatedAt, &contact.UpdatedAt,
		); err != nil {
			return nil, err
		}
		json.Unmarshal(attrsJSON, &contact.CustomAttributes)
		contacts = append(contacts, contact)
	}
	return contacts, rows.Err()
}

// Update atualiza um contato
func (r *ContactRepository) Update(ctx context.Context, contact *domain.Contact) error {
	attrsJSON, _ := json.Marshal(contact.CustomAttributes)
	query := `
		UPDATE contacts SET name = $2, email = $3, phone_number = $4, 
		       avatar_url = $5, custom_attributes = $6, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query,
		contact.ID, contact.Name, contact.Email, contact.PhoneNumber,
		contact.AvatarURL, attrsJSON,
	)
	return err
}

// Delete remove um contato
func (r *ContactRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM contacts WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// Search busca contatos por termo
func (r *ContactRepository) Search(ctx context.Context, term string, limit int) ([]*domain.Contact, error) {
	if limit <= 0 {
		limit = 20
	}
	query := `
		SELECT id, COALESCE(name, ''), COALESCE(email, ''), COALESCE(phone_number, ''),
		       COALESCE(avatar_url, ''), COALESCE(custom_attributes, '{}'), created_at, updated_at
		FROM contacts 
		WHERE name ILIKE $1 OR email ILIKE $1 OR phone_number ILIKE $1
		ORDER BY name LIMIT $2
	`
	rows, err := r.db.QueryContext(ctx, query, "%"+term+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contacts []*domain.Contact
	for rows.Next() {
		contact := &domain.Contact{}
		var attrsJSON []byte
		if err := rows.Scan(
			&contact.ID, &contact.Name, &contact.Email, &contact.PhoneNumber,
			&contact.AvatarURL, &attrsJSON, &contact.CreatedAt, &contact.UpdatedAt,
		); err != nil {
			return nil, err
		}
		json.Unmarshal(attrsJSON, &contact.CustomAttributes)
		contacts = append(contacts, contact)
	}
	return contacts, rows.Err()
}

// ContactInboxRepository repositorio de contact_inboxes
type ContactInboxRepository struct {
	db *sql.DB
}

// NewContactInboxRepository cria novo repositorio
func NewContactInboxRepository(db *sql.DB) *ContactInboxRepository {
	return &ContactInboxRepository{db: db}
}

// Create cria um contact_inbox
func (r *ContactInboxRepository) Create(ctx context.Context, ci *domain.ContactInbox) error {
	query := `
		INSERT INTO contact_inboxes (id, contact_id, inbox_id, source_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
	`
	_, err := r.db.ExecContext(ctx, query, ci.ID, ci.ContactID, ci.InboxID, ci.SourceID)
	return err
}

// GetByID busca por ID
func (r *ContactInboxRepository) GetByID(ctx context.Context, id string) (*domain.ContactInbox, error) {
	query := `
		SELECT id, contact_id, inbox_id, source_id, created_at, updated_at
		FROM contact_inboxes WHERE id = $1
	`
	ci := &domain.ContactInbox{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&ci.ID, &ci.ContactID, &ci.InboxID, &ci.SourceID, &ci.CreatedAt, &ci.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return ci, err
}

// GetBySourceID busca por source_id em um inbox
func (r *ContactInboxRepository) GetBySourceID(ctx context.Context, inboxID, sourceID string) (*domain.ContactInbox, error) {
	query := `
		SELECT id, contact_id, inbox_id, source_id, created_at, updated_at
		FROM contact_inboxes WHERE inbox_id = $1 AND source_id = $2
	`
	ci := &domain.ContactInbox{}
	err := r.db.QueryRowContext(ctx, query, inboxID, sourceID).Scan(
		&ci.ID, &ci.ContactID, &ci.InboxID, &ci.SourceID, &ci.CreatedAt, &ci.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return ci, err
}

// GetByContactID lista contact_inboxes de um contato
func (r *ContactInboxRepository) GetByContactID(ctx context.Context, contactID string) ([]*domain.ContactInbox, error) {
	query := `
		SELECT id, contact_id, inbox_id, source_id, created_at, updated_at
		FROM contact_inboxes WHERE contact_id = $1
	`
	rows, err := r.db.QueryContext(ctx, query, contactID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.ContactInbox
	for rows.Next() {
		ci := &domain.ContactInbox{}
		if err := rows.Scan(&ci.ID, &ci.ContactID, &ci.InboxID, &ci.SourceID, &ci.CreatedAt, &ci.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, ci)
	}
	return list, rows.Err()
}

// Delete remove um contact_inbox
func (r *ContactInboxRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM contact_inboxes WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// FindOrCreateBySourceID busca ou cria contact_inbox
func (r *ContactInboxRepository) FindOrCreateBySourceID(ctx context.Context, ci *domain.ContactInbox) (*domain.ContactInbox, error) {
	existing, err := r.GetBySourceID(ctx, ci.InboxID, ci.SourceID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}
	if err := r.Create(ctx, ci); err != nil {
		return nil, err
	}
	return ci, nil
}
