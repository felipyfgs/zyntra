package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"
)

var (
	ErrAPIKeyNotFound = errors.New("API key not found")
	ErrAPIKeyExpired  = errors.New("API key has expired")
	ErrAPIKeyRevoked  = errors.New("API key has been revoked")
)

// APIKey represents an API key for external integrations
type APIKey struct {
	ID          string     `json:"id" db:"id"`
	UserID      string     `json:"user_id" db:"user_id"`
	Name        string     `json:"name" db:"name"`
	KeyHash     string     `json:"-" db:"key_hash"`
	KeyPrefix   string     `json:"key_prefix" db:"key_prefix"` // First 8 chars for identification
	Permissions []string   `json:"permissions" db:"permissions"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	RevokedAt   *time.Time `json:"revoked_at,omitempty" db:"revoked_at"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty" db:"last_used_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

// APIKeyPermission represents available permissions
type APIKeyPermission string

const (
	PermissionReadChats      APIKeyPermission = "chats:read"
	PermissionWriteChats     APIKeyPermission = "chats:write"
	PermissionReadMessages   APIKeyPermission = "messages:read"
	PermissionWriteMessages  APIKeyPermission = "messages:write"
	PermissionReadContacts   APIKeyPermission = "contacts:read"
	PermissionWriteContacts  APIKeyPermission = "contacts:write"
	PermissionReadConnections  APIKeyPermission = "connections:read"
	PermissionWriteConnections APIKeyPermission = "connections:write"
	PermissionWebhooks       APIKeyPermission = "webhooks:manage"
	PermissionAll            APIKeyPermission = "*"
)

// AllPermissions returns all available permissions
func AllPermissions() []string {
	return []string{
		string(PermissionReadChats),
		string(PermissionWriteChats),
		string(PermissionReadMessages),
		string(PermissionWriteMessages),
		string(PermissionReadContacts),
		string(PermissionWriteContacts),
		string(PermissionReadConnections),
		string(PermissionWriteConnections),
		string(PermissionWebhooks),
	}
}

// GeneratedAPIKey is returned when creating a new API key (only time full key is shown)
type GeneratedAPIKey struct {
	APIKey
	Key string `json:"key"` // Full key, only shown once
}

// APIKeyService handles API key operations
type APIKeyService struct {
	db *sql.DB
}

// NewAPIKeyService creates a new API key service
func NewAPIKeyService(db *sql.DB) *APIKeyService {
	return &APIKeyService{db: db}
}

// GenerateAPIKey creates a new API key
func (s *APIKeyService) GenerateAPIKey(ctx context.Context, userID, name string, permissions []string, expiresIn *time.Duration) (*GeneratedAPIKey, error) {
	// Generate random key: zyn_xxxxxxxxxxxxxxxxxxxxxxxxxxxx
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, err
	}
	
	key := "zyn_" + hex.EncodeToString(keyBytes)
	keyPrefix := key[:12] // "zyn_" + first 8 chars
	
	// Hash the key for storage
	hash := sha256.Sum256([]byte(key))
	keyHash := hex.EncodeToString(hash[:])

	var expiresAt *time.Time
	if expiresIn != nil {
		exp := time.Now().Add(*expiresIn)
		expiresAt = &exp
	}

	// Generate UUID for ID
	id := generateUUID()

	apiKey := &APIKey{
		ID:          id,
		UserID:      userID,
		Name:        name,
		KeyHash:     keyHash,
		KeyPrefix:   keyPrefix,
		Permissions: permissions,
		ExpiresAt:   expiresAt,
		CreatedAt:   time.Now(),
	}

	// Insert into database
	query := `
		INSERT INTO api_keys (id, user_id, name, key_hash, key_prefix, permissions, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := s.db.ExecContext(ctx, query,
		apiKey.ID, apiKey.UserID, apiKey.Name, apiKey.KeyHash, apiKey.KeyPrefix,
		apiKey.Permissions, apiKey.ExpiresAt, apiKey.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &GeneratedAPIKey{
		APIKey: *apiKey,
		Key:    key,
	}, nil
}

// ValidateAPIKey validates an API key and returns its details
func (s *APIKeyService) ValidateAPIKey(ctx context.Context, key string) (*APIKey, error) {
	// Hash the provided key
	hash := sha256.Sum256([]byte(key))
	keyHash := hex.EncodeToString(hash[:])

	query := `
		SELECT id, user_id, name, key_hash, key_prefix, permissions, expires_at, revoked_at, last_used_at, created_at
		FROM api_keys
		WHERE key_hash = $1
	`

	var apiKey APIKey
	err := s.db.QueryRowContext(ctx, query, keyHash).Scan(
		&apiKey.ID, &apiKey.UserID, &apiKey.Name, &apiKey.KeyHash, &apiKey.KeyPrefix,
		&apiKey.Permissions, &apiKey.ExpiresAt, &apiKey.RevokedAt, &apiKey.LastUsedAt, &apiKey.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAPIKeyNotFound
		}
		return nil, err
	}

	// Check if revoked
	if apiKey.RevokedAt != nil {
		return nil, ErrAPIKeyRevoked
	}

	// Check if expired
	if apiKey.ExpiresAt != nil && time.Now().After(*apiKey.ExpiresAt) {
		return nil, ErrAPIKeyExpired
	}

	// Update last used timestamp
	go s.updateLastUsed(apiKey.ID)

	return &apiKey, nil
}

// RevokeAPIKey revokes an API key
func (s *APIKeyService) RevokeAPIKey(ctx context.Context, keyID, userID string) error {
	query := `
		UPDATE api_keys 
		SET revoked_at = $1 
		WHERE id = $2 AND user_id = $3 AND revoked_at IS NULL
	`
	result, err := s.db.ExecContext(ctx, query, time.Now(), keyID, userID)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrAPIKeyNotFound
	}

	return nil
}

// ListAPIKeys returns all API keys for a user
func (s *APIKeyService) ListAPIKeys(ctx context.Context, userID string) ([]*APIKey, error) {
	query := `
		SELECT id, user_id, name, key_prefix, permissions, expires_at, revoked_at, last_used_at, created_at
		FROM api_keys
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*APIKey
	for rows.Next() {
		var key APIKey
		err := rows.Scan(
			&key.ID, &key.UserID, &key.Name, &key.KeyPrefix,
			&key.Permissions, &key.ExpiresAt, &key.RevokedAt, &key.LastUsedAt, &key.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		keys = append(keys, &key)
	}

	return keys, nil
}

// HasPermission checks if the API key has the required permission
func (k *APIKey) HasPermission(required string) bool {
	for _, p := range k.Permissions {
		if p == string(PermissionAll) || p == required {
			return true
		}
		// Check wildcard permissions (e.g., "chats:*" matches "chats:read")
		if len(p) > 2 && p[len(p)-2:] == ":*" {
			prefix := p[:len(p)-1]
			if len(required) >= len(prefix) && required[:len(prefix)] == prefix {
				return true
			}
		}
	}
	return false
}

func (s *APIKeyService) updateLastUsed(keyID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `UPDATE api_keys SET last_used_at = $1 WHERE id = $2`
	s.db.ExecContext(ctx, query, time.Now(), keyID)
}

func generateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
